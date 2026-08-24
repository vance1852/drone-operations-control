package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"drone-operations-control/internal/service"
)

type InterventionWorker struct {
	service  *service.Service
	interval time.Duration
	log      *slog.Logger
}

func NewInterventionWorker(svc *service.Service, interval time.Duration, logger *slog.Logger) *InterventionWorker {
	if interval <= 0 {
		interval = time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &InterventionWorker{service: svc, interval: interval, log: logger}
}

func (w *InterventionWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if w.service == nil {
			return fmt.Errorf("safety_alert service is nil")
		}
		if err := w.reconcile(ctx); err != nil && ctx.Err() == nil {
			w.log.Error("safety_alert reconciliation failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *InterventionWorker) reconcile(ctx context.Context) error {
	items, err := w.service.DueInterventions(ctx, time.Now().UTC(), 100)
	if err != nil {
		return err
	}
	for _, item := range items {
		w.log.Warn("safety_alert is due", "drone_task_id", item.DroneTaskID, "kind", item.Kind, "due_at", item.DueAt)
	}
	return nil
}
