package middleware

import (
	"context"
	"fmt"
	"github.com/Rennbon/donself/service"
	"github.com/go-kit/kit/log"
	"time"
)

func WithLogging(svc service.DonselfService, logger log.Logger) service.DonselfService {
	m := new(logmw)
	m.next = svc
	m.logger = logger
	return m
}

type logmw struct {
	logger log.Logger
	next   service.DonselfService
}

func (o *logmw) GetAllMyTargets(ctx context.Context, page *service.Page) ([]*service.Target, error) {
	defer func(begin time.Time) {
		o.logger.Log(
			"method", "GetAllMyTargets",
			"input", "page",
			"took", time.Since(begin),
		)
	}(time.Now())
	fmt.Println("log middleware")
	return o.next.GetAllMyTargets(ctx, page)
}
