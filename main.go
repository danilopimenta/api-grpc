package main

import (
	"flag"
	"fmt"
	"github.com/danilopimenta/api-grpc/pb"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/danilopimenta/api-grpc/hi"
	"github.com/danilopimenta/api-grpc/hi/transport"
	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

const (
	defaultPort = "8080"
)

func main() {

	var (
		addr     = envString("PORT", defaultPort)
		grpcAddr = flag.String("grpc-addr", ":8082", "gRPC listen address")
		httpAddr = flag.String("http.addr", ":"+addr, "HTTP listen address")
	)

	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	httpLogger := log.With(logger, "component", "http")

	serverMux := http.NewServeMux()
	var hs hi.HiService

	hs = hi.NewService()

	grpcServer := transport.GRPCHandler(hs, logger)

	grpcListener, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		logger.Log("transport", "gRPC", "addr", *grpcAddr)
	}

	logger.Log("transport", "gRPC", "addr", *grpcAddr)

	baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
	pb.RegisterHiServer(baseServer, grpcServer)
	err = baseServer.Serve(grpcListener)
	if err != nil {
		logger.Log("transport", "gRPC", "addr", *grpcAddr)
	}

	serverMux.Handle("/hi", transport.HTTPHandler(hs, httpLogger))

	http.Handle("/", accessControl(serverMux))

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
