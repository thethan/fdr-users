package draft

import (
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	userEntities "github.com/thethan/fdr-users/pkg/users/entities"
	"go.elastic.co/apm"
	"math"
	"math/rand"
	"reflect"
	"time"
)

func NewService(logger log.Logger, repository draftRepository, broadcastRepo broadCastRepo) Service {
	return Service{logger: logger, draftRepo: repository, broadCastRepo: broadcastRepo}
}

type draftRepository interface {
	GetDraftResults(ctx context.Context, leagueKey string) ([]entities.DraftResult, error)
	GetLeague(ctx context.Context, leagueKey string) (entities.League, error)
	SaveLeague(ctx context.Context, league entities.League) (entities.League, error)
	ImportAllAvailablePlayers(ctx context.Context, gameID int, leagueKey string) error
	SaveDraftResultFromUser(ctx context.Context, league entities.League, user entities.User, team entities.Team, player entities.PlayerSeason, pick, round int) (*entities.DraftResult, error)
	GetTeamDraftResultsByTeam(ctx context.Context, leagueKey string) (map[string][]entities.DraftResult, error)
	SaveUserPlayerPreference(ctx context.Context, preference entities.UserPlayerPreference) error
	GetUserPlayerPreference(ctx context.Context, userGUID, leagueKey string) (entities.UserPlayerPreference, error)
}

type broadCastRepo interface {
	BroadCastDraftResult(ctx context.Context, league entities.League, user entities.User, team entities.Team, draftResult entities.DraftResult, pick, round int, rosters map[string]entities.Roster) error
	BroadCastLeagueInformation(ctx context.Context, league entities.League, message string, broadcastType entities.BroadcastType) error
	ChangeTeamName(ctx context.Context, league entities.League, user entities.User, team entities.Team) error
}

type Service struct {
	logger    log.Logger
	draftRepo draftRepository
	broadCastRepo
}

func (service *Service) ListDraftResults(ctx context.Context, leagueKey string) (*entities.League, []entities.DraftResult, error) {
	span, ctx := apm.StartSpan(ctx, "ListDraftResults", "service")
	defer func() {
		span.End()
	}()

	league, err := service.draftRepo.GetLeague(ctx, leagueKey)
	if err != nil {
		level.Error(service.logger).Log("message", "could not get league", "error", err)
		return nil, nil, err
	}
	results, err := service.draftRepo.GetDraftResults(ctx, leagueKey)
	if err != nil {
		level.Error(service.logger).Log("message", "could not get draft results", "error", err)

		return nil, nil, err
	}

	return &league, results, err
}

func (service *Service) GetTeamsDraftResults(ctx context.Context, leagueKey string) (map[string]entities.Roster, error) {
	span, ctx := apm.StartSpan(ctx, "GetTeamsDraftResults", "service")
	defer func() {
		span.End()
	}()

	league, err := service.draftRepo.GetLeague(ctx, leagueKey)
	if err != nil {
		level.Error(service.logger).Log("message", "could not get league", "error", err)
		return nil, err
	}

	return service.buildRosters(ctx, league)
}

func (service Service) buildRosters(ctx context.Context, league entities.League) (map[string]entities.Roster, error) {
	results, err := service.draftRepo.GetTeamDraftResultsByTeam(ctx, league.LeagueKey)
	if err != nil {
		level.Error(service.logger).Log("message", "could not get draft results", "error", err)

		return nil, err
	}

	rosters := make(map[string]entities.Roster, len(league.Teams))

	for teamKey, draftResults := range results {
		roster := makeRoster(league)
		teamRoster := buildTeamRoster(draftResults, roster)
		rosters[teamKey] = teamRoster
	}
	return rosters, nil
}

func makeRoster(league entities.League) entities.Roster {
	teamRoster := make(map[string]entities.RosterSlot, len(league.Settings.RosterPositions))
	for _, pos := range league.Settings.RosterPositions {
		teamRoster[pos.Position] = entities.RosterSlot{Count: pos.Count, DraftResults: []entities.DraftResult{}}
	}
	return entities.Roster{Roster: teamRoster}
}

func buildTeamRoster(teamDraftResults []entities.DraftResult, roster entities.Roster) entities.Roster {
	for _, teamDraftResult := range teamDraftResults {
		rosterSlotKey := teamDraftResult.Player[0].EligiblePositions[0]
		if roster.CanAddResult(rosterSlotKey) {
			roster = addToRoster(rosterSlotKey, roster, teamDraftResult)
			continue
		}

		if (rosterSlotKey == "RB" || rosterSlotKey == "WR" || rosterSlotKey == "TE") && roster.CanAddResult("W/R/T") {
			rosterSlotKey = "W/R/T"
			roster = addToRoster(rosterSlotKey, roster, teamDraftResult)
			continue
		}

		if !roster.CanAddResult(rosterSlotKey) {
			// add roster to bench
			rosterSlotKey = "BN"
			roster = addToRoster(rosterSlotKey, roster, teamDraftResult)
		}
	}
	return roster
}

func addToRoster(rosterSlotKey string, roster entities.Roster, result entities.DraftResult) entities.Roster {
	roster.AddResult(rosterSlotKey, result)
	return roster
}

