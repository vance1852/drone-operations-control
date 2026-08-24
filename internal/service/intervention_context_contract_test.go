package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/repository"
)

type cancellationRepository struct {
	repository.Repository
}

func (cancellationRepository) DueInterventions(ctx context.Context, _ time.Time, _ int) ([]repository.InterventionInput, error) {
	return nil, ctx.Err()
}

func (cancellationRepository) DueInterventionsDetached(context.Context, time.Time, int) ([]repository.InterventionInput, error) {
	return nil, nil
}

func TestDueInterventionQueryPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := New(cancellationRepository{}).DueInterventions(ctx, time.Now().UTC(), 10)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
