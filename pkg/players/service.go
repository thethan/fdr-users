package players

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
)

type GetUserRepo interface {
	GetCredentialInformation(ctx context.Context, session string) (entities.User, error)
}

type Repository interface {
	GetAvailablePlayersForDraft(ctx context.Context, gameID int, leagueKey string, limit, offset int, eligiblePositions []string, search string) ([]entities.LeaguePlayer, error)
	GetUserPlayerPreference(ctx context.Context, userGUID, leagueKey string) (entities.UserPlayerPreference, error)
}

type Service struct {
	logger   log.Logger
	repo     Repository
	userRepo GetUserRepo
}

func NewService(logger log.Logger, repository Repository) Service {
	return Service{
		logger:   logger,
		repo:     repository,
	}
}

func (s Service) GetAvailablePlayersForDraft(ctx context.Context, gameID int, leagueKey string, limit, offset int, eligiblePositions []string, search string) ([]entities.LeaguePlayer, error) {
	span, ctx := apm.StartSpan(ctx, "GetAvailablePlayersForDraft", "service")
	defer span.End()

	return s.repo.GetAvailablePlayersForDraft(ctx, gameID, leagueKey, limit, offset, eligiblePositions, "")
}

func (s Service) GetUserPlayerPreference(ctx context.Context, userGuid, leagueKey string) (entities.UserPlayerPreference, error) {

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

	pref, err := s.repo.GetUserPlayerPreference(ctx, userGuid, leagueKey)
	if err != nil {
		return pref, err
	}

	return pref, err
}
