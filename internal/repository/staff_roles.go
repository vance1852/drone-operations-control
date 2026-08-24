package repository

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (p *Postgres) ChangeFleetOperatorRole(ctx context.Context, id string, role domain.FleetOperatorRole) error {
	if err := (domain.FleetOperator{ID: id, Name: "valid", Role: role}).Validate(); err != nil {
		return err
	}
	result, err := p.pool.Exec(ctx, `UPDATE operators SET role=$1 WHERE id=$2`, role, id)
	if err != nil {
		return fmt.Errorf("change operator role: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrNotFound
	}
	return nil
}

func (p *Postgres) FleetOperatorsForRole(ctx context.Context, role domain.FleetOperatorRole) ([]domain.FleetOperator, error) {
	rows, err := p.pool.Query(ctx, `SELECT id,name,role FROM operators WHERE role=$1 ORDER BY name`, role)
	if err != nil {
		return nil, fmt.Errorf("operators for role: %w", err)
	}
	defer rows.Close()
	items := make([]domain.FleetOperator, 0)
	for rows.Next() {
		var item domain.FleetOperator
		if err := rows.Scan(&item.ID, &item.Name, &item.Role); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
