package background

import (
	"context"
	"time"
)

// noCancelCtx ignores any context cancellations.
type noCancelCtx struct {
	parent context.Context
}

func WithNoCancelContext(ctx context.Context) context.Context {
	return &noCancelCtx{parent: ctx}
}

func (c *noCancelCtx) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}

func (c *noCancelCtx) Done() <-chan struct{} {
	// From golang context package documentation:
	// Done may return nil if this context can never be canceled.
	return nil
}

func (*noCancelCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *noCancelCtx) Err() error {
	return nil
}
