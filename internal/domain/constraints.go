package domain

import (
	"fmt"
	"time"
)

type ConstraintSet struct {
	MaxDroneTasksPerDroneMissionBatch int
	MinimumRemainingTTL               time.Duration
	RequireHandoffChain               bool
}

func (c ConstraintSet) Validate() error {
	if c.MaxDroneTasksPerDroneMissionBatch < 1 {
		return fmt.Errorf("max drone_tasks per missionBatch must be positive: %w", ErrConflict)
	}
	if c.MinimumRemainingTTL < 0 {
		return fmt.Errorf("minimum ttl cannot be negative: %w", ErrConflict)
	}
	return nil
}

func (c ConstraintSet) AllowsDroneTask(task DroneTask, now time.Time) bool {
	if task.Status != DroneTaskAccepted {
		return false
	}
	return task.ExpiresAt.Sub(now) >= c.MinimumRemainingTTL
}

func SameMission(drone_tasks []DroneTask) bool {
	if len(drone_tasks) < 2 {
		return true
	}
	mission := drone_tasks[0].MissionID
	for _, task := range drone_tasks[1:] {
		if task.MissionID != mission {
			return false
		}
	}
	return true
}
