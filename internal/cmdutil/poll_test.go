package cmdutil

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPollUntil(t *testing.T) {
	t.Parallel()

	errNotYet := errors.New("not yet")

	t.Run("returns once the check passes", func(t *testing.T) {
		t.Parallel()
		calls := 0
		err := PollUntil(context.Background(), time.Second, time.Millisecond, time.Second, func(context.Context) error {
			calls++
			if calls < 3 {
				return errNotYet
			}
			return nil
		})
		if err != nil || calls != 3 {
			t.Fatalf("err = %v after %d calls, want nil after 3", err, calls)
		}
	})

	t.Run("timeout reports the last completed check", func(t *testing.T) {
		t.Parallel()
		err := PollUntil(context.Background(), 30*time.Millisecond, time.Millisecond, time.Second, func(ctx context.Context) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errNotYet
		})
		if !errors.Is(err, errNotYet) {
			t.Fatalf("err = %v, want the check's own error", err)
		}
	})

	t.Run("cancellation wins over the check error", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := PollUntil(ctx, time.Hour, time.Hour, time.Second, func(context.Context) error { return errNotYet })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	})
}
