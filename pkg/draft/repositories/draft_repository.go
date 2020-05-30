package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	pb "github.com/thethan/fdr_proto"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database string = "fdr"
const leaguesCollection string = "leagues"
const draftsCollection string = "draft_results"
const playersBySeason string = "players"
const playersByWeek string = "players"

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
		draftResultsCollection: "draft-results",
		userCollection:         userCollection,
		rosterCollection:       rosterCollection}
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

// {"teams.manager":{ $elemMatch: {"guid":"DPPQCXCRV75Z2LKJW5YRC7RAYM"}}}
func (m MongoRepository) SaveDraftResult(ctx context.Context, draftResult entities.DraftResult) error {
	span, ctx := apm.StartSpan(ctx, "SaveDraftResult", "repository.Mongo")
	defer span.End()

	collection := m.client.Database(database).Collection(draftsCollection)
	draftDBytes, err := bson.Marshal(&draftResult)
	var draftD bson.D
	err = bson.Unmarshal(draftDBytes, &draftD)
	if err != nil {
		level.Error(m.logger).Log("error", err, "could not make query", "guid", draftResult.UserGUID)

	}
	_, err = collection.UpdateOne(ctx, bson.M{"league_key": draftResult.LeagueKey, "player_key": draftResult.PlayerKey}, bson.D{{"$set", draftD}}, &options.UpdateOptions{
		Upsert: aws.Bool(true),
	})

	if err != nil {
		level.Error(m.logger).Log("error", err, "could not execute query", "guid", draftResult.UserGUID, "query", draftD)
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
	//res, err := collection.UpdateOne(ctx, updateMap, league, &options.UpdateOptions{
	//BypassDocumentValidation: aws.Bool(true),
	//})

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

func (m *MongoRepository) SaveLeague(ctx context.Context, leagueGroupd *entities.LeagueGroup) (*entities.LeagueGroup, error) {
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

		upRes := collection.FindOneAndReplace(ctx, leagueFilter, league, &updateOptions)

		fmt.Println(upRes)
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

func oldSaveUser() (interface{}, interface{}) {
	//fmt.Printf("%v", insertedResult.InsertedID)
	//draftID, _ := insertedResult.InsertedID.(primitive.ObjectID)
	//
	//userCollection := m.client.Database(m.database).Collection(m.userCollection)
	//
	//for _, user := range league.Teams {
	//	filter := bson.M{"_id": user.Email}
	//	var mongoUser User
	//	res := userCollection.FindOne(ctx, filter)
	//
	//	res.Decode(&mongoUser)
	//	if mongoUser.Email != "" {
	//		// transformPBUserToUser
	//		level.Debug(m.logger).Log("msg", "hitting ")
	//		update := bson.M{"$push": bson.M{"drafts": draftID}}
	//		_, err := userCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	//		if err != nil {
	//			return league, err
	//		}
	//	} else {
	//		mongoUser, _ := transformPBUserToUser(user)
	//		mongoUser.Commissioned = []primitive.ObjectID{}
	//		mongoUser.Drafts = []primitive.ObjectID{draftID}
	//
	//		insertOne, err := userCollection.InsertOne(ctx, mongoUser)
	//		if err != nil {
	//			return season, err
	//		}
	//		userID, _ := insertOne.InsertedID.(primitive.ObjectID)
	//		level.Debug(m.logger).Log("message", "new user inserted", "user_id", userID)
	//	}
	//
	//}
	//
	//for _, user := range season.Commissioners {
	//	filter := bson.M{"_id": user.Email}
	//	var mongoUser User
	//	res := userCollection.FindOne(ctx, filter)
	//
	//	res.Decode(&mongoUser)
	//	if mongoUser.Email != "" {
	//		// transformPBUserToUser
	//		update := bson.M{"$push": bson.M{"commissioned": draftID}}
	//		_, err := userCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	//		if err != nil {
	//			return season, err
	//		}
	//	} else {
	//		//
	//		mongoUser, _ = transformPBUserToUser(user)
	//		mongoUser.Commissioned = []primitive.ObjectID{draftID}
	//		mongoUser.Drafts = []primitive.ObjectID{}
	//		userID, _ := m.SaveUser(ctx, draftID, mongoUser)
	//		level.Debug(m.logger).Log("message", "new user inserted", "user_id", userID)
	//
	//	}
	//}
	//
	//rosterCollection := m.client.Database(m.database).Collection(m.rosterCollection)
	//
	//rosterRules := RosterRules{
	//	ID:          primitive.NewObjectID(),
	//	RosterRules: []RosterRule{},
	//}
	//
	//rrDoc := make([]bson.D, len(season.Roster))
	//for idx, rosterRule := range season.Roster {
	//	// transformPBUserToUser
	//	rrDoc[idx] = bson.D{{"position", int(rosterRule.Pick)}, {"max", rosterRule.Max}, {"starting", rosterRule.Starting}}
	//
	//	rosterRules.RosterRules = append(rosterRules.RosterRules, RosterRule{
	//		Pick: int(rosterRule.Pick),
	//		Starting: rosterRule.Starting,
	//		Max:      rosterRule.Max,
	//	})
	//
	//}
	//
	//bsonRosterQuery := bson.D{
	//	{"rosterrules", bson.D{{"$all", bson.A{rrDoc}}}},
	//}
	//
	//filter := bsonRosterQuery
	//findRoster := rosterCollection.FindOne(ctx, filter)
	//
	//findRoster.Decode(&rosterRules)
	//
	//if findRoster.Err() != nil {
	//	insertOne, err := rosterCollection.InsertOne(ctx, rosterRules)
	//	if err != nil {
	//		return season, err
	//	}
	//	rosterID, _ := insertOne.InsertedID.(primitive.ObjectID)
	//	collection.UpdateOne(ctx, bson.M{"_id": draftID}, bson.M{"$set": bson.M{"roster": rosterID}})
	//}
	//
	//// transform to response
	//newSeason, err := transformDraftToPBSeason(*draft)
	//participants, err := m.getUsersByType(ctx, userCollection, draft.ID, "drafts")
	//if err != nil {
	//	return newSeason, err
	//}
	//pbParticipants := make([]*pb.User, len(participants))
	//for idx := range participants {
	//	pbParticipant, _ := transformUserToPBUser(*participants[idx])
	//	pbParticipants[idx] = &pbParticipant
	//}
	//newSeason.Users = pbParticipants
	//
	//commissioners, err := m.getUsersByType(ctx, userCollection, draft.ID, "commissioned")
	//if err != nil {
	//	return newSeason, err
	//}
	//
	//pbCommissioners := make([]*pb.User, len(commissioners))
	//for idx := range commissioners {
	//	pbParticipant, _ := transformUserToPBUser(*commissioners[idx])
	//	pbCommissioners[idx] = &pbParticipant
	//}
	//newSeason.Commissioners = pbCommissioners
	//
	//// roster
	//newSeason.Roster, _ = transformRosterToRosterPB(rosterRules)
	//return newSeason, err
	return nil, nil
}

func (m MongoRepository) CreateDraft(ctx context.Context, season pb.Season) (pb.Season, error) {
	span, ctx := apm.StartSpan(ctx, "CreateDraft", "repository.Mongo")
	defer span.End()

	return pb.Season{}, nil
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

func transformDraftToPBSeason(draft Draft) (pb.Season, error) {
	//yearInt := strconv.Itoa(draft.Year)
	var timeHmmm *timestamp.Timestamp

	timeTime := draft.DraftTime.Time()
	timeHmmm = &timestamp.Timestamp{
		Seconds: timeTime.Unix(),
		Nanos:   int32(timeTime.UnixNano()),
	}

	season := pb.Season{
		ID: draft.ID.Hex(),
		//:      yearInt,
		League:    pb.League(draft.League),
		DraftType: pb.DraftType(draft.DraftType),
		DraftTime: timeHmmm,
	}

	for idx := range draft.Users {
		pbUser, _ := transformUserToPBUser(draft.Users[idx])
		season.Users = append(season.Users, &pbUser)
	}

	for idx := range draft.Commissioners {
		pbUser, _ := transformUserToPBUser(draft.Commissioners[idx])
		season.Commissioners = append(season.Users, &pbUser)
	}
	return season, nil
}

func transformPBUserToUser(user *pb.User) (User, error) {
	return User{
		Email: user.Email,
		Image: user.Image,
		Name:  user.Name,
	}, nil

}

func transformUserToPBUser(user User) (pb.User, error) {

	return pb.User{
		Email: user.Email,
		Image: user.Image,
		Name:  user.Name,
	}, nil

}

func transformRosterToRosterPB(rosterRules RosterRules) ([]*pb.RosterRules, error) {
	rules := make([]*pb.RosterRules, len(rosterRules.RosterRules))
	for idx, rosterRule := range rosterRules.RosterRules {
		rr := transformRosterRuleToPBRosterSlot(rosterRule)
		rules[idx] = &rr
	}
	return rules, nil
}

// transformRosterRuleToPBRosterSlot
func transformRosterRuleToPBRosterSlot(rule RosterRule) pb.RosterRules {
	return pb.RosterRules{
		Position: pb.PlayerPosition(rule.Position),
		Starting: rule.Starting,
		Max:      rule.Max,
	}
}
