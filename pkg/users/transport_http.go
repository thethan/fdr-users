// Code generated by truss. DO NOT EDIT.
// Rerunning truss will overwrite this file.
// Version: 8907ffca23
// Version Date: Wed Nov 27 21:28:21 UTC 2019

package users

// This file provides server-side bindings for the HTTP transport.
// It utilizes the transport/http.Server.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thethan/fdr-users/pkg/draft/entities"
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
func MakeHTTPHandler(endpoints Endpoints, m *mux.Router, authServerBefore httptransport.RequestFunc, options ...httptransport.ServerOption) *mux.Router {
	serverOptions := []httptransport.ServerOption{
		httptransport.ServerBefore(headersToContext),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerAfter(httptransport.SetContentType(contentType)),
	}
	serverOptions = append(serverOptions, options...)
	serverOptionsAuth := append(serverOptions, httptransport.ServerBefore(authServerBefore))

	m = m.PathPrefix("/users").Subrouter()
	m.Methods("POST").Path("/auth").Handler(httptransport.NewServer(
		endpoints.CreateEndpoint,
		DecodeHTTPCreateZeroRequest,
		EncodeHTTPGenericResponse,
		serverOptionsAuth...,
	))

	m.Methods("POST").Path("/search").Handler(httptransport.NewServer(
		endpoints.SearchEndpoint,
		DecodeHTTPSearchZeroRequest,
		EncodeHTTPGenericResponse,
		serverOptionsAuth...,
	))

	m.Methods("POST").Path("/login").Handler(httptransport.NewServer(
		endpoints.LoginEndpoint,
		DecodeHTTPLoginZeroRequest,
		EncodeHTTPGenericResponse,
		serverOptionsAuth...,
	))
	m.Methods("POST").Path("/credentials").Handler(httptransport.NewServer(
		endpoints.SaveCredentialEndpoint,
		DecodeHTTPSaveCredentialRequest,
		EncodeHTTPGenericResponse,
		serverOptionsAuth...,
	))
	m.Methods("POST").Path("/info").Handler(httptransport.NewServer(
		endpoints.SaveUserInfoEndpoint,
		DecodeHTTPSaveCredentialByUserIDRequest,
		EncodeHTTPSaveCredentialByUserIDRequest,
		serverOptions...,
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
	code := http.StatusInternalServerError
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

// DecodeHTTPCreateZeroRequest is a transport/http.DecodeRequestFunc that
// decodes a JSON-encoded create request from the HTTP request
// body. Primarily useful in a server.
func DecodeHTTPCreateZeroRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req pb.CreateUserRequest
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	if len(buf) > 0 {
		// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
		unmarshaller := jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		}
		if err = unmarshaller.Unmarshal(bytes.NewBuffer(buf), &req); err != nil {
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
	_ = pathParams

	queryParams := r.URL.Query()
	_ = queryParams

	return &req, err
}

// DecodeHTTPSearchZeroRequest is a transport/http.DecodeRequestFunc that
// decodes a JSON-encoded search request from the HTTP request
// body. Primarily useful in a server.
func DecodeHTTPSearchZeroRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req pb.ListUserRequest
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	if len(buf) > 0 {
		// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
		unmarshaller := jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		}
		if err = unmarshaller.Unmarshal(bytes.NewBuffer(buf), &req); err != nil {
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
	_ = pathParams

	queryParams := r.URL.Query()
	_ = queryParams

	return &req, err
}

// DecodeHTTPLoginZeroRequest is a transport/http.DecodeRequestFunc that
// decodes a JSON-encoded login request from the HTTP request
// body. Primarily useful in a server.
func DecodeHTTPLoginZeroRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req pb.LoginRequest
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	if len(buf) > 0 {
		// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
		unmarshaller := jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		}
		if err = unmarshaller.Unmarshal(bytes.NewBuffer(buf), &req); err != nil {
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
	_ = pathParams

	queryParams := r.URL.Query()
	_ = queryParams

	return &req, err
}

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

type UserCredentialRequest struct {
	UID          string `json:"uid"`
	Session      string `json:"session"`
	RefreshToken string `json:"refresh_token"`
	Guid         string `json:"guid"`

	Email string `json:"email"`
}

type UserCredentialResponse struct {
	UID          string `json:"uid"`
	Session      string `json:"session"`
	RefreshToken string `json:"refresh_token"`
	Guid         string `json:"guid"`
	Leagues      []*entities.LeagueGroup `json:"leagues"`
	Email        string `json:"email"`
}

func DecodeHTTPSaveCredentialByUserIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req UserCredentialRequest
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
	_ = pathParams

	queryParams := r.URL.Query()
	_ = queryParams

	return &req, err
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPSaveCredentialByUserIDRequest(_ context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.(*UserCredentialResponse)
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

func DecodeHTTPSaveCredentialRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req CredentialRequest
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read body of http request")
	}
	//if len(buf) > 0 {
	//	// AllowUnknownFields stops the unmarshaler from failing if the JSON contains unknown fields.
	//	unmarshaller := jsonpb.Unmarshaler{
	//		AllowUnknownFields: true,
	//	}
	//	if err = unmarshaller.Unmarshal(bytes.NewBuffer(buf), &req); err != nil {
	//		const size = 8196
	//		if len(buf) > size {
	//			buf = buf[:size]
	//		}
	//		return nil, httpError{errors.Wrapf(err, "request body '%s': cannot parse non-json request body", buf),
	//			http.StatusBadRequest,
	//			nil,
	//		}
	//	}
	//}

	err = json.Unmarshal(buf, &req)
	if err != nil {
		return nil, err
	}
	pathParams := mux.Vars(r)
	_ = pathParams

	queryParams := r.URL.Query()
	_ = queryParams

	return &req, err
}
