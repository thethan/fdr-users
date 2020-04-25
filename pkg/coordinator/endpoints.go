package coordinator

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/sirupsen/logrus"

	"go.elastic.co/apm"
)

type Endpoints struct {
	logrus logrus.FieldLogger
	ImportUserLeagues endpoint.Endpoint
}

func NewEndpoints(logger logrus.FieldLogger, service Service, authMiddleWare endpoint.Middleware) Endpoints {

	return Endpoints{
		logrus:            logger,
		ImportUserLeagues: authMiddleWare(makeImportUserLeagues(logger, service)),
	}
}

func makeImportUserLeagues(logger logrus.FieldLogger, service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "endpoint.importUserLeagues", "custom")
		defer span.End()

		return service.importUserLeagues(ctx, request)
	}
}

