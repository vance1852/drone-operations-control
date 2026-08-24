package service

import (
	"context"
	"time"

	"drone-operations-control/internal/repository"
)

func (s *Service) createReviewAlert(ctx context.Context, taskID string) (string, error) {
	return s.repo.CreateIntervention(ctx, repository.InterventionInput{
		DroneTaskID: taskID,
		Kind:        "reassess",
		Reason:      "risk score exceeded the alert threshold",
		DueAt:       s.now().Add(72 * time.Hour),
	})
}
