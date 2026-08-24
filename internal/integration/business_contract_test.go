package integration

import (
	"errors"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
)

func TestMissionProgressCountsEachDroneUnitRequirementOnce(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "progress"}, service.CreateMissionRequest{
		Code: "PROGRESS", Name: "Progress", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "P-1", RoomLabel: "A", RequiredTasks: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewPostgres(pool)
	for _, code := range []string{"P-S1", "P-S2"} {
		if _, err := repo.CreateDroneTask(ctx, repository.DroneTaskInput{MissionID: mission.Mission.ID, DroneUnitID: mission.DroneUnitIDs[0], TaskCode: code, ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	progress, err := repo.MissionProgress(ctx, mission.Mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.DroneUnits != 1 || progress.Required != 2 {
		t.Fatalf("progress=%+v", progress)
	}
}

func TestComplianceReportContainsOnlyItsMissionDroneTasks(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	create := func(code string) service.CreateMissionResponse {
		mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: code}, service.CreateMissionRequest{
			Code: code, Name: code, Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
			DroneUnits: []repository.DroneUnitInput{{Code: code + "-ROBOT", RoomLabel: "A", RequiredTasks: 1}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return mission
	}
	first, second := create("REPORT-A"), create("REPORT-B")
	repo := repository.NewPostgres(pool)
	for _, item := range []struct{ mission, drone, code string }{{first.Mission.ID, first.DroneUnitIDs[0], "REPORT-S1"}, {second.Mission.ID, second.DroneUnitIDs[0], "REPORT-S2"}} {
		if _, err := repo.CreateDroneTask(ctx, repository.DroneTaskInput{MissionID: item.mission, DroneUnitID: item.drone, TaskCode: item.code, ExpiresAt: now.Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	report, err := repo.ComplianceReport(ctx, first.Mission.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Expiring) != 1 || report.Expiring[0].MissionID != first.Mission.ID {
		t.Fatalf("expiring=%+v", report.Expiring)
	}
}

func TestDroneMissionBatchRejectsDroneTasksThatAreNotReadyForInProgress(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "missionBatch"}, service.CreateMissionRequest{
		Code: "ROUND-READY", Name: "DroneMissionBatch", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "B-1", RoomLabel: "A", RequiredTasks: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateDroneTask(ctx, service.RequestMeta{RequestID: "task"}, repository.DroneTaskInput{MissionID: mission.Mission.ID, DroneUnitID: mission.DroneUnitIDs[0], TaskCode: "B-S1", ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.CreateDroneMissionBatch(ctx, service.RequestMeta{RequestID: "missionBatch-create"}, repository.DroneMissionBatchInput{Code: "B-1", Method: "daily-drone", Capacity: 1}, []string{task.ID})
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("err=%v", err)
	}
	var mission_batches int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM mission_batches WHERE code='B-1'`).Scan(&mission_batches); err != nil {
		t.Fatal(err)
	}
	if mission_batches != 0 {
		t.Fatalf("partial missionBatch count=%d", mission_batches)
	}
}

func TestDroneMissionBatchAndAssignmentRejectSkippedStates(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	repo := repository.NewPostgres(pool)
	missionBatchID, err := repo.CreateDroneMissionBatch(ctx, repository.DroneMissionBatchInput{Code: "STATE-B", Method: "daily-drone", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CompleteDroneMissionBatch(ctx, missionBatchID, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("missionBatch err=%v", err)
	}
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repo)
	now := time.Now().UTC()
	mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "assignment"}, service.CreateMissionRequest{
		Code: "ASSIGN-STATE", Name: "Assignment", Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "A-1", RoomLabel: "A", RequiredTasks: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assignment := repository.NewAssignment(mission.Mission.ID, mission.DroneUnitIDs[0], operator, now, now.Add(time.Hour))
	if err := repo.CreateAssignment(ctx, assignment); err != nil {
		t.Fatal(err)
	}
	if err := repo.AdvanceAssignment(ctx, assignment.ID, "completed", 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("assignment err=%v", err)
	}
}

func TestDroneTaskRejectsDroneUnitFromAnotherMission(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	create := func(code string) service.CreateMissionResponse {
		mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: code}, service.CreateMissionRequest{
			Code: code, Name: code, Timezone: "UTC", StartsAt: now, EndsAt: now.Add(time.Hour), CreatedBy: operator,
			DroneUnits: []repository.DroneUnitInput{{Code: code + "-ROBOT", RoomLabel: "A", RequiredTasks: 1}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return mission
	}
	first := create("ROBOT-PLAN-A")
	second := create("ROBOT-PLAN-B")
	_, err := svc.CreateDroneTask(ctx, service.RequestMeta{RequestID: "mismatch"}, repository.DroneTaskInput{
		MissionID: first.Mission.ID, DroneUnitID: second.DroneUnitIDs[0], TaskCode: "MISMATCH-01", ExpiresAt: now.Add(time.Hour),
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error=%v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM drone_tasks WHERE task_code='MISMATCH-01'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("mismatched task persisted: %d", count)
	}
}

func TestExpiredDroneTaskReconciliationWritesAudit(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	operator := insertFleetOperator(t, ctx, pool, "drone_operator")
	svc := service.New(repository.NewPostgres(pool))
	now := time.Now().UTC()
	mission, err := svc.CreateMission(ctx, service.RequestMeta{RequestID: "expiry-mission"}, service.CreateMissionRequest{
		Code: "EXPIRY-PLAN", Name: "Expiry", Timezone: "UTC", StartsAt: now.Add(-time.Hour), EndsAt: now.Add(time.Hour), CreatedBy: operator,
		DroneUnits: []repository.DroneUnitInput{{Code: "EXPIRY-ROBOT", RoomLabel: "A", RequiredTasks: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var taskID string
	if err := pool.QueryRow(ctx, `INSERT INTO drone_tasks(id,mission_id,drone_id,task_code,status,expires_at,version) VALUES (gen_random_uuid(),$1,$2,'EXPIRED-01','queued',$3,1) RETURNING id`, mission.Mission.ID, mission.DroneUnitIDs[0], now.Add(-time.Minute)).Scan(&taskID); err != nil {
		t.Fatal(err)
	}
	result, err := svc.MarkExpiredDroneTasks(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.Marked != 1 || result.Failed != 0 {
		t.Fatalf("result=%+v", result)
	}
	var status string
	var audits int
	if err := pool.QueryRow(ctx, `SELECT status FROM drone_tasks WHERE id=$1`, taskID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE object_id=$1 AND action='expire'`, taskID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.DroneTaskRejected) || audits != 1 {
		t.Fatalf("status=%s audits=%d", status, audits)
	}
}
