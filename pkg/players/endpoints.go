package players

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	entities2 "github.com/thethan/fdr-users/pkg/users/entities"
	"go.elastic.co/apm"
	"sync"
)

type Endpoints struct {
	logger              log.Logger
	service             *Service
	GetAvailablePlayers endpoint.Endpoint
	GetLeagueDraft      endpoint.Endpoint
	SaveUsersDraftOrder endpoint.Endpoint
}

func NewEndpoint(logger log.Logger, service *Service, authMiddleWare endpoint.Middleware, getUserInfoMiddleWare endpoint.Middleware) Endpoints {
	return Endpoints{
		logger:              logger,
		service:             service,
		GetAvailablePlayers: authMiddleWare(getUserInfoMiddleWare(makeNewGetAvailablePlayers(logger, service))),
		SaveUsersDraftOrder: authMiddleWare(getUserInfoMiddleWare(makeSaveUserPlayerPrefEndpoint(logger, service))),
	}
}

type GetAvailablePlayersRequest struct {
	GameID    int
	LeagueKey string
	Limit     int
	Offset    int
	Positions []string
}

type GetAvailablePositionsForLeague struct {
	PlayerMap  map[string]entities.LeaguePlayer `json:"player_map"`
	PlayerKeys []string                         `json:"ap"`
	Players    []entities.LeaguePlayer          `json:"fdr-players-import"`
	Meta       Meta                             `json:"meta"`
	DoNotDraft []string                         `json:"dnd"`
	Pref       []string                         `json:"pref"`
}

type Meta struct {
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
	LeagueKey string   `json:"league_key"`
	Positions []string `json:"positions"`
}

func makeNewGetAvailablePlayers(logger log.Logger, service *Service, ) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		span, ctx := apm.StartSpan(ctx, "makeNewGetAvailablePlayers", "endpoint")
		defer span.End()

		req := request.(*GetAvailablePlayersRequest)
		players, err := service.GetAvailablePlayersForDraft(ctx, req.GameID, req.LeagueKey, req.Limit, req.Offset, req.Positions, "")
		if err != nil {
			return nil, err
		}

		userInterface := ctx.Value(auth.User)
		user, ok := userInterface.(*entities2.User)
		if !ok {
			return nil, errors.New("could not get user from context")
		}
		preference, err := service.GetUserPlayerPreference(ctx, user.GUID, req.LeagueKey)
		if err != nil {
			level.Debug(logger).Log("message", "could not get user preference", "error", err)
		}
		wg := &sync.WaitGroup{}

		playerMap := make(map[string]entities.LeaguePlayer, len(players))
		playerKeys := make([]string, len(players))
		playerKeyToIdx := make(map[string]int, len(players))
		playersDoNotDraft := make([]string, 0, len(preference.DoNotDraft))
		playersPref := make([]string, 0, len(preference.Preference))
		playersPrefMap := make(map[string]int, len(preference.Preference))
		playersDNDMap := make(map[string]int, len(preference.DoNotDraft))
		playersAvail := make([]string, 0, len(preference.Available))
		indexesToRemove := make([]int, 0)

		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for idx := range players {
				playerKeys[idx] = players[idx].Player.PlayerKey
				playerKeyToIdx[players[idx].Player.PlayerKey] = idx
			}
		}(wg)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for _, player := range players {
				playerMap[player.Player.PlayerKey] = player
			}
		}(wg)
		wg.Wait()

		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for _, playerKey := range preference.DoNotDraft {
				_, ok := playerMap[playerKey]
				if ok {
					playersDNDMap[playerKey] = len(playersDoNotDraft)
					playersDoNotDraft = append(playersDoNotDraft, playerKey)
					indexesToRemove = append(indexesToRemove, playerKeyToIdx[playerKey])
				}
			}
		}(wg)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for _, playerKey := range preference.Preference {
				_, ok := playerMap[playerKey]
				if ok {
					playersPrefMap[playerKey] = len(playersPref)

					playersPref = append(playersPref, playerKey)
					indexesToRemove = append(indexesToRemove, playerKeyToIdx[playerKey])
				}
			}
		}(wg)
		wg.Wait()
		//newPlayerKeysAvail := make([]string, 0, len(fdr-players-import))
		//wg.Add(1)
		//go func(wg *sync.WaitGroup) {
		//	defer wg.Done()
			for _, player := range players {
				playerKey := player.Player.PlayerKey
				_, ok := playerMap[playerKey]
				_, inDnd := playersDNDMap[playerKey]
				_, inPref := playersPrefMap[playerKey]

				if ok && !inPref && !inDnd {
					playersAvail = append(playersAvail, playerKey)
					//indexesToRemove = append(indexesToRemove, playerKeyToIdx[playerKey])
				}
			}
		//}(wg)
		//wg.Wait()

		/*for idx, idxToRemove := range indexesToRemove {
			playerKeys = append(playerKeys[:idxToRemove], playerKeys[idxToRemove+1-idx:]...)
		}*/
		//
		if len(preference.DoNotDraft) != 0 && len(preference.Preference) != 0 && len(preference.Available) != 0 {
			level.Debug(logger).Log("list", playerKeys, "ap", playersAvail)
			playerKeys = playersAvail
		}
		//level.Debug(logger).Log("list", playerKeys, "ap", playersAvail)

		response := GetAvailablePositionsForLeague{
			PlayerKeys: playersAvail,
			PlayerMap:  playerMap,
			Players:    players,
			DoNotDraft: playersDoNotDraft,
			Pref:       playersPref,
			Meta: Meta{
				PageSize: req.Limit,
				Page:     req.Offset / req.Limit,
			},
		}

		return &response, nil
	}
}

func makeSaveUserPlayerPrefEndpoint(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		span, ctx := apm.StartSpan(ctx, "makeSaveUserPlayerPrefEndpoint", "endpoint")
		defer span.End()

		return nil, nil
	}
}
