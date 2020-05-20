package league

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type SaveLeagueInfoIFace interface {
	SaveLeague(context.Context, *entities.LeagueGroup) (*entities.LeagueGroup, error)
	SavePlayers(context.Context, []entities.PlayerSeason) ([]entities.PlayerSeason, error)
	GetTeamsForManagers(ctx context.Context, guid string) ([]entities.League, error)
}

type ImportYahooData interface {
	GetUserLeagues(ctx context.Context, guid string) ([]*entities.LeagueGroup, error)
	ImportFromUser(ctx context.Context, userGames *yahoo.UserResourcesGameLeaguesResponse) ([]*entities.LeagueGroup, error)
}
type NewImporterService interface {
	NewImporterWithService(svc *yahoo.Service) ImportYahooData
}

type dataTransferObjects struct {
	leagueMu            *sync.Mutex
	leagueGroupMu       *sync.Mutex
	leagueTeamsMu       *sync.Mutex
	gameMapToIdx        map[int]int
	leagueMaps          map[int]int
	teamMaps            map[int]int
	userMaps            map[string]int // tied to email
	games               []*entities.Game
	leagues             []entities.League
	users               []*entities.User
	teams               []*entities.Team
	leagueGroups        []*entities.LeagueGroup
	leagueToMapGroupIds map[int]int
	leagueGroupMaps     map[int]int
}

func (d *dataTransferObjects) addLeague(league entities.League) {
	d.leagueMu.Lock()
	defer d.leagueMu.Unlock()
	d.leagues = append(d.leagues, league)
	d.leagueMaps[league.LeagueID] = len(d.leagues)
}

func (d *dataTransferObjects) addTeam(team entities.Team) {
	d.leagueTeamsMu.Lock()
	defer d.leagueTeamsMu.Unlock()

	for idx := range team.Manager {
		if useridx, ok := d.userMaps[team.Manager[idx].Email]; ok {
			team.Manager[idx] = *d.users[useridx]
		}
	}
	d.teams = append(d.teams, &team)
	d.teamMaps[team.TeamID] = len(d.teams)
}

func (d *dataTransferObjects) addUser(user entities.User) {
	d.leagueMu.Lock()
	defer d.leagueMu.Unlock()
	if _, ok := d.userMaps[user.ManagerID]; !ok {
		d.users = append(d.users, &user)
		d.userMaps[user.Email] = len(d.users)
	}
}

func (d *dataTransferObjects) addGames(game entities.Game) *entities.Game {
	d.leagueMu.Lock()
	defer d.leagueMu.Unlock()
	gameIdx, ok := d.gameMapToIdx[game.GameID]
	if !ok {
		currentGameLen := len(d.games)
		d.games[currentGameLen] = &game
		d.gameMapToIdx[game.GameID] = len(d.games)
		return &game
	}
	return d.games[gameIdx]
}

func (d *dataTransferObjects) addLeagueToLeagueGroup(parentLeagueID int, league entities.League) {
	d.leagueGroupMu.Lock()
	defer d.leagueGroupMu.Unlock()
	// find group IDx
	var leagueGroup *entities.LeagueGroup
	// see if already created by finding the index
	leagueToGroupIdx, ok := d.leagueToMapGroupIds[parentLeagueID]
	if !ok {
		leagueGroup = &entities.LeagueGroup{
			FirstLeagueID: parentLeagueID,
			Leagues:       make([]entities.League, 0, 400),
		}
		d.leagueGroups = append(d.leagueGroups, leagueGroup)
		d.leagueGroupMaps[parentLeagueID] = len(d.leagueGroups) - 1
		leagueToGroupIdx = d.leagueGroupMaps[parentLeagueID]

	} else {
		leagueGroup = d.leagueGroups[d.leagueGroupMaps[leagueToGroupIdx]]
	}

	d.leagueGroups[leagueToGroupIdx].Leagues = append(d.leagueGroups[leagueToGroupIdx].Leagues, league)
	d.leagueToMapGroupIds[league.LeagueID] = leagueToGroupIdx
}

type Importer struct {
	logger log.Logger

	yahooService *yahoo.Service
	repo         SaveLeagueInfoIFace
}

func NewImportService(logger log.Logger, yahooService *yahoo.Service, repo SaveLeagueInfoIFace) Importer {
	return Importer{
		logger:       logger,
		yahooService: yahooService,
		repo:         repo,
	}
}

