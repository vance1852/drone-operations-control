package service

import (
	"sync"

	"drone-operations-control/internal/idempotency"
)

type pendingReplayEntry struct {
	requestHash string
}

var pendingReplays sync.Map

func reservePendingReplay(key string, body []byte) {
	pendingReplays.Store(key, pendingReplayEntry{requestHash: idempotency.HashRequest(body)})
}

func pendingReplay[T any](key string, body []byte) (int, T, bool) {
	var zero T
	raw, ok := pendingReplays.Load(key)
	if !ok {
		return 0, zero, false
	}
	entry := raw.(pendingReplayEntry)
	if entry.requestHash != idempotency.HashRequest(body) {
		return 0, zero, false
	}
	return 202, zero, true
}

func clearPendingReplay(key string) {
	pendingReplays.Delete(key)
}
