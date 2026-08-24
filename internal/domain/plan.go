package domain

import (
	"fmt"
	"strings"
	"time"
)

type DroneUnit struct {
	ID            string `json:"id"`
	MissionID     string `json:"mission_id"`
	Code          string `json:"code"`
	RoomLabel     string `json:"room_label"`
	RequiredTasks int    `json:"required_tasks"`
	Completed     int    `json:"completed_tasks"`
}

func (s DroneUnit) Validate() error {
	if strings.TrimSpace(s.Code) == "" || strings.TrimSpace(s.RoomLabel) == "" {
		return fmt.Errorf("drone code and room_label are required: %w", ErrConflict)
	}
	if s.RequiredTasks < 1 {
		return fmt.Errorf("drone requires at least one task: %w", ErrConflict)
	}
	if s.Completed < 0 || s.Completed > s.RequiredTasks {
		return fmt.Errorf("drone completed task count is invalid: %w", ErrConflict)
	}
	return nil
}

func MissionExecutionAllowed(mission Mission, now time.Time) bool {
	if mission.Status != MissionActive {
		return false
	}
	if now.Before(mission.StartsAt) {
		return false
	}
	return now.Before(mission.EndsAt)
}

func (p Mission) CanExecuteAt(now time.Time) bool {
	return p.Status == MissionActive && !now.Before(p.StartsAt) && now.Before(p.EndsAt)
}

func (p Mission) RemainingWindow(now time.Time) time.Duration {
	if now.After(p.EndsAt) {
		return 0
	}
	return p.EndsAt.Sub(now)
}

func (p Mission) Summary() map[string]any {
	return map[string]any{"id": p.ID, "code": p.Code, "status": p.Status, "timezone": p.Timezone, "starts_at": p.StartsAt, "ends_at": p.EndsAt, "version": p.Version}
}
