package transports

import (
	"context"
	"github.com/Rennbon/donself/pb"
	"github.com/Rennbon/donself/service"
)

//主要做接口对外参数与service内部参数的转换
type Mapper interface {
	DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error)
	EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error)
}

type mapper struct {
}

func (o *mapper) DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error) {
	in := m.(*pb.AllMyTargetsRequest)
	return &service.Page{
		PageIndex: in.PageIndex,
		PageSize:  in.PageSize,
	}, nil
}

func (o *mapper) EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error) {
	in := m.([]*service.Target)
	out := new(pb.AllMyTargetsResponse)
	for _, v := range in {
		tmp := o.toPbTarget(v)
		out.Targets = append(out.Targets, tmp)
	}
	return out, nil
}

func (o *mapper) toTarget(m *pb.Target) *service.Target {
	return &service.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
func (o *mapper) toPbTarget(m *service.Target) *pb.Target {
	return &pb.Target{
		Id:     m.Id,
		Title:  m.Title,
		Score:  m.Score,
		Symbol: m.Symbol,
	}
}
