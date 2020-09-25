package yahoo

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.opencensus.io/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"strconv"
)

type YahooRepository struct {
	logger log.Logger
	conf   *oauth2.Config
}

func NewYahooRepository(logger log.Logger, conf *oauth2.Config) YahooRepository {
	return YahooRepository{logger: logger, conf: conf}
}

type ClientOption struct {
	HttpClient *http.Client
}

func (y YahooRepository) GetPlayerResourceStats(ctx context.Context, client *http.Client, playerKey string, week string) (*yahoo.PlayerResourcesStats, error) {
	ctx, span := trace.StartSpan(ctx, "GetPlayerResourceStats")
	span.AddAttributes(trace.StringAttribute("player_key", playerKey))
	span.AddAttributes(trace.StringAttribute("week", week))

	defer span.End()

	var weekString string
	if week != "" {
		weekString = fmt.Sprintf(";type=week;week=%s", week)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/stats%s", playerKey, weekString)

	req, _ := http.NewRequest("GET", url, nil)
	_, req = httptrace.W3C(ctx, req)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	v := yahoo.PlayerResourcesStats{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	return &v, nil
}

// GetGameResourcesPlayers
func (s *YahooRepository) GetGameResourcesPlayers(ctx context.Context, client *http.Client, gameKey int, start, count int) (yahoo.GameResourcePlayerResponse, error) {
	ctx, span := trace.StartSpan(ctx, "GetGameResourcesPlayers")
	span.AddAttributes(trace.StringAttribute("game_id", strconv.Itoa(gameKey)))
	span.AddAttributes(trace.Int64Attribute("offset", int64(count)))
	span.AddAttributes(trace.Int64Attribute("start", int64(start)))

	defer span.End()

	v := yahoo.GameResourcePlayerResponse{}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/game/%d/fdr-players-import/stats?start=%d&count=%d", gameKey, start, count)
	req, _ := http.NewRequest("GET", url, nil)
	_, req = httptrace.W3C(ctx, req)

	res, err := client.Do(req)

	if err != nil {
		return v, err
	}

	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return v, err
	}
	// transform response to games
	err = xml.Unmarshal(bytes, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return v, err
	}
	return v, nil
}
