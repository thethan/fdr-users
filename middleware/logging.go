package middleware

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LoggingMiddleware struct {}

func LogMiddleware(log log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			level.Info(log).Log("msg", "calling endpoint")
			defer level.Info(log).Log("msg", "calling endpoint")
			return next(ctx, request)
		}
	}
}
