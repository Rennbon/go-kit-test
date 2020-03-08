//这里相当与三层架构的业务逻辑层或者DDD模式的domain层
package service

import (
	"context"
)

type DonselfService interface {
	GetAllMyTargets(ctx context.Context, page *Page) ([]*Target, error)
}

type donselfService struct {
}

func NewDonselfService() DonselfService {
	return new(donselfService)
}

func (o *donselfService) GetAllMyTargets(ctx context.Context, page *Page) ([]*Target, error) {
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
