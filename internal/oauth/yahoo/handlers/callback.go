package handlers

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	http2 "github.com/go-kit/kit/transport/http"
	"github.com/thethan/fdr-users/internal/oauth/yahoo"
	"go.opentelemetry.io/otel/api/trace"

	"net/http"
)

func decodeYahooCallback(tracer trace.Tracer) http2.DecodeRequestFunc {

	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "decodeYahooCallback")

		defer span.End()
		return r, nil
	}
}

func encodeYahooCallback(tracer trace.Tracer) http2.EncodeResponseFunc {
	return func(ctx context.Context, writer http.ResponseWriter, i interface{}) error {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "encodeYahooCallback")
		defer span.End()

		writer.WriteHeader(http.StatusAccepted)

		return nil
	}
}

func makeYahooCallback(tracer trace.Tracer, service *yahoo.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "yahooCallback")
		defer span.End()
		r := request.(*http.Request)
		err = service.YahooCallback(ctx, r)
		return nil, err
	}
}
