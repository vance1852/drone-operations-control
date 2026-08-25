package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type ReportGenerator interface {
	Generate(context.Context, time.Time) error
}

type ReportWorker struct {
	generator ReportGenerator
	interval  time.Duration
	logger    *slog.Logger

	// active guards the single execution slot so that one worker never runs
	// more than one Generate at a time. wg tracks the in-flight generation so
	// that shutdown drains it instead of leaving an untracked task behind.
	mu     sync.Mutex
	active bool
	wg     sync.WaitGroup
}

func NewReportWorker(generator ReportGenerator, interval time.Duration, logger *slog.Logger) *ReportWorker {
	if interval <= 0 {
		interval = time.Hour
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ReportWorker{generator: generator, interval: interval, logger: logger}
}

func (w *ReportWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	// Wait for any in-flight Generate to finish (and release its snapshot/output
	// resources) before returning, so shutdown never leaves an untracked task.
	defer w.wg.Wait()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if w.generator != nil {
			w.launchReport(ctx, time.Now().UTC())
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
