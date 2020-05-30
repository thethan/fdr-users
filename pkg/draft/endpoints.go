package draft

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
)

type Endpoints struct {
	logger         log.Logger
	service        *Service
	ImportDraft    endpoint.Endpoint
	GetLeagueDraft endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, service *Service) Endpoints {
	e := Endpoints{
		logger:         logger,
		ImportDraft:    makeImportDraft(logger, service),
		GetLeagueDraft: makeGetDraftInfo(logger, service),
	}

	return e
}

func makeImportDraft(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "ImportDraft", "endpoint")
		defer span.End()

		return nil, nil
	}
}

type DraftResultResponse struct {
	League       *entities.League       `json:"league"`
	DraftResults []entities.DraftResult `json:"draft_results"`
}

type LeagueDraftRequest struct {
	LeagueKey string
}

func makeGetDraftInfo(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "GetDraftInfo", "endpoint")
		defer span.End()

		req, ok := request.(*LeagueDraftRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		league, results, err := service.ListDraftResults(ctx, req.LeagueKey)
		if err != nil{
			return nil, err
		}
		teamKeyToIdx := make(map[string]int, len(league.Teams))
		for idx := range league.Teams {
			teamKeyToIdx[league.Teams[idx].TeamKey] = idx
		}
		teams := make([]entities.Team, len(league.Teams))
		for idx, teamKey := range league.DraftOrder {
			teams[idx] = league.Teams[teamKeyToIdx[teamKey]]
		}
		league.TeamDraftOrder = teams
		return &DraftResultResponse{DraftResults: results, League: league}, err
	}
}
