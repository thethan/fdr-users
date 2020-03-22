// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

// Package grpc provides a gRPC client for the Users service.
package grpc

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	// This Service
	pb "github.com/thethan/fdr_proto"
)

// New returns an service backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn, options ...ClientOption) (pb.UsersServer, error) {
	var cc clientConfig

	for _, f := range options {
		err := f(&cc)
		if err != nil {
			return nil, errors.Wrap(err, "cannot apply option")
		}
	}

	clientOptions := []grpctransport.ClientOption{
		grpctransport.ClientBefore(
			contextValuesToGRPCMetadata(cc.headers)),
	}
	var createEndpoint endpoint.Endpoint
	{
		createEndpoint = grpctransport.NewClient(
			conn,
			"fdr.Users",
			"Create",
			EncodeGRPCCreateRequest,
			DecodeGRPCCreateResponse,
			pb.CreateUserResponse{},
			clientOptions...,
		).Endpoint()
	}

	var searchEndpoint endpoint.Endpoint
	{
		searchEndpoint = grpctransport.NewClient(
			conn,
			"fdr.Users",
			"Search",
			EncodeGRPCSearchRequest,
			DecodeGRPCSearchResponse,
			pb.ListUserResponse{},
			clientOptions...,
		).Endpoint()
	}

	var loginEndpoint endpoint.Endpoint
	{
		loginEndpoint = grpctransport.NewClient(
			conn,
			"fdr.Users",
			"Login",
			EncodeGRPCLoginRequest,
			DecodeGRPCLoginResponse,
			pb.LoginResponse{},
			clientOptions...,
		).Endpoint()
	}

	var userCredentials endpoint.Endpoint
	{
		loginEndpoint = grpctransport.NewClient(
			conn,
			"fdr.Users",
			"Credentials",
			EncodeGRPCLoginRequest,
			DecodeGRPCLoginResponse,
			pb.CredentialResponse{},
			clientOptions...,
		).Endpoint()
	}




	return users.Endpoints{
		CreateEndpoint: createEndpoint,
		SearchEndpoint: searchEndpoint,
		LoginEndpoint:  loginEndpoint,
	}, nil
}

// GRPC Client Decode

// DecodeGRPCCreateResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC create reply to a user-domain create response. Primarily useful in a client.
func DecodeGRPCCreateResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.CreateUserResponse)
	return reply, nil
}

// DecodeGRPCSearchResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC search reply to a user-domain search response. Primarily useful in a client.
func DecodeGRPCSearchResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.ListUserResponse)
	return reply, nil
}

// DecodeGRPCLoginResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC login reply to a user-domain login response. Primarily useful in a client.
func DecodeGRPCLoginResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.LoginResponse)
	return reply, nil
}

// GRPC Client Encode

// EncodeGRPCCreateRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain create request to a gRPC create request. Primarily useful in a client.
func EncodeGRPCCreateRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.CreateUserRequest)
	return req, nil
}

// EncodeGRPCSearchRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain search request to a gRPC search request. Primarily useful in a client.
func EncodeGRPCSearchRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.ListUserRequest)
	return req, nil
}

// EncodeGRPCLoginRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain login request to a gRPC login request. Primarily useful in a client.
func EncodeGRPCLoginRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.LoginRequest)
	return req, nil
}

type clientConfig struct {
	headers []string
}

// ClientOption is a function that modifies the client config
type ClientOption func(*clientConfig) error

func CtxValuesToSend(keys ...string) ClientOption {
	return func(o *clientConfig) error {
		o.headers = keys
		return nil
	}
}

func contextValuesToGRPCMetadata(keys []string) grpctransport.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		var pairs []string
		for _, k := range keys {
			if v, ok := ctx.Value(k).(string); ok {
				pairs = append(pairs, k, v)
			}
		}

		if pairs != nil {
			*md = metadata.Join(*md, metadata.Pairs(pairs...))
		}

		return ctx
	}
}
