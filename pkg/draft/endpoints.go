package draft

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	pb "github.com/thethan/fdr_proto"
)

type Endpoints struct {
	logger  log.Logger
	service *Service
	Create  endpoint.Endpoint
	List    endpoint.Endpoint
	Update  endpoint.Endpoint
}

func NewEndpoints(logger log.Logger, service *Service) Endpoints {
	e := Endpoints{
		logger: logger,
		Create: makeCreateEndpoint(logger, service),
		List:   makeListEndpoint(logger, service),
		Update: makeUpdateEndpoint(logger, service),
	}

	return e
}

func makeCreateEndpoint(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fmt.Printf("%+v", request)
		req, ok := request.(*pb.CreateDraftRequest)
		if !ok {
			return nil, errors.New("could not get season from request")
		}
		return service.CreateDraft(ctx, *req.Season)
	}
}

func makeListEndpoint(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		season, ok := request.(*pb.Season)
		if !ok {
			return nil, errors.New("could not get season")
		}
		return service.List(ctx, *season.Users[0])
	}
}

func makeUpdateEndpoint(logger log.Logger, service *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		season, ok := request.(*pb.Season)
		if !ok {
			return nil, errors.New("could not get season")
		}
		return service.UpdateDraft(ctx, *season)
	}
}
