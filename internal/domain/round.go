package domain

import (
	"fmt"
	"strings"
)

type DroneMissionBatch struct {
	ID         string
	Code       string
	Status     DroneMissionBatchStatus
	Method     string
	Capacity   int
	DroneTasks []string
	Version    int64
}

func (b DroneMissionBatch) Validate() error {
	if strings.TrimSpace(b.Code) == "" || strings.TrimSpace(b.Method) == "" {
		return fmt.Errorf("drone round code and method are required: %w", ErrConflict)
	}
	if b.Capacity < 1 {
		return fmt.Errorf("drone round capacity must be positive: %w", ErrConflict)
	}
	if len(b.DroneTasks) > b.Capacity {
		return ErrCapacityExceeded
	}
	return nil
}

func (b DroneMissionBatch) AddDroneTasks(ids []string) (DroneMissionBatch, error) {
	seen := make(map[string]struct{}, len(b.DroneTasks))
	for _, id := range b.DroneTasks {
		seen[id] = struct{}{}
	}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return DroneMissionBatch{}, fmt.Errorf("task id is empty: %w", ErrConflict)
		}
		if _, exists := seen[id]; exists {
			return DroneMissionBatch{}, fmt.Errorf("duplicate task in drone round: %w", ErrConflict)
		}
		seen[id] = struct{}{}
	}
	if len(b.DroneTasks)+len(ids) > b.Capacity {
		return DroneMissionBatch{}, ErrCapacityExceeded
	}
	b.DroneTasks = append(append([]string(nil), b.DroneTasks...), ids...)
	return b, nil
}
