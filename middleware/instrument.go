package middleware

import (
	"context"
	"fmt"
	"github.com/Rennbon/donself/service"
	"github.com/go-kit/kit/metrics"
	kitp "github.com/go-kit/kit/metrics/prometheus"
	basep "github.com/prometheus/client_golang/prometheus"
	"time"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           service.DonselfService
}

func WithMetric(svc service.DonselfService) service.DonselfService {
	m := new(instrumentingMiddleware)
	m.requestCount = kitp.NewCounterFrom(
		basep.CounterOpts{
			Namespace: "donself",
			Subsystem: "get_all_my_targets",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{})
	m.requestLatency = kitp.NewHistogramFrom(
		basep.HistogramOpts{
			Namespace: "donself",
			Subsystem: "get_all_my_targets",
			Name:      "request_latency",
			Help:      "Interface call time.",
		}, []string{})
	m.next = svc
	return m
}

func (o *instrumentingMiddleware) GetAllMyTargets(ctx context.Context, page *service.Page) ([]*service.Target, error) {
	defer func(begin time.Time) {
		o.requestCount.Add(1)
		o.requestLatency.Observe(time.Since(begin).Seconds())
	}(time.Now())
	fmt.Println("instrument middleware")
	return o.next.GetAllMyTargets(ctx, page)
}
