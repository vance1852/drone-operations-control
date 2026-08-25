package worker

import "time"

// snapshot returns a consistent (owner, until) pair.
//
// It must be taken under the lease mutex so that a concurrent Acquire
// (renew) or Release cannot update owner and until independently while the
// caller reads them. Without the lock two scheduling loops could observe a
// torn snapshot — e.g. the new owner paired with the old expiry — and reach
// contradictory conclusions about who holds the lease, causing duplicate
// execution of the same drone task under load.
func (l *Lease) snapshot() (string, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	owner, until := l.owner, l.until
	return owner, until
}
