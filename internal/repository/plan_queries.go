package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

type MissionFilter struct {
	Status domain.MissionStatus
	Search string
	Limit  int
	Offset int
}

func (p *Postgres) ListMissions(ctx context.Context, filter MissionFilter) ([]domain.Mission, int, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	where := "WHERE TRUE"
	args := []any{filter.Limit, filter.Offset}
	if filter.Status != "" {
		args = append(args, filter.Status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		where += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", len(args), len(args))
	}
	rows, err := p.pool.Query(ctx, fmt.Sprintf(`SELECT id,code,name,status,timezone,starts_at,ends_at,version,created_by FROM drone_missions %s ORDER BY starts_at DESC LIMIT $1 OFFSET $2`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list missions: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Mission, 0)
	for rows.Next() {
		var item domain.Mission
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Status, &item.Timezone, &item.StartsAt, &item.EndsAt, &item.Version, &item.CreatedBy); err != nil {
			return nil, 0, fmt.Errorf("scan mission: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	countArgs := args[2:]
	countWhere := "WHERE TRUE"
	if filter.Status != "" {
		countWhere += " AND status=$1"
	}
	if filter.Search != "" {
		index := 1
		if filter.Status != "" {
			index = 2
		}
		countWhere += fmt.Sprintf(" AND (code ILIKE $%d OR name ILIKE $%d)", index, index)
	}
	var total int
	if err := p.pool.QueryRow(ctx, "SELECT count(*) FROM drone_missions "+countWhere, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count missions: %w", err)
	}
	return items, total, nil
}

func (p *Postgres) ListMissionDroneUnits(ctx context.Context, missionID string) ([]domain.DroneUnit, error) {
	rows, err := p.pool.Query(ctx, `SELECT id,mission_id,code,room_label,required_tasks,completed_tasks FROM mission_drones WHERE mission_id=$1 ORDER BY code`, missionID)
	if err != nil {
		return nil, fmt.Errorf("list mission drones: %w", err)
	}
	defer rows.Close()
	items := make([]domain.DroneUnit, 0)
	for rows.Next() {
		var item domain.DroneUnit
		if err := rows.Scan(&item.ID, &item.MissionID, &item.Code, &item.RoomLabel, &item.RequiredTasks, &item.Completed); err != nil {
			return nil, fmt.Errorf("scan drone: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
