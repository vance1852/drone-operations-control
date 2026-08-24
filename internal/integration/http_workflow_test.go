package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/httpapi"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
)

func apiRequest(t *testing.T, handler http.Handler, method, path string, body any, operatorID string, wantStatus int, target any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "http-contract")
	if operatorID != "" {
		req.Header.Set("X-FleetOperator-ID", operatorID)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, path, res.Code, wantStatus, res.Body.String())
	}
	if res.Header().Get("X-Request-ID") != "http-contract" {
		t.Fatalf("%s %s request id=%q", method, path, res.Header().Get("X-Request-ID"))
	}
	if target != nil {
		if err := json.Unmarshal(res.Body.Bytes(), target); err != nil {
			t.Fatalf("%s %s decode response: %v body=%s", method, path, err, res.Body.String())
		}
	}
}

func createFleetOperatorHTTP(t *testing.T, handler http.Handler, name string, role domain.FleetOperatorRole) domain.FleetOperator {
	t.Helper()
	var operator domain.FleetOperator
	apiRequest(t, handler, http.MethodPost, "/v1/operators", map[string]any{"name": name, "role": role}, "", http.StatusCreated, &operator)
	if operator.ID == "" || operator.Role != role {
		t.Fatalf("operator=%+v", operator)
	}
	return operator
}

