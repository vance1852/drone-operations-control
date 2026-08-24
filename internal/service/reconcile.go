package service

import (
	"context"
	"fmt"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (s *Service) MarkExpiredDroneTasks(ctx context.Context, now time.Time, limit int) (repository.ReconcileResult, error) {
	repo, ok := s.repo.(interface {
		MarkExpiredDroneTasks(context.Context, time.Time, int) (repository.ReconcileResult, error)
	})
	if !ok {
		return repository.ReconcileResult{}, fmt.Errorf("reconcile repository unavailable")
	}
	return repo.MarkExpiredDroneTasks(ctx, now, limit)
}

func (s *Service) SearchDroneTasksAdvanced(ctx context.Context, request domain.SearchRequest) (repository.Page, error) {
	repo, ok := s.repo.(interface {
		SearchDroneTasks(context.Context, domain.SearchRequest) (repository.Page, error)
	})
	if !ok {
		return repository.Page{}, fmt.Errorf("search repository unavailable")
	}
	return repo.SearchDroneTasks(ctx, request)
}
