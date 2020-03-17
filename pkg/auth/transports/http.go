package transports

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/thethan/fdr-users/pkg/auth"
	"net/http"
	"os"
	"strings"
)

const contentType = "application/json; charset=utf-8"

func decodeFuncNothing(ctx context.Context, req *http.Request) (interface{}, error) {
	return nil, nil
}

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(endpoints auth.Endpoints, m *mux.Router, options ...httptransport.ServerOption) *mux.Router {
	serverOptions := []httptransport.ServerOption{
		httptransport.ServerBefore(headersToContext),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerAfter(httptransport.SetContentType(contentType)),
	}
	serverOptions = append(serverOptions, options...)

	m.Methods("GET", "POST").Path("/auth").Handler(
		httptransport.NewServer(
			endpoints.Index,
			decodeFuncNothing,
			auth.EncodeFuncTemplates,
			serverOptions...))

	m.Methods("GET", "POST").Path("/auth/{provider}/login").HandlerFunc(endpoints.Provider)

	m.Methods("GET", "POST").Path("/auth/{provider}").HandlerFunc(endpoints.CompleteUserAuthHandler)
	m.Methods("GET", "POST").Path("/auth/{provider}").HandlerFunc(endpoints.CompleteUserAuthHandler)

	m.Methods("GET", "POST").Path("/auth/{provider}/logout").HandlerFunc(endpoints.Logout)
	//m.Methods("GET", "POST").Path("/users/auth/login/{provider}").HandlerFunc(endpoints.Provider)

	//m.Methods("POST").Path("/login").Handler(httptransport.NewServer(
	//	endpoints.CompleteUserAuthEndpoint,
	//	DecodeHTTPLoginZeroRequest,
	//	EncodeHTTPGenericResponsefunc,
	//	serverOptions...,
	//))

	logger := log.NewJSONLogger(os.Stdout)
	m.NotFoundHandler = NotFoundHandler{log: logger}
	return m
}

type NotFoundHandler struct {
	m   mux.Router
	log log.Logger
}

type NotFoundInfo struct {
	Routes []Route     `json:"routes"`
	Header http.Header `json:"header"`
	Path   string      `json:"path"`
}

type Route struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Methods      []string `json:"methods"`
	HostTemplate string   `json:"host_template"`
	PathTemplate string   `json:"path_template"`
}

func (n NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	routes := make([]Route, 0)
	n.m.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		r := Route{}

		r.Name = route.GetName()
		r.Path, _ = route.GetPathRegexp()
		r.Methods, _ = route.GetMethods()
		r.HostTemplate, _ = route.GetHostTemplate()
		r.PathTemplate, _ = route.GetPathTemplate()

		routes = append(routes, r)
		return nil
	})

	//notFoundInfo := NotFoundInfo{}
	//notFoundInfo.Header = r.Header
	//notFoundInfo.Routes = routes
	//notFoundInfo.Path = r.URL.Path
	//n.log.Log("header", r.Header)
	//n.log.Log("url", r.URL.Path)
	//n.log.Log("host", r.URL.Host)

	w.WriteHeader(http.StatusNotFound)
	output, _ := json.Marshal(routes)
	_, _ = w.Write(output)

}

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

func DecodeHTTPLoginZeroRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func EncodeHTTPGenericResponsefunc(ctx context.Context, e http.ResponseWriter, value interface{}) error {
	return nil
}
