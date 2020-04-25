package transports

import (
	"context"
	"github.com/go-kit/kit/log"
	goKitGrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/coordinator"
	pb "github.com/thethan/fdr_proto"
	"go.elastic.co/apm/module/apmlogrus"
)

type Server struct {
	logger            log.Logger
	logrusLogger      logrus.FieldLogger
	importUserLeagues goKitGrpc.Handler
}

func NewServer(logger log.Logger, fieldLogger logrus.FieldLogger, endpoints coordinator.Endpoints) pb.ImportServer {
	return &Server{logger: logger, logrusLogger: fieldLogger,
		importUserLeagues: goKitGrpc.NewServer(
			endpoints.ImportUserLeagues,
			DecodeImportUserRequest(fieldLogger),
			EncodeDecodeImportUserRequest(fieldLogger),
		),
	}
}

func (s *Server) ImportUserLeagues(context.Context, *pb.ImportUserRequest) (*pb.ImportUserResponse, error) {
	return nil, nil
}

func DecodeImportUserRequest(logger logrus.FieldLogger) func(ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (i interface{}, err error) {
		fields := apmlogrus.TraceContext(ctx)

		logger.WithFields(fields).Info("decode the request")
		return nil, nil
	}
}

func EncodeDecodeImportUserRequest(logger logrus.FieldLogger) func(context.Context, interface{}) (response interface{}, err error) {
	return func(ctx context.Context, res interface{}) (response interface{}, err error) {
		fields := apmlogrus.TraceContext(ctx)
		logger.WithFields(fields).Info("encoding the response")
		return nil, nil
	}
}
