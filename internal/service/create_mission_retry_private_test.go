package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

type retryingCreateMissionRepository struct {
	repository.Repository
	attempt      int
	attemptIDs   []string
	committedIDs map[string]struct{}
}

func (r *retryingCreateMissionRepository) RunWithPolicy(_ context.Context, _ repository.TransactionPolicy, operation func(repository.Repository) error) error {
	for attempt := 1; attempt <= 2; attempt++ {
		r.attempt = attempt
		r.attemptIDs = nil
		if err := operation(r); err != nil {
			return err
		}
		if attempt == 1 {
			continue // The database discarded this attempt after a serialization conflict at commit.
		}
		for _, id := range r.attemptIDs {
			r.committedIDs[id] = struct{}{}
		}
		return nil
	}
	return fmt.Errorf("transaction retry did not commit")
}

func (r *retryingCreateMissionRepository) CreateMission(context.Context, *domain.Mission) error {
	return nil
}

func (r *retryingCreateMissionRepository) CreateDroneUnit(context.Context, repository.DroneUnitInput) (string, error) {
	id := fmt.Sprintf("attempt-%d-drone-%d", r.attempt, len(r.attemptIDs)+1)
	r.attemptIDs = append(r.attemptIDs, id)
	return id, nil
}

func (r *retryingCreateMissionRepository) WriteAudit(context.Context, repository.AuditInput) error {
	return nil
}

func TestCreateMissionReturnsOnlyCommittedDroneIDsAfterTransactionRetry(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	repo := &retryingCreateMissionRepository{committedIDs: map[string]struct{}{}}
	svc := New(repo).WithClock(func() time.Time { return now })

	result, err := svc.CreateMission(t.Context(), RequestMeta{RequestID: "retry-create"}, CreateMissionRequest{
		Code:      "MISSION-RETRY",
		Name:      "Retry Mission",
		Timezone:  "UTC",
		StartsAt:  now.Add(time.Hour),
		EndsAt:    now.Add(3 * time.Hour),
		CreatedBy: "operator-1",
		DroneUnits: []repository.DroneUnitInput{{
			Code:          "DRONE-01",
			RoomLabel:     "north-hangar",
			RequiredTasks: 1,
		}},
	})
	if err != nil {
		t.Fatalf("create mission after retry: %v", err)
	}
	if len(result.DroneUnitIDs) != len(repo.committedIDs) {
		t.Fatalf("response ids=%v committed ids=%v", result.DroneUnitIDs, repo.committedIDs)
	}
	for _, id := range result.DroneUnitIDs {
		if _, committed := repo.committedIDs[id]; !committed {
			t.Fatalf("response exposed rolled-back drone id %q; committed=%v", id, repo.committedIDs)
		}
	}
}
