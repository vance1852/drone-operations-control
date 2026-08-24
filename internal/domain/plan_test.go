package domain

import (
	"errors"
	"testing"
	"time"
)

func TestMissionWindowValidation(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	valid := Mission{Timezone: "Asia/Shanghai", StartsAt: now, EndsAt: now.Add(time.Hour), Status: MissionDraft}
	if err := valid.ValidateWindow(now); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.EndsAt = invalid.StartsAt
	if err := invalid.ValidateWindow(now); !errors.Is(err, ErrConflict) {
		t.Fatalf("equal window error = %v", err)
	}
	late := valid
	late.Status = MissionScheduled
	if err := late.ValidateWindow(now.Add(2 * time.Hour)); !errors.Is(err, ErrExpired) {
		t.Fatalf("late window error = %v", err)
	}
}

func TestMissionCollectionWindow(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	p := Mission{Status: MissionActive, StartsAt: start, EndsAt: start.Add(time.Hour)}
	if !p.CanExecuteAt(start) {
		t.Fatal("start boundary should be included")
	}
	if p.CanExecuteAt(start.Add(time.Hour)) {
		t.Fatal("end boundary should be excluded")
	}
	if p.RemainingWindow(start.Add(30*time.Minute)) != 30*time.Minute {
		t.Fatal("remaining window mismatch")
	}
	if p.RemainingWindow(start.Add(2*time.Hour)) != 0 {
		t.Fatal("expired window should be zero")
	}
}

func TestDroneUnitValidation(t *testing.T) {
	drone := DroneUnit{Code: "S-1", RoomLabel: "A-101", RequiredTasks: 2}
	if err := drone.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []DroneUnit{{RoomLabel: "x", RequiredTasks: 1}, {Code: "x", RequiredTasks: 0}, {Code: "x", RoomLabel: "x", RequiredTasks: 2, Completed: 3}} {
		if err := invalid.Validate(); !errors.Is(err, ErrConflict) {
			t.Fatalf("invalid drone error = %v", err)
		}
	}
}
