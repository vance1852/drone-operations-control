package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"drone-operations-control/internal/domain"
)

type DroneTaskExpiryRepository interface {
	ExpiringDroneTasks(context.Context, time.Time, int) ([]domain.DroneTask, error)
}

type DroneTaskExpiryReconciler struct {
	repo    DroneTaskExpiryRepository
	log     *slog.Logger
	metrics *Metrics
}

func NewDroneTaskExpiryReconciler(repo DroneTaskExpiryRepository, logger *slog.Logger, metrics *Metrics) *DroneTaskExpiryReconciler {
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &DroneTaskExpiryReconciler{repo: repo, log: logger, metrics: metrics}
}

func (r *DroneTaskExpiryReconciler) Reconcile(ctx context.Context, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.repo == nil {
		return fmt.Errorf("task expiry repository is nil")
	}
	r.metrics.RecordRun()
	items, err := r.repo.ExpiringDroneTasks(ctx, now, 100)
	if err != nil {
		r.metrics.RecordFailure()
		return err
	}
	r.metrics.RecordDue(len(items))
	for _, item := range items {
		r.log.Warn("task is near expiry", "drone_task_id", item.ID, "task_code", item.TaskCode, "expires_at", item.ExpiresAt)
	}
	return nil
}
