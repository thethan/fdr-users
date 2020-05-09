package draft

import (
	"context"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	pb "github.com/thethan/fdr_proto"
	"google.golang.org/grpc/metadata"
	"net/http"
)



// grpcServer implements the UsersServer interface
type grpcServer struct {
	create grpctransport.Handler
	list   grpctransport.Handler
	update grpctransport.Handler
}

func (g grpcServer) CreateDraft(ctx context.Context,req *pb.CreateDraftRequest) (*pb.CreateDraftResponse, error) {
	ctx, rep, err := g.create.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateDraftResponse), nil
}

func (g grpcServer) ListDraftsUserHasAccessTo(ctx context.Context, season *pb.Season) (*pb.Season, error) {
	ctx, rep, err := g.list.ServeGRPC(ctx, season)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.Season), nil
}

func (g grpcServer) UpdateDraft(ctx context.Context, req *pb.CreateDraftRequest) (*pb.CreateDraftResponse, error) {
	ctx, rep, err := g.update.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.CreateDraftResponse), nil
}

// MakeGRPCServer makes a set of endpoints available as a gRPC UsersServer.

// MakeGRPCServer makes a set of endpoints available as a gRPC UsersServer.
func MakeGRPCServer(endpoints Endpoints, options ...grpctransport.ServerOption) pb.DraftServer {
	serverOptions := []grpctransport.ServerOption{
		grpctransport.ServerBefore(metadataToContext),
	}
	serverOptions = append(serverOptions, options...)
	return &grpcServer{
		create: grpctransport.NewServer(
			endpoints.Create,
			DecodeGRPCCreateRequest,
			EncodeGRPCCreateResponse,
			serverOptions...,
		),
		list: grpctransport.NewServer(
			endpoints.List,
			DecodeGRPCListUserRequest,
			EncodeGRPCListUserResponse,
			serverOptions...,
		),

		// update and create are the same request
		update: grpctransport.NewServer(
			endpoints.Update,
			DecodeGRPCCreateRequest,
			EncodeGRPCCreateResponse,
			serverOptions...,
		),

	}
}

// DecodeGRPCCreateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC create request to a user-domain create request. Primarily useful in a server.
func DecodeGRPCCreateRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateDraftRequest)
	return req, nil
}

// EncodeGRPCCreateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain create response to a gRPC create reply. Primarily useful in a server.
func EncodeGRPCCreateResponse(_ context.Context, response interface{}) (interface{}, error) {
	
	season := response.(pb.Season)
	resp := &pb.CreateDraftResponse{
		Season:               &season,
	}
	return resp, nil
}


// DecodeGRPCCreateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC create request to a user-domain create request. Primarily useful in a server.
func DecodeGRPCListUserRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.Season)
	return req, nil
}

// EncodeGRPCCreateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain create response to a gRPC create reply. Primarily useful in a server.
func EncodeGRPCListUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.Season)
	return resp, nil
}

// DecodeGRPCCreateRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC create request to a user-domain create request. Primarily useful in a server.
func DecodeGRPCSearchRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateDraftResponse)
	return req, nil
}

// EncodeGRPCCreateResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain create response to a gRPC create reply. Primarily useful in a server.
func EncodeGRPCSearchResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.CreateDraftResponse)
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
