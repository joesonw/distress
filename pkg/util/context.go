package util

import (
	"context"
	"time"
)

func NewOptionalTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return OptionalTimeoutContext(context.Background(), timeout)
}

func OptionalTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc = func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	return ctx, cancel
}
