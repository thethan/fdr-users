package players

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
)

type Repository interface {
	GetAvailablePlayersForDraft(ctx context.Context, gameID int, leagueKey string, limit, offset int, eligiblePositions []string, search string) ([]entities.LeaguePlayer, error)
}

type Service struct {
	logger log.Logger
	repo   Repository
}

func NewService(logger log.Logger, repository Repository) Service {
	return Service{
		logger: logger,
		repo:   repository,
	}
}

func (s Service) GetAvailablePlayersForDraft(ctx context.Context, gameID int, leagueKey string, limit, offset int, eligiblePositions []string, search string) ([]entities.LeaguePlayer, error) {
	span, ctx := apm.StartSpan(ctx, "GetAvailablePlayersForDraft", "service")
	defer span.End()


	return s.repo.GetAvailablePlayersForDraft(ctx, gameID, leagueKey, limit, offset, eligiblePositions , "")
}