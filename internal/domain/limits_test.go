package domain

import (
	"errors"
	"testing"
)

func TestEnsureLimit(t *testing.T) {
	if err := EnsureLimit(2, 3, "drone_tasks"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLimit(4, 3, "drone_tasks"); !errors.Is(err, ErrCapacityExceeded) {
		t.Fatalf("error=%v", err)
	}
	if err := EnsureLimit(-1, 3, "drone_tasks"); !errors.Is(err, ErrCapacityExceeded) {
		t.Fatalf("negative error=%v", err)
	}
}

func TestProgressRemainingNeverNegative(t *testing.T) {
	if (MissionProgress{Required: 1, Completed: 3}).Remaining() != 0 {
		t.Fatal("remaining went negative")
	}
}
