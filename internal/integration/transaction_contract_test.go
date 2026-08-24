package integration

import (
	"testing"

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
