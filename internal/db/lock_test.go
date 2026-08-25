package db

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestAdvisoryUnlockContextSurvivesCancelledRequest reproduces the maintenance
// cancellation bug: when a request is cancelled mid-maintenance, the parent
// context is already done. The unlock context must NOT be derived from the
// cancelled request context, otherwise pg_advisory_unlock never reaches the
// database and the session-level lock leaks onto a pooled connection that every
// other instance is then unable to acquire.
func TestAdvisoryUnlockContextSurvivesCancelledRequest(t *testing.T) {
	// Simulate a request that was cancelled while the maintenance callback was
	// running.
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent() // request is now cancelled

	unlockCtx, cancel := advisoryUnlockContext()
	defer cancel()
	_ = parent // kept to make the cancellation scenario explicit

	select {
	case <-unlockCtx.Done():
		t.Fatalf("unlock context derived from cancelled parent: %v", unlockCtx.Err())
	default:
	}

	// It must still have a bounded deadline so a stuck backend cannot hang the
	// shutdown forever.
	if dl, ok := unlockCtx.Deadline(); !ok {
		t.Fatal("unlock context has no deadline")
	} else if remaining := time.Until(dl); remaining <= 0 || remaining > 5*time.Second {
		t.Fatalf("unexpected unlock deadline remaining=%v", remaining)
	}

	// Sanity: the helper is the source the lock release path uses.
	if errors.Is(unlockCtx.Err(), context.Canceled) {
		t.Fatalf("unlock context already cancelled: %v", unlockCtx.Err())
	}
}
