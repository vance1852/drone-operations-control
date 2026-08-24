package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSearchDescendingSortAndPagination(t *testing.T) {
	items := []DroneTask{{ID: "a", TaskCode: "A", Status: DroneTaskAccepted}, {ID: "c", TaskCode: "C", Status: DroneTaskAccepted}, {ID: "b", TaskCode: "B", Status: DroneTaskAccepted}}
	got := SearchDroneTasks(items, SearchRequest{Sort: SortCode, Desc: true, Limit: 2})
	if len(got) != 2 || got[0].ID != "c" || got[1].ID != "b" {
		t.Fatalf("got=%+v", got)
	}
}

func TestSameMissionRequiresAllDroneTasksToMatch(t *testing.T) {
	if !SameMission([]DroneTask{{MissionID: "p"}, {MissionID: "p"}}) {
		t.Fatal("same missions rejected")
	}
	if SameMission([]DroneTask{{MissionID: "p"}, {MissionID: "q"}}) {
		t.Fatal("different missions accepted")
	}
}

func TestInterventionRejectsTooOldDueDate(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	d := Intervention{DroneTaskID: "s", Kind: "reassess", Status: InterventionOpen, Reason: "bad", DueAt: now.Add(-25 * time.Hour)}
	if err := d.Validate(now); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestDroneMissionBatchRejectsBlankDroneTaskID(t *testing.T) {
	b := DroneMissionBatch{Code: "B", Method: "m", Capacity: 2}
	if _, err := b.AddDroneTasks([]string{""}); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestUTCWindowRejectsLocalTime(t *testing.T) {
	start := time.Date(2026, 8, 18, 9, 0, 0, 0, time.Local)
	end := start.Add(time.Hour)
	if err := ValidateUTCWindow(start, end); !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
}

func TestMissionSummaryContainsStableFields(t *testing.T) {
	p := Mission{ID: "p", Code: "P-1", Status: MissionDraft, Version: 3}
	summary := p.Summary()
	if summary["id"] != "p" || summary["version"] != int64(3) {
		t.Fatalf("summary=%+v", summary)
	}
}
