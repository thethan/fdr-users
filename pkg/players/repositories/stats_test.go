package repositories

import (
	"context"
	"github.com/go-kit/kit/log"
	pkgEntities "github.com/thethan/fdr-users/pkg/draft/entities"
	mongo2 "github.com/thethan/fdr-users/pkg/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"testing"
)

func TestStatsRepo_SavePlayerStats(t *testing.T) {
	type fields struct {
		logger log.Logger
		client *mongo.Client
	}
	type args struct {
		ctx       context.Context
		playerKey string
		season    string
		week      string
		stats     []pkgEntities.PlayerStat
	}

	logger := log.NewNopLogger()
	client, err := mongo2.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		t.Errorf("Could not reach mongo client. Error: %v", err)
		t.FailNow()
	}


	type testCase struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}
	tests := []testCase{
		{
			name:    "Simple insert into db",
			fields:  fields{
				logger: logger,
				client: client,
			},
			args:    args{
				ctx: context.TODO(),
				season: "390",
				week: "1",
				playerKey: "nfl.p.30121",
				stats: []pkgEntities.PlayerStat{
					{
						StatID: 1,
						Value: float64(1),
					},
					{
						StatID: 2,
						Value: float64(1),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Simple insert into db",
			fields:  fields{
				logger: logger,
				client: client,
			},
			args:    args{
				ctx: context.TODO(),
				season: "390",
				week: "2",
				playerKey: "nfl.p.30121",
				stats: []pkgEntities.PlayerStat{
					{
						StatID: 1,
						Value: float64(8),
					},
					{
						StatID: 2,
						Value: float64(8),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Simple insert into db",
			fields:  fields{
				logger: logger,
				client: client,
			},
			args:    args{
				ctx: context.TODO(),
				season: "390",
				week: "3",
				playerKey: "nfl.p.1",
				stats: []pkgEntities.PlayerStat{
					{
						StatID: 1,
						Value: float64(0.25),
					},
					{
						StatID: 2,
						Value: float64(64),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StatsRepo{
				logger: tt.fields.logger,
				client: tt.fields.client,
			}
			if err := s.SavePlayerStats(tt.args.ctx, tt.args.playerKey, tt.args.season, tt.args.week, tt.args.stats); (err != nil) != tt.wantErr {
				t.Errorf("SavePlayerStats() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
