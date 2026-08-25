package domain

import (
	"testing"
	"time"
)

func TestSearchDroneTasksUnknownSortUsesSafeDefault(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unknown sort panicked: %v", recovered)
		}
	}()
	items := []DroneTask{
		{ID: "task-b", MissionID: "mission-1", TaskCode: "B", ExpiresAt: time.Unix(20, 0)},
		{ID: "task-a", MissionID: "mission-1", TaskCode: "A", ExpiresAt: time.Unix(10, 0)},
	}
	got := SearchDroneTasks(items, SearchRequest{Sort: SortField("battery_level"), Limit: 10})
	if len(got) != 2 || got[0].ID != "task-a" || got[1].ID != "task-b" {
		t.Fatalf("unknown sort fallback result=%+v", got)
	}
}
