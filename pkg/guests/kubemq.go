package guests

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/kubemq-io/kubemq-go"
	"go.elastic.co/apm"
)

const draftPrefix = "guests"

type Repository struct {
	client  *kubemq.Client
	logger  log.Logger
	address string
}


func (r *Repository) StartChannelSubscription(ctx context.Context) (*kubemq.Client, error) {
	span, ctx := apm.StartSpan(ctx, "draftPrefix", "kubemq")
	defer span.End()

	sender, err := kubemq.NewClient(ctx,
		kubemq.WithAddress("localhost", 50000),
		kubemq.WithClientId("test-event-store-sender-id"),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC))
	if err != nil {
		panic(err)
	}
	return sender, err
}
