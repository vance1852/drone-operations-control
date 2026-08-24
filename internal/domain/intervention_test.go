package domain

import (
	"errors"
	"testing"
	"time"
)

func TestInterventionDueAndValidation(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	d := Intervention{DroneTaskID: "s1", Kind: "reassess", Status: InterventionOpen, Reason: "over limit", DueAt: now}
	if err := d.Validate(now); err != nil {
		t.Fatal(err)
	}
	if !d.IsDue(now) {
		t.Fatal("due safety_alert not detected")
	}
	closed := d
	closed.Status = InterventionClosed
	if closed.IsDue(now) {
		t.Fatal("closed safety_alert is due")
	}
	if err := closed.Validate(now); !errors.Is(err, ErrConflict) {
		t.Fatalf("closed timestamp error = %v", err)
	}
}
