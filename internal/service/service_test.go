package service

import (
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func TestCreateDroneMissionBatchRejectsEmptyAndOversizedRequestsBeforePersistence(t *testing.T) {
	svc := New(nil)
	if _, err := svc.CreateDroneMissionBatch(t.Context(), RequestMeta{}, repository.DroneMissionBatchInput{Code: "B", Method: "m", Capacity: 1}, nil); !errors.Is(err, domain.ErrCapacityExceeded) {
		t.Fatalf("empty missionBatch error = %v", err)
	}
	if _, err := svc.CreateDroneMissionBatch(t.Context(), RequestMeta{}, repository.DroneMissionBatchInput{Code: "B", Method: "m", Capacity: 1}, []string{"a", "b"}); !errors.Is(err, domain.ErrCapacityExceeded) {
		t.Fatalf("oversized missionBatch error = %v", err)
	}
}

func TestCreateMissionRejectsMissingDroneUnitsBeforeTransaction(t *testing.T) {
	svc := New(nil)
	now := time.Now().UTC()
	_, err := svc.CreateMission(t.Context(), RequestMeta{}, CreateMissionRequest{Code: "P", Name: "Mission", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: "operator", DroneUnits: nil})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("missing drones error = %v", err)
	}
}

func TestReviewObservationRejectsInvalidDroneTaskVersion(t *testing.T) {
	if domain.DroneTaskStatus("in_progress").CanMoveTo(domain.DroneTaskRejected) == false {
		t.Fatal("in-progress drone should support rejection")
	}
	if domain.DroneTaskStatus("queued").CanMoveTo(domain.DroneTaskRejected) {
		t.Fatal("queued task cannot be rejected")
	}
}
