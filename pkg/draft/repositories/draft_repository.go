package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	pb "github.com/thethan/fdr_proto"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

const database string = "fdr"
const leaguesCollection string = "leagues"
const draftsCollection string = "draft_results"
const playersBySeason string = "fdr-players-import"
const playersByWeek string = "fdr-players-import"
const userPreference string = "user_preferences"

type MongoRepository struct {
	logger                 log.Logger
	client                 *mongo.Client
	database               string
	draftCollection        string
	draftGroupCollection   string
	draftResultsCollection string
	userCollection         string
	rosterCollection       string
}

type Draft struct {
	ID                  primitive.ObjectID `bson:"_id"`
	Year                int
	League              int
	DraftType           int
	DraftTime           primitive.DateTime
	Users               []User
	Commissioners       []User
	Roster              primitive.ObjectID
	ExternalIdentifiers []ExternalIdentifiers
}

type ExternalIdentifiers struct {
	YearsActive      []int
	ExternalResource string
	ExternalID       string
}

type User struct {
	Name         string
	Image        string
	Email        string `bson:"_id"`
	Drafts       []primitive.ObjectID
	Commissioned []primitive.ObjectID
}

type RosterRules struct {
	ID          primitive.ObjectID `bson:"_id"`
	RosterRules []RosterRule
}

type RosterRule struct {
	Position int
	Starting int32
	Max      int32
}

func NewMongoRepository(logger log.Logger, client *mongo.Client, database string, draftCollection, userCollection, rosterCollection string) MongoRepository {
	return MongoRepository{logger: logger, client: client, database: database,
		draftCollection:        draftCollection,
		draftResultsCollection: "draft_results",
		userCollection:         userCollection,
		rosterCollection:       rosterCollection}
}

func getPlayerLeagueCollection(leagueKey string) string {
	tableVar := strings.Replace(leagueKey, ".", "", 2)
	return "players_per_league_" + tableVar
}

