package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
)

func (s *Service) MissionProgress(ctx context.Context, missionID string) (domain.MissionProgress, error) {
	if _, err := s.repo.GetMission(ctx, missionID); err != nil {
		return domain.MissionProgress{}, err
	}
	repo, ok := s.repo.(interface {
		MissionProgress(context.Context, string) (domain.MissionProgress, error)
	})
	if !ok {
		return domain.MissionProgress{}, fmt.Errorf("progress repository unavailable")
	}
	return repo.MissionProgress(ctx, missionID)
}

func (s *Service) AuditHistory(ctx context.Context, objectType, objectID string, limit int) ([]domain.AuditSummary, error) {
	repo, ok := s.repo.(interface {
		AuditHistory(context.Context, string, string, int) ([]domain.AuditSummary, error)
	})
	if !ok {
		return nil, fmt.Errorf("audit repository unavailable")
	}
	return repo.AuditHistory(ctx, objectType, objectID, limit)
}
