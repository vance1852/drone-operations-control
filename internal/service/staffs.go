package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"github.com/google/uuid"
)

func (s *Service) RegisterFleetOperator(ctx context.Context, meta RequestMeta, name string, role domain.FleetOperatorRole) (domain.FleetOperator, error) {
	operator := domain.FleetOperator{ID: uuid.NewString(), Name: name, Role: role}
	if err := operator.Validate(); err != nil {
		return domain.FleetOperator{}, err
	}
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CreateFleetOperator(context.Context, domain.FleetOperator) error
		})
		if !ok {
			return fmt.Errorf("operator repository unavailable")
		}
		if err := repo.CreateFleetOperator(ctx, operator); err != nil {
			return fmt.Errorf("register operator: %w", err)
		}
		return tx.WriteAudit(ctx, audit(meta, "operator", operator.ID, "create", "success", nil))
	})
	if err != nil {
		return domain.FleetOperator{}, err
	}
	return operator, nil
}

func (s *Service) LoadFleetOperator(ctx context.Context, id string) (domain.FleetOperator, error) {
	if repo, ok := s.repo.(interface {
		GetFleetOperator(context.Context, string) (domain.FleetOperator, error)
	}); ok {
		return repo.GetFleetOperator(ctx, id)
	}
	return domain.FleetOperator{}, fmt.Errorf("operator repository unavailable")
}

var _ repository.Repository = (*repository.Postgres)(nil)
