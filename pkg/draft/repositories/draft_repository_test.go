package repositories

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/thethan/fdr-draft/pkg/mongo"
	pb "github.com/thethan/fdr_proto"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

func TestCreatingNewMongodb(t *testing.T) {
	logger := log.NewNopLogger()
	t.Run("mongo", func(t *testing.T) {
		client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
		assert.Nil(t, err)
		if t.Failed() {
			t.FailNow()
		}
		filter := bson.D{}
		dbs, err := client.ListDatabaseNames(context.TODO(), filter)
		if err != nil {
			t.FailNow()
		}
		fmt.Printf("%+v\n", dbs)

		repo := NewMongoRepository(logger, client, "fdr_drafts", "draft", "fdr_user", "roster")
		ctx := context.Background()
		timestamp, _ := ptypes.TimestampProto(time.Now())
		season := pb.Season{
			Year:                 "2020",
			League:               1,
			DraftType:            0,
			DraftTime:            timestamp,
			Users: []*pb.User{
				&pb.User{
					Name:                 "Ethan",
					Email:                "ethan.totten@gmail.com",
				},
				&pb.User{
					Name:                 "Not Ethan",
					Email:                "ethantotten@gmail.com",
				},
				&pb.User{
					Name:                 "Not asdf",
					Email:                "asdf@gmail.com",
				},
			},
			Commissioners: []*pb.User{
				&pb.User{
					Name:                 "Not Ethan",
					Email:                "ethan.totten@gmail.com",
				},
			},
			Roster: []*pb.RosterRules{
				&pb.RosterRules{
					Position:             1,
					Starting:             1,
					Max:                  4,
				},
				&pb.RosterRules{
					Position:             2,
					Starting:             2,
					Max:                  6,
				},
				&pb.RosterRules{
					Position:             3,
					Starting:             3,
					Max:                  5,
				},
				&pb.RosterRules{
					Position:             4,
					Starting:             3,
					Max:                  5,
				},
			},
		}
		season, err = repo.CreateDraft(ctx, season)
		assert.Nil(t, err)
		assert.NotEqual(t, "", season.ID)
		assert.Equal(t, len(season.Commissioners), 1)
		assert.Equal(t, len(season.Users), 3)

	})
}


func TestListUsersDraft(t *testing.T) {
	logger := log.NewNopLogger()
	t.Run("mongo", func(t *testing.T) {
		client, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
		assert.Nil(t, err)
		if t.Failed() {
			t.FailNow()
		}

		repo := NewMongoRepository(logger, client, "drafts", "draft", "user", "rosters")
		ctx := context.Background()
		season := pb.Season{
			Year:                 "2020",
			League:               1,
			DraftType:            0,
			DraftTime:            nil,
			Users: []*pb.User{
				&pb.User{
					Name:                 "Ethan",
					Email:                "ethan.totten@gmail.com",
				},
				&pb.User{
					Name:                 "Not Ethan",
					Email:                "ethantotten@gmail.com",
				},
				&pb.User{
					Name:                 "Not asdf",
					Email:                "asdf@gmail.com",
				},
			},
			Commissioners: []*pb.User{
				&pb.User{
					Name:                 "Not Ethan",
					Email:                "ethan.totten@gmail.com",
				},
			},
			Roster: []*pb.RosterRules{
				&pb.RosterRules{
					Position:             1,
					Starting:             1,
					Max:                  4,
				},
				&pb.RosterRules{
					Position:             2,
					Starting:             2,
					Max:                  6,
				},
				&pb.RosterRules{
					Position:             3,
					Starting:             3,
					Max:                  5,
				},
				&pb.RosterRules{
					Position:             4,
					Starting:             3,
					Max:                  5,
				},
			},
		}
		seasons, err := repo.ListUserDrafts(ctx, *season.Users[0])
		assert.Nil(t, err)
		assert.NotEqual(t, "", seasons[0].ID)

	})
}
