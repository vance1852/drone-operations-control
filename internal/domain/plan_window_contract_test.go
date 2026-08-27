package domain

import (
	"testing"
	"time"
)

func TestCollectionRequiresAnOpenMissionWindow(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	mission := Mission{Status: MissionActive, StartsAt: now.Add(-2 * time.Hour), EndsAt: now.Add(-time.Hour)}
	if MissionExecutionAllowed(mission, now) {
		t.Fatal("collection allowed after mission window ended")
	}
}
