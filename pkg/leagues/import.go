package leagues

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"sync"
)

type SaveLeagueIFace interface {
	SaveLeague(ctx context.Context, group *LeagueGroup) (*LeagueGroup, error)
}

type dataTransferObjects struct {
	mu                  *sync.Mutex
	gameMapToIdx        map[int]int
	leagueMaps          map[int]int
	teamMaps            map[int]int
	userMaps            map[string]int // tied to email
	games               []*Game
	leagues             []*League
	users               []*User
	teams               []*Team
	leagueGroups        []*LeagueGroup
	leagueToMapGroupIds map[int]int
	leagueGroupMaps     map[int]int
}

func (d *dataTransferObjects) addLeague(league League) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.leagues = append(d.leagues, &league)
	d.leagueMaps[league.LeagueID] = len(d.leagues)
}

func (d *dataTransferObjects) addTeam(team Team) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for idx := range team.Manager {
		if useridx, ok := d.userMaps[team.Manager[idx].Email]; ok {
			team.Manager[idx] = *d.users[useridx]
		}
	}
	d.teams = append(d.teams, &team)
	d.teamMaps[team.TeamID] = len(d.teams)
}

func (d *dataTransferObjects) addUser(user User) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.userMaps[user.ManagerID]; !ok {
		d.users = append(d.users, &user)
		d.userMaps[user.Email] = len(d.users)
	}
}

func (d *dataTransferObjects) addGames(game Game) *Game {
	d.mu.Lock()
	defer d.mu.Unlock()
	gameIdx, ok := d.gameMapToIdx[game.GameID]
	if !ok {
		currentGameLen := len(d.games)
		d.games[currentGameLen] = &game
		d.gameMapToIdx[game.GameID] = len(d.games)
		return &game
	}
	return d.games[gameIdx]
}

func (d *dataTransferObjects) addLeagueToLeagueGroup(groupID int, league *League) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// find group IDx
	var leagueGroup *LeagueGroup
	// see if already created by finding the index
	leagueToGroupIdx, ok := d.leagueToMapGroupIds[groupID]
	if !ok {
		leagueGroup = &LeagueGroup{
			FirstLeagueID: groupID,
			Leagues:       make([]*League, 0, 400),
		}
		d.leagueGroups = append(d.leagueGroups, leagueGroup)
		d.leagueGroupMaps[groupID] = len(d.leagueGroups)-1
		leagueToGroupIdx = d.leagueGroupMaps[groupID]

	} else {
		leagueGroup = d.leagueGroups[d.leagueGroupMaps[leagueToGroupIdx]]
	}

	d.leagueGroups[leagueToGroupIdx].Leagues = append(d.leagueGroups[leagueToGroupIdx].Leagues, league)
	d.leagueToMapGroupIds[league.LeagueID] = leagueToGroupIdx
}

type User struct {
	Email          string `json:"email" bson:"_id"`
	Name           string `json:"name"`
	ManagerID      string `json:"manager_id"`
	Nickname       string `json:"nickname"`
	Guid           string `json:"guid"`
	IsCommissioner bool   `json:"is_commissioner"`
	IsCurrentLogin bool   `json:"is_current_login"`
	ImageURL       string `json:"image_url"`
	Teams          []Team
	Commissioned   []primitive.ObjectID
}

type LeagueGroup struct {
	ID            primitive.ObjectID `bson:"_id"`
	FirstLeagueID int
	Leagues       []*League
}

type League struct {
	ID             primitive.ObjectID `bson:"_id"`
	Name           string
	LeagueID       int
	LeagueGroup    primitive.ObjectID                 `bson:"league_group"`
	PreviousLeague *primitive.ObjectID                `bson:"previous_league,omitempty"`
	Settings       *LeagueSettings                    // year in which settings are for
	DraftResults   *yahoo.LeagueResourcesDraftResults // This is results by Year
	Teams          []Team                             // season is the key
	Game           Game                               // season is the key
}

