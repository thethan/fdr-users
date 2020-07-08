package kubemq

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestRepository_BroadCastDraftResult(t *testing.T) {
	type fields struct {
		client *kubemq.Client
		logger log.Logger
	}
	type args struct {
		ctx         context.Context
		league      entities.League
		user        entities.User
		team        entities.Team
		draftResult entities.DraftResult
		pick        int
		round       int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewKubeMQClient(context.TODO(), "localhost", 50000, uuid.New().String())
			r := &Repository{
				client: client,
				logger: log.NewNopLogger(),
			}
			league := entities.League{
				LeagueKey:      "399.l.19481",
				Name:           "",
				LeagueID:       0,
				LeagueGroup:    primitive.ObjectID{},
				PreviousLeague: nil,
				Settings:       nil,
				Teams:          nil,
				Game:           entities.Game{},
				DraftOrder:     nil,
				TeamDraftOrder: nil,
				DraftStarted:   false,
				DraftedCheck:   nil,
			}
			draftResult := entities.DraftResult{
				UserGUID:  "",
				PlayerKey: "",
				PlayerID:  0,
				LeagueKey: "",
				TeamKey:   "",
				Round:     0,
				Pick:      time.Now().Nanosecond(),
				Timestamp: time.Time{},
				GameID:    0,
				Player:    nil,
			}
			if err := r.BroadCastDraftResult(tt.args.ctx, league, tt.args.user, tt.args.team, draftResult, tt.args.pick, tt.args.round); (err != nil) != tt.wantErr {
				t.Errorf("BroadCastDraftResult() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}