package handlers

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/internal/oauth/yahoo"
	"go.opentelemetry.io/otel"
	"golang.org/x/oauth2"
)

type YahooOauthEndpoints struct {
	logger             log.Logger
	oauthConfig        *oauth2.Config
	YahooLoginEndpoint endpoint.Endpoint
	YahooCallback      endpoint.Endpoint
}

func NewYahooHandlersEndpoints(logger log.Logger, config *oauth2.Config, tracer otel.Tracer, service *yahoo.Service, authMiddleWare endpoint.Middleware) YahooOauthEndpoints {
	var (
		yahooCallback = authMiddleWare(makeYahooCallback(tracer, service))
		yahooRedirect = makeYahooRedirect(config, tracer)
	)

	return YahooOauthEndpoints{logger: logger, YahooCallback: yahooCallback, YahooLoginEndpoint: yahooRedirect}
}
