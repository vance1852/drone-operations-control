package worker

import "context"

func newAttemptContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}
