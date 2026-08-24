package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (s *Service) Authorize(ctx context.Context, operatorID, action string) error {
	operator, err := s.LoadFleetOperator(ctx, operatorID)
	if err != nil {
		return err
	}
	if !operator.Can(action) {
		return fmt.Errorf("operator cannot %s: %w", action, domain.ErrConflict)
	}
	return nil
}

func RequireSupervisor(ctx context.Context, s *Service, operatorID string) error {
	return s.Authorize(ctx, operatorID, "close_mission")
}

func RequireReviewer(ctx context.Context, s *Service, operatorID string) error {
	return s.Authorize(ctx, operatorID, "review_telemetry")
}
