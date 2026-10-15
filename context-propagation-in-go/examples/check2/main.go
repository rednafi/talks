package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type psp struct{}

func (psp) Charge(ctx context.Context, req string) error { return ctx.Err() }

func backoff(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func charge(ctx context.Context, p psp, req string) error {
	var err error
	for attempt := range 5 {
		if err = p.Charge(ctx, req); err == nil {
			return nil
		}
		select {
		case <-time.After(backoff(attempt)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

type user struct{}
type sessions struct{}

func (sessions) User(ctx context.Context, tok string) (user, error) { return user{}, nil }

type ctxKey struct{}

func WithUser(ctx context.Context, u user) context.Context { return context.WithValue(ctx, ctxKey{}, u) }

type Auth struct{ sessions sessions }

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.sessions.User(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			return
		}
		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	root, c0 := context.WithTimeout(context.Background(), 2*time.Second)
	defer c0()
	ctx, cancel := context.WithTimeout(root, 10*time.Second)
	defer cancel()
	d, _ := ctx.Deadline()
	fmt.Println("deadline in", time.Until(d).Round(time.Second))

	parent, pc := context.WithCancelCause(context.Background())
	child, stop := context.WithTimeout(parent, time.Minute)
	defer stop()
	pc(errors.New("order-service: shutting down"))
	fmt.Println(child.Err(), "|", context.Cause(child))

	c, cc := context.WithTimeout(context.Background(), time.Millisecond)
	defer cc()
	fmt.Println(charge(c, psp{}, "x"))

	ctx2, cancel2 := context.WithTimeoutCause(context.Background(), time.Millisecond,
		fmt.Errorf("inventory: budget spent: %w", context.DeadlineExceeded))
	defer cancel2()
	time.Sleep(5 * time.Millisecond)
	fmt.Println(errors.Is(context.Cause(ctx2), context.DeadlineExceeded))
	_ = exec.CommandContext
}
