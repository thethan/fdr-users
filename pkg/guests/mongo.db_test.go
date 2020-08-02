package guests

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	mongo2 "github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/test_helpers"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"testing"
	"time"
)

func TestMongoRepository_SaveGuest(t *testing.T) {
	type fields struct {
		logger log.Logger
		client *mongo.Client
	}
	type args struct {
		ctx   context.Context
		guest Guest
	}
	logger := test_helpers.LogrusLogger(t)
	client, err := mongo2.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	assert.Nil(t, err)
	if t.Failed() {
		t.FailNow()
	}

	mongoRepo := NewMongoRepository(logger, client)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "save guest",
			args: args{
				ctx:   context.Background(),
				guest: Guest{
					Name:             "some guest",
					Adults:           2,
					Children:         1,
					Email:            "guest@gmail.com",
					Attending:        true,
					VeganOptionCount: 1,
					CreatedAt:        time.Now(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mongoRepo
			if err := m.SaveGuest(tt.args.ctx, tt.args.guest); (err != nil) != tt.wantErr {
				t.Errorf("SaveGuest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}