type LeagueSettings struct {
	LeagueKey                  string           `xml:"league_key"`
	LeagueID                   int              `xml:"league_id"`
	Name                       string           `xml:"name"`
	URL                        string           `xml:"url"`
	LogoURL                    string           `xml:"logo_url"`
	Password                   string           `xml:"password"`
	DraftStatus                string           `xml:"draft_status"`
	NumTeams                   int              `xml:"num_teams"`
	EditKey                    string           `xml:"edit_key"`
	WeeklyDeadline             string           `xml:"weekly_deadline"`
	LeagueUpdateTimestamp      string           `xml:"league_update_timestamp"`
	LeagueType                 string           `xml:"league_type"`
	Renew                      string           `xml:"renew"`
	Renewed                    string           `xml:"renewed"`
	IrisGroupChatID            string           `xml:"iris_group_chat_id"`
	ShortInvitationURL         string           `xml:"short_invitation_url"`
	AllowAddToDlExtraPos       string           `xml:"allow_add_to_dl_extra_pos"`
	IsProLeague                string           `xml:"is_pro_league"`
	IsCashLeague               string           `xml:"is_cash_league"`
	CurrentWeek                string           `xml:"current_week"`
	StartWeek                  string           `xml:"start_week"`
	StartDate                  string           `xml:"start_date"`
	EndWeek                    string           `xml:"end_week"`
	EndDate                    string           `xml:"end_date"`
	GameCode                   string           `xml:"game_code"`
	Season                     string           `xml:"season"`
	ID                         int              `json:"id" bson:"_id"`
	DraftType                  string           `json:"draft_type"`
	IsAuctionDraft             bool             `json:"is_auction_draft"`
	ScoringType                string           `json:"scoring_type"`
	PersistentURL              string           `json:"persistent_url"`
	UsesPlayoff                string           `json:"uses_playoff"`
	HasPlayoffConsolationGames bool             `json:"has_playoff_consolation_games"`
	PlayoffStartWeek           string           `json:"playoff_start_week"`
	UsesPlayoffReseeding       bool             `json:"uses_playoff_reseeding"`
	UsesLockEliminatedTeams    bool             `json:"uses_lock_eliminated_teams"`
	NumPlayoffTeams            int              `json:"num_playoff_teams"`
	NumPlayoffConsolationTeams int              `json:"num_playoff_consolation_teams"`
	UsesRosterImport           bool             `json:"uses_roster_import"`
	RosterImportDeadline       string           `json:"roster_import_deadline"`
	WaiverType                 string           `json:"waiver_type"`
	WaiverRule                 string           `json:"waiver_rule"`
	UsesFaab                   bool             `json:"uses_faab"`
	DraftTime                  string           `json:"draft_time"`
	PostDraftPlayers           string           `json:"post_draft_players"`
	MaxTeams                   string           `json:"max_teams"`
	WaiverTime                 string           `json:"waiver_time"`
	TradeEndDate               string           `json:"trade_end_date"`
	TradeRatifyType            string           `json:"trade_ratify_type"`
	TradeRejectTime            string           `json:"trade_reject_time"`
	PlayerPool                 string           `json:"player_pool"`
	CantCutList                string           `json:"cant_cut_list"`
	IsPubliclyViewable         bool             `json:"is_publicly_viewable"`
	RosterPositions            []RosterPosition `json:"roster_positions"`
	StatCategories             []StatCategory   `json:"stat_categories"`
	StatModifiers              []StatModifier   `json:"stat_modifiers"`
	MaxAdds                    int              `json:"max_adds"`
	SeasonType                 string           `json:"season_type"`
	MinInningsPitched          string           `json:"min_innings_pitched"`
	UsesFractalPoints          bool             `json:"uses_fractal_points"`
	UsesNegativePoints         bool             `json:"uses_negative_points"`
}

type RosterPosition struct {
	Position     string `json:"position"`
	PositionType string `json:"position_type,omitempty"`
	Count        int    `json:"count"`
}

