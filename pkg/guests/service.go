package guests

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kubemq-io/kubemq-go"
	"go.elastic.co/apm"
	"time"
)

type SaveGuest interface {
	SaveGuest(ctx context.Context, guest Guest) error
	GetGuestList(ctx context.Context) ([]Guest, error)
}

type Service struct {
	logger       log.Logger
	channel      string
	kubemqClient *kubemq.Client
	saveGuest    SaveGuest
}

func NewService(logger log.Logger, kubemqClient *kubemq.Client, guest SaveGuest) Service {
	svc := Service{
		logger:       logger,
		kubemqClient: kubemqClient,
		saveGuest:    guest,
	}
	return svc
}

func (s *Service) Start(ctx context.Context) error {
	span, ctx := apm.StartSpan(ctx, "", "service")
	defer func() {
		span.End()
	}()
	errCh := make(chan error)
	time.Sleep(time.Second)

	eventsCh, err := s.kubemqClient.SubscribeToEventsStore(ctx, "guests", "", errCh, kubemq.StartFromNewEvents())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return s.Close()
		case event, _ := <-eventsCh:
			level.Debug(s.logger).Log("message", "message received from kubemq", "event", event.Body)
			var guest Guest
			err := json.Unmarshal(event.Body, &guest)
			if err != nil {
				level.Error(s.logger).Log("message", "could not assert guest was sent", "error", err)
				continue
			}
			err = s.saveGuest.SaveGuest(ctx, guest)
			if err != nil {
				level.Error(s.logger).Log("error", err, "message", "error in saving guest", "guest", guest)
			}
		}
	}

}

func (s *Service) Close() error {
	return s.kubemqClient.Close()
}

func (s *Service) GetRSVPList(ctx context.Context) ([]Guest, error) {
	return s.saveGuest.GetGuestList(ctx)
}

func (s *Service) SaveGuest(ctx context.Context, guest Guest) error {
	return s.saveGuest.SaveGuest(ctx, guest)
}
