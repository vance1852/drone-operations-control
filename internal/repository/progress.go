package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (p *Postgres) MissionProgress(ctx context.Context, missionID string) (domain.MissionProgress, error) {
	return missionProgress(ctx, p.pool, missionID)
}

func missionProgress(ctx context.Context, q sqler, missionID string) (domain.MissionProgress, error) {
	var progress domain.MissionProgress
	progress.MissionID = missionID
	err := q.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM mission_drones WHERE mission_id=$1),
		COALESCE((SELECT sum(required_tasks) FROM mission_drones WHERE mission_id=$1),0),
		count(*) FILTER (WHERE status='completed'),
		count(*) FILTER (WHERE status='accepted'),
		count(*) FILTER (WHERE status='in_progress'),
		count(*) FILTER (WHERE status='verified'),
		count(*) FILTER (WHERE status='rejected'),
		count(*) FILTER (WHERE status='archived')
		FROM drone_tasks WHERE mission_id=$1`, missionID).Scan(&progress.DroneUnits, &progress.Required, &progress.Completed, &progress.Accepted, &progress.InProgress, &progress.Verified, &progress.Rejected, &progress.Archived)
	if err != nil {
		return domain.MissionProgress{}, fmt.Errorf("mission progress: %w", err)
	}
	return progress, nil
}
