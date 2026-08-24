package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type DroneMissionBatchJob struct {
	ID          string
	Attempts    int
	MaxAttempts int
}
type DroneMissionBatchExecutor interface {
	Execute(context.Context, DroneMissionBatchJob) error
}

type DroneMissionBatchProcessor struct {
	executor DroneMissionBatchExecutor
	policy   RetryPolicy
	logger   *slog.Logger
	metrics  *Metrics
}

func NewDroneMissionBatchProcessor(executor DroneMissionBatchExecutor, policy RetryPolicy, logger *slog.Logger, metrics *Metrics) *DroneMissionBatchProcessor {
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		metrics = &Metrics{}
	}
	return &DroneMissionBatchProcessor{executor: executor, policy: policy, logger: logger, metrics: metrics}
}

func (p *DroneMissionBatchProcessor) Process(ctx context.Context, job DroneMissionBatchJob) error {
	if job.ID == "" {
		return fmt.Errorf("missionBatch job id is required")
	}
	if p.executor == nil {
		return fmt.Errorf("missionBatch executor is nil")
	}
	policy := p.policy
	if job.MaxAttempts > 0 && policy.Attempts > job.MaxAttempts {
		policy.Attempts = job.MaxAttempts
	}
	start := time.Now()
	err := RunWithRetry(ctx, policy, func(callCtx context.Context) error { job.Attempts++; return p.executor.Execute(callCtx, job) })
	p.metrics.RecordRun()
	if err != nil {
		p.metrics.RecordFailure()
		p.logger.Error("missionBatch job failed", "mission_batch_id", job.ID, "attempts", job.Attempts, "duration", time.Since(start), "error", err)
		return err
	}
	p.logger.Info("missionBatch job completed", "mission_batch_id", job.ID, "attempts", job.Attempts, "duration", time.Since(start))
	return nil
}
