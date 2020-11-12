package handlers

import (
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/thethan/fdr-users/internal/transports"
	"go.opentelemetry.io/otel"
	"net/http"
)

func MakeHTTPHandler(logger log.Logger, endpoints YahooOauthEndpoints, userRouter *mux.Router, authServerBefore httptransport.RequestFunc, tracer otel.Tracer, options ...httptransport.ServerOption) *mux.Router {
	serverOptions := []httptransport.ServerOption{
		httptransport.ServerBefore(transports.HeadersToContext),
		httptransport.ServerErrorEncoder(transports.ErrorEncoder),
		httptransport.ServerAfter(httptransport.SetContentType(transports.ContentType)),
	}
	serverOptions = append(serverOptions, options...)
	serverOptionsAuth := append(serverOptions, httptransport.ServerBefore(authServerBefore))

	userRouter.Methods(
		http.MethodGet).Path("/yahoo").Handler(
		httptransport.NewServer(
			endpoints.YahooLoginEndpoint,
			decodeYahooRedirect(tracer),
			encodeYahooRedirect(tracer),
			serverOptionsAuth...,
		))
	userRouter.Methods(
		http.MethodGet).Path("/auth").Handler(
		httptransport.NewServer(
			endpoints.YahooCallback,
			decodeYahooCallback(tracer),
			encodeYahooCallback(tracer),
			serverOptions...,
		))

	return userRouter
}
