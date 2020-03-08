package middleware

import (
	"context"
	"fmt"
	"github.com/Rennbon/donself/service"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type tracingMiddleware struct {
	next   service.DonselfService
	tracer stdopentracing.Tracer
}

func WithTrace(svc service.DonselfService, tracer stdopentracing.Tracer) service.DonselfService {
	m := new(tracingMiddleware)
	m.tracer = tracer
	m.next = svc
	return m
}

func (o *tracingMiddleware) GetAllMyTargets(ctx context.Context, page *service.Page) ([]*service.Target, error) {
	opentracing.TraceServer(o.tracer, "AllMyTargets")
	fmt.Println("trace middleware")
	return o.next.GetAllMyTargets(ctx, page)
}
