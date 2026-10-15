package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"example.com/ctxtalk/xctx"
)

var errSlow = errors.New("slow")

func checkInventory(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	// 1. Bare timeout.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err := checkInventory(ctx, time.Second)
		fmt.Println("1 bare:", err, "| cause:", context.Cause(ctx))
	}
	// 2. WithCancelCause, first cancel wins.
	{
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(fmt.Errorf("order %s: inventory check failed: %w", "ord-123", errSlow))
		cancel(errors.New("processOrder completed"))
		fmt.Println("2 cause:", context.Cause(ctx), "| err:", ctx.Err())
		fmt.Println("2 Is Canceled on cause:", errors.Is(context.Cause(ctx), context.Canceled))
	}
	// 3. WithTimeoutCause gotcha.
	{
		ctx, cancel := context.WithTimeoutCause(context.Background(), time.Second, errors.New("order ord-123: 5s timeout"))
		cancel()
		fmt.Println("3 normal return cause:", context.Cause(ctx))
		ctx2, cancel2 := context.WithTimeoutCause(context.Background(), 5*time.Millisecond, errors.New("order ord-123: 5s timeout"))
		defer cancel2()
		<-ctx2.Done()
		fmt.Println("3 timer cause:", context.Cause(ctx2), "| err:", ctx2.Err())
	}
	// 4. Stacked.
	{
		parent, cancelCause := context.WithCancelCause(context.Background())
		ctx, cancelTimeout := context.WithTimeoutCause(parent, time.Second, errors.New("timeout"))
		func() {
			defer cancelTimeout()
			defer cancelCause(errors.New("processOrder completed"))
		}()
		fmt.Println("4 stacked inner cause:", context.Cause(ctx), "| outer:", context.Cause(parent))
		_, hasDeadline := ctx.Deadline()
		fmt.Println("4 inner has deadline:", hasDeadline)
	}
	// 5. net/http returns the cause.
	{
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
		}))
		defer srv.Close()
		ctx, cancel := context.WithTimeoutCause(context.Background(), 20*time.Millisecond, errors.New("inventory: 20ms budget spent"))
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		_, err := http.DefaultClient.Do(req)
		fmt.Println("5 http err:", err)
		fmt.Println("5 Is DeadlineExceeded:", errors.Is(err, context.DeadlineExceeded))

		ctx2, cancel2 := xctx.WithTimeout(context.Background(), 20*time.Millisecond, errors.New("inventory: 20ms budget spent"))
		defer cancel2(errors.New("done"))
		req2, _ := http.NewRequestWithContext(ctx2, http.MethodGet, srv.URL, nil)
		_, err = http.DefaultClient.Do(req2)
		fmt.Println("5 xctx http err:", err)
		fmt.Println("5 xctx Is DeadlineExceeded:", errors.Is(err, context.DeadlineExceeded))
	}
	// 6. xctx normal return.
	{
		ctx, cancel := xctx.WithTimeout(context.Background(), time.Second, errors.New("GetOffers: timed out after 10s"))
		cancel(errors.New("GetOffers: completed"))
		fmt.Println("6 xctx cause:", context.Cause(ctx), "| Is Canceled:", errors.Is(context.Cause(ctx), context.Canceled), "| err:", ctx.Err())
		_, ok := ctx.Deadline()
		fmt.Println("6 deadline kept:", ok)
	}
	// 7. Key collision.
	{
		ctx := context.WithValue(context.Background(), "key", "from-main")
		ctx = context.WithValue(ctx, "key", "from-foo")
		fmt.Println("7 collision:", ctx.Value("key"))
		type userKey string
		ctx = context.WithValue(ctx, userKey("key"), "typed")
		fmt.Println("7 typed:", ctx.Value(userKey("key")), ctx.Value("key"))
	}
	// 8. WithoutCancel keeps values, drops deadline.
	{
		type reqIDKey struct{}
		base := context.WithValue(context.Background(), reqIDKey{}, "req-42")
		ctx, cancel := context.WithTimeout(base, time.Millisecond)
		cancel()
		d := context.WithoutCancel(ctx)
		_, ok := d.Deadline()
		fmt.Println("8 detached err:", d.Err(), "| value:", d.Value(reqIDKey{}), "| deadline:", ok)
	}
}
