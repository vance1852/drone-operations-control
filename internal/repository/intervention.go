package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CloseIntervention(ctx context.Context, id string, now time.Time) error {
	return closeIntervention(ctx, p.pool, id, now)
}
func (t *transaction) CloseIntervention(ctx context.Context, id string, now time.Time) error {
	return closeIntervention(ctx, t.tx, id, now)
}

func (p *Postgres) GetIntervention(ctx context.Context, id string) (domain.Intervention, error) {
	var d domain.Intervention
	err := p.pool.QueryRow(ctx, `SELECT id,drone_task_id,kind,status,reason,due_at,closed_at FROM safety_alerts WHERE id=$1`, id).Scan(&d.ID, &d.DroneTaskID, &d.Kind, &d.Status, &d.Reason, &d.DueAt, &d.ClosedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Intervention{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Intervention{}, fmt.Errorf("get safety_alert: %w", err)
	}
	return d, nil
}

func closeIntervention(ctx context.Context, q sqler, id string, now time.Time) error {
	result, err := q.Exec(ctx, `UPDATE safety_alerts SET status='closed',closed_at=$1 WHERE id=$2 AND status IN ('open','in_progress')`, now, id)
	if err != nil {
		return fmt.Errorf("close safety_alert: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (p *Postgres) MarkInterventionInProgress(ctx context.Context, id string) error {
	return markInterventionInProgress(ctx, p.pool, id)
}

func (t *transaction) MarkInterventionInProgress(ctx context.Context, id string) error {
	return markInterventionInProgress(ctx, t.tx, id)
}

func markInterventionInProgress(ctx context.Context, q sqler, id string) error {
	result, err := q.Exec(ctx, `UPDATE safety_alerts SET status='in_progress' WHERE id=$1 AND status='open'`, id)
	if err != nil {
		return fmt.Errorf("mark safety_alert in progress: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrConflict
	}
	return nil
}
