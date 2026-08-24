package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

// missionRetryTransaction is the per-attempt transaction handle. CreateDroneUnit
// returns attempt-scoped IDs so that IDs produced by a rolled-back attempt are
// distinguishable from the IDs produced by the attempt that finally commits.
type missionRetryTransaction struct {
	repository.Repository
	attempt  int
	droneIDs []string
}

func (t *missionRetryTransaction) CreateMission(context.Context, *domain.Mission) error { return nil }
func (t *missionRetryTransaction) CreateDroneUnit(_ context.Context, in repository.DroneUnitInput) (string, error) {
	id := fmt.Sprintf("attempt-%d-%s", t.attempt, in.Code)
	t.droneIDs = append(t.droneIDs, id)
	return id, nil
}
func (t *missionRetryTransaction) WriteAudit(context.Context, repository.AuditInput) error {
	return nil
}

// missionRetryRepository implements transactionPolicyRunner. It simulates a
// serializable transaction that conflicts (and is rolled back) on every attempt
// except the last one, mirroring the retry boundary in Postgres.RunWithPolicy
// without requiring a live database.
type missionRetryRepository struct {
	repository.Repository
	attempts int
}

func (r *missionRetryRepository) RunWithPolicy(_ context.Context, policy repository.TransactionPolicy, fn func(repository.Repository) error) error {
	for attempt := 1; attempt <= policy.RetryAttempts; attempt++ {
		r.attempts++
		tx := &missionRetryTransaction{attempt: attempt}
		if err := fn(tx); err != nil {
			return err
		}
		if attempt == policy.RetryAttempts {
			return nil // commit succeeds on the final attempt
		}
		// Otherwise the commit fails with a serializable conflict, the
		// attempt is rolled back, and the policy retries.
	}
	return nil
}

func TestCreateMissionReturnsOnlyCommittedDroneIDsAfterSerializationRetry(t *testing.T) {
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	repo := &missionRetryRepository{}
	resp, err := New(repo).WithClock(func() time.Time { return now }).CreateMission(
		t.Context(), RequestMeta{RequestID: "retry-plan"},
		CreateMissionRequest{
			Code: "RETRY", Name: "Retry", Timezone: "UTC",
			StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: "operator",
			DroneUnits: []repository.DroneUnitInput{
				{Code: "DRONE-A", RoomLabel: "Bay-1", RequiredTasks: 1},
				{Code: "DRONE-B", RoomLabel: "Bay-2", RequiredTasks: 1},
			},
		},
	)
	if err != nil {
		t.Fatalf("create mission: %v", err)
	}
	// The policy allows 3 attempts. The first two are rolled back by a
	// serializable conflict after generating drone IDs; only the final
	// attempt commits.
	if repo.attempts != 3 {
		t.Fatalf("transaction attempts=%d, want 3 (two rolled-back retries plus one commit)", repo.attempts)
	}
	want := []string{"attempt-3-DRONE-A", "attempt-3-DRONE-B"}
	if len(resp.DroneUnitIDs) != len(want) {
		t.Fatalf("drone ids=%v, want only the committed attempt's ids %v", resp.DroneUnitIDs, want)
	}
	for i := range want {
		if resp.DroneUnitIDs[i] != want[i] {
			t.Fatalf("drone ids=%v, want %v", resp.DroneUnitIDs, want)
		}
	}
}
