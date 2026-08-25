package worker

import (
	"context"
	"time"
)

// launchReport starts a single Generate run, but only if no other generation is
// already in flight for this worker. A worker owns at most one Generate at a
// time; if the previous round is still occupying the snapshot/output resources,
// the tick is skipped rather than overlapping. The in-flight run is tracked by
// w.wg so Run drains it during shutdown instead of orphaning the task.
func (w *ReportWorker) launchReport(ctx context.Context, at time.Time) {
	w.mu.Lock()
	if w.active {
		// Previous generation still holds the snapshot/output resources; skip
		// this tick to avoid concurrent writers overwriting each other.
		w.mu.Unlock()
		return
	}
	w.active = true
	w.mu.Unlock()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			w.mu.Lock()
			w.active = false
			w.mu.Unlock()
		}()
		if err := w.generator.Generate(ctx, at); err != nil && ctx.Err() == nil {
			w.logger.Error("report generation failed", "error", err)
		}
	}()
}
