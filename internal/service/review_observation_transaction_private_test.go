package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

type conflictingObservationReviewRepository struct {
	repository.Repository
	alerts []repository.InterventionInput
}

func (r *conflictingObservationReviewRepository) InTx(ctx context.Context, fn func(repository.Repository) error) error {
	return fn(r)
}
func (r *conflictingObservationReviewRepository) ObservationTaskID(context.Context, string) (string, error) {
	return "task-1", nil
}
func (r *conflictingObservationReviewRepository) ReviewObservationRecord(context.Context, string, bool, int64, time.Time) error {
	return nil
}
func (r *conflictingObservationReviewRepository) MoveDroneTask(context.Context, string, domain.DroneTaskStatus, int64, time.Time) error {
	return domain.ErrConflict
}
func (r *conflictingObservationReviewRepository) CreateIntervention(_ context.Context, in repository.InterventionInput) (string, error) {
	r.alerts = append(r.alerts, in)
	return "alert-1", nil
}
func (r *conflictingObservationReviewRepository) WriteAudit(context.Context, repository.AuditInput) error {
	return nil
}

func TestReviewObservationConflictDoesNotPersistSafetyAlert(t *testing.T) {
	repo := &conflictingObservationReviewRepository{}
	svc := New(repo).WithClock(func() time.Time { return time.Date(2026, 8, 24, 11, 0, 0, 0, time.UTC) })
	err := svc.ReviewObservation(t.Context(), RequestMeta{RequestID: "review-conflict"}, "telemetry-1", "task-1", false, 3, 7)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("review error=%v want conflict", err)
	}
	if len(repo.alerts) != 0 {
		t.Fatalf("failed review persisted safety alerts: %+v", repo.alerts)
	}
}
