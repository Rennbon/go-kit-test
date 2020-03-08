package application

import "context"

//主要做接口对外参数与service内部参数的转换
type Mapper interface {
	DecodeAllMyTargetsRequest(_ context.Context, m interface{}) (interface{}, error)
	EncodeAllMyTargetsResponse(_ context.Context, m interface{}) (interface{}, error)
}
