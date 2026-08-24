package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) AssignDroneUnit(ctx context.Context, meta RequestMeta, assignment domain.Assignment, operator domain.FleetOperator) error {
	if err := domain.CanAssign(operator, assignment); err != nil {
		return err
	}
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CreateAssignment(context.Context, domain.Assignment) error
			ValidateMissionDroneUnit(context.Context, string, string) error
		})
		if !ok {
			return fmt.Errorf("assignment repository unavailable")
		}
		if err := repo.ValidateMissionDroneUnit(ctx, assignment.MissionID, assignment.DroneUnitID); err != nil {
			return err
		}
		if err := repo.CreateAssignment(ctx, assignment); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "assignment", assignment.ID, "create", "success", nil))
	})
}

func (s *Service) AdvanceAssignment(ctx context.Context, meta RequestMeta, id, next string, version int64) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			AdvanceAssignment(context.Context, string, string, int64) error
		})
		if !ok {
			return fmt.Errorf("assignment repository unavailable")
		}
		if err := repo.AdvanceAssignment(ctx, id, next, version); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "assignment", id, next, "success", nil))
	})
}
