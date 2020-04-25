// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

package users

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats.

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport/http"
	pb "github.com/thethan/fdr_proto"
	"go.elastic.co/apm"
)

// Endpoints collects all of the endpoints that compose an add service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them into an Endpoints, and return it to the caller as a Service.
type Endpoints struct {
	CreateEndpoint         endpoint.Endpoint
	SearchEndpoint         endpoint.Endpoint
	LoginEndpoint          endpoint.Endpoint
	CredentialEndpoint     endpoint.Endpoint
	SaveCredentialEndpoint endpoint.Endpoint
	SaveUserInfoEndpoint   endpoint.Endpoint
	logger                 log.Logger
}

// Endpoints

func NewEndpoints(logger log.Logger, user GetUserInfo, saveInfo SaveUserInfo, authMiddleware endpoint.Middleware, serverBefore http.RequestFunc) Endpoints {
	// Business domain.
	var service usersService
	{
		service = NewService(logger, user, saveInfo)

	}

	// Endpoint domain.
	var (
		createEndpoint          = MakeCreateEndpoint(&service)
		searchEndpoint          = MakeSearchEndpoint(&service)
		loginEndpoint           = MakeLoginEndpoint(&service)
		getCredentialsEndpoint  = MakeCredentialsEndpoint(&service, user)
		saveCredentialsEndpoint = authMiddleware(MakeSaveCredentialsEndpoint(&service))
		saveUserInfoEndpoint    = MakeSaveInformationsEndpoint(&service)
	)

	endpoints := Endpoints{
		logger:                 logger,
		CreateEndpoint:         createEndpoint,
		SearchEndpoint:         searchEndpoint,
		LoginEndpoint:          loginEndpoint,
		CredentialEndpoint:     getCredentialsEndpoint,
		SaveCredentialEndpoint: saveCredentialsEndpoint,
		SaveUserInfoEndpoint:   saveUserInfoEndpoint,
	}

	// Wrap selected Endpoints with middlewares. See handlers/middlewares.go
	return endpoints
}

// Make Endpoints
func MakeCreateEndpoint(s pb.UsersServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "CreateEndpoint", "endpoint")
		defer span.End()

		req := request.(*pb.CreateUserRequest)
		v, err := s.Create(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeSearchEndpoint(s pb.UsersServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "ListEndpoint", "endpoint")
		defer span.End()

		req := request.(*pb.ListUserRequest)
		v, err := s.Search(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeLoginEndpoint(s pb.UsersServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "LoginEndpoint", "endpoint")
		defer span.End()

		req := request.(*pb.LoginRequest)
		v, err := s.Login(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeCredentialsEndpoint(s pb.UsersServer, user GetUserInfo) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "CredentialsEndpoint", "endpoint")
		defer span.End()

		req := request.(*pb.CredentialRequest)
		v, err := s.Credentials(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeSaveCredentialsEndpoint(s *usersService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveCredentialsEndpoint", "endpoint")
		defer span.End()

		req := request.(*pb.CredentialRequest)
		v, err := s.SaveYahooCredential(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func MakeSaveInformationsEndpoint(s *usersService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span, ctx := apm.StartSpan(ctx, "SaveInformationEndpoint", "endpoint")
		defer span.End()

		req := request.(*UserCredentialRequest)
		v, err := s.SaveFromUserID(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
