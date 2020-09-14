package importer

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/internal/yahoo/importer/entities"
	yahoo2 "github.com/thethan/fdr-users/internal/yahoo/importer/repositories/yahoo"
	pkgEntities "github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Queue interface {
	GetPlayerStats(ctx context.Context, userGuid, playerKey string, week int) error
	Start(context.Context, chan<- entities.Message)
}

type playersRepo interface {
	GetPlayersByRank(ctx context.Context, limit, offset int) ([]pkgEntities.PlayerSeason, error)
}

type oauthRepo interface {
	GetUserOAuthToken(ctx context.Context, guid string) (oauth2.Token, error)
}

type SavePlayerStats interface {
	SavePlayerStats(ctx context.Context, playerKey string, season, week string, stats []pkgEntities.PlayerStat) error
}

type YahooService interface {
	GetPlayerResourcesStats(ctx context.Context, playerKey string, week int, ) (*yahoo.PlayerResourcesStats, error)
}

type Service struct {
	logger      log.Logger
	playersRepo playersRepo
	queuer      Queue
	oauthRepo   oauthRepo
	statsRepo   SavePlayerStats

	yahooService *yahoo2.YahooRepository
	oauthConfig  *oauth2.Config
	clientGuid   map[string]*http.Client
}

// NewImporterClient will take
func NewImporterClient(logger log.Logger, playerRepo playersRepo, queuer Queue, yahooRepo *yahoo2.YahooRepository, statsRepo SavePlayerStats) Service {
	return Service{logger: logger, playersRepo: playerRepo, queuer: queuer, yahooService: yahooRepo, statsRepo: statsRepo}
}

func (c Service) QueuePlayers(ctx context.Context, guid string) {
	limit := 100
	offset := 0
	for {
		players, err := c.playersRepo.GetPlayersByRank(ctx, limit, offset)
		if err != nil {
			level.Error(c.logger).Log("message", "could not get players by rank", "error", err, "offset", offset, "limit", limit)
			return
		}

		for _, player := range players {
			for week := 1; week < 20; week++ {
				go func(player pkgEntities.PlayerSeason, week int) {
					_ = c.queuer.GetPlayerStats(ctx, guid, player.PlayerKey, week)
				}(player, week)
			}
		}
		// break here
		if len(players) < limit {
			level.Error(c.logger).Log("message", "finish sending to kubemq", "offset", offset, "limit", limit)
			return
		}
		offset += limit
	}
	return
}

func (s Service) Start(ctx context.Context, ) {
	messageChannel := make(chan entities.Message, 5)
	go func() {
		s.queuer.Start(ctx, messageChannel)
	}()

	mu := sync.RWMutex{}

	for {
		select {
		case <-ctx.Done():
			level.Info(s.logger).Log("message", "context is closed")
			return
		case msg := <-messageChannel:
			if client, ok := s.clientGuid[msg.Guid]; !ok || client == nil {
				token, err := s.oauthRepo.GetUserOAuthToken(ctx, msg.Guid)
				if err != nil {
					level.Error(s.logger).Log("message", "could not get oauth token")
					continue
				}

				mu.Lock()
				s.clientGuid[msg.Guid] = oauth2.NewClient(ctx, s.oauthConfig.TokenSource(ctx, &token))
				mu.Unlock()

				level.Error(s.logger).Log("message", "could not yahoo player stats")
			}

			mu.Lock()
			client := s.clientGuid[msg.Guid]
			mu.Unlock()

			err := s.getAndSavePlayerStats(ctx, client, msg.PlayerKey, msg.Week)
			if err != nil {
				mu.Lock()
				s.clientGuid[msg.Guid] = nil
				mu.Unlock()

				level.Error(s.logger).Log("message", "could not get oauth token")
				continue
			}
		}
	}
}

func (s Service) getAndSavePlayerStats(ctx context.Context, client *http.Client, playerKey string, week int) error {
	span, ctx := apm.StartSpan(ctx, "getAndSavePlayerStats", "service")
	defer span.End()

	resp, err := s.yahooService.GetPlayerResourceStats(ctx, client, playerKey, week)
	if err != nil {
		level.Error(s.logger).Log("message", "could not yahoo player stats", "player_key", playerKey, week)
		return err
	}

	playerStats := playerYahooStatToEntities(resp)
	games := strings.Split(playerKey, ".")
	err = s.statsRepo.SavePlayerStats(ctx, playerKey, games[0], strconv.Itoa(week), playerStats)
	if err != nil {
		level.Error(s.logger).Log("message", "could not save player stats", "error", err, "player_key", playerKey)
		return err
	}
	return nil
}

func playerYahooStatToEntities(stats *yahoo.PlayerResourcesStats) []pkgEntities.PlayerStat {
	playerStat := make([]pkgEntities.PlayerStat, len(stats.Stats.Stats))
	for idx, stat := range stats.Stats.Stats {
		value, _ := strconv.ParseFloat(stat.Value, 32)
		statIdInt, _ := strconv.Atoi(stat.StatID)

		playerStat[idx] = pkgEntities.PlayerStat{
			StatID: statIdInt,
			Value:  value,
		}
	}
	return playerStat
}
