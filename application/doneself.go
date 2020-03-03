package application

import (
	"context"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/pb"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	kitgrpc "github.com/go-kit/kit/transport/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
)

//需要实现pb定义的接口
type DoneselfServer struct {
	GetMyAllTargets kitgrpc.Handler
	mp              GrpcMapper
}

func NewDoneselfServer(svc domain.DonselfDomain, logger log.Logger, tracer stdopentracing.Tracer) pb.DoneselfServer {
	tp := NewTransports(svc, tracer)
	server := new(DoneselfServer)
	server.mp = new(mapper)

	options := []kitgrpc.ServerOption{
		kitgrpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	server.GetMyAllTargets = kitgrpc.NewServer(
		tp.AllMyTargetsEndpoint,
		server.mp.DecodeAllMyTargetsRequest,
		server.mp.EncodeAllMyTargetsResponse,
		append(options, kitgrpc.ServerBefore(opentracing.GRPCToContext(tracer, "allMyTargets", logger)))...,
	)
	return server
}

//pb接口实现
func (s *DoneselfServer) AllMyTargets(ctx context.Context, req *pb.AllMyTargetsRequest) (*pb.AllMyTargetsResponse, error) {
	_, rep, err := s.GetMyAllTargets.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.AllMyTargetsResponse), nil
}
