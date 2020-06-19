package repositories

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/test_helpers"
	"os"
	"testing"
)

func TestMongoRepository_GetTeamsForManagers(t *testing.T) {

	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	leagues, err := mongoRepo.GetTeamsForManagers(context.TODO(), "DPPQCXCRV75Z2LKJW5YRC7RAYM")
	assert.Nil(t, err)
	assert.True(t, len(leagues) > 0, "Could not find leagues")
}

func TestMongoRepository_GetPlayers(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	playerKeys := []string{"390.p.100020", "390.p.31013"}

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	players, err := mongoRepo.GetPlayers(context.TODO(), playerKeys)
	assert.Nil(t, err)
	assert.True(t, len(players) == len(playerKeys), "Could not find leagues")
	for _, player := range players {
		assert.NotEqual(t, "", player.PlayerKey)
	}
}

func TestMongoRepository_GetTeamsForLeague(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	leagueKey := "390.l.705710"

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	teams, err := mongoRepo.GetTeamsForLeague(context.TODO(), leagueKey)
	assert.Nil(t, err)
	assert.True(t, len(teams) == 10, "Could not find teams")

}

func TestMongoRepository_SaveDraftOrder(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	leagueKey := "390.l.705710"

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	err = mongoRepo.SaveDraftOrder(context.TODO(), leagueKey, []string{"first", "second", "tenth"})
	assert.Nil(t, err)

	//assert.True(t, len(teams) == 10, "Could not find teams" )

	league, err := mongoRepo.GetLeague(context.TODO(), leagueKey)
	assert.Nil(t, err)
	assert.NotEqual(t, "", league.LeagueKey)
	assert.True(t, league.DraftStarted, "no draft did not start")
	//league.DraftStarted = true
	//_, err = mongoRepo.SaveLeagueLeague(context.TODO(), league)
	//
	//league, err = mongoRepo.GetLeague(context.TODO(), leagueKey)
	//assert.True(t,league.DraftStarted, "no draft did not start")

	t.FailNow()

}

func TestMongoRepository_GetDraftResults(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	leagueKey := "390.l.705710"

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	draftResults, err := mongoRepo.GetDraftResults(context.TODO(), leagueKey)
	assert.Nil(t, err)
	assert.Equal(t, 150, len(draftResults))
	for idx := range draftResults {
		assert.NotNil(t, draftResults[idx].Player)
		if draftResults[idx].Player == nil {
			t.Log(fmt.Printf("%d %s", draftResults[idx].Pick, draftResults[idx].PlayerKey))
		}
		assert.NotEqual(t, "", draftResults[idx].Player[0].PlayerKey)
	}

	t.FailNow()

}

func TestMongoRepository_GetAvailablePlayersForDraft(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	leagueKey := "390.l.705710"
	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	t.Run("something", func(t *testing.T) {

		players, err := mongoRepo.GetAvailablePlayersForDraft(context.TODO(), 390, leagueKey, 150, 0, []string{}, "")
		assert.Nil(t, err)
		assert.Equal(t, 150, len(players))
		for idx := range players {
			assert.NotNil(t, players[idx].Player.PlayerID)
			assert.NotEqual(t, "", players[idx].Player.PlayerID)
			assert.Nil(t, players[idx].DraftResult)
		}
		t.FailNow()
	})
	t.Run("get players by positions", func(t *testing.T) {
		players, err := mongoRepo.GetAvailablePlayersForDraft(context.TODO(), 390, leagueKey, 150, 0, []string{"QB", "WR"}, "")
		assert.Nil(t, err)
		assert.Equal(t, 150, len(players))
		for idx := range players {
			var QBorWR bool
			if players[idx].Player.EligiblePositions[0] == "QB" || players[idx].Player.EligiblePositions[0] == "WR" {
				QBorWR = true
			}
			assert.True(t, QBorWR, "Neither QB Nor WR")
			t.Log(players[idx].Player.EligiblePositions[0])
		}
	})
}

func TestMongoRepository_ImportLeagueAvailablePlayers(t *testing.T) {
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}
	leagueKey := "399.l.19481"

	mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	err = mongoRepo.ImportAllAvailablePlayers(context.TODO(), 399, leagueKey)
	assert.Nil(t, err)

	t.Log(err)
	t.FailNow()

}


func TestMongoRepository_SaveDraftResults2(t *testing.T) {
	//logger := test_helpers.LogrusLogger(t)
	//client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	//assert.Nil(t, err)
	//if t.Failed() {
	//	t.FailNow()
	//}
	//leagueKey := "399.l.19481"
	//
	//mongoRepo := NewMongoRepository(logger, client, "fdr", "draft", "fdr_user", "roster")
	//draftResult := entities.DraftResult{
	//	UserGUID:  "MFG5HMFDHC634Q7W2FPKJBVTKY",
	//	PlayerKey: "",
	//	PlayerID:  0,
	//	LeagueKey: "",
	//	TeamKey:      "",
	//	Round:     0,
	//	Pick:      0,
	//	Timestamp: time.Time{},
	//	GameID:    0,
	//	Player:    nil,
	//}
	//// get league
	//league, err := mongoRepo.GetLeague(context.TODO(), leagueKey)
	//if err != nil {
	//	t.Fail()
	//	t.FailNow()
	//}



}
