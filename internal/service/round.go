package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) StartDroneMissionBatch(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.changeDroneMissionBatch(ctx, meta, id, version, domain.DroneMissionBatchRunning, "start")
}

func (s *Service) CompleteDroneMissionBatch(ctx context.Context, meta RequestMeta, id string, version int64) error {
	return s.changeDroneMissionBatch(ctx, meta, id, version, domain.DroneMissionBatchCompleted, "complete")
}

func (s *Service) CancelDroneMissionBatch(ctx context.Context, meta RequestMeta, id string, version int64) error {
	if err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CancelDroneMissionBatch(context.Context, string, int64) error
		})
		if !ok {
			return fmt.Errorf("missionBatch repository unavailable")
		}
		return repo.CancelDroneMissionBatch(ctx, id, version)
	}); err != nil {
		return err
	}
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			RestoreCancelledDroneTasks(context.Context, string) error
		})
		if !ok {
			return fmt.Errorf("missionBatch restoration repository unavailable")
		}
		if err := repo.RestoreCancelledDroneTasks(ctx, id); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "mission_batch", id, "cancel", "success", nil))
	})
}

func (s *Service) changeDroneMissionBatch(ctx context.Context, meta RequestMeta, id string, version int64, next domain.DroneMissionBatchStatus, action string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			StartDroneMissionBatch(context.Context, string, int64) error
			CompleteDroneMissionBatch(context.Context, string, int64) error
			CancelDroneMissionBatch(context.Context, string, int64) error
		})
		if !ok {
			return fmt.Errorf("missionBatch repository unavailable")
		}
		var err error
		switch next {
		case domain.DroneMissionBatchRunning:
			err = repo.StartDroneMissionBatch(ctx, id, version)
		case domain.DroneMissionBatchCompleted:
			err = repo.CompleteDroneMissionBatch(ctx, id, version)
		case domain.DroneMissionBatchCancelled:
			err = repo.CancelDroneMissionBatch(ctx, id, version)
		}
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "mission_batch", id, action, "success", nil))
	})
}
