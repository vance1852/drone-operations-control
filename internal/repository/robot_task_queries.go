package repository

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
)

type DroneTaskCounts struct {
	Queued     int
	Completed  int
	Transit    int
	Accepted   int
	InProgress int
	Verified   int
	Rejected   int
	Archived   int
}

func (p *Postgres) DroneTaskCounts(ctx context.Context, missionID string) (DroneTaskCounts, error) {
	var counts DroneTaskCounts
	err := p.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='queued'),count(*) FILTER (WHERE status='completed'),count(*) FILTER (WHERE status='device_transfer_pending'),count(*) FILTER (WHERE status='accepted'),count(*) FILTER (WHERE status='in_progress'),count(*) FILTER (WHERE status='verified'),count(*) FILTER (WHERE status='rejected'),count(*) FILTER (WHERE status='archived') FROM drone_tasks WHERE mission_id=$1`, missionID).Scan(&counts.Queued, &counts.Completed, &counts.Transit, &counts.Accepted, &counts.InProgress, &counts.Verified, &counts.Rejected, &counts.Archived)
	if err != nil {
		return DroneTaskCounts{}, fmt.Errorf("task counts: %w", err)
	}
	return counts, nil
}

func (p *Postgres) ExpiringDroneTasks(ctx context.Context, before time.Time, limit int) ([]domain.DroneTask, error) {
	return p.expiringDroneTasks(ctx, "", before, limit)
}

func (p *Postgres) ExpiringDroneTasksForMission(ctx context.Context, missionID string, before time.Time, limit int) ([]domain.DroneTask, error) {
	return p.expiringDroneTasks(ctx, missionID, before, limit)
}

func (p *Postgres) expiringDroneTasks(ctx context.Context, missionID string, before time.Time, limit int) ([]domain.DroneTask, error) {
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	query := `SELECT id,mission_id,drone_id,task_code,status,completed_at,accepted_at,expires_at,version FROM drone_tasks WHERE status NOT IN ('verified','archived') AND expires_at <= $1`
	args := []any{before, limit}
	if missionID != "" {
		query += " AND mission_id=$3"
		args = append(args, missionID)
	}
	query += " ORDER BY expires_at LIMIT $2"
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("expiring drone_tasks: %w", err)
	}
	defer rows.Close()
	items := make([]domain.DroneTask, 0)
	for rows.Next() {
		var task domain.DroneTask
		if err := rows.Scan(&task.ID, &task.MissionID, &task.DroneUnitID, &task.TaskCode, &task.Status, &task.CompletedAt, &task.AcceptedAt, &task.ExpiresAt, &task.Version); err != nil {
			return nil, fmt.Errorf("scan expiring task: %w", err)
		}
		items = append(items, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read expiring drone_tasks: %w", err)
	}
	return items, nil
}
