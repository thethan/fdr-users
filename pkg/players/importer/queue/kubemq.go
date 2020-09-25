package queue

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
	"github.com/thethan/fdr-users/pkg/players/entities"
	"go.elastic.co/apm/module/apmgrpc"
	"time"
)

type MessageType int

const (
	PlayerStatsMessage MessageType = iota
)

const (
	playerChannel      string = "player_to_be_imported"
	playerStatsChannel string = "player_stats_to_be_imported"
)

type Queue struct {
	logger               log.Logger
	client               *kubemq.Client
	clientID             uuid.UUID
	BatchID              *uint64
	messageTypeToChannel map[MessageType]string

	startChannel       chan *kubemq.EventStore
	resultsChannel     chan *kubemq.EventStoreResult
	errorChannel       chan error
	eventChannel       <-chan *kubemq.EventStoreReceive
	eventPlayerChannel <-chan *kubemq.EventStoreReceive
}

func NewQueueImportStats(logger log.Logger, client *kubemq.Client) Queue {
	errorChan := make(chan error)
	startChannel := make(chan *kubemq.EventStore)
	resultChannel := make(chan *kubemq.EventStoreResult)

	q := Queue{
		logger:   logger,
		client:   client,
		clientID: uuid.New(),

		// send orders
		startChannel:   startChannel,
		errorChannel:   errorChan,
		resultsChannel: resultChannel,
	}

	return q
}

func (q *Queue) StartWorker(ctx context.Context, messageChannel chan<- entities.ImportPlayerStat) error {
	eventChannel, err := q.client.SubscribeToEventsStore(ctx, playerStatsChannel, "fdr-users", q.errorChannel, kubemq.StartFromTime(time.Now()))
	if err != nil {
		level.Error(q.logger).Log("message", "error in subscribing to events", "error", err)
		return err
	}
	// event channel
	q.eventChannel = eventChannel
	for {
		select {
		case <-ctx.Done():
			return nil
		case res := <-q.eventChannel:
			var stats entities.ImportPlayerStat
			err := json.Unmarshal(res.Body, &stats)
			if err != nil {
				level.Error(q.logger).Log("message", "error in marshalling json to import player stat", "sequence_id", res.Sequence, "id", res.Id)
				continue
			}
			messageChannel <- stats
		default:
			//level.Debug(q.logger).Log("message", "nothing to report on", "func", "startworker")
		}
	}
}

func (q *Queue) StartPlayerWorker(ctx context.Context, messageChannel chan<- entities.ImportPlayer) error {
	eventChannel, err := q.client.SubscribeToEventsStore(ctx, playerChannel, "fdr-users", q.errorChannel, kubemq.StartFromTime(time.Now()))
	if err != nil {
		level.Error(q.logger).Log("message", "error in subscribing to events", "error", err)
		return err
	}
	// event channel
	q.eventPlayerChannel = eventChannel
	for {
		select {
		case <-ctx.Done():
			return nil
		case res := <-q.eventChannel:
			var stats entities.ImportPlayer
			err := json.Unmarshal(res.Body, &stats)
			if err != nil {
				level.Error(q.logger).Log("message", "error in marshalling json to import player stat", "sequence_id", res.Sequence, "id", res.Id)
				continue
			}
			messageChannel <- stats
		default:
			//level.Debug(q.logger).Log("message", "nothing to report on", "func", "startworker")
		}
	}
}
func (q *Queue) SendPlayerStatMessage(ctx context.Context, stats entities.ImportPlayerStat) error {
	statBytes, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	apmgrpc.NewUnaryClientInterceptor()
	eventStore := kubemq.EventStore{
		Id:       uuid.New().String(),
		Channel:  playerStatsChannel,
		Body:     statBytes,
		ClientId: q.clientID.String(),
	}
	q.startChannel <- &eventStore

	return nil
}

func (q *Queue) SendPlayerMessage(ctx context.Context, stats entities.ImportPlayer) error {
	statBytes, err := json.Marshal(&stats)
	if err != nil {
		return err
	}
	apmgrpc.NewUnaryClientInterceptor()
	eventStore := kubemq.EventStore{
		Id:       uuid.New().String(),
		Channel:  playerChannel,
		Body:     statBytes,
		ClientId: q.clientID.String(),
	}
	q.startChannel <- &eventStore

	return nil
}

func (q *Queue) ReceiveMessages(ctx context.Context, messageChannel chan<- entities.ImportPlayerStat) error {
	ticker := time.NewTicker(time.Second / 2)
	for {
		select {
		case <-ctx.Done():
			level.Debug(q.logger).Log("message", "Closing ReceiveMessages")
			ticker.Stop()
			return nil
		case <-ticker.C:
		}
	}

}