func newDTO() dataTransferObjects {
	return dataTransferObjects{
		leagueMu:            &sync.Mutex{},
		leagueGroupMu:       &sync.Mutex{},
		leagueTeamsMu:       &sync.Mutex{},
		games:               make([]*entities.Game, 0, 400),
		leagues:             make([]entities.League, 0, 1000),
		leagueGroups:        make([]*entities.LeagueGroup, 0, 1000),
		leagueGroupMaps:     map[int]int{},
		leagueToMapGroupIds: map[int]int{},
		leagueMaps:          make(map[int]int),
		teamMaps:            make(map[int]int),
		userMaps:            make(map[string]int),
		users:               make([]*entities.User, 0, 1000),
		teams:               make([]*entities.Team, 0, 1000),
	}
}

func (i *Importer) NewImporterWithService(svc *yahoo.Service) ImportYahooData {
	i.yahooService = svc
	return i
}

func (i *Importer) GetUserLeagues(ctx context.Context, guid string) ([]*entities.LeagueGroup, error) {
	dto := newDTO()
	leagues, err := i.repo.GetTeamsForManagers(ctx, guid)
	if err != nil {
		level.Error(i.logger).Log("message", "could not get user teams", "error", err, "guid", guid)
	}

	for i := len(leagues) -1; i >= 0; i-- {
		dto.addLeague(leagues[i])
		parentLeagueID := 0
		ids := strings.Split(leagues[i].Settings.Renew, "_")
		if len(ids) > 1 {
			parentLeagueID, _ = strconv.Atoi(ids[1])
			leagues[i].PreviousLeague = &parentLeagueID
		}

		// check if parentID exists
		dto.addLeagueToLeagueGroup(parentLeagueID, leagues[i])
	}
	return dto.leagueGroups, nil
}

func (i *Importer) ImportFromUser(ctx context.Context, userGames *yahoo.UserResourcesGameLeaguesResponse) ([]*entities.LeagueGroup, error) {
	span, ctx := apm.StartSpan(ctx, "ImportFromUser", "service")
	defer func() {
		span.End()
	}()

	leagueChan := make(chan entities.League, 100)
	a := newDTO()
	dto := &a
	wg := &sync.WaitGroup{}

	go func() {
		for {
			select {
			case league := <-leagueChan:

				dto.addLeague(league)
				parentLeagueID := 0
				ids := strings.Split(league.Settings.Renew, "_")
				if len(ids) > 1 {
					parentLeagueID, _ = strconv.Atoi(ids[1])
					league.PreviousLeague = &parentLeagueID
				}

				// check if parentID exists
				dto.addLeagueToLeagueGroup(parentLeagueID, league)
				wg.Done()
			}
		}
	}()

	for _, yahooGame := range userGames.Users.User.Games.Game {
		game := transformGameResponseToGame(yahooGame)
		dto.games = append(dto.games, &game)
		for _, leagueRes := range yahooGame.Leagues {
			wg.Add(1)
			err := i.getLeague(ctx, game, leagueRes, leagueChan)
			if err != nil {
				level.Error(i.logger).Log("message", "error in getting league information", "league_key", leagueRes.LeagueKey, "error", err)
			}

		}
	}
	wg.Wait()
	// sort dtos
	for idx, leagueGroup := range dto.leagueGroups {
		sort.SliceStable(leagueGroup.Leagues, func(i, j int) bool {
			return leagueGroup.Leagues[i].Game.GameID < leagueGroup.Leagues[j].Game.GameID
		})

		dto.leagueGroups[idx] = leagueGroup
	}

	leagueGroupsWg := &sync.WaitGroup{}
	leagueGroupsWg.Add(len(dto.leagueGroups))
	for _, leagueGroup := range dto.leagueGroups {
		// add league group to the first one
		go func(leagueGroup *entities.LeagueGroup) {
			_, err := i.repo.SaveLeague(ctx, leagueGroup)
			if err != nil {
				level.Error(i.logger).Log("error", err)
			}
			leagueGroupsWg.Done()
		}(leagueGroup)
	}
	leagueGroupsWg.Wait()

	return dto.leagueGroups, nil

}

