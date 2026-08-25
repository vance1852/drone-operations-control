package db

import (
	"context"
	"os"
	"testing"
)

func TestAdvisoryLockReleasesWhenRequestContextIsCanceled(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL is required")
	}
	holder, err := Open(t.Context(), databaseURL, 1, 0)
	if err != nil {
		t.Fatalf("open holder pool: %v", err)
	}
	defer holder.Close()
	observer, err := Open(t.Context(), databaseURL, 1, 0)
	if err != nil {
		t.Fatalf("open observer pool: %v", err)
	}
	defer observer.Close()
	ctx, cancel := context.WithCancel(t.Context())
	const key int64 = 932606
	if err := AdvisoryLock(ctx, holder, key, func() error {
		cancel()
		return nil
	}); err != nil {
		t.Fatalf("advisory lock operation: %v", err)
	}
	var acquired bool
	if err := observer.QueryRow(t.Context(), `SELECT pg_try_advisory_lock($1)`, key).Scan(&acquired); err != nil {
		t.Fatalf("try lock from independent session: %v", err)
	}
	if !acquired {
		t.Fatal("request cancellation left advisory lock held by pooled session")
	}
	if _, err := observer.Exec(t.Context(), `SELECT pg_advisory_unlock($1)`, key); err != nil {
		t.Fatalf("release observer lock: %v", err)
	}
}
