//传输层，选择传输协议用
package transports

import (
	"context"
	"errors"
	"github.com/Rennbon/donself/common"
	"github.com/Rennbon/donself/middleware"
	"github.com/Rennbon/donself/service"
	//"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	//"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
	//"github.com/sony/gobreaker"
	//"time"
)

//对外暴露内层，这里能通过接口方式桥接http,grpc等，
type Transports struct {
	svc                  service.DonselfService
	AllMyTargetsEndpoint endpoint.Endpoint
}

//svc 为业务逻辑层，只需要锚定业务即可，无所考虑等和对外暴露结构相关逻辑，
//所以这里的入参都是内部实体，竟可能不要用api传进来的实体，否则会污染代码，
//不适于目前的设计模式，对于拆分解耦也不利。
func NewTransports(svc service.DonselfService, tracer stdopentracing.Tracer, logger log.Logger) *Transports {

	tp := new(Transports)
	//下面的先执行
	svc = middleware.WithCircuitBreaker(svc)
	svc = middleware.WithMetric(svc)
	svc = middleware.WithLogging(svc, logger)
	svc = middleware.WithTrace(svc, tracer)
	tp.svc = svc
	tp.AllMyTargetsEndpoint = tp.makeAllMyTargetsEndpoint()
	/*	tp.AllMyTargetsEndpoint = circuitbreaker.Gobreaker(
		gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "AllMyTargets",
			MaxRequests: 2,               //half-open状态下允许放行请求
			Interval:    0,               //重置计数周期
			Timeout:     time.Second * 5, //open状态下切入half-open周期，成功才切，不成功继续。
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 5
			},
		}),
	)(tp.AllMyTargetsEndpoint)*/
	//tp.AllMyTargetsEndpoint = opentracing.TraceServer(tracer, "AllMyTargets")(tp.AllMyTargetsEndpoint)
	return tp
}

func (o *Transports) makeAllMyTargetsEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*service.Page)
		ctx, cancelFunc := common.Context()
		defer cancelFunc()
		//这里用来调试circuit
		if req.PageIndex > 50 && req.PageIndex < 100 {
			return nil, errors.New("手动错误！！！")
		}
		v, err := o.svc.GetAllMyTargets(ctx, req)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}
