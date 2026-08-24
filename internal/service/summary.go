package service

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
)

type Summary struct {
	Mission  domain.Mission
	Progress domain.MissionProgress
	Counts   any
}

func (s *Service) MissionSummary(ctx context.Context, missionID string) (Summary, error) {
	mission, err := s.repo.GetMission(ctx, missionID)
	if err != nil {
		return Summary{}, err
	}
	progress, err := s.MissionProgress(ctx, missionID)
	if err != nil {
		return Summary{}, fmt.Errorf("summary progress: %w", err)
	}
	return Summary{Mission: mission, Progress: progress, Counts: progress}, nil
}

func (s *Service) ExpiringDroneTasks(ctx context.Context, beforeUnix int64, limit int) ([]domain.DroneTask, error) {
	repo, ok := s.repo.(interface {
		ExpiringDroneTasks(context.Context, time.Time, int) ([]domain.DroneTask, error)
	})
	if !ok {
		return nil, fmt.Errorf("task query repository unavailable")
	}
	return repo.ExpiringDroneTasks(ctx, time.Unix(beforeUnix, 0).UTC(), limit)
}
