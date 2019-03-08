package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"flag"
	"github.com/danilopimenta/micro-api-rpc/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"github.com/go-kit/kit/log"

	kithttp "github.com/go-kit/kit/transport/http"
)

const (
	defaultPort = "8080"
)

func main() {

	var (
		addr = envString("PORT", defaultPort)

		httpAddr = flag.String("http.addr", ":"+addr, "HTTP listen address")
	)

	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	httpLogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()
	var hs service.HiService

	hs = service.NewService()

	mux.Handle("/hi", microFramework(hs, httpLogger))

	http.Handle("/", accessControl(mux))

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening")
		errs <- http.ListenAndServe(*httpAddr, nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func microFramework(hs service.HiService, logger log.Logger) http.Handler {

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

type sayHiRequest struct {
}

type hiResponse struct {
	Say string `json:"say"`
	Err error  `json:"error,omitempty"`
}

func makeHiEndpoint(s service.HiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(sayHiRequest)

		return hiResponse{Say: s.Hi(), Err: nil}, nil
	}
}

func decodeSayHiRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return sayHiRequest{}, nil
}

type errorer interface {
	error() error
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// ErrInvalidArgument is returned when one or more arguments are invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
