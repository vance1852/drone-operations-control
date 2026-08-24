package service

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) OpenIntervention(ctx context.Context, meta RequestMeta, in repository.InterventionInput) (string, error) {
	if in.DueAt.IsZero() {
		in.DueAt = s.now().Add(72 * time.Hour)
	}
	safety_alert := domain.Intervention{DroneTaskID: in.DroneTaskID, Kind: in.Kind, Status: domain.InterventionOpen, Reason: in.Reason, DueAt: in.DueAt}
	if err := safety_alert.Validate(s.now()); err != nil {
		return "", err
	}
	var id string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		var err error
		id, err = tx.CreateIntervention(ctx, in)
		if err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "open", "success", nil))
	})
	return id, err
}

func (s *Service) CloseIntervention(ctx context.Context, meta RequestMeta, id string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			CloseIntervention(context.Context, string, time.Time) error
		})
		if !ok {
			return fmt.Errorf("safety_alert repository unavailable")
		}
		if err := repo.CloseIntervention(ctx, id, s.now()); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "close", "success", nil))
	})
}

func (s *Service) MarkInterventionInProgress(ctx context.Context, meta RequestMeta, id string) error {
	return s.repo.InTx(ctx, func(tx repository.Repository) error {
		repo, ok := tx.(interface {
			MarkInterventionInProgress(context.Context, string) error
		})
		if !ok {
			return fmt.Errorf("safety_alert repository unavailable")
		}
		if err := repo.MarkInterventionInProgress(ctx, id); err != nil {
			return err
		}
		return tx.WriteAudit(ctx, audit(meta, "safety_alert", id, "start", "success", nil))
	})
}
