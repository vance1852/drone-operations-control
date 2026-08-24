package domain

import (
	"fmt"
	"strings"
	"time"
)

type Handoff struct {
	DroneTaskID string
	From        string
	To          string
	Location    string
	RecordedAt  time.Time
}

func (c Handoff) Validate() error {
	if strings.TrimSpace(c.DroneTaskID) == "" || strings.TrimSpace(c.To) == "" || strings.TrimSpace(c.Location) == "" {
		return fmt.Errorf("device_transfer task, receiver and location are required: %w", ErrConflict)
	}
	if !c.RecordedAt.IsZero() && c.RecordedAt.After(time.Now().UTC().Add(5*time.Minute)) {
		return fmt.Errorf("device_transfer timestamp is in the future: %w", ErrConflict)
	}
	return nil
}

func (s DroneTask) CanBePerformed(now time.Time) error {
	if s.Status != DroneTaskAccepted {
		return fmt.Errorf("task is not accepted: %w", ErrInvalidTransition)
	}
	if now.After(s.ExpiresAt) {
		return ErrExpired
	}
	return nil
}

func (s DroneTask) CanBeArchived() bool {
	return s.Status == DroneTaskVerified || s.Status == DroneTaskRejected
}

func ObservationDecision(riskScore, alertThreshold float64) (ObservationStatus, error) {
	if alertThreshold < 0 {
		return "", fmt.Errorf("alert threshold cannot be negative: %w", ErrConflict)
	}
	if riskScore > alertThreshold {
		return ObservationRejected, nil
	}
	return ObservationVerified, nil
}
