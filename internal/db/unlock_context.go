package db

import (
	"context"
	"time"
)

func advisoryUnlockContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 5*time.Second)
}
