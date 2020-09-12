package league

import (
	"context"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm"
	"golang.org/x/oauth2"
)

func (i Importer) ImportGamePlayers(ctx context.Context, gameID int) error {
	span, ctx := apm.StartSpan(ctx, "ImportGamePlayers", "service.importer")
	defer span.End()
	// count is 25
	counter := 0
	count := 25
	var fin bool
	for !fin {
		startingCounter := count * counter
		res, err := i.yahooService.GetGameResourcesPlayers(ctx, gameID, startingCounter, count)
		if err != nil {
			// complain
			level.Error(i.logger).Log("error", err)
			return err

		}
		if len(res.Game.Players.Player) < 25 {
			fin = true
		}
		players := transformPlayerResponseToPlayers(ctx, gameID, res, startingCounter)
		_, err = i.repo.SavePlayers(ctx, players)
		if err != nil {
			level.Error(i.logger).Log("message", "could not save players", "error", err)
		}
		counter++
	}

	return nil
}

type ImportStats struct {
	Token     oauth2.Token
	PlayerKey string
}

func (i Importer) ImportPlayerStats(ctx context.Context, gameID int) error {
	//span, ctx := apm.StartSpan(ctx, "ImportPlayerStats", "service.importer")
	//defer span.End()
	//// count is 25
	//counter := 1
	//count := 17
	//
	//users.OauthConfig.Client(ctx, )
	//var fin bool
	//for !fin {
	//	startingCounter := count * counter
	//	res, err := i.yahooService.GetPlayerResourcesStats(ctx, gameID, startingCounter, count)
	//	if err != nil {
	//		// complain
	//		level.Error(i.logger).Log("error", err)
	//		return err
	//
	//	}
	//	if len(res.Game.Players.Player) < 25 {
	//		fin = true
	//	}
	//	players := transformPlayerResponseToPlayers(ctx, gameID, res, startingCounter)
	//	_, err = i.repo.SavePlayers(ctx, players)
	//	if err != nil {
	//		level.Error(i.logger).Log("message", "could not save players", "error", err)
	//	}
	//	counter++
	//}

	return nil
}

func (i Importer) ImportGamePlayersUserHasAccessTo(ctx context.Context, guid string) error {
	span, ctx := apm.StartSpan(ctx, "ImportGamePlayersUserHasAccessTo", "service")
	defer func() {
		span.End()
	}()

	gameHashMap := make(map[int]bool)
	leagues, err := i.repo.GetTeamsForManagers(ctx, guid)
	if err != nil {
		return err
	}
	for _, league := range leagues {
		if imported, ok := gameHashMap[league.Game.GameID]; !imported || !ok {
			err = i.ImportGamePlayers(ctx, league.Game.GameID)
			if err != nil {
				level.Error(i.logger).Log("message", "error in importing game", "err", err, "game_id", league.Game.GameID)
			}
			gameHashMap[league.Game.GameID] = true
		}
	}

	return nil
}

func transformPlayerResponseToPlayers(ctx context.Context, gameID int, gamePlayers yahoo.GameResourcePlayerResponse, startingCounter int) []entities.PlayerSeason {
	span, ctx := apm.StartSpan(ctx, "transformPlayerResponseToPlayers", "service.importer")
	defer span.End()

	players := make([]entities.PlayerSeason, len(gamePlayers.Game.Players.Player))
	for idx, yahooPlayer := range gamePlayers.Game.Players.Player {
		player := transformPlayerResponseToPlayer(ctx, gameID, yahooPlayer, (idx+1)+startingCounter)
		player.SeasonStats = transformYahooSeasonStats(ctx, yahooPlayer)
		players[idx] = player
	}

	return players
}

func transformPlayerResponseToPlayer(ctx context.Context, GameID int, yahooPlayer yahoo.GameResourcePlayerStats, rank int) entities.PlayerSeason {
	span, ctx := apm.StartSpan(ctx, "transformPlayerResponseToPlayer", "service.importer")
	defer span.End()

	playerRank := make(map[string]int)
	playerRank["yahoo"] = rank

	player := entities.PlayerSeason{
		PlayerKey: yahooPlayer.PlayerKey,
		PlayerID:  yahooPlayer.PlayerID,
		GameID:    GameID,
		Name: entities.PlayerName{
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
		ByeWeeks:              entities.PlayerByeWeeks{Week: yahooPlayer.ByeWeeks.Week},
		UniformNumber:         yahooPlayer.UniformNumber,
		DisplayPosition:       yahooPlayer.DisplayPosition,
		Headshot: entities.PlayerHeadshot{
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

func transformYahooSeasonStats(ctx context.Context, yahooPlayer yahoo.GameResourcePlayerStats) []entities.PlayerStat {
	playerStats := make([]entities.PlayerStat, len(yahooPlayer.PlayerStats.Stats))
	for idx, stat := range yahooPlayer.PlayerStats.Stats {
		playerStats[idx] = entities.PlayerStat{
			StatID: stat.StatID,
			Value:  stat.Value,
		}
	}
	return playerStats

}
