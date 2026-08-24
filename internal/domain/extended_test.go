package domain

import (
	"errors"
	"testing"
	"time"
)

func TestFleetOperatorRolesAndPermissions(t *testing.T) {
	cases := []struct {
		role       FleetOperatorRole
		permission Permission
		want       bool
	}{
		{RoleDroneOperator, PermissionDroneTaskComplete, true},
		{RoleDroneOperator, PermissionObservationReview, false},
		{RoleTelemetryOperator, PermissionObservationRecord, true},
		{RoleQualityReviewer, PermissionObservationReview, true},
		{RoleSafetySupervisor, PermissionInterventionClose, true},
	}
	for _, tc := range cases {
		operator := FleetOperator{ID: "op", Name: "FleetOperator", Role: tc.role}
		if got := operator.Has(tc.permission); got != tc.want {
			t.Errorf("role=%s permission=%s got=%v want=%v", tc.role, tc.permission, got, tc.want)
		}
	}
}

func TestFleetOperatorValidationRejectsUnknownRole(t *testing.T) {
	if err := (FleetOperator{ID: "op", Name: "FleetOperator", Role: "unknown"}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestAssignmentLifecycle(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	a := Assignment{ID: "a", MissionID: "p", DroneUnitID: "s", FleetOperatorID: "o", StartsAt: start, EndsAt: start.Add(time.Hour), Status: "queued"}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	if !a.CanMoveTo("active") || a.CanMoveTo("completed") {
		t.Fatal("assignment transition graph is wrong")
	}
	if a.ActiveAt(start) {
		t.Fatal("queued assignment is not active")
	}
	a.Status = "active"
	if !a.ActiveAt(start.Add(time.Minute)) {
		t.Fatal("active assignment not active in window")
	}
	if a.ActiveAt(start.Add(time.Hour)) {
		t.Fatal("end boundary should be inactive")
	}
}

func TestDroneTaskFilterAndSearch(t *testing.T) {
	drone_tasks := []DroneTask{
		{ID: "2", MissionID: "p", TaskCode: "S-002", Status: DroneTaskAccepted, ExpiresAt: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)},
		{ID: "1", MissionID: "p", TaskCode: "S-001", Status: DroneTaskCompleted, ExpiresAt: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)},
		{ID: "3", MissionID: "q", TaskCode: "S-003", Status: DroneTaskAccepted, ExpiresAt: time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)},
	}
	request := SearchRequest{Filter: DroneTaskFilter{MissionID: " p ", Search: "002"}, Sort: SortExpiry, Limit: 10}
	items := SearchDroneTasks(drone_tasks, request)
	if len(items) != 1 || items[0].ID != "2" {
		t.Fatalf("search result = %+v", items)
	}
	if !(DroneTaskFilter{Status: DroneTaskAccepted}).Matches(drone_tasks[0]) {
		t.Fatal("status filter did not match")
	}
}

func TestBulkValidationDetectsDuplicateAndExpiredInput(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	valid := []DroneTaskRequest{{MissionID: "p", DroneUnitID: "s", TaskCode: "S-1", ExpiresAt: now.Add(time.Hour)}}
	if err := ValidateBulkRequests(valid, now); err != nil {
		t.Fatal(err)
	}
	duplicate := append(valid, valid[0])
	if err := ValidateBulkRequests(duplicate, now); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error = %v", err)
	}
	expired := []DroneTaskRequest{{MissionID: "p", DroneUnitID: "s", TaskCode: "S-2", ExpiresAt: now.Add(-time.Second)}}
	if err := ValidateBulkRequests(expired, now); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired error = %v", err)
	}
}

func TestStateMachinesReachExpectedStates(t *testing.T) {
	missionStates := DefaultMissionMachine().Reachable("draft")
	if len(missionStates) != 4 {
		t.Fatalf("mission states = %v", missionStates)
	}
	if err := DefaultDroneTaskMachine().ValidatePath([]string{"queued", "completed", "device_transfer_pending", "accepted", "in_progress", "rejected", "archived"}); err != nil {
		t.Fatal(err)
	}
	if err := DefaultDroneMissionBatchMachine().ValidatePath([]string{"queued", "completed"}); err == nil {
		t.Fatal("invalid missionBatch path accepted")
	}
}

func TestConstraintSetAndRedaction(t *testing.T) {
	if err := (ConstraintSet{MaxDroneTasksPerDroneMissionBatch: 2, MinimumRemainingTTL: time.Hour}).Validate(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	task := DroneTask{Status: DroneTaskAccepted, ExpiresAt: now.Add(2 * time.Hour)}
	if !(ConstraintSet{MaxDroneTasksPerDroneMissionBatch: 2, MinimumRemainingTTL: time.Hour}).AllowsDroneTask(task, now) {
		t.Fatal("valid task rejected")
	}
	if RedactTaskCode("S-1234") != "S-**34" {
		t.Fatalf("redaction mismatch")
	}
	if RedactRoomLabel("North Gate") != "N***e" {
		t.Fatalf("room_label redaction mismatch")
	}
}

func TestValidationHelpers(t *testing.T) {
	if err := ValidateBusinessCode("PLAN-001"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBusinessCode("bad code"); !errors.Is(err, ErrConflict) {
		t.Fatalf("code error = %v", err)
	}
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	if err := ValidateUTCWindow(start, start.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePositiveVersion(0); !errors.Is(err, ErrConflict) {
		t.Fatalf("version error = %v", err)
	}
	if err := ValidatePage(0, 101); !errors.Is(err, ErrConflict) {
		t.Fatalf("page error = %v", err)
	}
	if err := ValidateObservation(1, 2, "mg/L"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReason("ok"); !errors.Is(err, ErrConflict) {
		t.Fatalf("reason error = %v", err)
	}
}
