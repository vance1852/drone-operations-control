package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) ListFleetOperators(ctx context.Context, role domain.FleetOperatorRole, limit, offset int) ([]domain.FleetOperator, int, error) {
	repo, ok := s.repo.(interface {
		ListFleetOperators(context.Context, domain.FleetOperatorRole, int, int) ([]domain.FleetOperator, int, error)
	})
	if !ok {
		return nil, 0, fmt.Errorf("operator query repository unavailable")
	}
	items, total, err := repo.ListFleetOperators(ctx, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return append([]domain.FleetOperator(nil), items...), total, nil
}

func (s *Service) RenameFleetOperator(ctx context.Context, meta RequestMeta, id, name string) error {
	if err := validateCode(name, "operator name"); err != nil {
		return err
	}
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			RenameFleetOperator(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("operator mutation repository unavailable")
		}
		if err := repo.RenameFleetOperator(ctx, id, name); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "operator", id, "rename", "success", nil))
	})
}
