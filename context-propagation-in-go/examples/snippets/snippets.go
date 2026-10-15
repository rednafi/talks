// Package snippets holds the code shown on the slides, with the elided
// parts filled in, so that every slide snippet compiles and passes vet.
package snippets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Order struct{}

type Store struct{}

func (s *Store) Order(ctx context.Context, id string) (Order, error) { return Order{}, nil }

type pool struct{}

func (pool) Submit(func()) {}

type auditLog struct{}

func (auditLog) Write(ctx context.Context, id string) {}

type Handler struct {
	pool  pool
	audit auditLog
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	detached := context.WithoutCancel(r.Context())
	h.pool.Submit(func() {
		ctx, cancel := context.WithTimeout(detached, 5*time.Second)
		defer cancel()
		h.audit.Write(ctx, orderID)
	})
}

func checkInventory(ctx context.Context, id string) error { return ctx.Err() }
func chargePayment(ctx context.Context, id string) error  { return ctx.Err() }
func shipOrder(ctx context.Context, id string) error      { return ctx.Err() }

func processOrderWrapped(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := checkInventory(ctx, id); err != nil {
		return fmt.Errorf("check inventory for %s: %w", id, err)
	}
	if err := chargePayment(ctx, id); err != nil {
		return fmt.Errorf("charge payment for %s: %w", id, err)
	}
	return shipOrder(ctx, id)
}

func processOrderCause(ctx context.Context, id string) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.New("processOrder: completed"))

	if err := checkInventory(ctx, id); err != nil {
		cancel(fmt.Errorf("inventory %s: %w", id, err))
		return err
	}
	return nil
}

func processOrderTimeoutCause(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second,
		fmt.Errorf("processOrder %s: 5s timeout", id))
	defer cancel()
	return checkInventory(ctx, id)
}

func processOrderStacked(ctx context.Context, id string) error {
	ctx, cancelCause := context.WithCancelCause(ctx)
	ctx, cancelTimeout := context.WithTimeoutCause(ctx, 5*time.Second,
		fmt.Errorf("processOrder %s: 5s timeout", id))
	defer cancelTimeout()
	defer cancelCause(errors.New("processOrder: completed"))

	if err := checkInventory(ctx, id); err != nil {
		cancelCause(fmt.Errorf("processOrder %s: inventory: %w", id, err))
		return err
	}
	return nil
}

func logCause(ctx context.Context, logger *slog.Logger) {
	if err := ctx.Err(); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "request failed",
			slog.Any("err", err),
			slog.Any("cause", context.Cause(ctx)),
		)
	}
}

type ctxKey struct{}

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func IDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok
}

func httpCause(ctx context.Context, url string) bool {
	ctx, cancel := context.WithTimeoutCause(ctx, 20*time.Millisecond,
		errors.New("inventory: budget spent"))
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	_, err := http.DefaultClient.Do(req)

	return errors.Is(err, context.DeadlineExceeded)
}
