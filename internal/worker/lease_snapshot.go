package worker

import "time"

func (l *Lease) snapshot() (string, time.Time) {
	return l.owner, l.until
}
