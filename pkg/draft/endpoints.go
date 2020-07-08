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
	logger          log.Logger
	service         *Service
	ImportDraft     endpoint.Endpoint
	GetLeagueDraft  endpoint.Endpoint
	SaveDraftResult endpoint.Endpoint
	GetTeamRoster   endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, service *Service) Endpoints {
	e := Endpoints{
		logger:          logger,
		ImportDraft:     makeImportDraft(logger, service),
		GetLeagueDraft:  makeGetDraftInfo(logger, service),
		SaveDraftResult: makeSaveDraftResult(logger, service),
		GetTeamRoster:   makeGetTeamDraftRoster(logger, service),
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

type DraftTeamRostersResponse struct {
	Teams map[string]entities.Roster `json:"rosters"`
}

type LeagueDraftRequest struct {
	LeagueKey string
}

type User struct {
	Email string `json:"email"`
	Guid  string `json:"guid"`
	UID   string `json:"uid"`
}

type SaveDraftResultRequest struct {
	User   entities.User         `json:"user"`
	League entities.League       `json:"league"`
	Player entities.PlayerSeason `json:"player"`
	Team   entities.Team         `json:"team"`
	Pick   int                   `json:"pick"`
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
		if err != nil {
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

func makeGetTeamDraftRoster(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "GetTeamDraftRoster", "endpoint")
		defer span.End()

		req, ok := request.(*LeagueDraftRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		rosters, err := service.GetTeamsDraftResults(ctx, req.LeagueKey)
		if err != nil {
			return nil, err
		}

		return &DraftTeamRostersResponse{Teams: rosters}, err
	}
}

func makeSaveDraftResult(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveDraftRequest", "endpoint")
		defer span.End()

		req, ok := request.(*SaveDraftResultRequest)
		if !ok {
			level.Error(logger).Log("message", "could not get request")
			return nil, errors.New("bad request for get draft")
		}

		result, err := service.SaveDraftRequest(ctx, req.User, req.League, req.Team, req.Player, req.Pick)
		if err != nil {
			return nil, err
		}
		return &DraftResultResponse{
			DraftResults: []entities.DraftResult{*result},
		}, err
	}
}
