package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) ValidateMissionDroneUnit(ctx context.Context, missionID, droneID string) error {
	return validateMissionDroneUnit(ctx, p.pool, missionID, droneID)
}

func (t *transaction) ValidateMissionDroneUnit(ctx context.Context, missionID, droneID string) error {
	return validateMissionDroneUnit(ctx, t.tx, missionID, droneID)
}

func validateMissionDroneUnit(ctx context.Context, q sqler, missionID, droneID string) error {
	var exists bool
	if err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM mission_drones WHERE id=$1 AND mission_id=$2)`, droneID, missionID).Scan(&exists); err != nil {
		return fmt.Errorf("validate mission drone: %w", err)
	}
	if !exists {
		return fmt.Errorf("drone does not belong to mission: %w", domain.ErrConflict)
	}
	return nil
}

func (p *Postgres) ValidateObservationTarget(ctx context.Context, taskID, missionBatchID string) error {
	return validateObservationTarget(ctx, p.pool, taskID, missionBatchID)
}

func (t *transaction) ValidateObservationTarget(ctx context.Context, taskID, missionBatchID string) error {
	return validateObservationTarget(ctx, t.tx, taskID, missionBatchID)
}

func validateObservationTarget(ctx context.Context, q sqler, taskID, missionBatchID string) error {
	var taskStatus, missionBatchStatus string
	err := q.QueryRow(ctx, `SELECT s.status,b.status FROM mission_batch_tasks bs JOIN drone_tasks s ON s.id=bs.drone_task_id JOIN mission_batches b ON b.id=bs.mission_batch_id WHERE bs.drone_task_id=$1 AND bs.mission_batch_id=$2`, taskID, missionBatchID).Scan(&taskStatus, &missionBatchStatus)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("task is not attached to missionBatch: %w", domain.ErrConflict)
	}
	if err != nil {
		return fmt.Errorf("validate telemetry target: %w", err)
	}
	if taskStatus != string(domain.DroneTaskInProgress) || missionBatchStatus != string(domain.DroneMissionBatchRunning) {
		return fmt.Errorf("task and drone round are not ready for an telemetry: %w", domain.ErrInvalidTransition)
	}
	return nil
}

func (p *Postgres) ObservationTaskID(ctx context.Context, telemetryID string) (string, error) {
	return telemetryDroneTaskID(ctx, p.pool, telemetryID)
}

func (t *transaction) ObservationTaskID(ctx context.Context, telemetryID string) (string, error) {
	return telemetryDroneTaskID(ctx, t.tx, telemetryID)
}

func telemetryDroneTaskID(ctx context.Context, q sqler, telemetryID string) (string, error) {
	var taskID string
	if err := q.QueryRow(ctx, `SELECT drone_task_id FROM telemetry_events WHERE id=$1`, telemetryID).Scan(&taskID); err == pgx.ErrNoRows {
		return "", domain.ErrNotFound
	} else if err != nil {
		return "", fmt.Errorf("get telemetry task: %w", err)
	}
	return taskID, nil
}
