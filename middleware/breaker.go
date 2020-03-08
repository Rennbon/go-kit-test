package middleware

import (
	"context"
	"fmt"
	"github.com/Rennbon/donself/service"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/sony/gobreaker"
	"time"
)

type breakerMiddleware struct {
	next service.DonselfService
}

func WithCircuitBreaker(svc service.DonselfService) service.DonselfService {
	m := new(breakerMiddleware)
	m.next = svc
	return m
}

func (o *breakerMiddleware) GetAllMyTargets(ctx context.Context, page *service.Page) ([]*service.Target, error) {
	circuitbreaker.Gobreaker(
		gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "AllMyTargets",
			MaxRequests: 2,               //half-open状态下允许放行请求
			Interval:    0,               //重置计数周期
			Timeout:     time.Second * 5, //open状态下切入half-open周期，成功才切，不成功继续。
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 5
			},
		}))
	fmt.Println("breaker middleware")
	return o.next.GetAllMyTargets(ctx, page)
}
