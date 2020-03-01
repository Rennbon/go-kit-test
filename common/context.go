package common

import (
	"context"
	"time"
)

func Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Minute)
}