func (m MongoRepository) ImportAllAvailablePlayers(ctx context.Context, gameID int, leagueKey string) error {
	newTable := getPlayerLeagueCollection(leagueKey)

	tableCollection := m.client.Database(database).Collection(newTable)

	_ = tableCollection.Drop(ctx)
	models := []mongo.IndexModel{
		{
			Keys: bson.M{
				"fdr-players-import.name.full": "text",
			},
		},
		{
			Keys: bson.M{
				"league_key": 1,
			},
		},
		{
			Keys: bson.M{
				"fdr-players-import.player_key": 1,
			},
		},
		{
			Keys: bson.M{
				"fdr-players-import.eligiblepositions": 1,
			},
		},
		{
			Keys: bson.M{
				"draft_results.team_key": 1,
			},
		},
		{
			Keys: bson.M{
				"draft_results.user_guid": 1,
			},
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := tableCollection.Indexes().CreateMany(ctx, models, opts)
	if err != nil {
		return err
	}
	// create Index

	pipeline := mongo.Pipeline{
		bson.D{{"$project", bson.M{"_id": 0, "league_key": 1, "game.game_id": 1, "leagues": "$$ROOT"}}},
		bson.D{{"$lookup", bson.M{"from": playersBySeason, "localField": "game.game_id", "foreignField": "game_id", "as": "player"}}},
		bson.D{{"$unwind", bson.M{"path": "$player", "preserveNullAndEmptyArrays": false}}},
		bson.D{{"$lookup", bson.M{"from": draftsCollection, "localField": "player._id", "foreignField": "player_key", "as": "draft_results"}}},
		bson.D{{"$unwind", bson.M{"path": "$draft_results", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$match", bson.M{"league_key": leagueKey}}},
		bson.D{{"$out", newTable}},
	}
	collection := m.client.Database(database).Collection(leaguesCollection)
	findOptions := &options.AggregateOptions{
		AllowDiskUse: aws.Bool(true),
	}
	_, err = collection.Aggregate(ctx, pipeline, findOptions)

	return err
}

//

type QueryBson struct {
	ID           primitive.ObjectID    `bson:"_id"`
	Leagues      entities.League       `bson:"leagues"`
	Players      entities.PlayerSeason `bson:"fdr-players-import"`
	DraftResults entities.DraftResult  `bson:"draft_results"`
}

func (m MongoRepository) GetAvailablePlayersForDraft(ctx context.Context, gameID int, leagueKey string, limit, offset int, eligiblePositions []string, search string) ([]entities.LeaguePlayer, error) {
	span, ctx := apm.StartSpan(ctx, "GetAvailablePlayersForDraft", "repository.Mongo")
	defer span.End()

	newTable := getPlayerLeagueCollection(leagueKey)

	collection := m.client.Database(database).Collection(newTable)
	findOptions := &options.AggregateOptions{}
	elements := make([]bson.E, 0, 4)

	elements = append(elements, bson.E{Key: "draft_results", Value: bson.M{"$exists": false}})
	elements = append(elements, bson.E{Key: "league_key", Value: leagueKey})

	if len(eligiblePositions) > 0 {
		elements = append(elements, bson.E{Key: "player.eligiblepositions", Value: bson.M{"$in": eligiblePositions}})
	}

	if search != "" {
		elements = append(elements, bson.E{Key: "player.name.full", Value: bson.M{"$regex": fmt.Sprintf(".*%s.*", search)}})
	}

	skip := offset * limit
	pipeline := mongo.Pipeline{
		bson.D{{"$match", elements}},
		bson.D{{"$skip", skip}},
		bson.D{{"$limit", limit}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline, findOptions)
	if err != nil {
		level.Error(m.logger).Log("message", "could not get fdr-players-import results", "error", err, "game_id", gameID)
		return []entities.LeaguePlayer{}, nil
	}

	players := make([]entities.LeaguePlayer, limit)
	err = cursor.All(ctx, &players)
	if err != nil {
		return []entities.LeaguePlayer{}, nil
	}

	return players, nil
}

func (m MongoRepository) GetUserPlayerPreference(ctx context.Context, userGUID, leagueKey string) (entities.UserPlayerPreference, error) {
	span, ctx := apm.StartSpan(ctx, "GetUserPlayerPreference", "repository.Mongo")
	defer span.End()
	collection := m.client.Database(database).Collection(userPreference)
	findQuert := bson.M{"league_key": leagueKey, "user_guid": userGUID}

	res := collection.FindOne(ctx, findQuert)

	var playerPreference entities.UserPlayerPreference
	err := res.Decode(&playerPreference)
	return playerPreference, err
}

func (m MongoRepository) SaveUserPlayerPreference(ctx context.Context, preference entities.UserPlayerPreference) error {
	span, ctx := apm.StartSpan(ctx, "SaveUserPlayerPreference", "repository.Mongo")
	defer span.End()
	collection := m.client.Database(database).Collection(userPreference)

	res := bson.M{"$set": preference}
	filter := bson.M{"league_key": preference.LeagueKey, "user_guid": preference.UserID}
	_, err := collection.UpdateOne(ctx, filter, res, &options.UpdateOptions{
		Upsert: aws.Bool(true),
	})

	return err
}

func (m MongoRepository) GetDraftResults(ctx context.Context, leagueKey string) ([]entities.DraftResult, error) {
	span, ctx := apm.StartSpan(ctx, "GetDraftResults", "repository.Mongo")
	defer span.End()

	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.M{"league_key": leagueKey}}},
		bson.D{{"$lookup", bson.M{"from": playersBySeason, "localField": "player_key", "foreignField": "_id", "as": "player"}}},
	}
	collection := m.client.Database(database).Collection(draftsCollection)
	findOptions := &options.AggregateOptions{}
	cursor, err := collection.Aggregate(ctx, pipeline, findOptions)
	if err != nil {
		level.Error(m.logger).Log("message", "could not get draft results", "error", err, "league_key", leagueKey)
		return nil, err
	}

	var bsonResults []bson.M
	err = cursor.All(ctx, &bsonResults)
	if err != nil {
		return nil, err
	}
	draftResults := make([]entities.DraftResult, len(bsonResults))
	for idx, bsonM := range bsonResults {
		var draftResult entities.DraftResult
		bsonBytes, err := bson.Marshal(&bsonM)
		if err != nil {
			return nil, err
		}

		err = bson.Unmarshal(bsonBytes, &draftResult)
		if err != nil {
			return nil, err
		}
		draftResults[idx] = draftResult
	}

	return draftResults, nil
}

func (m MongoRepository) GetTeamDraftResultsByTeam(ctx context.Context, leagueKey string) (map[string][]entities.DraftResult, error) {
	span, ctx := apm.StartSpan(ctx, "GetTeamDraftResultsByTeam", "repository.Mongo")
	defer span.End()

	draftResults, err := m.GetDraftResults(ctx, leagueKey)
	if err != nil {
		level.Error(m.logger).Log("message", "error in getting draft results", "err", err)
		return nil, err
	}
	if len(draftResults) == 0 {
		return map[string][]entities.DraftResult{}, nil
	}

	teams := make(map[string][]entities.DraftResult, len(draftResults[0].League.Teams))
	for _, draftResult := range draftResults {
		teamResults, ok := teams[draftResult.TeamKey]
		if !ok {
			teamResults = newPlayerSeasonSlice()
		}
		teams[draftResult.TeamKey] = append(teamResults, draftResult)
	}

	return teams, nil
}

func newPlayerSeasonSlice() []entities.DraftResult {
	return make([]entities.DraftResult, 0)
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveDraftResultFromUser(ctx context.Context, league entities.League, user entities.User, team entities.Team, player entities.PlayerSeason, pick, round int) (*entities.DraftResult, error) {
	span, ctx := apm.StartSpan(ctx, "SaveDraftResult", "repository.Mongo")
	defer span.End()

	draftResult := entities.DraftResult{
		UserGUID:  user.Guid,
		PlayerKey: player.PlayerKey,
		PlayerID:  player.PlayerID,
		LeagueKey: league.LeagueKey,
		TeamKey:   team.TeamKey,
		Round:     round,
		Pick:      pick,
		Timestamp: time.Now(),
		GameID:    league.Game.GameID,
		Player:    []*entities.PlayerSeason{&player},
	}

	err := m.SaveDraftResult(ctx, draftResult)
	if err != nil {
		level.Error(m.logger).Log("message", "error in saving draft result", "err", err)
		return nil, err
	}

	newTable := getPlayerLeagueCollection(league.LeagueKey)

	collection := m.client.Database(database).Collection(newTable)

	res := bson.M{"$set": bson.M{"draft_results": draftResult}}
	filter := bson.M{"player._id": player.PlayerKey}
	insertResult, err := collection.UpdateOne(ctx, filter, res)
	if err != nil {
		level.Error(m.logger).Log("error", err, "message", "could not execute query", "guid", draftResult.UserGUID, "league_key", league.LeagueKey)
		return nil, err
	}

	level.Debug(m.logger).Log("message", "insert draft result", "upsert_count", insertResult.UpsertedCount, "upsert_id", insertResult.UpsertedID, "player_key", draftResult.PlayerKey)
	if insertResult.ModifiedCount == 0 {
		return nil, errors.New("did not update")
	}
	return &draftResult, nil
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveDraftResult(ctx context.Context, draftResult entities.DraftResult) error {
	span, ctx := apm.StartSpan(ctx, "SaveDraftResult", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(draftsCollection)
	draftDBytes, err := bson.Marshal(&draftResult)
	var draftD bson.D
	err = bson.Unmarshal(draftDBytes, &draftD)
	if err != nil {
		level.Error(m.logger).Log("error", err, "message", "could not make query", "guid", draftResult.UserGUID)

	}
	_, err = collection.InsertOne(ctx, draftD, &options.InsertOneOptions{})

	if err != nil {
		level.Error(m.logger).Log("error", err, "message", "could not execute query", "guid", draftResult.UserGUID, "query", draftD)
		return err
	}
	return nil
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveDraftOrder(ctx context.Context, leagueKey string, teamOrder []string) error {
	span, ctx := apm.StartSpan(ctx, "SaveDraftOrder", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)

	res, err := collection.UpdateOne(ctx, bson.M{"league_key": leagueKey}, bson.M{"$set": bson.M{"draft_order": teamOrder}})
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", leagueKey)
		return err
	}

	level.Debug(m.logger).Log("message", "updated matched count", "league_key", leagueKey, "matched", res.MatchedCount, "modified", res.ModifiedCount)
	return err
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveLeagueLeague(ctx context.Context, league entities.League) (entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "SaveLeagueLeague", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)
	var bsonMap, updateMap bson.M
	queryString := fmt.Sprintf(`{"league_key": "%s"}`, league.LeagueKey)
	err := json.Unmarshal([]byte(queryString), &bsonMap)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", league.LeagueKey, "query", queryString)
		return entities.League{}, err
	}

	res, err := collection.UpdateOne(ctx, updateMap, bson.M{"$set": bson.M{"draft_started": league.DraftStarted}})

	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", league.LeagueKey, "query", queryString)
		return entities.League{}, err
	}

	level.Debug(m.logger).Log("message", "updated matched count", "league_key", league.LeagueKey, "query", queryString, "matched", res.MatchedCount, "modified", res.ModifiedCount)
	return league, err
}

func (m MongoRepository) GetLeague(ctx context.Context, leagueKey string) (entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "GetLeague", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)
	var findQuert bson.M
	queryString := fmt.Sprintf(`{"league_key": "%s"}`, leagueKey)
	err := json.Unmarshal([]byte(queryString), &findQuert)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", leagueKey, "query", queryString)
		return entities.League{}, err
	}

	res := collection.FindOne(ctx, findQuert)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", leagueKey, "query", queryString)
		return entities.League{}, err
	}
	var league entities.League
	err = res.Decode(&league)
	return league, err
}

//
func (m MongoRepository) SaveLeague(ctx context.Context, league entities.League) (entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "GetLeague", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)
	var findQuert bson.M
	queryString := fmt.Sprintf(`{"league_key": "%s"}`, league.LeagueKey)
	err := json.Unmarshal([]byte(queryString), &findQuert)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", league.LeagueKey, "query", queryString)
		return entities.League{}, err
	}

	_, err = collection.UpdateOne(ctx, findQuert, league)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", league.LeagueKey, "query", queryString)
		return entities.League{}, err
	}

	return league, err
}

