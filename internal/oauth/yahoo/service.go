package yahoo

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	trace2 "go.opentelemetry.io/otel/api/trace"
	"golang.org/x/oauth2"
	"net/http"
)

type Service struct {
	logger          log.Logger
	tracer          otel.Tracer
	oauthConfig     *oauth2.Config
	oauthRepository OauthRepository
}

func NewOauthYahooService(logger log.Logger, tracer otel.Tracer, oauthConfig *oauth2.Config, repo OauthRepository) Service {
	return Service{
		logger:          logger,
		tracer:          tracer,
		oauthConfig:     oauthConfig,
		oauthRepository: repo,
	}
}

type OauthRepository interface {
	SaveOauthToken(ctx context.Context, uuid string, token oauth2.Token) error
}

const guidKey string = "xoauth_yahoo_guid"
const state string = "state"
const code string = "code"

func (service Service) YahooCallback(ctx context.Context, r *http.Request) error {
	ctx, span := service.tracer.Start(ctx, "YahooCallback")
	defer span.End()

	values := r.URL.Query()
	state, ok := values[state]
	if !ok || len(state) != 0 {
		return errors.New("could not get state from request")
	}
	code, ok := values[code]
	if !ok || len(code) != 0 {
		return errors.New("could not get code from request")
	}

	token, err := service.getUserInfo(ctx, service.logger, state[0], code[0])
	if err != nil {
		level.Error(service.logger).Log("msg", "could not  get User from yahoo")
		return err
	}
	level.Debug(service.logger).Log("msg", "token received", "token", token)

	// get the raw key from oauth
	guidInterface := token.Extra(guidKey)
	guid, ok := guidInterface.(string)
	if !ok {
		_ = level.Error(service.logger).Log("msg", "Could not get key from")
	}
	_ = level.Debug(service.logger).Log("msg", "about to save guid", "guid", guid)
	go func() {
		// todo queue importer
		// QueuePlayers
	}()
	return service.oauthRepository.SaveOauthToken(ctx, guid, *token)
}

func (service Service) getUserInfo(ctx context.Context, logger log.Logger, state string, code string) (*oauth2.Token, error) {
	ctx, span := service.tracer.Start(ctx, "getUserInfo", trace2.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	// @todo make this into an auth state from the actual user and not just
	//if state != oauthStateString {
	if state == "" {
		return nil, fmt.Errorf("invalid OauthConfig state")
	}
	token, err := service.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	level.Debug(logger).Log("msg", "token received", "token", token)
	return token, nil
}

// save token info here.
//func (service Service) getUserGameInfoFromYahoo(ctx context.Context, token *oauth2.Token) ([]byte, error) {
//	ctx, span := service.tracer.Start(ctx, "getUserGameInfoFromYahoo", trace2.WithSpanKind(trace.SpanKindClient))
//	defer span.End()
//
//	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games", nil)
//	if err != nil {
//		return nil, fmt.Errorf("could not create http request: %v", err)
//	}
//	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
//	client := http.Client{}
//	response, err := client.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
//	}
//	defer response.Body.Close()
//	contents, err := ioutil.ReadAll(response.Body)
//
//	span.AddEvent(ctx, "response got")
//
//	if err != nil {
//		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
//	}
//	return contents, nil
//}
