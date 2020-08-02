package guests

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"go.elastic.co/apm"
	"time"
)

type Endpoints struct {
	logger             log.Logger
	GetPureTottenRSVP  endpoint.Endpoint
	SavePureTottenRSVP endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, service Service, authMiddleWare endpoint.Middleware) Endpoints {
	return Endpoints{
		logger:            logger,
		GetPureTottenRSVP: authMiddleWare(makeGetGuestResponses(logger, service)),
		SavePureTottenRSVP: makeSaveGuest(logger, service),
	}
}

func makeGetGuestResponses(logger log.Logger, service Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "GetGuestResponses", "endpoint")
		defer span.End()

		guests, err := service.GetRSVPList(ctx)
		if err != nil {
			return nil, err
		}
		return guests, err
	}
}

func makeSaveGuest(logger log.Logger, service Service) endpoint.Endpoint {
	return func(ctx context.Context, reqInt interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveGuest", "endpoint")
		defer span.End()

		req, ok := reqInt.(*GuestRequest)
		if !ok {
			return nil, errors.New("could not marshal error")
		}

		guest := Guest{
			Name:             req.Name,
			Adults:           req.Adults,
			Children:         req.Children,
			Email:            req.Email,
			Attending:        req.Attending,
			VeganOptionCount: req.VeganOptionCount,
			CreatedAt:        time.Now(),
		}

		err = service.SaveGuest(ctx, guest)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}
