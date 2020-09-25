package importer

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"net/http"
)

type YahooRepository interface {
	GetPlayerResourceStats(ctx context.Context, client *http.Client, playerKey string, week int) (*yahoo.PlayerResourcesStats, error)
}

// get from yahoo
type worker struct {
	logger log.Logger
	repo   YahooRepository
}

