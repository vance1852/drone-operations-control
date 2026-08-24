package service

import (
	"context"
	"errors"
	"testing"

	"drone-operations-control/internal/idempotency"
)

type emptyIdempotencyStore struct{}

func (emptyIdempotencyStore) Get(context.Context, string) (idempotency.Record, bool, error) {
	return idempotency.Record{}, false, nil
}
func (emptyIdempotencyStore) Put(context.Context, idempotency.Record) error { return nil }

func TestReplayOrFailedCreateDoesNotPoisonRetry(t *testing.T) {
	type response struct {
		ID string `json:"id"`
	}
	store := emptyIdempotencyStore{}
	calls := 0
	create := func() (int, response, error) {
		calls++
		if calls == 1 {
			return 0, response{}, errors.New("temporary mission validation dependency failure")
		}
		return 201, response{ID: "task-created-on-retry"}, nil
	}
	const key = "private-failed-create-retry-009"
	body := []byte(`{"mission_id":"mission-9"}`)
	if _, _, err := ReplayOr(t.Context(), store, key, body, create); err == nil {
		t.Fatal("first create unexpectedly succeeded")
	}
	code, got, err := ReplayOr(t.Context(), store, key, body, create)
	if err != nil {
		t.Fatalf("retry create: %v", err)
	}
	if calls != 2 || code != 201 || got.ID != "task-created-on-retry" {
		t.Fatalf("retry calls=%d code=%d response=%+v", calls, code, got)
	}
}
