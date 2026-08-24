package domain

import (
	"errors"
	"testing"
	"time"
)

func TestMissionTransitions(t *testing.T) {
	cases := []struct {
		from, to MissionStatus
		want     bool
	}{
		{MissionDraft, MissionScheduled, true},
		{MissionScheduled, MissionActive, true},
		{MissionActive, MissionClosed, true},
		{MissionDraft, MissionActive, false},
		{MissionClosed, MissionDraft, false},
	}
	for _, tc := range cases {
		if got := tc.from.CanMoveTo(tc.to); got != tc.want {
			t.Errorf("%s -> %s = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestDroneTaskMoveSetsTimestampsAndVersion(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DroneTask{ID: "s1", TaskCode: "S-1", Status: DroneTaskQueued, ExpiresAt: now.Add(time.Hour), Version: 3}
	updated, err := task.Move(DroneTaskCompleted, now)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != DroneTaskCompleted || updated.Version != 4 || updated.CompletedAt == nil || !updated.CompletedAt.Equal(now) {
		t.Fatalf("unexpected collected task: %+v", updated)
	}
}

func TestDroneTaskRejectsInvalidMoveAndExpiredDroneTask(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DroneTask{TaskCode: "S-1", Status: DroneTaskQueued, ExpiresAt: now.Add(-time.Minute), Version: 1}
	if _, err := task.Move(DroneTaskAccepted, now); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("invalid transition error = %v", err)
	}
	task.Status = DroneTaskCompleted
	if _, err := task.Move(DroneTaskHandoffPending, now); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired error = %v", err)
	}
}

func TestObservationOutcomeUsesInclusiveThreshold(t *testing.T) {
	if ObservationVerified != ObservationStatus("verified") {
		t.Fatal("approved value changed")
	}
	if ObservationStatus(ObservationVerified).Outcome(10, 10) != ObservationVerified {
		t.Fatal("value at limit should be approved")
	}
	if ObservationStatus(ObservationVerified).Outcome(10.01, 10) != ObservationRejected {
		t.Fatal("value over limit should be rejected")
	}
}
