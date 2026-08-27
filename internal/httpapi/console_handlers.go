package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drone-operations-control/internal/console"
)

func (a *API) registerConsoleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", a.consoleLogin)
	mux.HandleFunc("GET /api/auth/info", a.consoleUserInfo)
	mux.HandleFunc("POST /api/auth/logout", a.consoleLogout)
	mux.HandleFunc("GET /api/dashboard/stats", a.consoleDashboard)
	mux.HandleFunc("GET /api/drone/page", a.consoleDronePage)
	mux.HandleFunc("GET /api/drone/list", a.consoleDroneList)
	mux.HandleFunc("GET /api/drone/{id}", a.consoleDroneByID)
	mux.HandleFunc("POST /api/drone", a.consoleCreateDrone)
	mux.HandleFunc("PUT /api/drone", a.consoleUpdateDrone)
	mux.HandleFunc("DELETE /api/drone/{id}", a.consoleDeleteDrone)
	mux.HandleFunc("GET /api/operator/page", a.consoleOperatorPage)
	mux.HandleFunc("GET /api/operator/list", a.consoleOperatorList)
	mux.HandleFunc("POST /api/operator", a.consoleCreateOperator)
	mux.HandleFunc("PUT /api/operator", a.consoleUpdateOperator)
	mux.HandleFunc("DELETE /api/operator/{id}", a.consoleDeleteOperator)
	mux.HandleFunc("GET /api/capability/page", a.consoleCapabilityPage)
	mux.HandleFunc("GET /api/capability/list", a.consoleCapabilityList)
	mux.HandleFunc("POST /api/capability", a.consoleCreateCapability)
	mux.HandleFunc("PUT /api/capability", a.consoleUpdateCapability)
	mux.HandleFunc("DELETE /api/capability/{id}", a.consoleDeleteCapability)
	mux.HandleFunc("GET /api/command/page", a.consoleCommandPage)
	mux.HandleFunc("POST /api/command", a.consoleCreateCommand)
	mux.HandleFunc("PUT /api/command/status", a.consoleUpdateCommandStatus)
	mux.HandleFunc("GET /api/telemetry/page", a.consoleTelemetryPage)
	mux.HandleFunc("GET /api/telemetry/drone/{droneId}", a.consoleTelemetryByDrone)
	mux.HandleFunc("POST /api/telemetry", a.consoleCreateTelemetry)
	mux.HandleFunc("GET /api/log/page", a.consoleLogPage)
}

type consoleEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func consoleSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(consoleEnvelope{Code: 200, Message: "success", Data: data})
}

func consoleError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(consoleEnvelope{Code: status, Message: err.Error()})
}

func decodeConsole(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		consoleError(w, http.StatusBadRequest, fmt.Errorf("请求数据格式错误"))
		return false
	}
	return true
}

func consolePageParams(r *http.Request) (int, int) {
	current, _ := strconv.Atoi(r.URL.Query().Get("current"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	return current, size
}

func consoleClientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if found {
		return host
	}
	return r.RemoteAddr
}

func (a *API) recordConsoleOperation(r *http.Request, operation string) {
	_ = a.consoleStore.WriteLog(r.Context(), operation, r.Method+" "+r.URL.Path, consoleClientIP(r))
}

func (a *API) consoleLogin(w http.ResponseWriter, r *http.Request) {
	var request struct{ Username, Password string }
	if !decodeConsole(w, r, &request) {
		return
	}
	user, token, expiresAt, err := a.consoleStore.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		consoleError(w, http.StatusUnauthorized, err)
		return
	}
	a.recordConsoleOperation(r, "用户登录")
	consoleSuccess(w, map[string]any{"token": token, "expiresAt": expiresAt, "user": user})
}

func (a *API) consoleUserInfo(w http.ResponseWriter, r *http.Request) {
	user, err := a.consoleStore.SessionUser(r.Context(), bearerToken(r))
	if err != nil {
		consoleError(w, http.StatusUnauthorized, err)
		return
	}
	consoleSuccess(w, user)
}

