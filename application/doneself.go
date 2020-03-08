package application

import (
	"context"
	"github.com/Rennbon/donself/pb"
	"github.com/Rennbon/donself/service"
	"github.com/Rennbon/donself/transports"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	kitgrpc "github.com/go-kit/kit/transport/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
)

//需要实现pb定义的接口
type DoneselfServer struct {
	GetMyAllTargets kitgrpc.Handler
	grpcmp          Mapper
}

func NewDoneselfServer(svc service.DonselfService, logger log.Logger, tracer stdopentracing.Tracer) pb.DoneselfServer {
	tp := transports.NewTransports(svc, tracer, logger)
	server := new(DoneselfServer)
	server.grpcmp = new(grpcMapper)

	options := []kitgrpc.ServerOption{
		kitgrpc.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}
	server.GetMyAllTargets = kitgrpc.NewServer(
		tp.AllMyTargetsEndpoint,
		server.grpcmp.DecodeAllMyTargetsRequest,
		server.grpcmp.EncodeAllMyTargetsResponse,
		append(options,
			kitgrpc.ServerBefore(opentracing.GRPCToContext(tracer, "AllMyTargets", logger)),
		)...,
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