func TestHTTPWorkflowCoversPublicBackendRoutes(t *testing.T) {
	pool, ctx := openDatabase(t)
	defer pool.Close()
	now := time.Now().UTC().Truncate(time.Second)
	svc := service.New(repository.NewPostgres(pool)).WithClock(func() time.Time { return now })
	handler := httpapi.New(svc, pool.Ping).Handler()

	apiRequest(t, handler, http.MethodGet, "/healthz", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/readyz", nil, "", http.StatusOK, nil)

	field := createFleetOperatorHTTP(t, handler, "DroneOperator Lin", domain.RoleDroneOperator)
	telemetryOperator := createFleetOperatorHTTP(t, handler, "TelemetryOperator Zhao", domain.RoleTelemetryOperator)
	reviewer := createFleetOperatorHTTP(t, handler, "Safety Reviewer Chen", domain.RoleQualityReviewer)
	apiRequest(t, handler, http.MethodPost, "/v1/operators/"+field.ID+"/rename", map[string]any{"name": "Senior DroneOperator Lin"}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/operators?role=drone_operator", nil, "", http.StatusOK, nil)

	request := map[string]any{
		"code": "HTTP-PLAN", "name": "HTTP workflow", "timezone": "UTC",
		"starts_at": now.Add(-time.Minute), "ends_at": now.Add(24 * time.Hour), "created_by": field.ID,
		"drones": []map[string]any{{"code": "ROBOT-01", "room_label": "A-101", "required_tasks": 2}},
	}
	var mission service.CreateMissionResponse
	apiRequest(t, handler, http.MethodPost, "/v1/missions", request, field.ID, http.StatusCreated, &mission)
	if mission.Mission.ID == "" || len(mission.DroneUnitIDs) != 1 {
		t.Fatalf("mission=%+v", mission)
	}
	missionID, droneID := mission.Mission.ID, mission.DroneUnitIDs[0]
	apiRequest(t, handler, http.MethodGet, "/v1/missions?search=HTTP", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/missions/"+missionID+"/drones", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/missions/"+missionID+"/schedule", map[string]any{"version": 1}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/missions/"+missionID+"/activate", map[string]any{"version": 2}, field.ID, http.StatusOK, nil)

	var assignment domain.Assignment
	apiRequest(t, handler, http.MethodPost, "/v1/assignments", map[string]any{
		"mission_id": missionID, "drone_id": droneID, "operator_id": field.ID,
		"starts_at": now.Add(-time.Minute), "ends_at": now.Add(time.Hour),
	}, field.ID, http.StatusCreated, &assignment)
	apiRequest(t, handler, http.MethodPost, "/v1/assignments/"+assignment.ID+"/advance", map[string]any{"status": "active", "version": 1}, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/assignments/"+assignment.ID+"/advance", map[string]any{"status": "completed", "version": 2}, field.ID, http.StatusOK, nil)

	createDroneTask := func(code string) domain.DroneTask {
		var task domain.DroneTask
		apiRequest(t, handler, http.MethodPost, "/v1/drone_tasks", map[string]any{
			"mission_id": missionID, "drone_id": droneID, "task_code": code, "expires_at": now.Add(12 * time.Hour),
		}, field.ID, http.StatusCreated, &task)
		apiRequest(t, handler, http.MethodPost, "/v1/drone_tasks/"+task.ID+"/complete", map[string]any{"version": 1}, field.ID, http.StatusOK, nil)
		apiRequest(t, handler, http.MethodPost, "/v1/drone_tasks/"+task.ID+"/device_transfer", map[string]any{
			"to_operator": field.ID, "location": "A-101", "recorded_at": now, "version": 2,
		}, field.ID, http.StatusOK, nil)
		apiRequest(t, handler, http.MethodPost, "/v1/drone_tasks/"+task.ID+"/accept", map[string]any{
			"from_operator": field.ID, "to_operator": telemetryOperator.ID, "location": "East drone station", "recorded_at": now.Add(time.Minute), "version": 3,
		}, telemetryOperator.ID, http.StatusOK, nil)
		return task
	}

	first := createDroneTask("HTTP-TASK-01")
	apiRequest(t, handler, http.MethodGet, "/v1/drone_tasks?mission_id="+missionID+"&status=accepted", nil, "", http.StatusOK, nil)
	var missionBatch map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/mission_batches", map[string]any{
		"code": "HTTP-ROUND-01", "method": "evening-drone", "capacity": 1, "drone_task_ids": []string{first.ID},
	}, telemetryOperator.ID, http.StatusCreated, &missionBatch)
	missionBatchID := missionBatch["id"]
	apiRequest(t, handler, http.MethodPost, "/v1/mission_batches/"+missionBatchID+"/start?version=1", nil, telemetryOperator.ID, http.StatusOK, nil)
	var telemetry map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/telemetry", map[string]any{
		"drone_task_id": first.ID, "mission_batch_id": missionBatchID, "recorded_by": telemetryOperator.ID,
		"risk_score": 3.5, "scale": "drone-risk", "alert_threshold": 5.0, "observed_at": now,
	}, telemetryOperator.ID, http.StatusCreated, &telemetry)
	apiRequest(t, handler, http.MethodPost, "/v1/telemetry/"+telemetry["id"]+"/review", map[string]any{
		"drone_task_id": first.ID, "accepted": true, "telemetry_version": 1, "task_version": 5,
	}, reviewer.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/mission_batches/"+missionBatchID+"/complete?version=2", nil, telemetryOperator.ID, http.StatusOK, nil)

	var safety_alert map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/safety_alerts", map[string]any{
		"drone_task_id": first.ID, "kind": "close_record", "reason": "scheduled record closure", "due_at": now.Add(time.Hour),
	}, field.ID, http.StatusCreated, &safety_alert)
	apiRequest(t, handler, http.MethodPost, "/v1/safety_alerts/"+safety_alert["id"]+"/start", nil, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/safety_alerts/"+safety_alert["id"]+"/close", nil, field.ID, http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/drone_tasks/"+first.ID+"/archive", map[string]any{"version": 6}, field.ID, http.StatusOK, nil)

	second := createDroneTask("HTTP-TASK-02")
	var cancelDroneMissionBatch map[string]string
	apiRequest(t, handler, http.MethodPost, "/v1/mission_batches", map[string]any{
		"code": "HTTP-ROUND-02", "method": "evening-drone", "capacity": 1, "drone_task_ids": []string{second.ID},
	}, telemetryOperator.ID, http.StatusCreated, &cancelDroneMissionBatch)
	apiRequest(t, handler, http.MethodPost, "/v1/mission_batches/"+cancelDroneMissionBatch["id"]+"/cancel?version=1", nil, telemetryOperator.ID, http.StatusOK, nil)
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM drone_tasks WHERE id=$1`, second.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != string(domain.DroneTaskAccepted) {
		t.Fatalf("cancelled missionBatch left task status=%s", status)
	}
	var collected int
	if err := pool.QueryRow(ctx, `SELECT completed_tasks FROM mission_drones WHERE id=$1`, droneID).Scan(&collected); err != nil {
		t.Fatal(err)
	}
	if collected != 2 {
		t.Fatalf("drone completed_tasks=%d want=2", collected)
	}

	apiRequest(t, handler, http.MethodGet, "/v1/missions/"+missionID+"/progress", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, "/v1/missions/"+missionID+"/report", nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodGet, fmt.Sprintf("/v1/audit/drone_task/%s?limit=20", first.ID), nil, "", http.StatusOK, nil)
	apiRequest(t, handler, http.MethodPost, "/v1/missions/"+missionID+"/close", map[string]any{"version": 3}, field.ID, http.StatusOK, nil)
}
