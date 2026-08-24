package service

import (
	"context"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/pagination"
)

func (s *Service) SearchDroneTasks(ctx context.Context, query pagination.Query, missionID string, status domain.DroneTaskStatus) (pagination.Page[domain.DroneTask], error) {
	query = pagination.Normalize(query.Offset, query.Limit)
	page, err := s.repo.ListDroneTasks(ctx, query.Offset, query.Limit, missionID, status)
	if err != nil {
		return pagination.Page[domain.DroneTask]{}, err
	}
	return pagination.From(page.Items, page.Total, query), nil
}
