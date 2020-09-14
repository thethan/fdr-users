package kubemq

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kubemq "github.com/kubemq-io/protobuf/go"
	"github.com/thethan/fdr-users/internal/yahoo/importer/entities"
	"go.elastic.co/apm"
	"strconv"
	"time"
)

type Repository struct {
	logger  log.Logger
	channel string
	kubemq  kubemq.KubemqClient
}

func NewImportPlayerStatsQueue(logger log.Logger, channel string, kubemq kubemq.KubemqClient) Repository {
	return Repository{
		logger:  logger,
		channel: channel,
		kubemq:  kubemq,
	}
}

func (r Repository) GetPlayerStats(ctx context.Context, guid, playerKey string, week int) {

}

func (r Repository) Start(ctx context.Context, messageChannel chan<- entities.Message) {
	span, ctx := apm.StartSpan(ctx, "Start", "importer.service")
	defer span.End()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req := kubemq.ReceiveQueueMessagesRequest{
				RequestID:           strconv.Itoa(time.Now().Nanosecond()),
				ClientID:            r.channel,
				Channel:             r.channel,
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     1,
			}
			receiveResult, err := r.kubemq.ReceiveQueueMessages(ctx, &req)

			if err != nil {
				level.Error(r.logger).Log("message", "could not get message", "error", err, "channel", r.channel)
			}
			if receiveResult != nil {

				for _, msg := range receiveResult.Messages {
					var obj entities.Message
					json.Unmarshal(msg.Body, &obj)

					messageChannel <- obj
				}
			}

		}
	}
}
