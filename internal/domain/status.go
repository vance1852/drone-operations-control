package domain

import (
	"fmt"
	"time"
)

type MissionStatus string

const (
	MissionDraft     MissionStatus = "draft"
	MissionScheduled MissionStatus = "scheduled"
	MissionActive    MissionStatus = "active"
	MissionClosed    MissionStatus = "closed"
)

func (s MissionStatus) CanMoveTo(next MissionStatus) bool {
	switch s {
	case MissionDraft:
		return next == MissionScheduled
	case MissionScheduled:
		return next == MissionActive || next == MissionClosed
	case MissionActive:
		return next == MissionClosed
	default:
		return false
	}
}

type DroneTaskStatus string

const (
	DroneTaskQueued         DroneTaskStatus = "queued"
	DroneTaskCompleted      DroneTaskStatus = "completed"
	DroneTaskHandoffPending DroneTaskStatus = "device_transfer_pending"
	DroneTaskAccepted       DroneTaskStatus = "accepted"
	DroneTaskInProgress     DroneTaskStatus = "in_progress"
	DroneTaskVerified       DroneTaskStatus = "verified"
	DroneTaskRejected       DroneTaskStatus = "rejected"
	DroneTaskArchived       DroneTaskStatus = "archived"
)

func (s DroneTaskStatus) CanMoveTo(next DroneTaskStatus) bool {
	switch s {
	case DroneTaskQueued:
		return next == DroneTaskCompleted
	case DroneTaskCompleted:
		return next == DroneTaskHandoffPending
	case DroneTaskHandoffPending:
		return next == DroneTaskAccepted
	case DroneTaskAccepted:
		return next == DroneTaskInProgress
	case DroneTaskInProgress:
		return next == DroneTaskVerified || next == DroneTaskRejected
	case DroneTaskRejected:
		return next == DroneTaskArchived
	case DroneTaskVerified:
		return next == DroneTaskArchived
	default:
		return false
	}
}

type Mission struct {
	ID        string        `json:"id"`
	Code      string        `json:"code"`
	Name      string        `json:"name"`
	Status    MissionStatus `json:"status"`
	Timezone  string        `json:"timezone"`
	StartsAt  time.Time     `json:"starts_at"`
	EndsAt    time.Time     `json:"ends_at"`
	Version   int64         `json:"version"`
	CreatedBy string        `json:"created_by"`
}

func (p Mission) ValidateWindow(now time.Time) error {
	if p.EndsAt.Before(p.StartsAt) || p.EndsAt.Equal(p.StartsAt) {
		return fmt.Errorf("mission end must be after start: %w", ErrConflict)
	}
	if p.Timezone == "" {
		return fmt.Errorf("timezone is required: %w", ErrConflict)
	}
	if now.After(p.EndsAt) && p.Status != MissionClosed {
		return fmt.Errorf("mission window has elapsed: %w", ErrExpired)
	}
	return nil
}

type DroneTask struct {
	ID          string          `json:"id"`
	MissionID   string          `json:"mission_id"`
	DroneUnitID string          `json:"drone_id"`
	TaskCode    string          `json:"task_code"`
	Status      DroneTaskStatus `json:"status"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	AcceptedAt  *time.Time      `json:"accepted_at,omitempty"`
	ExpiresAt   time.Time       `json:"expires_at"`
	Version     int64           `json:"version"`
}

func EligibleForExecution(status DroneTaskStatus, expiresAt, now time.Time) bool {
	if status != DroneTaskAccepted {
		return false
	}
	if expiresAt.Before(now) {
		return false
	}
	return true
}

func (s DroneTask) Move(next DroneTaskStatus, now time.Time) (DroneTask, error) {
	if !s.Status.CanMoveTo(next) {
		return DroneTask{}, fmt.Errorf("%s -> %s: %w", s.Status, next, ErrInvalidTransition)
	}
	if now.After(s.ExpiresAt) && next != DroneTaskArchived {
		return DroneTask{}, fmt.Errorf("task %s expired: %w", s.TaskCode, ErrExpired)
	}
	s.Status = next
	s.Version++
	if next == DroneTaskCompleted {
		s.CompletedAt = &now
	}
	if next == DroneTaskAccepted {
		s.AcceptedAt = &now
	}
	return s, nil
}

type DroneMissionBatchStatus string

const (
	DroneMissionBatchQueued    DroneMissionBatchStatus = "queued"
	DroneMissionBatchRunning   DroneMissionBatchStatus = "running"
	DroneMissionBatchCompleted DroneMissionBatchStatus = "completed"
	DroneMissionBatchCancelled DroneMissionBatchStatus = "cancelled"
)

func (s DroneMissionBatchStatus) CanMoveTo(next DroneMissionBatchStatus) bool {
	switch s {
	case DroneMissionBatchQueued:
		return next == DroneMissionBatchRunning || next == DroneMissionBatchCancelled
	case DroneMissionBatchRunning:
		return next == DroneMissionBatchCompleted || next == DroneMissionBatchCancelled
	default:
		return false
	}
}

type ObservationStatus string

const (
	ObservationPending  ObservationStatus = "pending"
	ObservationVerified ObservationStatus = "verified"
	ObservationRejected ObservationStatus = "rejected"
)

func (r ObservationStatus) Outcome(riskScore, alertThreshold float64) ObservationStatus {
	if riskScore > alertThreshold {
		return ObservationRejected
	}
	return ObservationVerified
}
