package integration

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"drone-operations-control/internal/db"
	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
	"github.com/google/uuid"
)

func openDatabase(t *testing.T) (*db.Pool, context.Context) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is required for PostgreSQL integration tests")
	}
	ctx := context.Background()
	pool, err := db.Open(ctx, url, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE TABLE audit_events, safety_alerts, telemetry_events, mission_batch_tasks, mission_batches, device_transfer_events, drone_tasks, mission_drones, drone_missions, operators, idempotency_keys CASCADE`); err != nil {
		pool.Close()
		t.Fatal(err)
	}
	return pool, ctx
}

func insertFleetOperator(t *testing.T, ctx context.Context, pool *db.Pool, role string) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO operators(id,name,role) VALUES ($1,$2,$3)`, id, "Test FleetOperator", role); err != nil {
		t.Fatal(err)
	}
	return id
}

func createWorkflow(t *testing.T, ctx context.Context, svc *service.Service, operator string) (service.CreateMissionResponse, domain.DroneTask) {
	t.Helper()
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	svc.WithClock(func() time.Time { return now })
	mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "req-create"}, service.CreateMissionRequest{
		Code: "PLAN-001", Name: "North river survey", Timezone: "Asia/Shanghai", StartsAt: now, EndsAt: now.Add(24 * time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "N-01", RoomLabel: "A-101", RequiredTasks: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ScheduleMission(ctx, service.RequestMeta{RequestID: "req-schedule"}, mission.Mission.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.ActivateMission(ctx, service.RequestMeta{RequestID: "req-collect"}, mission.Mission.ID, 2); err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateDroneTask(ctx, service.RequestMeta{RequestID: "req-task"}, repository.DroneTaskInput{MissionID: mission.Mission.ID, DroneUnitID: mission.DroneUnitIDs[0], TaskCode: "S-001", ExpiresAt: now.Add(12 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteDroneTask(ctx, service.RequestMeta{RequestID: "req-collect-task"}, task.ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.HandoffDroneTask(ctx, service.RequestMeta{RequestID: "req-device_transfer"}, repository.HandoffInput{DroneTaskID: task.ID, To: operator, Location: "A-101", RecordedAt: now}, 2); err != nil {
		t.Fatal(err)
	}
	if err := svc.AcceptDroneTask(ctx, service.RequestMeta{RequestID: "req-receive"}, repository.HandoffInput{DroneTaskID: task.ID, To: operator, Location: "Drone bay", RecordedAt: now}, 3); err != nil {
		t.Fatal(err)
	}
	page, err := svc.ListDroneTasks(ctx, 0, 10, mission.Mission.ID, domain.DroneTaskAccepted)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("task listing = %+v, %v", page, err)
	}
	return mission, page.Items[0]
}

func TestDroneTaskWorkflowPersistsAcrossOperations(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	mission, task := createWorkflow(t, ctx, svc, operator)
	if mission.Mission.Status != domain.MissionDraft {
		t.Fatalf("created mission status = %s", mission.Mission.Status)
	}
	if task.Status != domain.DroneTaskAccepted {
		t.Fatalf("task status = %s", task.Status)
	}
	if task.Version != 4 {
		t.Fatalf("task version = %d", task.Version)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1`, task.ID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 4 {
		t.Fatalf("audit count = %d, want 4", auditCount)
	}
}

func TestMissionCreationRollsBackWhenASecondDroneUnitViolatesConstraint(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "safety_supervisor")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	_, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "req-rollback"}, service.CreateMissionRequest{
		Code: "PLAN-ROLLBACK", Name: "Rollback test", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "GOOD", RoomLabel: "A", RequiredTasks: 1}, {Code: "BAD", RoomLabel: "B", RequiredTasks: 0}},
	})
	if err == nil {
		t.Fatal("invalid second drone was accepted")
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drone_missions WHERE code='PLAN-ROLLBACK'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("mission survived rollback: %d", count)
	}
}

func TestConcurrentDroneTaskTransitionAllowsOnlyOneVersionWinner(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	repo := repository.NewPostgres(pool)
	mission, task := createWorkflow(t, ctx, svc, operator)
	_ = mission
	// The workflow leaves the task at received/version 4. Two workers race to start testing.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- repo.MoveDroneTask(ctx, task.ID, domain.DroneTaskInProgress, 4, time.Now().UTC())
		}()
	}
	wg.Wait()
	close(results)
	var success, conflict int
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrConflict) {
			conflict++
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}

func TestMigrationAndStateSurviveDatabaseReopen(t *testing.T) {
	pool, ctx := openDatabase(t)
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	_, task := createWorkflow(t, ctx, svc, operator)
	pool.Close()
	url := os.Getenv("DATABASE_URL")
	reopened, err := db.Open(ctx, url, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := repository.NewPostgres(reopened).GetDroneTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TaskCode != "S-001" || loaded.Status != domain.DroneTaskAccepted {
		t.Fatalf("reopened task = %+v", loaded)
	}
}
