package service

import (
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func TestAuthorizeRequiresRepositoryAndPermission(t *testing.T) {
	svc := New(nil)
	if err := svc.Authorize(t.Context(), "operator", "complete"); err == nil {
		t.Fatal("authorization succeeded without repository")
	}
}

func TestOpenInterventionRejectsOldDueTime(t *testing.T) {
	svc := New(nil).WithClock(func() time.Time { return time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC) })
	_, err := svc.OpenIntervention(t.Context(), RequestMeta{}, repository.InterventionInput{DroneTaskID: "s1", Kind: "reassess", Reason: "bad", DueAt: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestValidateBulkForDroneUnitRejectsMixedDroneUnits(t *testing.T) {
	svc := New(nil).WithClock(time.Now)
	requests := []domain.DroneTaskRequest{{MissionID: "p", DroneUnitID: "s1", TaskCode: "S-1", ExpiresAt: time.Now().Add(time.Hour)}, {MissionID: "p", DroneUnitID: "s2", TaskCode: "S-2", ExpiresAt: time.Now().Add(time.Hour)}}
	if err := svc.ValidateBulkForDroneUnit(requests, "s1"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}
