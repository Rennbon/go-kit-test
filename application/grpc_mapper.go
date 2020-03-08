package application

import (
	"context"
	"github.com/Rennbon/donself/pb"
	"github.com/Rennbon/donself/service"
)

//grpc pb decode and encode
type grpcMapper struct {
}

func (o *grpcMapper) DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error) {
	in := m.(*pb.AllMyTargetsRequest)
	return &service.Page{
		PageIndex: in.PageIndex,
		PageSize:  in.PageSize,
	}, nil
}

func (o *grpcMapper) EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error) {
	in := m.([]*service.Target)
	out := new(pb.AllMyTargetsResponse)
	for _, v := range in {
		tmp := o.toPbTarget(v)
		out.Targets = append(out.Targets, tmp)
	}
	return out, nil
}

func (o *grpcMapper) toTarget(m *pb.Target) *service.Target {
	return &service.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
func (o *grpcMapper) toPbTarget(m *service.Target) *pb.Target {
	return &pb.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
