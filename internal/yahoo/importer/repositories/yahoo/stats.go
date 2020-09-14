package yahoo

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
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

func (y YahooRepository) GetPlayerResourceStats(ctx context.Context, client *http.Client, playerKey string, week int) (*yahoo.PlayerResourcesStats, error) {
	span, ctx := apm.StartSpan(ctx, "GetPlayerResourcesStats", "client")

	defer span.End()

	var weekString string
	if week > 0 {
		weekString = fmt.Sprintf(";type=week;week=%d", week)
	}
	url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/player/%s/stats%s", playerKey, weekString)
	res, err := client.Get(url)
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
