package service

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
)

func (s *Service) ComplianceReport(ctx context.Context, missionID string) (domain.ComplianceReport, error) {
	if _, err := s.repo.GetMission(ctx, missionID); err != nil {
		return domain.ComplianceReport{}, err
	}
	repo, ok := s.repo.(interface {
		ComplianceReport(context.Context, string, time.Time) (domain.ComplianceReport, error)
	})
	if !ok {
		return domain.ComplianceReport{}, fmt.Errorf("report repository unavailable")
	}
	return repo.ComplianceReport(ctx, missionID, s.now())
}

func (s *Service) PublicDroneTask(task domain.DroneTask) map[string]any {
	return map[string]any{"id": task.ID, "task_code": domain.RedactTaskCode(task.TaskCode), "status": task.Status, "expires_at": task.ExpiresAt, "version": task.Version}
}
