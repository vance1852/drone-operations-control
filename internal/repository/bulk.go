package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (p *Postgres) CreateDroneTasksBulk(ctx context.Context, inputs []DroneTaskInput) ([]domain.DroneTask, error) {
	var drone_tasks []domain.DroneTask
	err := p.InTx(ctx, func(tx Repository) error {
		for _, input := range inputs {
			task, err := tx.CreateDroneTask(ctx, input)
			if err != nil {
				return fmt.Errorf("create task %s: %w", input.TaskCode, err)
			}
			drone_tasks = append(drone_tasks, task)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return append([]domain.DroneTask(nil), drone_tasks...), nil
}

func ValidateBulkCapacity(inputs []DroneTaskInput, capacity int) error {
	if capacity < 1 || len(inputs) > capacity {
		return domain.ErrCapacityExceeded
	}
	return nil
}
