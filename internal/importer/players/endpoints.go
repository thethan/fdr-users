package players

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"go.opentelemetry.io/otel"
)

type Endpoints struct {
	logger log.Logger
	tracer otel.Tracer
	meter  *otel.Meter

	QueuePlayersImport endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, tracer otel.Tracer, metrics *otel.Meter) Endpoints {
	return Endpoints{
		logger:             logger,
		tracer:             tracer,
		meter:              metrics,
		QueuePlayersImport: makeQueuePlayerEndpoint(),
	}
}

func makeQueuePlayerEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return nil, nil
	}
}
