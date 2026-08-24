package repository

import (
	"context"
	"errors"
	"fmt"

	"drone-operations-control/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateFleetOperator(ctx context.Context, operator domain.FleetOperator) error {
	return createFleetOperator(ctx, p.pool, operator)
}
func (t *transaction) CreateFleetOperator(ctx context.Context, operator domain.FleetOperator) error {
	return createFleetOperator(ctx, t.tx, operator)
}

func (p *Postgres) GetFleetOperator(ctx context.Context, id string) (domain.FleetOperator, error) {
	return getFleetOperator(ctx, p.pool, id)
}
func (t *transaction) GetFleetOperator(ctx context.Context, id string) (domain.FleetOperator, error) {
	return getFleetOperator(ctx, t.tx, id)
}

func createFleetOperator(ctx context.Context, q sqler, operator domain.FleetOperator) error {
	if err := operator.Validate(); err != nil {
		return err
	}
	_, err := q.Exec(ctx, `INSERT INTO operators(id,name,role) VALUES ($1,$2,$3)`, operator.ID, operator.Name, operator.Role)
	return wrapWrite(err)
}

func getFleetOperator(ctx context.Context, q sqler, id string) (domain.FleetOperator, error) {
	var operator domain.FleetOperator
	err := q.QueryRow(ctx, `SELECT id,name,role FROM operators WHERE id=$1`, id).Scan(&operator.ID, &operator.Name, &operator.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FleetOperator{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.FleetOperator{}, fmt.Errorf("get operator: %w", err)
	}
	return operator, nil
}
