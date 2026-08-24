package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	"drone-operations-control/internal/repository"
)

func TestRepositoryInputsDecodePublicJSONFields(t *testing.T) {
	observedAt := "2026-08-18T09:00:00Z"

	var task repository.DroneTaskInput
	decodeContractJSON(t, `{"mission_id":"mission-1","drone_id":"drone-1","task_code":"S-1","expires_at":"2026-08-18T10:00:00Z"}`, &task)
	if task.MissionID != "mission-1" || task.DroneUnitID != "drone-1" || task.TaskCode != "S-1" || task.ExpiresAt.IsZero() {
		t.Fatalf("task input = %+v", task)
	}

	var missionBatch missionBatchRequest
	decodeContractJSON(t, `{"code":"ROUND-1","method":"evening-drone","capacity":2,"drone_task_ids":["task-1"]}`, &missionBatch)
	if missionBatch.Code != "ROUND-1" || missionBatch.Method != "evening-drone" || missionBatch.Capacity != 2 || len(missionBatch.DroneTaskIDs) != 1 {
		t.Fatalf("missionBatch input = %+v", missionBatch)
	}

	var telemetry repository.ObservationInput
	decodeContractJSON(t, `{"drone_task_id":"task-1","mission_batch_id":"round-1","recorded_by":"telemetryOperator-1","risk_score":2.5,"scale":"drone-risk","alert_threshold":5,"observed_at":"`+observedAt+`"}`, &telemetry)
	if telemetry.DroneTaskID != "task-1" || telemetry.DroneMissionBatchID != "round-1" || telemetry.RecorderID != "telemetryOperator-1" || telemetry.ObservedAt.IsZero() {
		t.Fatalf("telemetry input = %+v", telemetry)
	}

	var safety_alert repository.InterventionInput
	decodeContractJSON(t, `{"drone_task_id":"task-1","kind":"repeat_drone","reason":"verification","due_at":"2026-08-18T11:00:00Z"}`, &safety_alert)
	if safety_alert.DroneTaskID != "task-1" || safety_alert.Kind != "repeat_drone" || safety_alert.DueAt.IsZero() {
		t.Fatalf("safety_alert input = %+v", safety_alert)
	}

	if telemetry.ObservedAt.Format(time.RFC3339) != observedAt {
		t.Fatalf("observed_at = %s", telemetry.ObservedAt.Format(time.RFC3339))
	}
}

func decodeContractJSON(t *testing.T, value string, destination any) {
	t.Helper()
	if err := json.Unmarshal([]byte(value), destination); err != nil {
		t.Fatal(err)
	}
}
