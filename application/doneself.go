package application

import (
	"context"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/pb"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

//需要实现pb定义的接口
type DoneselfServer struct {
	GetMyAllTargets kitgrpc.Handler
	mp              GrpcMapper
}

func NewDoneselfServer(svc domain.DonselfDomain) pb.DoneselfServer {
	tp := NewTransports(svc)
	server := new(DoneselfServer)
	server.mp = new(mapper)

	server.GetMyAllTargets = kitgrpc.NewServer(
		tp.AllMyTargetsEndpoint,
		server.mp.DecodeAllMyTargetsRequest,
		server.mp.EncodeAllMyTargetsResponse,
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
