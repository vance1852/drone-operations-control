package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) VerifyHandoff(ctx context.Context, taskID string) ([]repository.HandoffRecord, error) {
	repo, ok := s.repo.(interface {
		ListHandoff(context.Context, string) ([]repository.HandoffRecord, error)
	})
	if !ok {
		return nil, fmt.Errorf("device_transfer repository unavailable")
	}
	items, err := repo.ListHandoff(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if err := repository.ValidateHandoffSequence(items); err != nil {
		return nil, err
	}
	return append([]repository.HandoffRecord(nil), items...), nil
}

func (s *Service) HandoffChecked(ctx context.Context, meta RequestMeta, in repository.HandoffInput, version int64) error {
	if in.RecordedAt.IsZero() {
		in.RecordedAt = s.now()
	}
	device_transfer := domain.Handoff{DroneTaskID: in.DroneTaskID, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt}
	if err := device_transfer.Validate(); err != nil {
		return err
	}
	return s.HandoffDroneTask(ctx, meta, in, version)
}

func (s *Service) AcceptChecked(ctx context.Context, meta RequestMeta, in repository.HandoffInput, version int64) error {
	if in.RecordedAt.IsZero() {
		in.RecordedAt = s.now()
	}
	device_transfer := domain.Handoff{DroneTaskID: in.DroneTaskID, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt}
	if err := device_transfer.Validate(); err != nil {
		return err
	}
	return s.AcceptDroneTask(ctx, meta, in, version)
}
