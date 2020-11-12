package coordinator

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/league"
	"go.elastic.co/apm"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
)

type Endpoints struct {
	logrus                logrus.FieldLogger
	ImportUserLeagues     endpoint.Endpoint
	ImportGamePlayers     endpoint.Endpoint
	ImportGamePlayerStats endpoint.Endpoint
}

func NewEndpoints(logger logrus.FieldLogger, service league.Importer, authMiddleWare endpoint.Middleware) Endpoints {

	return Endpoints{
		logrus:                logger,
		ImportUserLeagues:     authMiddleWare(makeImportUserLeagues(logger, service)),
		ImportGamePlayers:     authMiddleWare(makeImportGamePlayers(logger, service)),
		ImportGamePlayerStats: authMiddleWare(makeImportGamePlayersStats(logger, service)),
	}
}

func makeImportUserLeagues(logger logrus.FieldLogger, service league.Importer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "endpoint.importUserLeagues", "custom")
		defer span.End()

		//req, ok := request.(ImportGamePlayersRequest)
		//if !ok {
		//	return nil, errors.New("bad request")
		//}

		_, err = service.ImportLeagueFromUser(ctx)
		return nil, err
	}
}

type ImportGamePlayersRequest struct {
	GameID int
}

func makeImportGamePlayers(logger logrus.FieldLogger, service league.Importer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "endpoint.importGamePlayers", "custom")
		defer span.End()

		req, ok := request.(ImportGamePlayersRequest)
		if !ok {
			return nil, errors.New("bad request")
		}

		err = service.ImportGamePlayers(ctx, req.GameID)
		return nil, err
	}
}

func makeImportGamePlayersStats(logger logrus.FieldLogger, service league.Importer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span := trace.SpanFromContext(ctx)
		defer span.End()

		req, ok := request.(ImportGamePlayersRequest)
		span.AddEvent(ctx, "makeImportGamePlayersStats", label.Int("game_id", req.GameID))

		span.SetAttribute("game_id", req.GameID)

		if !ok {
			return nil, errors.New("bad request")
		}

		err = service.ImportPlayerStats(ctx, req.GameID)
		return nil, err
	}
}
