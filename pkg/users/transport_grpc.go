// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

package users

// This file provides server-side bindings for the gRPC transport.
// It utilizes the transport/grpc.Server.

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"

	grpctransport "github.com/go-kit/kit/transport/grpc"

	// This Service
	pb "github.com/thethan/fdr_proto"
)

// MakeGRPCServer makes a set of endpoints available as a gRPC UsersServer.
func MakeGRPCServer(endpoints Endpoints, options ...grpctransport.ServerOption) pb.UsersServer {
	serverOptions := []grpctransport.ServerOption{
		grpctransport.ServerBefore(metadataToContext),
	}
	serverOptions = append(serverOptions, options...)
	return &grpcServer{
		// fdr-users

		create: grpctransport.NewServer(
			endpoints.CreateEndpoint,
			DecodeGRPCCreateRequest,
			EncodeGRPCCreateResponse,
			serverOptions...,
		),
		search: grpctransport.NewServer(
			endpoints.SearchEndpoint,
			DecodeGRPCSearchRequest,
			EncodeGRPCSearchResponse,
			serverOptions...,
		),
		login: grpctransport.NewServer(
			endpoints.LoginEndpoint,
			DecodeGRPCLoginRequest,
			EncodeGRPCLoginResponse,
			serverOptions...,
		),

		credentials: grpctransport.NewServer(
			endpoints.CredentialEndpoint,
			DecodeGRPCCredentialsRequest,
			EncodeGRPCCredentialsResponse,
			serverOptions...,
		),
	}
}

// grpcServer implements the UsersServer interface
type grpcServer struct {
	create         grpctransport.Handler
	search         grpctransport.Handler
	login          grpctransport.Handler
	credentials grpctransport.Handler
}

// Methods for grpcServer to implement UsersServer interface

func (s *grpcServer) Create(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	_, rep, err := s.create.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateUserResponse), nil
}

func (s *grpcServer) Search(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserResponse, error) {
	_, rep, err := s.search.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ListUserResponse), nil
}

func (s *grpcServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	_, rep, err := s.login.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.LoginResponse), nil
}

func (s *grpcServer) Credentials(ctx context.Context, req *pb.CredentialRequest) (*pb.CredentialResponse, error) {
	_, rep, err := s.credentials.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CredentialResponse), nil
}

// Server Decode

// DecodeGRPCCreateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC create request to a user-domain create request. Primarily useful in a server.
func DecodeGRPCCreateRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateUserRequest)
	return req, nil
}

// DecodeGRPCSearchRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC search request to a user-domain search request. Primarily useful in a server.
func DecodeGRPCSearchRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ListUserRequest)
	return req, nil
}

// DecodeGRPCLoginRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC login request to a user-domain login request. Primarily useful in a server.
func DecodeGRPCLoginRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.LoginRequest)
	return req, nil
}

// DecodeGRPCCredentialsRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC login request to a user-domain login request. Primarily useful in a server.
func DecodeGRPCCredentialsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CredentialRequest)
	return req, nil
}

// Server Encode

// EncodeGRPCCreateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain create response to a gRPC create reply. Primarily useful in a server.
func EncodeGRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.CreateUserResponse)
	return resp, nil
}

// EncodeGRPCSearchResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain search response to a gRPC search reply. Primarily useful in a server.
func EncodeGRPCSearchResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.ListUserResponse)
	return resp, nil
}

// EncodeGRPCLoginResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain login response to a gRPC login reply. Primarily useful in a server.
func EncodeGRPCLoginResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.LoginResponse)
	return resp, nil
}

// EncodeGRPCCredentialsResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain login response to a gRPC login reply. Primarily useful in a server.
func EncodeGRPCCredentialsResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.CredentialResponse)
	return resp, nil
}

// Helpers

func metadataToContext(ctx context.Context, md metadata.MD) context.Context {
	for k, v := range md {
		if v != nil {
			// The key is added both in metadata format (k) which is all lower
			// and the http.CanonicalHeaderKey of the key so that it can be
			// accessed in either format
			ctx = context.WithValue(ctx, k, v[0])
			ctx = context.WithValue(ctx, http.CanonicalHeaderKey(k), v[0])
		}
	}

	return ctx
}
