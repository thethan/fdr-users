package kubemq

import (
	"context"
	"github.com/kubemq-io/kubemq-go"
)


func NewKubeMQClient(ctx context.Context, address string, port int, clientID string) (*kubemq.Client, error) {
	return kubemq.NewClient(ctx,
		kubemq.WithAddress(address, port),
		kubemq.WithClientId(clientID),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC))
}