func (i *Importer) ImportLeagueFromUser(ctx context.Context) ([]*entities.LeagueGroup, error) {
	span, ctx := apm.StartSpan(ctx, "ImportLeagueFromUsers", "service")
	defer func() {
		span.End()
	}()
	// string is yahoo's guid
	//var userTeams map[string]Team
	res, err := i.yahooService.GetUserResourcesGameLeaguesResponse(ctx)
	if err != nil {
		level.Error(i.logger).Log("message", "could not get GetUserResourcesGameLeaguesResponse from yahoo", "error", err)
		return []*entities.LeagueGroup{}, err
	}

	return i.ImportFromUser(ctx, res)
}

func (i *Importer) getLeague(ctx context.Context, game entities.Game, leagueRes yahoo.YahooLeague, leagueChan chan entities.League) error {
	level.Debug(i.logger).Log("name", leagueRes.Name)
	var league entities.League

	league.Game = game
	league.Name = leagueRes.Name
	// get league settings
	yahooSettings, err := i.yahooService.GetLeagueResourcesSettings(ctx, leagueRes.LeagueKey)
	if err != nil {
		level.Error(i.logger).Log("message", "error in getting league Resource setting", "error", err)
		return err
	}
	settings := transformYahooLeagueSettingsToLeagueSettings(yahooSettings)
	league.Settings = &settings
	league.LeagueID = settings.LeagueID
	league.LeagueKey = settings.LeagueKey
	// get league standings
	yahooStandings, err := i.yahooService.GetLeagueResourcesStandings(ctx, leagueRes.LeagueKey)
	if err != nil {
		level.Error(i.logger).Log("message", "error in getting league Resource standings", "error", err)
		return err
	}

	teams := transformYahooStandingsToStandings(yahooStandings)
	league.Teams = teams

	leagueChan <- league
	return nil
}

func (i *Importer) ImportTeamsFromUser(ctx context.Context) {
	res, err := i.yahooService.GetUserResourcesGameTeams(ctx)
	if err != nil {
		return
	}

	// Assuming that games are not
	dto := newDTO()
	// Game loop
	for _, gameRes := range res.Users.User.Games.Game {
		// add game to dto
		game := transformGameResponseToGame(gameRes)
		dto.games = append(dto.games, &game)

	}

}

func intToBool(i int) bool {
	return i == 1
}

func transformGameResponseToGame(res yahoo.YahooGame) entities.Game {
	return entities.Game{
		GameID:             res.GameID,
		GameKey:            res.GameKey,
		Name:               res.Name,
		Code:               res.Code,
		Type:               res.Type,
		URL:                res.URL,
		Season:             res.Season,
		IsRegistrationOver: intToBool(res.IsRegistrationOver),
		IsGameOver:         intToBool(res.IsGameOver),
		IsOffseason:        intToBool(res.IsOffseason),
	}
}

func transformTeamResponseToTeam(res yahoo.TeamResponse) entities.Team {
	return entities.Team{}
}

//func transformLeagueResponseToTeam(res yahoo.YahooLeague) League {
//	return League{
//		ID:           primitive.ObjectID{},
//		Settings:     nil,
//		DraftResults: nil,
//		Teams:        make([]Team),
//		Game:         Game{},
//	}
//}

