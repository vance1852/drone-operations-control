package integration

import (
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
	"github.com/google/uuid"
)

func TestFleetOperatorCreationRollsBackWhenAuditWriteFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	svc := service.New(repository.NewPostgres(pool))
	missingFleetOperator := uuid.NewString()
	_, err := svc.RegisterFleetOperator(ctx, service.RequestMeta{RequestID: "rollback", FleetOperatorID: &missingFleetOperator}, "Rollback FleetOperator", domain.RoleSafetySupervisor)
	if err == nil {
		t.Fatal("operator creation succeeded despite rejected audit event")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM operators WHERE name='Rollback FleetOperator'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("operator survived audit rollback: %d", count)
	}
}

func TestDroneMissionBatchTransitionRollsBackWhenAuditWriteFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	missionBatchID, err := repo.CreateDroneMissionBatch(ctx, repository.DroneMissionBatchInput{Code: "ROLLBACK-ROUND", Method: "daily-drone", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	missingFleetOperator := uuid.NewString()
	err = service.New(repo).StartDroneMissionBatch(ctx, service.RequestMeta{RequestID: "rollback", FleetOperatorID: &missingFleetOperator}, missionBatchID, 1)
	if err == nil {
		t.Fatal("missionBatch transition succeeded despite rejected audit event")
	}
	var status string
	var version int64
	if err := pool.QueryRow(ctx, `SELECT status,version FROM mission_batches WHERE id=$1`, missionBatchID).Scan(&status, &version); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.DroneMissionBatchQueued) || version != 1 {
		t.Fatalf("missionBatch survived audit rollback: status=%s version=%d", status, version)
	}
}

func TestDroneMissionBatchCancelRollsBackWhenTaskRestoreFails(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	future := time.Now().UTC().Add(12 * time.Hour)
	missionID := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO drone_missions(id,code,name,status,timezone,starts_at,ends_at,version,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,1,$8)`,
		missionID, "CANCEL-MISSION", "Cancel rollback mission", string(domain.MissionActive), "UTC", time.Now().UTC().Add(-time.Hour), future, operator); err != nil {
		t.Fatal(err)
	}
	droneID := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO mission_drones(id,mission_id,code,room_label,required_tasks) VALUES ($1,$2,$3,$4,1)`,
		droneID, missionID, "DRONE-CANCEL", "A-101"); err != nil {
		t.Fatal(err)
	}
	taskID := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO drone_tasks(id,mission_id,drone_id,task_code,status,expires_at,version) VALUES ($1,$2,$3,$4,$5,$6,1)`,
		taskID, missionID, droneID, "CANCEL-TASK", string(domain.DroneTaskInProgress), future); err != nil {
		t.Fatal(err)
	}
	missionBatchID, err := repo.CreateDroneMissionBatch(ctx, repository.DroneMissionBatchInput{Code: "CANCEL-ROLLBACK", Method: "daily-drone", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO mission_batch_tasks(mission_batch_id,drone_task_id) VALUES ($1,$2)`, missionBatchID, taskID); err != nil {
		t.Fatal(err)
	}
	missingFleetOperator := uuid.NewString()
	err = service.New(repo).CancelDroneMissionBatch(ctx, service.RequestMeta{RequestID: "req-cancel", FleetOperatorID: &missingFleetOperator}, missionBatchID, 1)
	if err == nil {
		t.Fatal("missionBatch cancel succeeded despite rejected audit event")
	}
	var batchStatus, taskStatus string
	var batchVersion, taskVersion int64
	if err := pool.QueryRow(ctx, `SELECT status,version FROM mission_batches WHERE id=$1`, missionBatchID).Scan(&batchStatus, &batchVersion); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT status,version FROM drone_tasks WHERE id=$1`, taskID).Scan(&taskStatus, &taskVersion); err != nil {
		t.Fatal(err)
	}
	if batchStatus != string(domain.DroneMissionBatchQueued) || batchVersion != 1 {
		t.Fatalf("cancelled missionBatch survived audit rollback: status=%s version=%d", batchStatus, batchVersion)
	}
	if taskStatus != string(domain.DroneTaskInProgress) || taskVersion != 1 {
		t.Fatalf("task changed during cancelled missionBatch rollback: status=%s version=%d", taskStatus, taskVersion)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1`, missionBatchID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 {
		t.Fatalf("cancel audit survived rollback: count=%d", auditCount)
	}
}
