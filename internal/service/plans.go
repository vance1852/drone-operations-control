package service

import (
	"context"
	"fmt"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) ListMissions(ctx context.Context, filter repository.MissionFilter) ([]domain.Mission, int, error) {
	repo, ok := s.repo.(interface {
		ListMissions(context.Context, repository.MissionFilter) ([]domain.Mission, int, error)
	})
	if !ok {
		return nil, 0, fmt.Errorf("mission query repository unavailable")
	}
	items, total, err := repo.ListMissions(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return append([]domain.Mission(nil), items...), total, nil
}

func (s *Service) ListMissionDroneUnits(ctx context.Context, missionID string) ([]domain.DroneUnit, error) {
	repo, ok := s.repo.(interface {
		ListMissionDroneUnits(context.Context, string) ([]domain.DroneUnit, error)
	})
	if !ok {
		return nil, fmt.Errorf("drone query repository unavailable")
	}
	items, err := repo.ListMissionDroneUnits(ctx, missionID)
	if err != nil {
		return nil, err
	}
	return append([]domain.DroneUnit(nil), items...), nil
}
