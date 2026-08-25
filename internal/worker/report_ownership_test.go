package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// blockingGenerator blocks each Generate until release is closed, and counts
// concurrent and total invocations so tests can assert single-ownership.
type blockingGenerator struct {
	mu          sync.Mutex
	concurrent  int
	maxParallel int
	started     atomic.Int64
	release     chan struct{}
}

func newBlockingGenerator() *blockingGenerator {
	return &blockingGenerator{release: make(chan struct{})}
}

func (g *blockingGenerator) Generate(_ context.Context, _ time.Time) error {
	g.mu.Lock()
	g.concurrent++
	if g.concurrent > g.maxParallel {
		g.maxParallel = g.concurrent
	}
	g.mu.Unlock()

	g.started.Add(1)
	<-g.release // block until the test lets this run finish

	g.mu.Lock()
	g.concurrent--
	g.mu.Unlock()
	return nil
}

func (g *blockingGenerator) maxObservedParallelism() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.maxParallel
}

// TestReportWorkerRunsAtMostOneGenerateAtATime verifies the execution
// ownership: when a Generate takes longer than the scheduling interval, the
// worker skips subsequent ticks instead of launching overlapping generators
// that would write the same report and clobber each other.
func TestReportWorkerRunsAtMostOneGenerateAtATime(t *testing.T) {
	gen := newBlockingGenerator()
	w := NewReportWorker(gen, 5*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Wait for at least one generation to start, then let the ticker fire
	// several times while it is still in flight.
	if waitErr := waitForCount(&gen.started, 1, time.Second); waitErr != nil {
		cancel()
		t.Fatalf("first Generate did not start: %v", waitErr)
	}
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Millisecond)
	}

	// Release the in-flight generation and stop the worker.
	close(gen.release)
	cancel()
	if err := <-errCh; err != nil && err != context.Canceled {
		t.Fatalf("run error=%v", err)
	}

	if gen.maxObservedParallelism() > 1 {
		t.Fatalf("overlapping generators observed: max parallel=%d", gen.maxObservedParallelism())
	}
}

// TestReportWorkerShutdownDrainsInFlightGenerate verifies that Run does not
// return while a Generate is still running, so shutdown never leaves an
// untracked task holding snapshot/output resources.
func TestReportWorkerShutdownDrainsInFlightGenerate(t *testing.T) {
	gen := newBlockingGenerator()
	w := NewReportWorker(gen, 5*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- w.Run(ctx) }()

	// Wait for a generation to start (it is blocked), then cancel the context.
	if waitErr := waitForCount(&gen.started, 1, time.Second); waitErr != nil {
		cancel()
		t.Fatalf("Generate did not start: %v", waitErr)
	}
	cancel()

	// Run must still be blocked: it must not return before Generate finishes.
	select {
	case err := <-errCh:
		t.Fatalf("Run returned before in-flight Generate finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	// Now let the in-flight Generate finish; Run should return promptly.
	close(gen.release)
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Fatalf("run error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after Generate finished")
	}
}

func waitForCount(c *atomic.Int64, want int64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.Load() >= want {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return context.DeadlineExceeded
}
