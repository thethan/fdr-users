package auth

import (
	"context"
	"github.com/go-kit/kit/log"

	"net/http"
)

func NewEndpoints(logger log.Logger) Endpoints {
	return Endpoints{logger: logger}
}

// Endpoints in this package are not standard go-kut packages.
// They align with the mux handler. These are not
type Endpoints struct {
	logger log.Logger
}

func decodeFuncNothing(ctx context.Context, req *http.Request) (interface{}, error) {
	return nil, nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.

func EncodeFuncTemplates(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	return nil
}

// @todo move this to an svc
func (e Endpoints) Index(ctx context.Context, req interface{}) (interface{}, error) {
	return nil, nil
}

// @todo move this to an endpoint
func (e Endpoints) CompleteUserAuthHandler(res http.ResponseWriter, req *http.Request) {

}

// @todo move this to an endpoint
func (e Endpoints) Logout(res http.ResponseWriter, req *http.Request) {

}

// @todo move this to an endpoint
func (e Endpoints) Provider(res http.ResponseWriter, req *http.Request) {
}


func (e Endpoints) CompleteUserAuthEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	return nil, nil
}