func (s *Service) OpenDraft(ctx context.Context, leagueKey string) (*entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "OpenDraft", "service")
	defer func() {
		span.End()
	}()

	league, err := s.draftRepo.GetLeague(ctx, leagueKey)
	if err != nil {
		level.Error(s.logger).Log("message", "could get draft", "error", err)
		return nil, err
	}
	if !isUserCommissioner(ctx, league) {
		return nil, errors.New("user is not commissioner")
	}
	league.DraftStarted = true

	league, err = s.draftRepo.SaveLeague(ctx, league)
	if err != nil {
		level.Error(s.logger).Log("message", "could not save draft", "error", err)
		return nil, err
	}
	err = s.broadCastRepo.BroadCastLeagueInformation(ctx, league, "league is opened", 0)

	if err != nil {
		level.Error(s.logger).Log("message", "could not save draft", "error", err)
		return nil, err
	}
	return &league, err
}

func isUserCommissioner(ctx context.Context, league entities.League) bool {
	userInterface := ctx.Value(auth.User)
	reflect.TypeOf(userInterface)
	user, ok := userInterface.(*userEntities.User)
	if !ok {
		return false
	}
	for _, team := range league.Teams {
		for _, manager := range team.Manager {
			if user.GUID == manager.Guid {
				return manager.IsCommissioner
			}
		}
	}
	return false
}

func (s *Service) ShuffleOrder(ctx context.Context, leagueKey string) (*entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "OpenDraft", "service")
	defer func() {
		span.End()
	}()

	league, err := s.draftRepo.GetLeague(ctx, leagueKey)
	if err != nil {
		level.Error(s.logger).Log("message", "could get draft", "error", err)
		return nil, err
	}

	if !isUserCommissioner(ctx, league) {
		return nil, errors.New("user is not commissioner")
	}

	if league.DraftStarted {
		_ = level.Error(s.logger).Log("message", "draft is started and could not change order", "error", errors.New("could not change order"))
	}

	draftOrder := league.DraftOrder
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(draftOrder), func(i, j int) { draftOrder[i], draftOrder[j] = draftOrder[j], draftOrder[i] })
	league.DraftOrder = draftOrder

	league, err = s.draftRepo.SaveLeague(ctx, league)
	if err != nil {
		level.Error(s.logger).Log("message", "could not save draft", "error", err)
		return nil, err
	}
	err = s.broadCastRepo.BroadCastLeagueInformation(ctx, league, "draft order change", 0)

	if err != nil {
		level.Error(s.logger).Log("message", "could not save draft", "error", err)
		return nil, err
	}
	return &league, err
}

// let numRounds = 0
//        league.settings.roster_positions.map((position, idx) => {
//            numRounds += position.count
//        })
//        return numRounds
func getRound(pick int, league entities.League) float64 {
	count := 0
	positions := league.Settings.RosterPositions
	for _, pos := range positions {
		count += pos.Count
	}
	something := float64(pick) / float64(len(league.DraftOrder))
	roundMod := math.Ceil(something)

	return roundMod
}

func (service *Service) SaveDraftRequest(ctx context.Context, user entities.User, reqKey entities.League, team entities.Team, player entities.PlayerSeason, pick int) (*entities.DraftResult, error) {
	span, ctx := apm.StartSpan(ctx, "SaveDraftRequest", "service")
	span.Context.SetLabel("user_id", user.Guid)
	span.Context.SetLabel("league_key", reqKey.LeagueKey)
	defer func() {
		span.End()
	}()
	// get league
	league, err := service.draftRepo.GetLeague(ctx, reqKey.LeagueKey)
	if err != nil {
		level.Error(service.logger).Log("message", "could not get key", "err", err, "league_key", reqKey.LeagueKey)
		return nil, err
	}
	round := getRound(pick, league)

	draftResult, err := service.draftRepo.SaveDraftResultFromUser(ctx, league, user, team, player, pick, int(round))
	if err != nil {
		level.Error(service.logger).Log("message", "error in updating draft result", "error", err)

		return nil, &ErrorUpdateDraft{}
	}

	rosters, err := service.buildRosters(ctx, league)
	if err != nil {
		return nil, err
	}

	err = service.broadCastRepo.BroadCastDraftResult(ctx, league, user, team, *draftResult, pick, int(round), rosters)
	if err != nil {
		level.Error(service.logger).Log("message", "error in broadcasting", "error", err)
		return nil, &ErrorUpdateDraft{}
	}
	return draftResult, err
}

func (service *Service) SaveUserPlayerPreference(ctx context.Context, preference entities.UserPlayerPreference) error {
	span, ctx := apm.StartSpan(ctx, "SaveUserPlayerPreference", "service")
	defer func(span *apm.Span) {
		span.Context.SetTag("user_guid", preference.UserID)
		span.Context.SetTag("league_key", preference.LeagueKey)
		span.End()
	}(span)

	return service.draftRepo.SaveUserPlayerPreference(ctx, preference)
}

func (service *Service) GetUserPlayerPreference(ctx context.Context, userGuid, leagueKey string) (entities.UserPlayerPreference, error) {
	span, ctx := apm.StartSpan(ctx, "GetUserPlayerPreference", "service")
	defer func(span *apm.Span) {
		span.Context.SetTag("user_guid", userGuid)
		span.Context.SetTag("league_key", leagueKey)
		span.End()
	}(span)
	//
	//wg := &sync.WaitGroup{}
	//wg.Add(2)
	//wg.Wait()

	pref, err := service.draftRepo.GetUserPlayerPreference(ctx, userGuid, leagueKey)
	if err != nil {
		return pref, err
	}

	return pref, err
}
