package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-operations-control/internal/console"
	"drone-operations-control/internal/httpapi"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
)

type consoleResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func consoleRequest(t *testing.T, handler http.Handler, method, path string, body any, target any) {
	t.Helper()
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, payload)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("%s %s status=%d body=%s", method, path, response.Code, response.Body.String())
	}
	var envelope consoleResponse
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Code != 200 || envelope.Message != "success" {
		t.Fatalf("response=%+v", envelope)
	}
	if target != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, target); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOriginalConsolePagesUseGoAndPostgreSQLContracts(t *testing.T) {
	pool, _ := openDatabase(t)
	defer pool.Close()
	store := console.NewStore(pool)
	handler := httpapi.New(service.New(repository.NewPostgres(pool)), pool.Ping).WithConsole(store).Handler()

	var login struct {
		Token string       `json:"token"`
		User  console.User `json:"user"`
	}
	consoleRequest(t, handler, http.MethodPost, "/api/auth/login", map[string]any{"username": "admin", "password": "admin123"}, &login)
	if login.Token == "" || login.User.RealName != "系统管理员" {
		t.Fatalf("login=%+v", login)
	}

	drone := console.Drone{Name: "接口测试无人机设备", ModelClass: 1, SerialNumber: "RBT-TEST-001", Endpoint: "10.0.0.18", TelemetryStatus: 1, Status: 1}
	consoleRequest(t, handler, http.MethodPost, "/api/drone", drone, &drone)
	if drone.ID == "" {
		t.Fatal("drone id is empty")
	}

	operator := console.Operator{Name: "接口测试飞行调度员", Gender: 2, Phone: "13800139999", Skills: "日常飞行任务", Status: 1}
	consoleRequest(t, handler, http.MethodPost, "/api/operator", operator, &operator)
	capabilityItem := console.CapabilityItem{Name: "接口测试服务", Description: "飞行任务流程验证", Price: 80, Duration: 30, Status: 1}
	consoleRequest(t, handler, http.MethodPost, "/api/capability", capabilityItem, &capabilityItem)

	appointment := time.Now().UTC().Add(time.Hour).Format("2006-01-02 15:04:05")
	var command console.Command
	consoleRequest(t, handler, http.MethodPost, "/api/command", map[string]any{
		"droneId": drone.ID, "capabilityId": capabilityItem.ID, "operatorId": operator.ID,
		"appointmentTime": appointment, "remark": "接口测试",
	}, &command)
	if command.ID == "" || command.Status != 0 {
		t.Fatalf("command=%+v", command)
	}
	consoleRequest(t, handler, http.MethodPut, "/api/command/status", map[string]any{"id": command.ID, "status": 1}, nil)

	battery, motorTemperature, jointLoad := 82.0, 36.5, 0.42
	record := console.TelemetryRecord{DroneID: drone.ID, BatteryLevel: &battery, MotorTemperature: &motorTemperature, JointLoad: &jointLoad, Remark: "状态平稳"}
	consoleRequest(t, handler, http.MethodPost, "/api/telemetry", record, &record)
	if record.ID == "" {
		t.Fatal("telemetry record id is empty")
	}

	var page console.Page[console.Command]
	consoleRequest(t, handler, http.MethodGet, "/api/command/page?current=1&size=20", nil, &page)
	if page.Total < 1 {
		t.Fatalf("command page=%+v", page)
	}
	var stats console.DashboardStats
	consoleRequest(t, handler, http.MethodGet, "/api/dashboard/stats", nil, &stats)
	if stats.DroneCount < 1 || stats.OperatorCount < 1 || stats.PendingCommands < 1 {
		t.Fatalf("stats=%+v", stats)
	}
}
