package transport

import (
	"context"

	"github.com/danilopimenta/micro-api-rpc/hi"
	"github.com/danilopimenta/micro-api-rpc/pb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)


type grpcServer struct {
	saying    grpctransport.Handler
}

type Set struct {
	HiEndpoint    endpoint.Endpoint
}

func GRPCHandler(hs hi.HiService, logger log.Logger) pb.HiServer {
	endpoints := NewEndpoint(hs)
	opts := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}
	return &grpcServer{
		saying: grpctransport.NewServer(
			endpoints.HiEndpoint,
			decodeGRPCHiRequest,
			encodeGRPCHiResponse,
			opts...,
		),
	}
}
// SumRequest collects the request parameters for the Sum method.
type hiRequest struct {
	Say string
}


func (s *grpcServer) Saying(ctx context.Context, req *pb.SayingRequest) (*pb.SayingResponse, error) {
	_, rep, err := s.saying.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SayingResponse), nil
}


// decodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func decodeGRPCHiRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SayingRequest)
	return hiRequest{Say: string(req.Say)}, nil
}

// encodeGRPCSumResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sum response to a gRPC sum reply. Primarily useful in a server.
func encodeGRPCHiResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(hiResponse)
	return &pb.SayingResponse{Say: string(resp.Say)}, nil
}

// MakeSumEndpoint constructs a Sum endpoint wrapping the service.
func makeHiEndpointRpc(s hi.HiService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(hiRequest)

		return hiResponse{Say: s.Hi(), Err: nil}, nil
	}
}

func NewEndpoint(hs hi.HiService) Set {
	var hiEndpoint endpoint.Endpoint
	{
		hiEndpoint = makeHiEndpointRpc(hs)
	}

	return Set{
		HiEndpoint: hiEndpoint,
	}
}