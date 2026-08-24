package domain

import (
	"fmt"
	"strings"
	"time"
)

type DroneTaskRequest struct {
	MissionID   string
	DroneUnitID string
	TaskCode    string
	ExpiresAt   time.Time
}

type BulkItemResult struct {
	Index       int    `json:"index"`
	TaskCode    string `json:"task_code"`
	DroneTaskID string `json:"drone_task_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (r DroneTaskRequest) Validate(now time.Time) error {
	if strings.TrimSpace(r.MissionID) == "" || strings.TrimSpace(r.DroneUnitID) == "" || strings.TrimSpace(r.TaskCode) == "" {
		return fmt.Errorf("task mission, drone and code are required: %w", ErrConflict)
	}
	if r.ExpiresAt.Before(now) {
		return ErrExpired
	}
	return nil
}

func ValidateBulkRequests(requests []DroneTaskRequest, now time.Time) error {
	seen := make(map[string]struct{}, len(requests))
	for _, request := range requests {
		if err := request.Validate(now); err != nil {
			return err
		}
		if _, ok := seen[request.TaskCode]; ok {
			return fmt.Errorf("duplicate external code %s: %w", request.TaskCode, ErrConflict)
		}
		seen[request.TaskCode] = struct{}{}
	}
	return nil
}
