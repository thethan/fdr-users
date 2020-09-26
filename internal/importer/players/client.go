package players

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	yahoo2 "github.com/thethan/fdr-users/internal/importer/repositories/yahoo"
	pkgEntities "github.com/thethan/fdr-users/pkg/draft/entities"
	entities2 "github.com/thethan/fdr-users/pkg/players/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Queue interface {
	SendPlayerStatMessage(ctx context.Context, stats entities2.ImportPlayerStat) error
	SendPlayerMessage(ctx context.Context, stats entities2.ImportPlayer) error
	ReceiveMessages(context.Context, chan<- entities2.ImportPlayerStat) error
}

type playersRepo interface {
	GetPlayersByRank(ctx context.Context, limit, offset, gameID int) ([]pkgEntities.PlayerSeason, error)
	SavePlayers(context.Context, []pkgEntities.PlayerSeason) ([]pkgEntities.PlayerSeason, error)
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
	mu           *sync.RWMutex
}

// NewImporterClient will take
func NewImporterClient(logger log.Logger, playerRepo playersRepo, queuer Queue, yahooRepo *yahoo2.YahooRepository, statsRepo SavePlayerStats) Service {
	return Service{logger: logger, playersRepo: playerRepo, queuer: queuer, yahooService: yahooRepo, statsRepo: statsRepo, mu: &sync.RWMutex{}}
}

// QueuePlayers
func (c Service) QueueAllPlayers(ctx context.Context, guid string, gameID int) {
	counter := 0
	count := 25
	for counter < 60 {
		startingCounter := count * counter
		importer := entities2.ImportPlayer{
			Guid:   guid,
			GameID: gameID,
			Offset: startingCounter,
			Limit:  count,
		}

		err := c.queuer.SendPlayerMessage(ctx, importer)
		if err != nil {
			level.Error(c.logger).Log("message", "could not save", "error", err, "offset", startingCounter, "limit", count, "game_id", gameID)
		}
		counter++
	}
	return
}

func transformPlayerResponseToPlayers(ctx context.Context, gameID int, gamePlayers yahoo.GameResourcePlayerResponse, startingCounter int) []pkgEntities.PlayerSeason {
	span, ctx := apm.StartSpan(ctx, "transformPlayerResponseToPlayers", "service.importer")
	defer span.End()

	players := make([]pkgEntities.PlayerSeason, len(gamePlayers.Game.Players.Player))
	for idx, yahooPlayer := range gamePlayers.Game.Players.Player {
		player := transformPlayerResponseToPlayer(ctx, gameID, yahooPlayer, (idx+1)+startingCounter)
		player.SeasonStats = transformYahooSeasonStats(ctx, yahooPlayer)
		players[idx] = player
	}

	return players
}

func transformPlayerResponseToPlayer(ctx context.Context, GameID int, yahooPlayer yahoo.GameResourcePlayerStats, rank int) pkgEntities.PlayerSeason {
	span, ctx := apm.StartSpan(ctx, "transformPlayerResponseToPlayer", "service.importer")
	defer span.End()

	playerRank := make(map[string]int)
	playerRank["yahoo"] = rank

	player := pkgEntities.PlayerSeason{
		PlayerKey: yahooPlayer.PlayerKey,
		PlayerID:  yahooPlayer.PlayerID,
		GameID:    GameID,
		Name: pkgEntities.PlayerName{
			Full:       yahooPlayer.Name.Full,
			First:      yahooPlayer.Name.First,
			Last:       yahooPlayer.Name.Last,
			AsciiFirst: yahooPlayer.Name.AsciiFirst,
			AsciiLast:  yahooPlayer.Name.AsciiLast,
		},
		EditorialPlayerKey:    yahooPlayer.EditorialPlayerKey,
		EditorialTeamKey:      yahooPlayer.EditorialTeamKey,
		EditorialTeamFullName: yahooPlayer.EditorialTeamFullName,
		EditorialTeamAbbr:     yahooPlayer.EditorialTeamAbbr,
		ByeWeeks:              pkgEntities.PlayerByeWeeks{Week: yahooPlayer.ByeWeeks.Week},
		UniformNumber:         yahooPlayer.UniformNumber,
		DisplayPosition:       yahooPlayer.DisplayPosition,
		Headshot: pkgEntities.PlayerHeadshot{
			URL:  yahooPlayer.Headshot.URL,
			Size: yahooPlayer.Headshot.Size,
		},
		ImageURL:          yahooPlayer.ImageURL,
		IsUndroppable:     intToBool(yahooPlayer.IsUndroppable),
		PositionType:      yahooPlayer.PositionType,
		EligiblePositions: []string{yahooPlayer.EligiblePositions.Position},
		Ranks:             playerRank,
	}
	return player
}
func intToBool(i int) bool {
	return i == 1
}

