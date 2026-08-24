package integration

import (
	"strings"
	"testing"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
)

func TestCancelDroneMissionBatchRollsBackWhenTaskRestoreFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	_, task := createWorkflow(t, ctx, svc, operator)
	if _, err := pool.Exec(ctx, `UPDATE drone_tasks SET expires_at=now()+interval '1 hour' WHERE id=$1`, task.ID); err != nil {
		t.Fatalf("extend task expiry: %v", err)
	}
	roundID, err := svc.CreateDroneMissionBatch(ctx, service.RequestMeta{RequestID: "private-round-create"}, repository.DroneMissionBatchInput{Code: "PRIVATE-CANCEL-ROUND", Method: "atomic-cancel", Capacity: 1}, []string{task.ID})
	if err != nil {
		t.Fatalf("create round: %v", err)
	}

	if _, err := pool.Exec(ctx, `CREATE OR REPLACE FUNCTION private_reject_round_restore() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.status = 'in_progress' AND NEW.status = 'accepted' THEN
    RAISE EXCEPTION 'injected task restore failure';
  END IF;
  RETURN NEW;
END $$`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `CREATE TRIGGER private_round_restore_failure BEFORE UPDATE ON drone_tasks FOR EACH ROW EXECUTE FUNCTION private_reject_round_restore()`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DROP TRIGGER IF EXISTS private_round_restore_failure ON drone_tasks`)
		_, _ = pool.Exec(ctx, `DROP FUNCTION IF EXISTS private_reject_round_restore()`)
	})

	err = svc.CancelDroneMissionBatch(ctx, service.RequestMeta{RequestID: "private-cancel-failure"}, roundID, 1)
	if err == nil {
		t.Fatal("cancellation unexpectedly succeeded while task restoration failed")
	}
	if !strings.Contains(err.Error(), "injected task restore failure") {
		t.Fatalf("cancellation lost restore error: %v", err)
	}
	var roundStatus, taskStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM mission_batches WHERE id=$1`, roundID).Scan(&roundStatus); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT status FROM drone_tasks WHERE id=$1`, task.ID).Scan(&taskStatus); err != nil {
		t.Fatal(err)
	}
	if roundStatus != string(domain.DroneMissionBatchQueued) || taskStatus != string(domain.DroneTaskInProgress) {
		t.Fatalf("failed cancellation persisted round=%s task=%s", roundStatus, taskStatus)
	}
	var cancelAudits int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_type='mission_batch' AND object_id=$1 AND action='cancel' AND outcome='success'`, roundID).Scan(&cancelAudits); err != nil {
		t.Fatal(err)
	}
	if cancelAudits != 0 {
		t.Fatalf("failed cancellation wrote %d success audits", cancelAudits)
	}

	if _, err := pool.Exec(ctx, `DROP TRIGGER private_round_restore_failure ON drone_tasks`); err != nil {
		t.Fatal(err)
	}
	if err := svc.CancelDroneMissionBatch(ctx, service.RequestMeta{RequestID: "private-cancel-retry"}, roundID, 1); err != nil {
		t.Fatalf("retry cancellation: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT status FROM mission_batches WHERE id=$1`, roundID).Scan(&roundStatus); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT status FROM drone_tasks WHERE id=$1`, task.ID).Scan(&taskStatus); err != nil {
		t.Fatal(err)
	}
	if roundStatus != string(domain.DroneMissionBatchCancelled) || taskStatus != string(domain.DroneTaskAccepted) {
		t.Fatalf("successful cancellation round=%s task=%s", roundStatus, taskStatus)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_type='mission_batch' AND object_id=$1 AND action='cancel' AND outcome='success'`, roundID).Scan(&cancelAudits); err != nil {
		t.Fatal(err)
	}
	if cancelAudits != 1 {
		t.Fatalf("successful cancellation audits=%d", cancelAudits)
	}
}
