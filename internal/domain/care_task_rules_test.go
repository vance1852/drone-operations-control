package domain

import (
	"errors"
	"testing"
	"time"
)

func TestHandoffValidation(t *testing.T) {
	c := Handoff{DroneTaskID: "s1", To: "operator-2", Location: "lab", RecordedAt: time.Now().UTC()}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Handoff{DroneTaskID: "s1"}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("missing device_transfer fields = %v", err)
	}
}

func TestDroneTaskInProgressRequiresAcceptedAndUnexpired(t *testing.T) {
	now := time.Now().UTC()
	s := DroneTask{Status: DroneTaskAccepted, ExpiresAt: now.Add(time.Hour)}
	if err := s.CanBePerformed(now); err != nil {
		t.Fatal(err)
	}
	s.Status = DroneTaskCompleted
	if err := s.CanBePerformed(now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("status error = %v", err)
	}
	s.Status = DroneTaskAccepted
	if err := s.CanBePerformed(now.Add(2 * time.Hour)); !errors.Is(err, ErrExpired) {
		t.Fatalf("expiry error = %v", err)
	}
}

func TestObservationDecisionRejectsNegativeLimit(t *testing.T) {
	status, err := ObservationDecision(1, 2)
	if err != nil || status != ObservationVerified {
		t.Fatalf("decision = %s, %v", status, err)
	}
	if _, err := ObservationDecision(1, -1); !errors.Is(err, ErrConflict) {
		t.Fatalf("negative limit error = %v", err)
	}
}
