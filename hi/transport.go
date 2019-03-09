package hi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	kithttp "github.com/go-kit/kit/transport/http"
)

type sayHiRequest struct {
}

type hiResponse struct {
	Say string `json:"say"`
	Err error  `json:"error,omitempty"`
}

type errorer interface {
	error() error
}

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

func Handler(hs HiService, logger log.Logger) http.Handler {

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(encodeError),
	}

	hiHandler := kithttp.NewServer(
		makeHiEndpoint(hs),
		decodeSayHiRequest,
		encodeResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/hi", hiHandler).Methods("GET")

	return r
}

func makeHiEndpoint(s HiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(sayHiRequest)

		return hiResponse{Say: s.Hi(), Err: nil}, nil
	}
}

func decodeSayHiRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return sayHiRequest{}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
