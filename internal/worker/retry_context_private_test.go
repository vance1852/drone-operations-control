package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunWithRetryGivesEachAttemptLiveContext(t *testing.T) {
	calls := 0
	err := RunWithRetry(t.Context(), RetryPolicy{
		Attempts:  2,
		BaseDelay: time.Nanosecond,
		MaxDelay:  time.Nanosecond,
	}, func(ctx context.Context) error {
		calls++
		if calls == 1 {
			return errors.New("temporary telemetry gateway failure")
		}
		if err := ctx.Err(); err != nil {
			t.Fatalf("retry attempt received canceled context: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retry should recover from a temporary failure: %v", err)
	}
	if calls != 2 {
		t.Fatalf("executor calls=%d want=2", calls)
	}
}
