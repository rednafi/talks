// Package xctx is a trimmed copy of the idea behind Pedregal's
// libraries/go/xctx: every cancellation carries a cause that also wraps the
// matching context sentinel.
package xctx

import (
	"context"
	"fmt"
	"time"
)

// WithCancel is like context.WithCancelCause, but every cause wraps
// context.Canceled.
func WithCancel(parent context.Context) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	return ctx, func(cause error) {
		cancel(fmt.Errorf("%w: %w", cause, context.Canceled))
	}
}

// WithTimeout is like context.WithTimeoutCause, but the returned cancel
// function takes a cause too.
func WithTimeout(parent context.Context, d time.Duration, timeoutCause error) (context.Context, context.CancelCauseFunc) {
	return WithDeadline(parent, time.Now().Add(d), timeoutCause)
}

// WithDeadline is like context.WithDeadlineCause, but the returned cancel
// function takes a cause too.
func WithDeadline(parent context.Context, d time.Time, deadlineCause error) (context.Context, context.CancelCauseFunc) {
	deadlineCtx, deadlineCancel := context.WithDeadlineCause(parent, d,
		fmt.Errorf("%w: %w", deadlineCause, context.DeadlineExceeded))
	causeCtx, causeCancel := WithCancel(deadlineCtx)
	return causeCtx, func(cause error) {
		causeCancel(cause)
		deadlineCancel()
	}
}