func transformYahooLeagueSettingsToLeagueSettings(response *yahoo.LeagueResourcesSettingsResponse) entities.LeagueSettings {
	leagueSettings := entities.LeagueSettings{
		LeagueKey:             response.League.LeagueKey,
		LeagueID:              response.League.LeagueID,
		Name:                  response.League.Name,
		URL:                   response.League.URL,
		LogoURL:               response.League.LogoURL,
		Password:              response.League.Password,
		DraftStatus:           response.League.DraftStatus,
		NumTeams:              response.League.NumTeams,
		EditKey:               response.League.EditKey,
		WeeklyDeadline:        response.League.WeeklyDeadline,
		LeagueUpdateTimestamp: response.League.LeagueUpdateTimestamp,
		LeagueType:            response.League.LeagueType,
		Renew:                 response.League.Renew,
		Renewed:               response.League.Renewed,
		IrisGroupChatID:       response.League.IrisGroupChatID,
		ShortInvitationURL:    response.League.ShortInvitationURL,
		AllowAddToDlExtraPos:  response.League.AllowAddToDlExtraPos,
		IsProLeague:           response.League.IsProLeague,
		IsCashLeague:          response.League.IsCashLeague,
		CurrentWeek:           response.League.CurrentWeek,
		StartWeek:             response.League.StartWeek,
		StartDate:             response.League.StartDate,
		EndWeek:               response.League.EndWeek,
		EndDate:               response.League.EndDate,
		GameCode:              response.League.GameCode,
		Season:                response.League.Season,
		MaxAdds:               response.League.MaxTrades,
		SeasonType:            "",
		MinInningsPitched:     "",


		//ID:                         response.League.LeagueID,
		DraftType:                  response.League.Settings.DraftType,
		IsAuctionDraft:             intToBool(response.League.Settings.IsAuctionDraft),
		ScoringType:                response.League.Settings.ScoringType,
		PersistentURL:              response.League.Settings.PersistentURL,
		UsesPlayoff:                response.League.Settings.UsesPlayoff,
		HasPlayoffConsolationGames: intToBool(response.League.Settings.HasPlayoffConsolationGames),
		PlayoffStartWeek:           response.League.Settings.PlayoffStartWeek,
		UsesPlayoffReseeding:       intToBool(response.League.Settings.UsesPlayoffReseeding),
		UsesLockEliminatedTeams:    intToBool(response.League.Settings.UsesLockEliminatedTeams),
		NumPlayoffTeams:            response.League.Settings.NumPlayoffTeams,
		NumPlayoffConsolationTeams: response.League.Settings.NumPlayoffConsolationTeams,
		UsesRosterImport:           intToBool(response.League.Settings.UsesRosterImport),
		RosterImportDeadline:       response.League.Settings.RosterImportDeadline,
		WaiverType:                 response.League.Settings.WaiverType,
		WaiverRule:                 response.League.Settings.WaiverRule,
		UsesFaab:                   intToBool(response.League.Settings.UsesFaab),
		DraftTime:                  response.League.Settings.DraftPickTime,
		PostDraftPlayers:           response.League.Settings.PostDraftPlayers,
		MaxTeams:                   response.League.Settings.MaxTeams,
		WaiverTime:                 response.League.Settings.WaiverTime,
		TradeEndDate:               response.League.Settings.TradeEndDate,
		TradeRatifyType:            response.League.Settings.TradeRejectTime,
		TradeRejectTime:            response.League.Settings.TradeRejectTime,
		PlayerPool:                 response.League.Settings.PlayerPool,
		CantCutList:                response.League.Settings.CantCutList,
		IsPubliclyViewable:         intToBool(response.League.Settings.IsPubliclyViewable),
		RosterPositions:            transformYahooRosterPositionsToRosterPositions(response.League.Settings.RosterPositions),
		StatCategories:             transformYahooStatCategoriesToStatCategories(response.League.Settings.StatCategories.Stats.Stat),
		StatModifiers:              transformYahooStatModifiers(response.League.Settings.StatModifiers.Stats.Stat),
		//MaxAdds:                     response.League.Settings.Ma,
		//SeasonType:                  response.League.Settings.Se,
		//MinInningsPitched:           response.League.Settings.MinInningsPitched,
		UsesFractalPoints:  intToBool(response.League.UsesFractionalPoints),
		UsesNegativePoints: intToBool(response.League.UsesNegativePoints),
	}

	return leagueSettings
}

//
func transformYahooRosterPositionsToRosterPositions(yahooRosterPositions []yahoo.ResponseRosterPosition) []entities.RosterPosition {
	rosterPositions := make([]entities.RosterPosition, len(yahooRosterPositions))

	for idx, yahooRosPos := range yahooRosterPositions {
		rosterPositions[idx] = entities.RosterPosition{
			Position:     yahooRosPos.Position,
			PositionType: yahooRosPos.PositionType,
			Count:        yahooRosPos.Count,
		}
	}

	return rosterPositions
}

func transformYahooStatCategoriesToStatCategories(yahooStatCategories []yahoo.ResponseStatCategory) []entities.StatCategory {
	stateCategories := make([]entities.StatCategory, len(yahooStatCategories))

	for idx, yahooStatCategories := range yahooStatCategories {
		stateCategories[idx] = entities.StatCategory{
			StatID:            yahooStatCategories.StatID,
			Enabled:           intToBool(yahooStatCategories.Enabled),
			Name:              yahooStatCategories.Name,
			DisplayName:       yahooStatCategories.DisplayName,
			SortOrder:         yahooStatCategories.SortOrder,
			PositionType:      yahooStatCategories.PositionType,
			StatPositionTypes: transformYahooPositionTypes(yahooStatCategories.StatPositionTypes),
		}
	}

	return stateCategories
}

