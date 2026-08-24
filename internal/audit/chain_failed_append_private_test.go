package audit

import (
	"testing"
	"time"
)

func TestChainFailedAppendDoesNotChangeNextEventHead(t *testing.T) {
	first := Event{RequestID: "req-1", ObjectType: "mission", ObjectID: "mission-1", Action: "create", Outcome: "success", CreatedAt: time.Unix(1, 0).UTC()}
	next := Event{RequestID: "req-2", ObjectType: "mission", ObjectID: "mission-2", Action: "schedule", Outcome: "success", CreatedAt: time.Unix(2, 0).UTC()}
	var uninterrupted Chain
	var afterFailure Chain
	if _, err := uninterrupted.Append(first); err != nil {
		t.Fatalf("append first event to control chain: %v", err)
	}
	if _, err := afterFailure.Append(first); err != nil {
		t.Fatalf("append first event to exercised chain: %v", err)
	}
	invalid := Event{RequestID: "req-bad", ObjectType: "mission", ObjectID: "mission-bad", Action: "update", Outcome: "failure", Detail: map[string]any{"stream": make(chan int)}, CreatedAt: time.Unix(3, 0).UTC()}
	if _, err := afterFailure.Append(invalid); err == nil {
		t.Fatal("append with unsupported detail unexpectedly succeeded")
	}
	want, err := uninterrupted.Append(next)
	if err != nil {
		t.Fatalf("append next event to control chain: %v", err)
	}
	got, err := afterFailure.Append(next)
	if err != nil {
		t.Fatalf("append next event after rejected event: %v", err)
	}
	if got != want {
		t.Fatalf("failed append changed future chain head: got=%s want=%s", got, want)
	}
}
