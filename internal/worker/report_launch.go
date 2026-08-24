package worker

import (
	"context"
	"time"
)

func (w *ReportWorker) launchReport(ctx context.Context, at time.Time) {
	go func() {
		if err := w.generator.Generate(ctx, at); err != nil && ctx.Err() == nil {
			w.logger.Error("report generation failed", "error", err)
		}
	}()
}
