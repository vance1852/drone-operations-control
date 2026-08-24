package domain

import (
	"fmt"
	"strings"
	"time"
)

type InterventionStatus string

const (
	InterventionOpen       InterventionStatus = "open"
	InterventionInProgress InterventionStatus = "in_progress"
	InterventionClosed     InterventionStatus = "closed"
)

type Intervention struct {
	ID          string
	DroneTaskID string
	Kind        string
	Status      InterventionStatus
	Reason      string
	DueAt       time.Time
	ClosedAt    *time.Time
}

func (d Intervention) Validate(now time.Time) error {
	if d.DroneTaskID == "" || strings.TrimSpace(d.Kind) == "" || strings.TrimSpace(d.Reason) == "" {
		return fmt.Errorf("safety_alert fields are required: %w", ErrConflict)
	}
	switch d.Kind {
	case "reassess", "repeat_drone", "safety_adjustment", "close_record":
	default:
		return fmt.Errorf("safety_alert kind is invalid: %w", ErrConflict)
	}
	if d.DueAt.Before(now.Add(-24 * time.Hour)) {
		return fmt.Errorf("safety_alert due time is too old: %w", ErrConflict)
	}
	if d.Status == InterventionClosed && d.ClosedAt == nil {
		return fmt.Errorf("closed safety_alert needs closed_at: %w", ErrConflict)
	}
	return nil
}

func (d Intervention) IsDue(now time.Time) bool {
	return d.Status != InterventionClosed && !d.DueAt.After(now)
}
