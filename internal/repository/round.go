package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (p *Postgres) StartDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, p.pool, id, domain.DroneMissionBatchRunning, version)
}
func (p *Postgres) CompleteDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, p.pool, id, domain.DroneMissionBatchCompleted, version)
}
func (p *Postgres) CancelDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, p.pool, id, domain.DroneMissionBatchCancelled, version)
}
func (t *transaction) StartDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, t.tx, id, domain.DroneMissionBatchRunning, version)
}
func (t *transaction) CompleteDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, t.tx, id, domain.DroneMissionBatchCompleted, version)
}
func (t *transaction) CancelDroneMissionBatch(ctx context.Context, id string, version int64) error {
	return changeDroneMissionBatch(ctx, t.tx, id, domain.DroneMissionBatchCancelled, version)
}

func (p *Postgres) RestoreCancelledDroneTasks(ctx context.Context, id string) error {
	return restoreCancelledDroneTasks(ctx, p.pool, id)
}

func (t *transaction) RestoreCancelledDroneTasks(ctx context.Context, id string) error {
	return restoreCancelledDroneTasks(ctx, t.tx, id)
}

func restoreCancelledDroneTasks(ctx context.Context, q sqler, id string) error {
	_, err := q.Exec(ctx, `UPDATE drone_tasks s SET status='accepted',version=s.version+1
		FROM mission_batch_tasks bs, mission_batches b
		WHERE bs.mission_batch_id=b.id AND b.id=$1 AND b.status='cancelled'
		AND s.id=bs.drone_task_id AND s.status='in_progress'`, id)
	if err != nil {
		return fmt.Errorf("restore cancelled missionBatch tasks: %w", err)
	}
	return nil
}

func changeDroneMissionBatch(ctx context.Context, q sqler, id string, status domain.DroneMissionBatchStatus, version int64) error {
	if status == domain.DroneMissionBatchCancelled {
		result, err := q.Exec(ctx, `UPDATE mission_batches SET status='cancelled',version=version+1,completed_at=now()
			WHERE id=$1 AND version=$2 AND status IN ('queued','running')
			AND NOT EXISTS (SELECT 1 FROM telemetry_events WHERE mission_batch_id=$1)`, id, version)
		if err != nil {
			return fmt.Errorf("cancel missionBatch: %w", err)
		}
		if result.RowsAffected() != 1 {
			return domain.ErrConflict
		}
		return nil
	}
	allowedFrom := []string{}
	switch status {
	case domain.DroneMissionBatchRunning:
		allowedFrom = []string{string(domain.DroneMissionBatchQueued)}
	case domain.DroneMissionBatchCompleted:
		allowedFrom = []string{string(domain.DroneMissionBatchRunning)}
	default:
		return domain.ErrInvalidTransition
	}
	result, err := q.Exec(ctx, `UPDATE mission_batches SET status=$1,version=version+1,started_at=CASE WHEN $1='running' THEN now() ELSE started_at END,completed_at=CASE WHEN $1='completed' THEN now() ELSE completed_at END
		WHERE id=$2 AND version=$3 AND status=ANY($4)
		AND ($1 <> 'completed' OR NOT EXISTS (
			SELECT 1 FROM mission_batch_tasks bs
			LEFT JOIN telemetry_events r ON r.mission_batch_id=bs.mission_batch_id AND r.drone_task_id=bs.drone_task_id
			WHERE bs.mission_batch_id=$2 AND (r.id IS NULL OR r.status='pending')
		))`, status, id, version, allowedFrom)
	if err != nil {
		return fmt.Errorf("change missionBatch: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
