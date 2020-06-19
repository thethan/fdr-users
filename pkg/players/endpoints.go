package players

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.elastic.co/apm"
)

type Endpoints struct {
	logger              log.Logger
	service             *Service
	GetAvailablePlayers endpoint.Endpoint
	GetLeagueDraft      endpoint.Endpoint
}

func NewEndpoint(logger log.Logger, service *Service,  authMiddleWare endpoint.Middleware) Endpoints {
	return Endpoints{
		logger:              logger,
		service:             service,
		GetAvailablePlayers: authMiddleWare(makeNewGetAvailablePlayers(logger, service)),
	}
}

type GetAvailablePlayersRequest struct {
	GameID    int
	LeagueKey string
	Limit     int
	Offset    int
	Positions []string
}

type GetAvailablePositionsForLeague struct {
	Players []entities.LeaguePlayer `json:"players"`
	Meta    Meta                    `json:"meta"`
}

type Meta struct {
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
	LeagueKey string   `json:"league_key"`
	Positions []string `json:"positions"`
}

func makeNewGetAvailablePlayers(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		span, ctx := apm.StartSpan(ctx, "makeNewGetAvailablePlayers", "endpoint")
		defer span.End()

		req := request.(*GetAvailablePlayersRequest)
		players, err := service.GetAvailablePlayersForDraft(ctx, req.GameID, req.LeagueKey, req.Limit, req.Offset, req.Positions, "")
		if err != nil {
			return nil, err
		}
		response := GetAvailablePositionsForLeague{
			Players: players,
			Meta: Meta{
				PageSize: req.Limit,
				Page:     req.Offset / req.Limit,
			},
		}

		return &response, nil
	}
}
