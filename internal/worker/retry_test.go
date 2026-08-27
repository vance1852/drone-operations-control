package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryPolicyCapsExponentialDelay(t *testing.T) {
	p := RetryPolicy{BaseDelay: time.Second, MaxDelay: 3 * time.Second}
	if p.Delay(1) != time.Second || p.Delay(2) != 2*time.Second || p.Delay(3) != 3*time.Second {
		t.Fatalf("delays are incorrect")
	}
}

func TestRunWithRetryStopsAfterSuccess(t *testing.T) {
	attempts := 0
	err := RunWithRetry(context.Background(), RetryPolicy{Attempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}, func(context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil || attempts != 2 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestRunWithRetryPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunWithRetry(ctx, RetryPolicy{Attempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, func(context.Context) error { t.Fatal("operation ran after cancel"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

func TestRunWithRetryGivesEachAttemptFreshContext(t *testing.T) {
	var seen []error
	err := RunWithRetry(context.Background(), RetryPolicy{Attempts: 3, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}, func(attemptCtx context.Context) error {
		seen = append(seen, attemptCtx.Err())
		if len(seen) < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	for i, cErr := range seen {
		if cErr != nil {
			t.Fatalf("attempt %d ran with already-cancelled context: %v", i+1, cErr)
		}
	}
}
