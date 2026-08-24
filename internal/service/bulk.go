package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) CreateDroneTasksBulk(ctx context.Context, meta RequestMeta, requests []domain.DroneTaskRequest) ([]domain.BulkItemResult, error) {
	now := s.now()
	if len(requests) == 0 {
		return nil, fmt.Errorf("at least one task is required: %w", domain.ErrConflict)
	}
	if err := domain.ValidateBulkRequests(requests, now); err != nil {
		return nil, err
	}
	inputs := make([]repository.DroneTaskInput, len(requests))
	for i, request := range requests {
		inputs[i] = repository.DroneTaskInput{MissionID: request.MissionID, DroneUnitID: request.DroneUnitID, TaskCode: request.TaskCode, ExpiresAt: request.ExpiresAt}
	}
	var drone_tasks []domain.DroneTask
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		for _, input := range inputs {
			placement, ok := tx.(interface {
				ValidateMissionDroneUnit(context.Context, string, string) error
			})
			if !ok {
				return fmt.Errorf("task placement repository unavailable")
			}
			if err := placement.ValidateMissionDroneUnit(ctx, input.MissionID, input.DroneUnitID); err != nil {
				return err
			}
			task, err := tx.CreateDroneTask(ctx, input)
			if err != nil {
				return err
			}
			drone_tasks = append(drone_tasks, task)
		}
		return tx.WriteAudit(ctx, audit(meta, "task_missionBatch", requests[0].MissionID, "create_bulk", "success", nil))
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.BulkItemResult, len(drone_tasks))
	for i, task := range drone_tasks {
		result[i] = domain.BulkItemResult{Index: i, TaskCode: task.TaskCode, DroneTaskID: task.ID}
	}
	return result, nil
}

func (s *Service) ValidateBulkForDroneUnit(requests []domain.DroneTaskRequest, droneID string) error {
	if err := domain.ValidateBulkRequests(requests, s.now()); err != nil {
		return err
	}
	for _, request := range requests {
		if request.DroneUnitID != droneID {
			return fmt.Errorf("task drone differs from missionBatch drone: %w", domain.ErrConflict)
		}
	}
	return nil
}
