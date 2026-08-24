package repository

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
)

func (p *Postgres) OpenInterventionCount(ctx context.Context, missionID string) (int, error) {
	var count int
	if err := p.pool.QueryRow(ctx, `SELECT count(*) FROM safety_alerts d JOIN drone_tasks s ON s.id=d.drone_task_id WHERE s.mission_id=$1 AND d.status IN ('open','in_progress')`, missionID).Scan(&count); err != nil {
		return 0, fmt.Errorf("open safety_alert count: %w", err)
	}
	return count, nil
}

func (p *Postgres) ComplianceReport(ctx context.Context, missionID string, now time.Time) (domain.ComplianceReport, error) {
	progress, err := p.MissionProgress(ctx, missionID)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	expiring, err := p.ExpiringDroneTasksForMission(ctx, missionID, now.Add(48*time.Hour), 100)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	count, err := p.OpenInterventionCount(ctx, missionID)
	if err != nil {
		return domain.ComplianceReport{}, err
	}
	return domain.ComplianceReport{MissionID: missionID, GeneratedAt: now.UTC(), Progress: progress, Expiring: expiring, OpenInterventions: count}, nil
}
