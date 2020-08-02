package kubemq

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kubemq-io/kubemq-go"
)


func NewKubeMQClient(ctx context.Context, address string, port int, clientID string) (*kubemq.Client, error) {
	clientID = uuid.New().String()
	fmt.Println(address, port, )
	return kubemq.NewClient(ctx,
		kubemq.WithAddress(address, port),
		kubemq.WithClientId(clientID),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC))
}