func transformYahooSeasonStats(ctx context.Context, yahooPlayer yahoo.GameResourcePlayerStats) []pkgEntities.PlayerStat {
	playerStats := make([]pkgEntities.PlayerStat, len(yahooPlayer.PlayerStats.Stats))
	for idx, stat := range yahooPlayer.PlayerStats.Stats {
		playerStats[idx] = pkgEntities.PlayerStat{
			StatID: stat.StatID,
			Value:  float64(stat.Value),
		}
	}
	return playerStats
}

// QueuePlayersStats
func (c Service) QueuePlayersStats(ctx context.Context, guid string, gameID int) {
	limit := 100
	offset := 0
	for {
		players, err := c.playersRepo.GetPlayersByRank(ctx, limit, 0, offset)
		if err != nil {
			level.Error(c.logger).Log("message", "could not get fdr-players-import by rank", "error", err, "offset", offset, "limit", limit)
			return
		}

		for _, player := range players {
			for week := 1; week < 20; week++ {
				go func(player pkgEntities.PlayerSeason, week int) {
					// make player leu
					stat := entities2.ImportPlayerStat{
						Guid:        guid,
						PlayerKey:   player.EditorialPlayerKey,
						Week:        strconv.Itoa(week),
						Season:      strconv.Itoa(player.GameID),
						PlayerStats: nil,
					}
					_ = c.queuer.SendPlayerStatMessage(ctx, stat)
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
}

// Start starts the import of user's playuers
func (s Service) Start(ctx context.Context, messageChannel <-chan entities2.ImportPlayerStat) error {
	errorChan := make(chan error)

	for {
		select {
		case <-ctx.Done():
			level.Info(s.logger).Log("message", "context is closed")
			return nil
		case err := <-errorChan:
			panic("Error in reading from channel")
			return err
		case msg := <-messageChannel:
			client := s.getClientService(ctx, msg.Guid)

			err := s.getAndSavePlayerStats(ctx, client, msg.PlayerKey, msg.Week)
			if err != nil {
				s.mu.Lock()
				s.clientGuid[msg.Guid] = nil
				s.mu.Unlock()

				level.Error(s.logger).Log("message", "could not get oauth token")
				continue
			}
		}
	}
}

// Start starts the import of user's playuers
func (s Service) StartPlayersWorker(ctx context.Context, messageChannel <-chan entities2.ImportPlayer) error {
	errorChan := make(chan error)

	for {
		select {
		case <-ctx.Done():
			level.Info(s.logger).Log("message", "context is closed")
			return nil
		case err := <-errorChan:
			panic("Error in reading from channel")
			return err
		case msg, ok := <-messageChannel:
			if !ok {
				continue
			}
			err := s.getAndSavePlayers(ctx, msg)
			if err != nil {
				s.mu.Lock()
				s.clientGuid[msg.Guid] = nil
				s.mu.Unlock()

				level.Error(s.logger).Log("message", "could not get oauth token")
				continue
			}
		}
	}
}

func (s *Service) getClientService(ctx context.Context, guid string) *http.Client {
	if client, ok := s.clientGuid[guid]; !ok || client == nil {
		token, err := s.oauthRepo.GetUserOAuthToken(ctx, guid)
		if err != nil {
			level.Error(s.logger).Log("message", "could not get oauth token")
		}

		s.mu.Lock()
		s.clientGuid[guid] = oauth2.NewClient(ctx, s.oauthConfig.TokenSource(ctx, &token))
		s.mu.Unlock()
		return client
	}

	s.mu.Lock()
	client := s.clientGuid[guid]
	s.mu.Unlock()
	return client
}

func (s Service) getAndSavePlayerStats(ctx context.Context, client *http.Client, playerKey string, week string) error {

	level.Debug(s.logger).Log("message", "getAndSavePlayerStats", "player_key", playerKey, "week")

	resp, err := s.yahooService.GetPlayerResourceStats(ctx, client, playerKey, week)
	if err != nil {
		level.Error(s.logger).Log("message", "could not yahoo player stats", "player_key", playerKey, week)
		return err
	}

	playerStats := playerYahooStatToEntities(resp)
	games := strings.Split(playerKey, ".")
	err = s.statsRepo.SavePlayerStats(ctx, playerKey, games[0], week, playerStats)
	if err != nil {
		level.Error(s.logger).Log("message", "could not save player stats", "error", err, "player_key", playerKey)
		return err
	}
	return nil
}

func (c Service) getAndSavePlayers(ctx context.Context, msg entities2.ImportPlayer) error {

	client := c.getClientService(ctx, msg.Guid)
	res, err := c.yahooService.GetGameResourcesPlayers(ctx, client, msg.GameID, msg.Offset, msg.Limit)
	if err != nil {
		level.Error(c.logger).Log("message", "could not get fdr-players-import by rank", "error", err, "offset", msg.Offset, "limit", msg.Limit)
		return err
	}

	players := transformPlayerResponseToPlayers(ctx, msg.GameID, res, msg.Limit)
	_, err = c.playersRepo.SavePlayers(ctx, players)
	return err
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
