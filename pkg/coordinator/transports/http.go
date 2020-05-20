package transports

import (
	"context"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/coordinator"
	"go.elastic.co/apm/module/apmlogrus"
	"net/http"
	"strconv"
)

const GameID string = "game_id"

func NewHTTPServer(fieldLogger logrus.FieldLogger, m *mux.Router, endpoints coordinator.Endpoints, authServerBefore httptransport.RequestFunc) *mux.Router {
	options := []httptransport.ServerOption{
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
		httptransport.ServerBefore(authServerBefore),
	}

	m.Methods(http.MethodGet).PathPrefix("/import/leagues").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAlreadyReported)
	})
	m.Methods(http.MethodGet).PathPrefix("/import/players").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAlreadyReported)
	})

	m.Methods("POST").Path("/import/leagues").Handler(httptransport.NewServer(
		endpoints.ImportUserLeagues,
		DecodeHTTPImportUserRequest(fieldLogger),
		EncodeHTTPDecodeImportUserRequest(fieldLogger),
		options...,
	))
	m.Methods("POST").Path("/import/games/{game_id}/players").Handler(httptransport.NewServer(
		endpoints.ImportGamePlayers,
		DecodeHTTPImportGamePlayersRequest(fieldLogger),
		EncodeHTTPDecodeImportGamePlayersRequest(fieldLogger),
		options...,
	))
	return m
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


func DecodeHTTPImportGamePlayersRequest(logger logrus.FieldLogger) httptransport.DecodeRequestFunc {
	return func(ctx context.Context, req *http.Request) (request interface{}, err error) {
		fields := apmlogrus.TraceContext(ctx)
		vars := mux.Vars(req)

		var ImportGamePlayersRequest coordinator.ImportGamePlayersRequest
		gameID, _ := strconv.Atoi(vars[GameID])
		ImportGamePlayersRequest.GameID = gameID
		logger.WithFields(fields).Info("decode the request")
		return ImportGamePlayersRequest, nil
	}
}

func EncodeHTTPDecodeImportGamePlayersRequest(logger logrus.FieldLogger) httptransport.EncodeResponseFunc {
	return func(ctx context.Context, r http.ResponseWriter, res interface{}) error {
		fields := apmlogrus.TraceContext(ctx)
		logger.WithFields(fields).Info("encoding the response")
		return nil
	}
}