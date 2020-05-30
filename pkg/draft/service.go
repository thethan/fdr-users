package draft

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
)

func NewService(logger log.Logger, repository draftRepository, ) Service {
	return Service{logger: logger, draftRepo: repository}
}

type draftRepository interface {
	GetDraftResults(ctx context.Context, leagueKey string) ([]entities.DraftResult, error)
	GetLeague(ctx context.Context, leagueKey string) (entities.League, error)
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
