package transports

import (
	"context"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"go.elastic.co/apm"
	"net/http"
)

func ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	span, ctx := apm.StartSpan(ctx, "errorEncoder", "response")
	defer span.End()

	body, _ := json.Marshal(errorWrapper{Error: err.Error()})
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			body = jsonBody
		}
	}
	w.Header().Set("Content-Type", ContentType)
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

const ContentType = "application/json; charset=utf-8"

type errorWrapper struct {
	Error string `json:"error"`
}
