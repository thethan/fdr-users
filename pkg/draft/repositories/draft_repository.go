package repositories

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes/timestamp"
	pb "github.com/thethan/fdr_proto"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

type MongoRepository struct {
	logger           log.Logger
	client           *mongo.Client
	database         string
	draftCollection  string
	userCollection   string
	rosterCollection string
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
	return MongoRepository{logger: logger, client: client, database: database, draftCollection: draftCollection, userCollection: userCollection, rosterCollection: rosterCollection}
}



func (m MongoRepository) CreateDraft(ctx context.Context, season pb.Season) (pb.Season, error) {
	span, ctx := apm.StartSpan(ctx, "CreateDraft", "repository.Mongo")
	defer span.End()

	// save only email
	collection := m.client.Database("fdr_drafts").Collection(m.draftCollection)

	draft, err := transformPBSeasonToDraft(season)
	if err != nil {
		level.Error(m.logger).Log("message", "could not transform season to draft", "error", err)
		return season, err
	}

	insertedResult, err := collection.InsertOne(ctx, draft)
	if err != nil {
		level.Error(m.logger).Log("message", "could not save draft", "error", err)
		return season, err
	}

	fmt.Printf("%v", insertedResult.InsertedID)
	draftID, _ := insertedResult.InsertedID.(primitive.ObjectID)

	season.ID = draftID.String()

	userCollection := m.client.Database(m.database).Collection(m.userCollection)

	for _, user := range season.Users {
		filter := bson.M{"_id": user.Email}
		var mongoUser User
		res := userCollection.FindOne(ctx, filter)

		res.Decode(&mongoUser)
		if mongoUser.Email != "" {
			// transformPBUserToUser
			level.Debug(m.logger).Log("msg", "hitting ")
			update := bson.M{"$push": bson.M{"drafts": draftID}}
			_, err := userCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
			if err != nil {
				return season, err
			}
		} else {
			mongoUser, _ := transformPBUserToUser(user)
			mongoUser.Commissioned = []primitive.ObjectID{}
			mongoUser.Drafts = []primitive.ObjectID{draftID}

			insertOne, err := userCollection.InsertOne(ctx, mongoUser)
			if err != nil {
				return season, err
			}
			userID, _ := insertOne.InsertedID.(primitive.ObjectID)
			level.Debug(m.logger).Log("message", "new user inserted", "user_id", userID)
		}

	}

	for _, user := range season.Commissioners {
		filter := bson.M{"_id": user.Email}
		var mongoUser User
		res := userCollection.FindOne(ctx, filter)

		res.Decode(&mongoUser)
		if mongoUser.Email != "" {
			// transformPBUserToUser
			update := bson.M{"$push": bson.M{"commissioned": draftID}}
			_, err := userCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
			if err != nil {
				return season, err
			}
		} else {
			//
			mongoUser, _ = transformPBUserToUser(user)
			mongoUser.Commissioned = []primitive.ObjectID{draftID}
			mongoUser.Drafts = []primitive.ObjectID{}
			userID, _ := m.SaveUser(ctx, draftID, mongoUser)
			level.Debug(m.logger).Log("message", "new user inserted", "user_id", userID)

		}
	}

	rosterCollection := m.client.Database(m.database).Collection(m.rosterCollection)

	rosterRules := RosterRules{
		ID:          primitive.NewObjectID(),
		RosterRules: []RosterRule{},
	}

	rrDoc := make([]bson.D, len(season.Roster))
	for idx, rosterRule := range season.Roster {
		// transformPBUserToUser
		rrDoc[idx] = bson.D{{"position", int(rosterRule.Position)}, {"max", rosterRule.Max}, {"starting", rosterRule.Starting}}

		rosterRules.RosterRules = append(rosterRules.RosterRules, RosterRule{
			Position: int(rosterRule.Position),
			Starting: rosterRule.Starting,
			Max:      rosterRule.Max,
		})

	}

	bsonRosterQuery := bson.D{
		{"rosterrules", bson.D{{"$all", bson.A{rrDoc}}}},
	}

	filter := bsonRosterQuery
	findRoster := rosterCollection.FindOne(ctx, filter)

	findRoster.Decode(&rosterRules)

	if findRoster.Err() != nil {
		insertOne, err := rosterCollection.InsertOne(ctx, rosterRules)
		if err != nil {
			return season, err
		}
		rosterID, _ := insertOne.InsertedID.(primitive.ObjectID)
		collection.UpdateOne(ctx, bson.M{"_id": draftID}, bson.M{"$set": bson.M{"roster": rosterID}})
	}

	// transform to response
	newSeason, err := transformDraftToPBSeason(*draft)
	participants, err := m.getUsersByType(ctx, userCollection, draft.ID, "drafts")
	if err != nil {
		return newSeason, err
	}
	pbParticipants := make([]*pb.User, len(participants))
	for idx := range participants {
		pbParticipant, _ := transformUserToPBUser(*participants[idx])
		pbParticipants[idx] = &pbParticipant
	}
	newSeason.Users = pbParticipants

	commissioners, err := m.getUsersByType(ctx, userCollection, draft.ID, "commissioned")
	if err != nil {
		return newSeason, err
	}

	pbCommissioners := make([]*pb.User, len(commissioners))
	for idx := range commissioners {
		pbParticipant, _ := transformUserToPBUser(*commissioners[idx])
		pbCommissioners[idx] = &pbParticipant
	}
	newSeason.Commissioners = pbCommissioners

	// roster
	newSeason.Roster, _ = transformRosterToRosterPB(rosterRules)
	return newSeason, err
}

func (repo *MongoRepository) getUserCollection(ctx context.Context) *mongo.Collection {
	return repo.client.Database(repo.database).Collection(repo.userCollection)
}

func (repo *MongoRepository) SaveUser(ctx context.Context,  draftID primitive.ObjectID, user User) (primitive.ObjectID, error) {
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

func transformPBSeasonToDraft(season pb.Season) (*Draft, error) {
	yearInt, err := strconv.Atoi(season.Year)
	if err != nil {
		return nil, err
	}

	var draftTime time.Time
	if season.DraftTime != nil {
		timeHmmm := time.Unix(season.DraftTime.GetSeconds(), int64(season.DraftTime.Nanos))
		draftTime = timeHmmm
	}

	draft := Draft{
		ID:        primitive.NewObjectID(),
		Year:      yearInt,
		League:    int(season.League),
		DraftType: int(season.DraftType),
		DraftTime: primitive.NewDateTimeFromTime(draftTime),
	}
	//if len(season.ExternalIdentifiers) > 0 {}
	return &draft, nil
}

func transformDraftToPBSeason(draft Draft) (pb.Season, error) {
	yearInt := strconv.Itoa(draft.Year)
	var timeHmmm *timestamp.Timestamp

	timeTime := draft.DraftTime.Time()
	timeHmmm = &timestamp.Timestamp{
		Seconds: timeTime.Unix(),
		Nanos:   int32(timeTime.UnixNano()),
	}

	season := pb.Season{
		ID:        draft.ID.Hex(),
		Year:      yearInt,
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