type PositionType struct {
	PositionType      string `json:"position_type"`
	IsOnlyDisplayStat bool   `json:"is_only_display_stat"`
}
type StatCategory struct {
	StatID            int            `json:"stat_id"`
	Enabled           bool           `json:"enabled"`
	Name              string         `json:"name"`
	DisplayName       string         `json:"display_name"`
	SortOrder         int            `json:"sort_order"`
	PositionType      string         `json:"position_type"`
	StatPositionTypes []PositionType `json:"stat_position_types"`
	IsOnlyDisplayStat string         `json:"is_only_display_stat,omitempty"`
}

type StatModifier struct {
	StatID  int     `json:"stat_id"`
	Value   float32 `json:"value"`
	Bonuses *Bonus  `json:"bonuses,omitempty"`
}

type Bonus struct {
	Target float32
	Points float32
}

type LeagueStandings struct {
	TeamKey          string `json:"team_key"`
	TeamID           string `json:"team_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	TeamLogo         string `json:"team_logo"`
	WaiverPriority   int    `json:"waiver_priority"`
	NumberOfMoves    string `json:"number_of_moves"`
	NumberOfTrades   string `json:"number_of_trades"`
	ClinchedPlayoffs int    `json:"clinched_playoffs,omitempty"`
	Managers         []User `json:"managers"`
}

type Game struct {
	GameID             int    `json:"game_id"`
	GameKey            string `json:"game_key"`
	Name               string `json:"name"`
	Code               string `json:"code"`
	Type               string `json:"type"`
	URL                string `json:"url"`
	Season             int    `json:"season"`
	IsRegistrationOver bool   `json:"is_registration_over"`
	IsGameOver         bool   `json:"is_game_over"`
	IsOffseason        bool   `json:"is_offseason"`
}

type TeamLogo struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

type RosterAdds struct {
	CoverageType  string `json:"coverage_type"`
	CoverageValue int    `json:"coverage_value"`
	Value         int    `json:"value"`
}

type Team struct {
	League                primitive.ObjectID `json:"league_id"`
	TeamKey               string             `json:"team_key"`
	TeamID                int                `json:"team_id"`
	Name                  string             `json:"name"`
	IsOwnedByCurrentLogin bool               `json:"is_owned_by_current_login"`
	URL                   string             `json:"url"`
	TeamLogo              TeamLogo           `json:"team_logo"`
	WaiverPriority        int                `json:"waiver_priority"`
	NumberOfMoves         int                `json:"number_of_moves"`
	NumberOfTrades        int                `json:"number_of_trades"`
	RosterAdds            RosterAdds         `json:"roster_adds"`
	LeagueScoringType     string             `json:"league_scoring_type"`
	HasDraftGrade         bool               `json:"has_draft_grade"`
	DraftGrade            string             `json:"draft_grade"`
	Standing              Standings          `json:"standing"`
	Manager               []User             `json:"managers"`
}

type Outcome struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	Ties       int `json:"ties"`
	Percentage int `json:"percentage"`
}

type Standings struct {
	Rank          int     `json:"rank"`
	PlayoffSeed   int     `json:"playoff_seed"`
	OutcomeTotals Outcome `json:"outcome_totals"`
	GamesBack     string  `json:"games_back"`
	PointsFor     float32 `json:"points_for"`
	PointsAgainst float32 `json:"points_against"`
}

type Importer struct {
	logger log.Logger

	yahooService *yahoo.Service
}

func NewImportService(logger log.Logger, yahooService *yahoo.Service) Importer {
	return Importer{
		logger:       logger,
		yahooService: yahooService,
	}
}

func newDTO() dataTransferObjects {
	return dataTransferObjects{
		mu:                  &sync.Mutex{},
		games:               make([]*Game, 0, 400),
		leagues:             make([]*League, 0, 1000),
		leagueGroups:        make([]*LeagueGroup, 0, 1000),
		leagueGroupMaps:     map[int]int{},
		leagueToMapGroupIds: map[int]int{},
		leagueMaps:          make(map[int]int),
		teamMaps:            make(map[int]int),
		userMaps:            make(map[string]int),
		users:               make([]*User, 0, 1000),
		teams:               make([]*Team, 0, 1000),
	}
}

func (i *Importer) ImportLeagueFromUser(ctx context.Context) ([]*LeagueGroup, error) {
	span, ctx := apm.StartSpan(ctx, "ImportLeagueFromUsers", "service")
	defer func() {
		span.End()
	}()
	// string is yahoo's guid
	//var userTeams map[string]Team
	res, err := i.yahooService.GetUserResourcesGameLeaguesResponse(ctx)
	if err != nil {
		level.Error(i.logger).Log("message", "could not get GetUserResourcesGameLeaguesResponse from yahoo", "error", err)
		return []*LeagueGroup{}, err
	}

	a := newDTO()
	dto := &a

	for _, gameRes := range res.Users.User.Games.Game {
		// add game to dto
		game := transformGameResponseToGame(gameRes)
		dto.games = append(dto.games, &game)

		//yahooLeagues := make([]yahoo.YahooLeague, res.Users.User.Games.Count)
		for _, leagueRes := range gameRes.Leagues {
			dto, err = i.getLeague(ctx, game, leagueRes, dto)
			if err != nil {
				level.Error(i.logger).Log("message", "error in getting league information", "league_key", leagueRes.LeagueKey, "error", err)
				continue
			}
		}
	}

	// filter out the composition for dto
	for _, league := range dto.leagues {

		// find league group
		parentLeagueID := 0
		ids := strings.Split(league.Settings.Renewed, "_")
		if len(ids) > 1 {
			parentLeagueID, _ = strconv.Atoi(ids[1])
		}

		// check if parentID exists
		dto.addLeagueToLeagueGroup(parentLeagueID, league)

	}
	return dto.leagueGroups, nil

}

func (i *Importer) getLeague(ctx context.Context, game Game, leagueRes yahoo.YahooLeague, dto *dataTransferObjects) (*dataTransferObjects, error) {
	level.Debug(i.logger).Log("name", leagueRes.Name)
	var league League

	league.Game = game
	league.Name = leagueRes.Name
	// get league settings
	yahooSettings, err := i.yahooService.GetLeagueResourcesSettings(ctx, leagueRes.LeagueKey)
	if err != nil {
		level.Error(i.logger).Log("message", "error in getting league Resource setting", "error", err)
		return dto, err
	}
	settings := transformYahooLeagueSettingsToLeagueSettings(yahooSettings)
	league.Settings = &settings
	league.LeagueID = settings.LeagueID
	// get league standings
	yahooStandings, err := i.yahooService.GetLeagueResourcesStandings(ctx, leagueRes.LeagueKey)
	if err != nil {
		level.Error(i.logger).Log("message", "error in getting league Resource standings", "error", err)
		return dto, err
	}

	teams := transformYahooStandingsToStandings(yahooStandings)
	league.Teams = teams
	dto.addLeague(league)
	return dto, nil

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

func transformGameResponseToGame(res yahoo.YahooGame) Game {
	return Game{
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

func transformTeamResponseToTeam(res yahoo.TeamResponse) Team {
	return Team{}
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

func transformYahooLeagueSettingsToLeagueSettings(response *yahoo.LeagueResourcesSettingsResponse) LeagueSettings {
	leagueSettings := LeagueSettings{
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


		ID:                         response.League.LeagueID,
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
func transformYahooRosterPositionsToRosterPositions(yahooRosterPositions []yahoo.ResponseRosterPosition) []RosterPosition {
	rosterPositions := make([]RosterPosition, len(yahooRosterPositions))

	for idx, yahooRosPos := range yahooRosterPositions {
		rosterPositions[idx] = RosterPosition{
			Position:     yahooRosPos.Position,
			PositionType: yahooRosPos.PositionType,
			Count:        yahooRosPos.Count,
		}
	}

	return rosterPositions
}

func transformYahooStatCategoriesToStatCategories(yahooStatCategories []yahoo.ResponseStatCategory) []StatCategory {
	stateCategories := make([]StatCategory, len(yahooStatCategories))

	for idx, yahooStatCategories := range yahooStatCategories {
		stateCategories[idx] = StatCategory{
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

func transformYahooPositionTypes(yahooPositionTypes []yahoo.ResponsePositionType) []PositionType {
	positionTypes := make([]PositionType, len(yahooPositionTypes))

	for idx, positionType := range yahooPositionTypes {
		positionTypes[idx] = PositionType{
			PositionType:      positionType.PositionType,
			IsOnlyDisplayStat: intToBool(positionType.IsOnlyDisplayStat),
		}
	}
	return positionTypes
}

func transformYahooStatModifiers(yahooStatModifiers []yahoo.StatModifier) []StatModifier {
	statModifiers := make([]StatModifier, len(yahooStatModifiers))

	for idx, yahooStatModifier := range yahooStatModifiers {
		statModifiers[idx] = StatModifier{
			StatID:  yahooStatModifier.StatID,
			Value:   yahooStatModifier.Value,
			Bonuses: transformYahooStatModifierBonusToBonus(yahooStatModifier.Bonus),
		}
	}
	return statModifiers
}

func transformYahooStatModifierBonusToBonus(yahooBonus *yahoo.Bonus) *Bonus {
	if yahooBonus == nil {
		return nil
	}

	return &Bonus{
		Target: yahooBonus.Target,
		Points: yahooBonus.Points,
	}
}

func transformManagerToUser(manager yahoo.Manager) User {
	return User{
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

func transformYahooStandingsToStandings(standings *yahoo.LeagueResourcesStandingsResponse) []Team {

	teams := make([]Team, standings.League.NumTeams)

	for idx, yahooTeam := range standings.League.Standings.Teams.Team {
		user := transformManagerToUser(yahooTeam.Managers.Manager)
		teams[idx] = Team{

			League:                primitive.NewObjectID(),
			TeamKey:               yahooTeam.TeamKey,
			TeamID:                yahooTeam.TeamID,
			Name:                  yahooTeam.Name,
			IsOwnedByCurrentLogin: intToBool(yahooTeam.IsOwnedByCurrentLogin),
			URL:                   yahooTeam.URL,
			TeamLogo:              TeamLogo{},
			WaiverPriority:        yahooTeam.WaiverPriority,
			NumberOfMoves:         yahooTeam.NumberOfMoves,
			NumberOfTrades:        yahooTeam.NumberOfTrades,
			RosterAdds: RosterAdds{
				CoverageType:  yahooTeam.RosterAdds.CoverageType,
				CoverageValue: yahooTeam.RosterAdds.CoverageValue,
				Value:         yahooTeam.RosterAdds.Value,
			},
			LeagueScoringType: yahooTeam.LeagueScoringType,
			HasDraftGrade:     intToBool(yahooTeam.HasDraftGrade),
			DraftGrade:        yahooTeam.DraftGrade,
			Standing: Standings{
				Rank:          yahooTeam.TeamStandings.Rank,
				PlayoffSeed:   yahooTeam.TeamStandings.PlayoffSeed,
				PointsAgainst: yahooTeam.TeamStandings.PointsAgainst,
				PointsFor:     yahooTeam.TeamStandings.PointsFor,
				OutcomeTotals: Outcome{
					Wins:       yahooTeam.TeamStandings.OutcomeTotals.Wins,
					Losses:     yahooTeam.TeamStandings.OutcomeTotals.Losses,
					Ties:       yahooTeam.TeamStandings.OutcomeTotals.Ties,
					Percentage: makePercentIntoInt(yahooTeam.TeamStandings.OutcomeTotals.Percentage),
				},
			},
			Manager: []User{user},
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