func (m MongoRepository) GetTeamsForLeague(ctx context.Context, leagueKey string) ([]entities.Team, error) {

	span, ctx := apm.StartSpan(ctx, "GetTeamsForManagers", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)

	// get newest to oldest
	// will have to flip at some place
	cursor := collection.FindOne(ctx, bson.M{"league_key": leagueKey})
	if cursor.Err() != nil {
		level.Error(m.logger).Log("message", "cursor error", "league_key", leagueKey, "error", cursor.Err())
		return []entities.Team{}, cursor.Err()
	}
	var league entities.League
	err := cursor.Decode(&league)

	if err != nil {
		level.Error(m.logger).Log("message", "could not marshal league", "league_key", leagueKey, "error", err)
		return []entities.Team{}, err
	}

	return league.Teams, nil
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) GetPlayers(ctx context.Context, playerKeys []string) ([]entities.PlayerSeason, error) {
	span, ctx := apm.StartSpan(ctx, "GetPlayers", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(playersBySeason)
	result := bson.A{}
	for _, e := range playerKeys {
		result = append(result, e)
	}
	query := bson.M{"_id": bson.M{"$in":
	result,
	},
	}

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query")
		return []entities.PlayerSeason{}, err
	}

	players := make([]entities.PlayerSeason, len(playerKeys))
	var i int
	for cursor.Next(ctx) {
		var player entities.PlayerSeason

		err = cursor.Decode(&player)
		if err != nil {
			level.Error(m.logger).Log("message", "could not marshal league", "error", err)
			return players, err
		}
		players[i] = player
		i++
	}

	return players, nil
}
func (m MongoRepository) GetPlayersByRank(ctx context.Context, limit, offset int) ([]entities.PlayerSeason, error) {
	span, ctx := apm.StartSpan(ctx, "GetPlayersByRank", "repository.Mongo")
	defer func() {
		span.End()
	}()

	collection := m.client.Database(database).Collection(playersBySeason)

	filter := bson.D{}
	findOptions := options.Find()

	findOptions.SetSort(bson.D{{"game_id", -1}, {"ranks.yahoo", 1}})
	findOptions.SetLimit(int64(limit))

	results, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		level.Error(m.logger).Log("message", "could not get users ", "error", err)
		return nil, err
	}
	players := make([]entities.PlayerSeason, 0, limit)
	for results.Next(ctx) {
		var player entities.PlayerSeason

		err = results.Decode(&player)
		if err != nil {
			level.Error(m.logger).Log("message", "could not marshal league", "error", err)
			return players, err
		}
		_ = level.Debug(m.logger).Log("message", "name", "name", player.Name.Full)
		players = append(players, player)

	}

	return players, err
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveDraftResults(ctx context.Context, draftResults []entities.DraftResult) error {
	span, ctx := apm.StartSpan(ctx, "SaveDraft", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(draftsCollection)

	_, err := collection.UpdateMany(ctx, bson.M{"league_key": draftResults[0].LeagueKey}, draftResults, &options.UpdateOptions{
		Upsert: aws.Bool(true),
	})

	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "league_key", draftResults[0].LeagueKey)
		return err
	}
	return nil
}

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) GetTeamsForManagers(ctx context.Context, guid string) ([]entities.League, error) {
	span, ctx := apm.StartSpan(ctx, "GetTeamsForManagers", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(leaguesCollection)
	var bsonMap bson.M
	queryString := fmt.Sprintf(`{"teams.manager": {"$elemMatch":{"guid": "%s"}}}`, guid)
	err := json.Unmarshal([]byte(queryString), &bsonMap)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "guid", guid, "query", queryString)
		return []entities.League{}, err
	}

	findOptions := options.Find()
	// get newest to oldest
	// will have to flip at some place
	findOptions.SetSort(bson.D{{"game.game_id", -1}})

	cursor, err := collection.Find(ctx, bsonMap, findOptions)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not find leagues", "guid", guid)
		return []entities.League{}, err
	}

	leagues := make([]entities.League, 0)
	for cursor.Next(ctx) {
		var league entities.League

		err = cursor.Decode(&league)
		if err != nil {
			level.Error(m.logger).Log("message", "could not marshal league", "error", err)
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (m *MongoRepository) SaveLeagueLeagueGroup(ctx context.Context, leagueGroupd *entities.LeagueGroup) (*entities.LeagueGroup, error) {
	span, ctx := apm.StartSpan(ctx, "CreateLeague", "repository.Mongo")
	defer span.End()

	// save only email
	collection := m.client.Database(database).Collection(leaguesCollection)
	// find previous leagueGroupd if any

	leagueGroupFilter := bson.M{"league_id": leagueGroupd.FirstLeagueID}

	res := collection.FindOne(ctx, leagueGroupFilter)
	parentLeague := entities.League{}

	//
	res.Decode(&parentLeague)
	//league.LeagueGroup = primitive.NewObjectID()
	leagGroupID := primitive.NewObjectID()
	if parentLeague.LeagueID != 0 {
		leagGroupID = parentLeague.LeagueGroup
	}

	for _, league := range leagueGroupd.Leagues {
		league.LeagueGroup = leagGroupID
		// insert one
		leagueFilter := bson.M{"league_key": league.LeagueKey}
		updateOptions := options.FindOneAndReplaceOptions{
			Upsert: aws.Bool(true),
		}

		_ = collection.FindOneAndReplace(ctx, leagueFilter, league, &updateOptions)

	}

	return leagueGroupd, nil
}

func (m *MongoRepository) SavePlayers(ctx context.Context, players []entities.PlayerSeason) ([]entities.PlayerSeason, error) {
	span, ctx := apm.StartSpan(ctx, "SavePlayers", "repository.Mongo")
	span.Context.SetLabel("player_count", len(players))
	defer span.End()

	// save only email
	collection := m.client.Database(database).Collection(playersBySeason)
	// find previous leagueGroupd if any
	updateOptions := &options.UpdateOptions{Upsert: aws.Bool(true)}

	for _, player := range players {
		playerLeagueFinder := bson.M{"_id": player.PlayerKey}

		playersInterface := player
		_, err := collection.UpdateOne(ctx, playerLeagueFinder, bson.D{{"$set", playersInterface}}, updateOptions)
		if err != nil {
			level.Error(m.logger).Log("message", "could not save player", "error", err, "player_key", player.PlayerKey)
		}
	}

	return players, nil
}

func (repo *MongoRepository) getUserCollection(ctx context.Context) *mongo.Collection {
	return repo.client.Database(repo.database).Collection(repo.userCollection)
}

func (repo *MongoRepository) SaveUser(ctx context.Context, draftID primitive.ObjectID, user User) (primitive.ObjectID, error) {
	span, ctx := apm.StartSpan(ctx, "SaveUser", "db")
	defer span.End()

	userCollection := repo.getUserCollection(ctx)

	insertOne, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	userID, _ := insertOne.InsertedID.(primitive.ObjectID)
	level.Debug(repo.logger).Log("message", "new user inserted", "user_id", userID)
	return userID, nil
}

// type will either be commissioned or drafts
func (m MongoRepository) getUsersByType(ctx context.Context, collection *mongo.Collection, draftID primitive.ObjectID, stringType string) ([]*User, error) {
	filter := bson.D{
		{stringType, draftID},
	}
	mongoUsers, err := collection.Find(ctx, filter)
	if err != nil {
		return []*User{}, nil
	}

	var users []*User

	for mongoUsers.Next(ctx) {
		var user User

		err = mongoUsers.Decode(&user)
		if err != nil {
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}

func (m MongoRepository) ListUserDrafts(ctx context.Context, pbUser pb.User) ([]pb.Season, error) {
	userCollection := m.client.Database(m.database).Collection(m.userCollection)
	filter := bson.M{
		"email": pbUser.Email,
	}

	res := userCollection.FindOne(ctx, filter)
	if res.Err() != nil {
		return []pb.Season{}, res.Err()
	}
	var user User
	res.Decode(&user)

	return []pb.Season{}, nil
}
