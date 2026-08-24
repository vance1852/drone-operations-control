package domain

import (
	"errors"
	"testing"
)

func TestDroneMissionBatchAddsDroneTasksWithoutSharingSlice(t *testing.T) {
	b := DroneMissionBatch{Code: "B-1", Method: "water-ph", Capacity: 3, DroneTasks: []string{"s1"}}
	updated, err := b.AddDroneTasks([]string{"s2"})
	if err != nil {
		t.Fatal(err)
	}
	updated.DroneTasks[0] = "changed"
	if b.DroneTasks[0] != "s1" {
		t.Fatal("missionBatch input slice was polluted")
	}
	if len(updated.DroneTasks) != 2 {
		t.Fatalf("drone_tasks = %v", updated.DroneTasks)
	}
}

func TestDroneMissionBatchRejectsDuplicateAndCapacityOverflow(t *testing.T) {
	b := DroneMissionBatch{Code: "B-1", Method: "water-ph", Capacity: 2, DroneTasks: []string{"s1"}}
	if _, err := b.AddDroneTasks([]string{"s1"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate error = %v", err)
	}
	if _, err := b.AddDroneTasks([]string{"s2", "s3"}); !errors.Is(err, ErrCapacityExceeded) {
		t.Fatalf("capacity error = %v", err)
	}
}

func TestDroneMissionBatchValidation(t *testing.T) {
	if err := (DroneMissionBatch{Code: "B", Method: "m", Capacity: 1}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (DroneMissionBatch{Code: "B", Method: "m", Capacity: 0}).Validate(); !errors.Is(err, ErrConflict) {
		t.Fatalf("invalid capacity error = %v", err)
	}
}