func transformYahooPositionTypes(yahooPositionTypes []yahoo.ResponsePositionType) []entities.PositionType {
	positionTypes := make([]entities.PositionType, len(yahooPositionTypes))

	for idx, positionType := range yahooPositionTypes {
		positionTypes[idx] = entities.PositionType{
			PositionType:      positionType.PositionType,
			IsOnlyDisplayStat: intToBool(positionType.IsOnlyDisplayStat),
		}
	}
	return positionTypes
}

func transformYahooStatModifiers(yahooStatModifiers []yahoo.StatModifier) []entities.StatModifier {
	statModifiers := make([]entities.StatModifier, len(yahooStatModifiers))

	for idx, yahooStatModifier := range yahooStatModifiers {
		statModifiers[idx] = entities.StatModifier{
			StatID:  yahooStatModifier.StatID,
			Value:   yahooStatModifier.Value,
			Bonuses: transformYahooStatModifierBonusToBonus(yahooStatModifier.Bonus),
		}
	}
	return statModifiers
}

func transformYahooStatModifierBonusToBonus(yahooBonus *yahoo.Bonus) *entities.Bonus {
	if yahooBonus == nil {
		return nil
	}

	return &entities.Bonus{
		Target: yahooBonus.Target,
		Points: yahooBonus.Points,
	}
}

func transformManagerToUser(manager yahoo.Manager) entities.User {
	return entities.User{
		Email:     manager.Email,
		Name:      manager.Nickname,
		ManagerID: manager.ManagerID,
		Nickname:  manager.Nickname,
		Guid:      manager.GUID,
		//IsCommissioner: ,
		ImageURL: manager.ImageURL,
		//Teams:          nil,
		//Commissioned:   nil,
	}
}

func transformYahooStandingsToStandings(standings *yahoo.LeagueResourcesStandingsResponse) []entities.Team {

	teams := make([]entities.Team, standings.League.NumTeams)

	for idx, yahooTeam := range standings.League.Standings.Teams.Team {
		user := transformManagerToUser(yahooTeam.Managers.Manager)
		teams[idx] = entities.Team{

			TeamKey:               yahooTeam.TeamKey,
			TeamID:                yahooTeam.TeamID,
			Name:                  yahooTeam.Name,
			IsOwnedByCurrentLogin: intToBool(yahooTeam.IsOwnedByCurrentLogin),
			URL:                   yahooTeam.URL,
			TeamLogo:              entities.TeamLogo{},
			WaiverPriority:        yahooTeam.WaiverPriority,
			NumberOfMoves:         yahooTeam.NumberOfMoves,
			NumberOfTrades:        yahooTeam.NumberOfTrades,
			RosterAdds: entities.RosterAdds{
				CoverageType:  yahooTeam.RosterAdds.CoverageType,
				CoverageValue: yahooTeam.RosterAdds.CoverageValue,
				Value:         yahooTeam.RosterAdds.Value,
			},
			LeagueScoringType: yahooTeam.LeagueScoringType,
			HasDraftGrade:     intToBool(yahooTeam.HasDraftGrade),
			DraftGrade:        yahooTeam.DraftGrade,
			Standing: entities.Standings{
				Rank:          yahooTeam.TeamStandings.Rank,
				PlayoffSeed:   yahooTeam.TeamStandings.PlayoffSeed,
				PointsAgainst: yahooTeam.TeamStandings.PointsAgainst,
				PointsFor:     yahooTeam.TeamStandings.PointsFor,
				OutcomeTotals: entities.Outcome{
					Wins:       yahooTeam.TeamStandings.OutcomeTotals.Wins,
					Losses:     yahooTeam.TeamStandings.OutcomeTotals.Losses,
					Ties:       yahooTeam.TeamStandings.OutcomeTotals.Ties,
					Percentage: makePercentIntoInt(yahooTeam.TeamStandings.OutcomeTotals.Percentage),
				},
			},
			Manager: []entities.User{user},
		}
	}
	return teams
}

func makePercentIntoInt(perc string) int {
	percSplits := strings.SplitAfter(perc, ".")
	if len(percSplits) < 2 {
		return 0
	}
	percInt, _ := strconv.Atoi(percSplits[1])
	return percInt
}
