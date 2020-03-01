package application

import (
	"context"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/pb"
)

type GrpcMapper interface {
	DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error)
	EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error)
}

type mapper struct {
}

func (o *mapper) DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error) {
	in := m.(*pb.AllMyTargetsRequest)
	return &domain.Page{
		PageIndex: in.PageIndex,
		PageSize:  in.PageSize,
	}, nil
}

func (o *mapper) EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error) {
	in := m.([]*domain.Target)
	out := new(pb.AllMyTargetsResponse)
	for _, v := range in {
		tmp := o.toPbTarget(v)
		out.Targets = append(out.Targets, tmp)
	}
	return out, nil
}

func (o *mapper) toTarget(m *pb.Target) *domain.Target {
	return &domain.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
func (o *mapper) toPbTarget(m *domain.Target) *pb.Target {
	return &pb.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
