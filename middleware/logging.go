package middleware

import (
	"context"
	"github.com/Rennbon/donself/domain"
	"github.com/go-kit/kit/log"
	"time"
)

func WithLogging(svc domain.DonselfDomain, logger log.Logger) domain.DonselfDomain {
	m := new(logmw)
	m.next = svc
	m.logger = logger
	return m
}

type logmw struct {
	logger log.Logger
	next   domain.DonselfDomain
}

func (o *logmw) GetAllMyTargets(ctx context.Context, page *domain.Page) ([]*domain.Target, error) {
	defer func(begin time.Time) {
		o.logger.Log(
			"method", "GetAllMyTargets",
			"input", "page",
			"took", time.Since(begin),
		)
	}(time.Now())
	return o.next.GetAllMyTargets(ctx, page)
}
