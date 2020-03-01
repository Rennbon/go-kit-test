package domain

import (
	"context"
)

type DonselfDomain interface {
	GetAllMyTargets(ctx context.Context, page *Page) ([]*Target, error)
}

type donselfDomain struct {
}

func NewDonselfDomain() DonselfDomain {
	return new(donselfDomain)
}

func (o *donselfDomain) GetAllMyTargets(ctx context.Context, page *Page) ([]*Target, error) {
	arr := []*Target{
		{
			Id:     1,
			Title:  "标题",
			Score:  50,
			Symbol: "symbol.jpg",
		},
	}
	return arr, nil
}
