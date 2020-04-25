package transports

import (
	"context"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/coordinator"
	"go.elastic.co/apm/module/apmlogrus"
	"net/http"
)

func NewHTTPServer(fieldLogger logrus.FieldLogger, endpoints coordinator.Endpoints, authServerBefore httptransport.RequestFunc) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(authServerBefore),
	}
	return httptransport.NewServer(
		endpoints.ImportUserLeagues,
		DecodeHTTPImportUserRequest(fieldLogger),
		EncodeHTTPDecodeImportUserRequest(fieldLogger),
		options...,

	)

}

func DecodeHTTPImportUserRequest(logger logrus.FieldLogger) httptransport.DecodeRequestFunc {
	return func(ctx context.Context, req *http.Request) (request interface{}, err error) {
		fields := apmlogrus.TraceContext(ctx)

		logger.WithFields(fields).Info("decode the request")
		return nil, nil
	}
}

func EncodeHTTPDecodeImportUserRequest(logger logrus.FieldLogger) httptransport.EncodeResponseFunc {
	return func(ctx context.Context, r http.ResponseWriter, res interface{}) error {
		fields := apmlogrus.TraceContext(ctx)
		logger.WithFields(fields).Info("encoding the response")
		return nil
	}
}
