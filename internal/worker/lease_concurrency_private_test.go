package worker

import (
	"sync"
	"testing"
	"time"
)

func TestLeaseRenewalAndOwnershipReadShareSynchronization(t *testing.T) {
	var lease Lease
	base := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	if !lease.Acquire("worker-a", base, time.Minute) {
		t.Fatal("initial lease acquisition failed")
	}
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 20000; i++ {
			lease.Acquire("worker-a", base.Add(time.Duration(i)*time.Nanosecond), time.Minute)
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 20000; i++ {
			lease.HeldBy("worker-a", base.Add(time.Duration(i)*time.Nanosecond))
		}
	}()
	close(start)
	workers.Wait()
	finalNow := base.Add(2 * time.Hour)
	if !lease.Acquire("worker-final", finalNow, time.Minute) {
		t.Fatal("final lease acquisition failed")
	}
	if !lease.HeldBy("worker-final", finalNow) {
		t.Fatal("final owner is not visible after concurrent renewals")
	}
}
