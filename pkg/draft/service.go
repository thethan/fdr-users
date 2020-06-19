package draft

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
	"math"
)

func NewService(logger log.Logger, repository draftRepository, ) Service {
	return Service{logger: logger, draftRepo: repository}
}

type draftRepository interface {
	GetDraftResults(ctx context.Context, leagueKey string) ([]entities.DraftResult, error)
	GetLeague(ctx context.Context, leagueKey string) (entities.League, error)
	ImportAllAvailablePlayers(ctx context.Context, gameID int, leagueKey string) error
	SaveDraftResultFromUser(ctx context.Context, league entities.League, user entities.User, team entities.Team, player entities.PlayerSeason, pick, round int) (*entities.DraftResult, error)
}

type Service struct {
	logger    log.Logger
	draftRepo draftRepository
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

func (service *Service) OpenDraft(ctx context.Context, leagueKey string) error {
	span, ctx := apm.StartSpan(ctx, "OpenDraft", "service")
	defer func() {
		span.End()
	}()

	// make
	return nil
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
	fmt.Printf("%v \n", round)
	draftResult, err := service.draftRepo.SaveDraftResultFromUser(ctx, league, user, team, player, pick, int(round))
	if err != nil {

		return nil, err
	}
	return draftResult, err
	//return &entities.DraftResult{
	//	UserGUID:  leaguePlayer.DraftResult.UserGUID,
	//	PlayerKey: leaguePlayer.DraftResult.PlayerKey,
	//	PlayerID:  leaguePlayer.DraftResult.PlayerID,
	//	LeagueKey: leaguePlayer.DraftResult.LeagueKey,
	//	TeamKey:   leaguePlayer.DraftResult.TeamKey,
	//	Round:     leaguePlayer.DraftResult.Round,
	//	Pick:      leaguePlayer.DraftResult.Pick,
	//	Timestamp: leaguePlayer.DraftResult.Timestamp,
	//	GameID:    leaguePlayer.League.Game.GameID,
	//	//LeaguePlayer: &leaguePlayer,
	//	Player: []*entities.PlayerSeason{&leaguePlayer.Player},
	//}, err
}
