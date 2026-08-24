package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type blockingReportGenerator struct {
	active      atomic.Int32
	started     chan struct{}
	overlap     chan struct{}
	release     chan struct{}
	startedOnce sync.Once
	overlapOnce sync.Once
}

func (g *blockingReportGenerator) Generate(ctx context.Context, _ time.Time) error {
	active := g.active.Add(1)
	defer g.active.Add(-1)
	g.startedOnce.Do(func() { close(g.started) })
	if active > 1 {
		g.overlapOnce.Do(func() { close(g.overlap) })
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-g.release:
		return nil
	}
}

func TestReportWorkerDoesNotOverlapSlowGenerations(t *testing.T) {
	generator := &blockingReportGenerator{started: make(chan struct{}), overlap: make(chan struct{}), release: make(chan struct{})}
	worker := NewReportWorker(generator, time.Millisecond, nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	select {
	case <-generator.started:
	case <-time.After(time.Second):
		cancel()
		t.Fatal("report worker did not start generation")
	}
	select {
	case <-generator.overlap:
		cancel()
		close(generator.release)
		t.Fatal("report worker started a second Generate before the first completed")
	case <-time.After(40 * time.Millisecond):
		cancel()
		close(generator.release)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("report worker did not stop after cancellation")
	}
}
