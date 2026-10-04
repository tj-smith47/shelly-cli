package cmdutil

import (
	"context"
	"time"
)

// PollUntil calls check every interval until it returns nil, the timeout
// passes, or ctx is cancelled. Each call gets at most probeTimeout, so one
// stalled attempt cannot use up the whole wait.
//
// It returns nil once check succeeds and ctx.Err() when ctx is cancelled. On
// timeout it returns the error of the last check that ran to completion, which
// says why the condition did not hold; callers tell the two failures apart with
// ctx.Err().
func PollUntil(ctx context.Context, timeout, interval, probeTimeout time.Duration, check func(context.Context) error) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastErr error
	for {
		probeCtx, probeCancel := context.WithTimeout(waitCtx, probeTimeout)
		err := check(probeCtx)
		probeCancel()
		if err == nil {
			return nil
		}
		// A check cut short by the overall timeout only says "deadline exceeded";
		// the answer from the last completed check explains the timeout better.
		if lastErr == nil || waitCtx.Err() == nil {
			lastErr = err
		}

		select {
		case <-waitCtx.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return lastErr
		case <-time.After(interval):
		}
	}
}
