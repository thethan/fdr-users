package handlers

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	http2 "github.com/go-kit/kit/transport/http"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"golang.org/x/oauth2"
	"net/http"
)

func decodeYahooRedirect(tracer trace.Tracer) http2.DecodeRequestFunc {

	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "decodeYahooRedirect")
		defer span.End()
		return r, nil
	}
}

func encodeYahooRedirect(tracer trace.Tracer) http2.EncodeResponseFunc {
	return func(ctx context.Context, writer http.ResponseWriter, i interface{}) error {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "encodeYahooRedirect")
		defer span.End()

		r := i.(yahooRedirect)
		http.Redirect(writer, r.r, r.url, http.StatusTemporaryRedirect)
		return nil
	}
}

type yahooRedirect struct {
	url string
	r   *http.Request
}

func makeYahooRedirect(config *oauth2.Config, tracer trace.Tracer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "yahooRedirect")
		defer span.End()
		r := request.(*http.Request)
		userInterface := ctx.Value(auth.User)
		user, ok := userInterface.(entities.User)
		if !ok {
			return nil, errors.New("could not get user from auth")
		}
		url := config.AuthCodeURL(user.Guid)
		return yahooRedirect{
			url: url,
			r:   r,
		}, nil

	}
}
