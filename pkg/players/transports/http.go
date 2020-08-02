// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

package transports

// This file provides server-side bindings for the HTTP transport.
// It utilizes the transport/http.Server.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/thethan/fdr-users/pkg/draft/entities"
	"github.com/thethan/fdr-users/pkg/players"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"context"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	// This service
	pb "github.com/thethan/fdr_proto"
)

const contentType = "application/json; charset=utf-8"
const leagueIdParam = "leagueId"
const userGuid = "userGuid"

var (
	_ = fmt.Sprint
	_ = bytes.Compare
	_ = strconv.Atoi
	_ = httptransport.NewServer
	_ = ioutil.NopCloser
	_ = pb.NewUsersClient
	_ = io.Copy
	_ = errors.Wrap
)

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(logger log.Logger, endpoints players.Endpoints, m *mux.Router, authServerBefore httptransport.RequestFunc, options ...httptransport.ServerOption) *mux.Router {
	serverOptions := []httptransport.ServerOption{
		httptransport.ServerBefore(headersToContext),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerAfter(httptransport.SetContentType(contentType)),
	}
	serverOptions = append(serverOptions, options...)
	serverOptionsAuth := append(serverOptions, httptransport.ServerBefore(authServerBefore))

	m = m.PathPrefix("/leagues").Subrouter()
	m.Methods(http.MethodGet).Path("/{" + leagueIdParam + "}/players").Handler(httptransport.NewServer(
		endpoints.GetAvailablePlayers,
		DecodeHTTPGetAvailablePlayersForLeague,
		EncodeHTTPGetAvailablePlayers,
		serverOptionsAuth...,
	))
	return m

	m.Methods(http.MethodPut).Path("/{+" + userGuid + "}/{" + leagueIdParam + "}/players").Handler(httptransport.NewServer(
		endpoints.SaveUsersDraftOrder,
		DecodeHTTPSavePlayersOrders,
		EncodeHTTPGetAvailablePlayers,
		serverOptionsAuth...,
	))
	return m
}

// ErrorEncoder writes the error to the ResponseWriter, by default a content
// type of application/json, a body of json with key "error" and the value
// error.Error(), and a status code of 500. If the error implements Headerer,
// the provided headers will be applied to the response. If the error
// implements json.Marshaler, and the marshaling succeeds, the JSON encoded
// form of the error will be used. If the error implements StatusCoder, the
// provided StatusCode will be used instead of 500.
func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	body, _ := json.Marshal(errorWrapper{Error: err.Error()})
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			body = jsonBody
		}
	}
	w.Header().Set("Content-Type", contentType)
	if headerer, ok := err.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	code := http.StatusBadRequest
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	w.Write(body)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// httpError satisfies the Headerer and StatusCoder interfaces in
// package github.com/go-kit/kit/transport/http.
type httpError struct {
	error
	statusCode int
	headers    map[string][]string
}

func (h httpError) StatusCode() int {
	return h.statusCode
}

func (h httpError) Headers() http.Header {
	return h.headers
}

// Server Decode

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	marshaller := jsonpb.Marshaler{
		EmitDefaults: false,
		OrigName:     true,
	}

	return marshaller.Marshal(w, response.(proto.Message))
}

// Helper functions

func headersToContext(ctx context.Context, r *http.Request) context.Context {
	for k, _ := range r.Header {
		// The key is added both in http format (k) which has had
		// http.CanonicalHeaderKey called on it in transport as well as the
		// strings.ToLower which is the grpc metadata format of the key so
		// that it can be accessed in either format
		ctx = context.WithValue(ctx, k, r.Header.Get(k))
		ctx = context.WithValue(ctx, strings.ToLower(k), r.Header.Get(k))
	}

	// Tune specific change.
	// also add the request url
	ctx = context.WithValue(ctx, "request-url", r.URL.Path)
	ctx = context.WithValue(ctx, "transport", "HTTPJSON")

	return ctx
}

func DecodeHTTPGetAvailablePlayersForLeague(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req players.GetAvailablePlayersRequest
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	if len(buf) > 0 {
		// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
		if err = json.Unmarshal(buf, &req); err != nil {
			const size = 8196
			if len(buf) > size {
				buf = buf[:size]
			}
			return nil, httpError{errors.Wrapf(err, "request body '%s': cannot parse non-json request body", buf),
				http.StatusBadRequest,
				nil,
			}
		}
	}

	pathParams := mux.Vars(r)
	leagueKey, ok := pathParams[leagueIdParam]
	if !ok {
		return nil, errors.New("bad request")
	}

	gameIDs := strings.Split(".l.", leagueKey)
	if len(gameIDs) == 0 {
		return nil, errors.New("could not get game id from league")
	}
	req.GameID, _ = strconv.Atoi(gameIDs[0])

	req.LeagueKey = leagueKey

	queryParams := r.URL.Query()
	// limit
	limtString := queryParams.Get("limit")
	limit, _ := strconv.Atoi(limtString)
	if limit == 0 {
		limit = 50
	}
	req.Limit = limit
	// offset
	offsetString := queryParams.Get("offset")
	offset, _ := strconv.Atoi(offsetString)
	req.Offset = offset
	// positions
	positions := queryParams.Get("positions")
	if len(positions) > 0 {
		pos := strings.Split(",", positions)
		req.Positions = pos
	}

	return &req, err
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGetPlayersOrders(_ context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.(*entities.UserPlayerPreference)
	if !ok {
		return errors.New("could not get user player order ")
	}
	bytesJson, err := json.Marshal(&res)
	if err != nil {
		return err
	}
	w.Write(bytesJson)
	return nil
}


func DecodeHTTPSavePlayersOrders(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req entities.UserPlayerPreference
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	if len(buf) > 0 {
		// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
		if err = json.Unmarshal(buf, &req); err != nil {
			const size = 8196
			if len(buf) > size {
				buf = buf[:size]
			}
			return nil, httpError{errors.Wrapf(err, "request body '%s': cannot parse non-json request body", buf),
				http.StatusBadRequest,
				nil,
			}
		}
	}

	pathParams := mux.Vars(r)
	leagueKey, ok := pathParams[leagueIdParam]
	if !ok {
		return nil, errors.New("bad request")
	}

	gameIDs := strings.Split(".l.", leagueKey)
	if len(gameIDs) == 0 {
		return nil, errors.New("could not get game id from league")
	}
	req.GameID = gameIDs[0]
	req.LeagueKey = leagueKey

	return &req, err
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGetAvailablePlayers(_ context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.(*players.GetAvailablePositionsForLeague)
	if !ok {
		return errors.New("could not get user Credentials Response ")
	}
	bytesJson, err := json.Marshal(&res)
	if err != nil {
		return err
	}
	w.Write(bytesJson)
	return nil
}
