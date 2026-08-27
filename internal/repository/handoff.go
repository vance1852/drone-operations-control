package repository

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
)

type HandoffRecord struct {
	ID          string
	DroneTaskID string
	From        *string
	To          string
	Location    string
	RecordedAt  time.Time
	Note        string
}

func (p *Postgres) ListHandoff(ctx context.Context, taskID string) ([]HandoffRecord, error) {
	rows, err := p.pool.Query(ctx, `SELECT id,drone_task_id,from_operator,to_operator,location,recorded_at,note FROM device_transfer_events WHERE drone_task_id=$1 ORDER BY recorded_at,id`, taskID)
	if err != nil {
		return nil, fmt.Errorf("list device_transfer: %w", err)
	}
	defer rows.Close()
	items := make([]HandoffRecord, 0)
	for rows.Next() {
		var item HandoffRecord
		if err := rows.Scan(&item.ID, &item.DroneTaskID, &item.From, &item.To, &item.Location, &item.RecordedAt, &item.Note); err != nil {
			return nil, fmt.Errorf("scan device_transfer: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ValidateHandoffSequence(items []HandoffRecord) error {
	for i := 1; i < len(items); i++ {
		if items[i].RecordedAt.Before(items[i-1].RecordedAt) {
			return fmt.Errorf("device_transfer sequence is not chronological: %w", domain.ErrConflict)
		}
		if items[i].From == nil || *items[i].From != items[i-1].To {
			return fmt.Errorf("device_transfer chain has a broken device_transfer: %w", domain.ErrConflict)
		}
	}
	return nil
}
