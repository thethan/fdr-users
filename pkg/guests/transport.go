package guests

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"strings"

	"errors"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

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

type errorWrapper struct {
	Error string `json:"error"`
}

const contentType = "application/json; charset=utf-8"

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

// MakeHTTPHandler returns a handler that makes a set of endpoints available
// on predefined paths.
func MakeHTTPHandler(logger log.Logger, endpoints Endpoints, m *mux.Router, authServerBefore httptransport.RequestFunc, options ...httptransport.ServerOption) *mux.Router {
	serverOptions := []httptransport.ServerOption{
		httptransport.ServerBefore(headersToContext),
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerAfter(httptransport.SetContentType(contentType)),
	}
	serverOptions = append(serverOptions, options...)
	serverOptionsAuth := append(serverOptions, httptransport.ServerBefore(authServerBefore))

	m = m.PathPrefix("/puretotten").Subrouter()
	m.Methods(http.MethodGet).Path("/rsvp").Handler(httptransport.NewServer(
		endpoints.GetPureTottenRSVP,
		DecodeHTTPGetRSVP,
		EncodeRSVPCSV,
		serverOptionsAuth...,
	))

	m.Methods(http.MethodPost).Path("/rsvp").Handler(httptransport.NewServer(
		endpoints.SavePureTottenRSVP,
		DecodeHTTPPOSTRSVP,
		EncodeSaveResponse,
		serverOptionsAuth...,
	))

	return m
}

func DecodeHTTPGetRSVP(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	return nil, nil
}

type GuestRequest struct {
	*Guest
	Token string `json:"token"`
}

func DecodeHTTPPOSTRSVP(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req GuestRequest

	//fmt.Printf("%s\n", string(r.Body.))
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, err
	}
	return &req, nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeSaveResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {

	w.Header().Set("Content-Type", "application/json") // setting the content type header to text/csv
	w.WriteHeader(http.StatusAccepted)
	return nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeRSVPCSV(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res, ok := response.([]Guest)
	if !ok {
		return errors.New("could not get user Credentials Response ")
	}
	records := convertGuestsToStringSlice(res)
	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename=PureTottenRSVPList.csv")
	wr := csv.NewWriter(w)
	for _, record := range records {
		_ = wr.Write(record)
	}
	wr.Flush()

	return nil
}

func convertGuestsToStringSlice(guests []Guest) [][]string {
	rowsSlice := make([][]string, len(guests)+1)

	// csv
	// name, email, adults, children, attending, vegan option, signed up
	rowsSlice[0] = []string{"Name", "Email", "Adults", "Children", "Attending", "Vegan Options", "Created At"}
	for idx := range guests {
		rowsSlice[idx+1] = convertGuestToSliceOfStrings(guests[idx])
	}
	return rowsSlice
}

func convertGuestToSliceOfStrings(guest Guest) []string {
	row := make([]string, 7)
	row[0] = guest.Name
	row[1] = guest.Email
	row[2] = strconv.Itoa(guest.Adults)
	row[3] = strconv.Itoa(guest.Children)
	row[4] = strconv.FormatBool(guest.Attending)
	row[5] = strconv.Itoa(guest.VeganOptionCount)
	row[6] = guest.CreatedAt.Format(time.RFC3339)
	return row
}
