package db

import (
	"context"
	"time"
)

// advisoryUnlockContext returns an independent, bounded context for releasing a
// session-level advisory lock. It is derived from context.Background rather than
// the request context on purpose: when a request is cancelled, the unlock must
// still reach the database so the lock is freed before the connection returns to
// the pool. Deriving from the already-cancelled request context would leave the
// advisory lock held on a pooled connection, blocking every other instance that
// waits on the same maintenance key.
func advisoryUnlockContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