func (a *API) consoleLogout(w http.ResponseWriter, r *http.Request) {
	if err := a.consoleStore.RevokeSession(r.Context(), bearerToken(r)); err != nil {
		consoleError(w, http.StatusUnauthorized, err)
		return
	}
	consoleSuccess(w, nil)
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(value) > 7 && strings.EqualFold(value[:7], "Bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func (a *API) consoleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := a.consoleStore.Dashboard(r.Context())
	if err != nil {
		consoleError(w, http.StatusInternalServerError, err)
		return
	}
	consoleSuccess(w, stats)
}

func (a *API) consoleDronePage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.DronePage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleDroneList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.DroneList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleDroneByID(w http.ResponseWriter, r *http.Request) {
	item, err := a.consoleStore.DroneByID(r.Context(), r.PathValue("id"))
	consoleWriteResult(w, item, err)
}

func (a *API) consoleCreateDrone(w http.ResponseWriter, r *http.Request) {
	var item console.Drone
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	created, err := a.consoleStore.CreateDrone(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增无人机设备")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateDrone(w http.ResponseWriter, r *http.Request) {
	var item console.Drone
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	err := a.consoleStore.UpdateDrone(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑无人机设备")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteDrone(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteDrone(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除无人机设备")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleOperatorPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.OperatorPage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleOperatorList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.OperatorList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleCreateOperator(w http.ResponseWriter, r *http.Request) {
	var item console.Operator
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	created, err := a.consoleStore.CreateOperator(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增飞行调度员")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateOperator(w http.ResponseWriter, r *http.Request) {
	var item console.Operator
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	err := a.consoleStore.UpdateOperator(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑飞行调度员")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteOperator(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteOperator(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除飞行调度员")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleCapabilityPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.CapabilityPage(r.Context(), current, size, r.URL.Query().Get("keyword"))
	consoleWriteResult(w, page, err)
}

func (a *API) consoleCapabilityList(w http.ResponseWriter, r *http.Request) {
	items, err := a.consoleStore.CapabilityList(r.Context())
	consoleWriteResult(w, items, err)
}

func (a *API) consoleCreateCapability(w http.ResponseWriter, r *http.Request) {
	var item console.CapabilityItem
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	created, err := a.consoleStore.CreateCapability(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增服务")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateCapability(w http.ResponseWriter, r *http.Request) {
	var item console.CapabilityItem
	if !decodeConsole(w, r, &item) || !validateConsoleName(w, item.Name) {
		return
	}
	err := a.consoleStore.UpdateCapability(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "编辑服务")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleDeleteCapability(w http.ResponseWriter, r *http.Request) {
	err := a.consoleStore.DeleteCapability(r.Context(), r.PathValue("id"))
	if err == nil {
		a.recordConsoleOperation(r, "删除服务")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleCommandPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.CommandPage(r.Context(), current, size)
	consoleWriteResult(w, page, err)
}

type consoleCommandRequest struct {
	DroneID         string  `json:"droneId"`
	CapabilityID    string  `json:"capabilityId"`
	OperatorID      *string `json:"operatorId"`
	AppointmentTime string  `json:"appointmentTime"`
	Remark          string  `json:"remark"`
}

func (a *API) consoleCreateCommand(w http.ResponseWriter, r *http.Request) {
	var request consoleCommandRequest
	if !decodeConsole(w, r, &request) {
		return
	}
	appointment, err := parseConsoleTime(request.AppointmentTime)
	if err != nil {
		consoleError(w, http.StatusBadRequest, err)
		return
	}
	created, err := a.consoleStore.CreateCommand(r.Context(), console.Command{DroneID: request.DroneID, CapabilityID: request.CapabilityID, OperatorID: request.OperatorID, AppointmentTime: appointment, Remark: request.Remark})
	if err == nil {
		a.recordConsoleOperation(r, "创建订单")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleUpdateCommandStatus(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
	}
	if !decodeConsole(w, r, &request) {
		return
	}
	err := a.consoleStore.UpdateCommandStatus(r.Context(), request.ID, request.Status)
	if err == nil {
		a.recordConsoleOperation(r, "更新订单状态")
	}
	consoleWriteResult(w, nil, err)
}

func (a *API) consoleTelemetryPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.TelemetryPage(r.Context(), current, size, "")
	consoleWriteResult(w, page, err)
}

func (a *API) consoleTelemetryByDrone(w http.ResponseWriter, r *http.Request) {
	page, err := a.consoleStore.TelemetryPage(r.Context(), 1, 100, r.PathValue("droneId"))
	consoleWriteResult(w, page.Records, err)
}

func (a *API) consoleCreateTelemetry(w http.ResponseWriter, r *http.Request) {
	var item console.TelemetryRecord
	if !decodeConsole(w, r, &item) {
		return
	}
	created, err := a.consoleStore.CreateTelemetry(r.Context(), item)
	if err == nil {
		a.recordConsoleOperation(r, "新增健康记录")
	}
	consoleWriteResult(w, created, err)
}

func (a *API) consoleLogPage(w http.ResponseWriter, r *http.Request) {
	current, size := consolePageParams(r)
	page, err := a.consoleStore.LogPage(r.Context(), current, size)
	consoleWriteResult(w, page, err)
}

func consoleWriteResult(w http.ResponseWriter, data any, err error) {
	if err != nil {
		consoleError(w, http.StatusConflict, err)
		return
	}
	consoleSuccess(w, data)
}

func validateConsoleName(w http.ResponseWriter, name string) bool {
	if strings.TrimSpace(name) == "" {
		consoleError(w, http.StatusBadRequest, fmt.Errorf("名称不能为空"))
		return false
	}
	return true
}

func parseConsoleTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
	}
	return nil, fmt.Errorf("预约时间格式错误")
}
