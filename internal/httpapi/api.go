package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drone-operations-control/internal/console"
	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
	"github.com/google/uuid"
)

type API struct {
	service      *service.Service
	ready        func(context.Context) error
	consoleStore *console.Store
}

func New(svc *service.Service, ready func(context.Context) error) *API {
	return &API{service: svc, ready: ready}
}

func (a *API) WithConsole(store *console.Store) *API {
	a.consoleStore = store
	return a
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /readyz", a.readyz)
	mux.HandleFunc("POST /v1/missions", a.createMission)
	mux.HandleFunc("POST /v1/missions/{id}/schedule", a.scheduleMission)
	mux.HandleFunc("POST /v1/missions/{id}/activate", a.activateMission)
	mux.HandleFunc("POST /v1/missions/{id}/close", a.closeMission)
	mux.HandleFunc("GET /v1/missions/{id}/progress", a.missionProgress)
	mux.HandleFunc("GET /v1/missions", a.listMissions)
	mux.HandleFunc("GET /v1/missions/{id}/drones", a.listMissionDroneUnits)
	mux.HandleFunc("GET /v1/missions/{id}/report", a.complianceReport)
	mux.HandleFunc("POST /v1/operators", a.createFleetOperator)
	mux.HandleFunc("GET /v1/operators", a.listFleetOperators)
	mux.HandleFunc("POST /v1/operators/{id}/rename", a.renameFleetOperator)
	mux.HandleFunc("POST /v1/assignments", a.createAssignment)
	mux.HandleFunc("POST /v1/assignments/{id}/advance", a.advanceAssignment)
	mux.HandleFunc("POST /v1/drone_tasks", a.createDroneTask)
	mux.HandleFunc("POST /v1/drone_tasks/{id}/complete", a.completeDroneTask)
	mux.HandleFunc("POST /v1/drone_tasks/{id}/device_transfer", a.transferDroneTask)
	mux.HandleFunc("POST /v1/drone_tasks/{id}/accept", a.acceptDroneTask)
	mux.HandleFunc("POST /v1/drone_tasks/{id}/archive", a.archiveDroneTask)
	mux.HandleFunc("GET /v1/drone_tasks", a.listDroneTasks)
	mux.HandleFunc("POST /v1/mission_batches", a.createDroneMissionBatch)
	mux.HandleFunc("POST /v1/mission_batches/{id}/start", a.startDroneMissionBatch)
	mux.HandleFunc("POST /v1/mission_batches/{id}/complete", a.completeDroneMissionBatch)
	mux.HandleFunc("POST /v1/mission_batches/{id}/cancel", a.cancelDroneMissionBatch)
	mux.HandleFunc("POST /v1/telemetry", a.submitObservation)
	mux.HandleFunc("POST /v1/telemetry/{id}/review", a.reviewObservation)
	mux.HandleFunc("POST /v1/safety_alerts", a.openIntervention)
	mux.HandleFunc("POST /v1/safety_alerts/{id}/start", a.startIntervention)
	mux.HandleFunc("POST /v1/safety_alerts/{id}/close", a.closeIntervention)
	mux.HandleFunc("GET /v1/audit", a.queryAudit)
	mux.HandleFunc("GET /v1/audit/{object_type}/{object_id}", a.auditHistory)
	if a.consoleStore != nil {
		a.registerConsoleRoutes(mux)
	}
	return requestMiddleware(mux)
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (a *API) readyz(w http.ResponseWriter, r *http.Request) {
	if a.ready != nil {
		if err := a.ready(r.Context()); err != nil {
			writeError(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type missionRequest struct {
	Code       string                      `json:"code"`
	Name       string                      `json:"name"`
	Timezone   string                      `json:"timezone"`
	StartsAt   time.Time                   `json:"starts_at"`
	EndsAt     time.Time                   `json:"ends_at"`
	CreatedBy  string                      `json:"created_by"`
	DroneUnits []repository.DroneUnitInput `json:"drones"`
}

func (a *API) createMission(w http.ResponseWriter, r *http.Request) {
	var in missionRequest
	if !decode(w, r, &in) {
		return
	}
	result, err := a.service.CreateMission(r.Context(), meta(r), service.CreateMissionRequest{Code: in.Code, Name: in.Name, Timezone: in.Timezone, StartsAt: in.StartsAt, EndsAt: in.EndsAt, CreatedBy: in.CreatedBy, DroneUnits: in.DroneUnits})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

type versionRequest struct {
	Version int64 `json:"version"`
}

func (a *API) scheduleMission(w http.ResponseWriter, r *http.Request) {
	a.advanceMission(w, r, domain.MissionScheduled)
}
func (a *API) activateMission(w http.ResponseWriter, r *http.Request) {
	a.advanceMission(w, r, domain.MissionActive)
}
func (a *API) closeMission(w http.ResponseWriter, r *http.Request) {
	a.advanceMission(w, r, domain.MissionClosed)
}

func (a *API) advanceMission(w http.ResponseWriter, r *http.Request, next domain.MissionStatus) {
	var in versionRequest
	if !decode(w, r, &in) {
		return
	}
	var err error
	switch next {
	case domain.MissionScheduled:
		err = a.service.ScheduleMission(r.Context(), meta(r), r.PathValue("id"), in.Version)
	case domain.MissionActive:
		err = a.service.ActivateMission(r.Context(), meta(r), r.PathValue("id"), in.Version)
	case domain.MissionClosed:
		err = a.service.CloseMission(r.Context(), meta(r), r.PathValue("id"), in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) createDroneTask(w http.ResponseWriter, r *http.Request) {
	var in repository.DroneTaskInput
	if !decode(w, r, &in) {
		return
	}
	result, err := a.service.CreateDroneTask(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) completeDroneTask(w http.ResponseWriter, r *http.Request) {
	a.moveDroneTask(w, r, domain.DroneTaskCompleted)
}

func (a *API) archiveDroneTask(w http.ResponseWriter, r *http.Request) {
	a.moveDroneTask(w, r, domain.DroneTaskArchived)
}

func (a *API) moveDroneTask(w http.ResponseWriter, r *http.Request, next domain.DroneTaskStatus) {
	var in versionRequest
	if !decode(w, r, &in) {
		return
	}
	var err error
	if next == domain.DroneTaskCompleted {
		err = a.service.CompleteDroneTask(r.Context(), meta(r), r.PathValue("id"), in.Version)
	} else {
		err = a.service.ArchiveDroneTask(r.Context(), meta(r), r.PathValue("id"), in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) transferDroneTask(w http.ResponseWriter, r *http.Request) {
	a.device_transfer(w, r, domain.DroneTaskHandoffPending)
}

func (a *API) acceptDroneTask(w http.ResponseWriter, r *http.Request) {
	a.device_transfer(w, r, domain.DroneTaskAccepted)
}

type device_transferRequest struct {
	From       *string   `json:"from_operator"`
	To         string    `json:"to_operator"`
	Location   string    `json:"location"`
	RecordedAt time.Time `json:"recorded_at"`
	Note       string    `json:"note"`
	Version    int64     `json:"version"`
}

func (a *API) device_transfer(w http.ResponseWriter, r *http.Request, next domain.DroneTaskStatus) {
	var in device_transferRequest
	if !decode(w, r, &in) {
		return
	}
	if in.RecordedAt.IsZero() {
		in.RecordedAt = time.Now().UTC()
	}
	input := repository.HandoffInput{DroneTaskID: r.PathValue("id"), From: in.From, To: in.To, Location: in.Location, RecordedAt: in.RecordedAt, Note: in.Note}
	var err error
	if next == domain.DroneTaskHandoffPending {
		err = a.service.HandoffChecked(r.Context(), meta(r), input, in.Version)
	} else {
		err = a.service.AcceptChecked(r.Context(), meta(r), input, in.Version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(next)})
}

func (a *API) listDroneTasks(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	page, err := a.service.ListDroneTasks(r.Context(), offset, limit, r.URL.Query().Get("mission_id"), domain.DroneTaskStatus(r.URL.Query().Get("status")))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

type missionBatchRequest struct {
	repository.DroneMissionBatchInput
	DroneTaskIDs []string `json:"drone_task_ids"`
}

func (a *API) createDroneMissionBatch(w http.ResponseWriter, r *http.Request) {
	var in missionBatchRequest
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.CreateDroneMissionBatch(r.Context(), meta(r), in.DroneMissionBatchInput, append([]string(nil), in.DroneTaskIDs...))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) submitObservation(w http.ResponseWriter, r *http.Request) {
	var in repository.ObservationInput
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.SubmitObservation(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

type reviewRequest struct {
	DroneTaskID        string `json:"drone_task_id"`
	Accepted           bool   `json:"accepted"`
	ObservationVersion int64  `json:"telemetry_version"`
	DroneTaskVersion   int64  `json:"task_version"`
}

func (a *API) reviewObservation(w http.ResponseWriter, r *http.Request) {
	var in reviewRequest
	if !decode(w, r, &in) {
		return
	}
	err := a.service.ReviewObservation(r.Context(), meta(r), r.PathValue("id"), in.DroneTaskID, in.Accepted, in.ObservationVersion, in.DroneTaskVersion)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reviewed"})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, fmt.Errorf("invalid json: %w", domain.ErrConflict))
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, fmt.Errorf("request body must contain one json value: %w", domain.ErrConflict))
		return false
	}
	return true
}

func meta(r *http.Request) service.RequestMeta {
	operator := strings.TrimSpace(r.Header.Get("X-FleetOperator-ID"))
	var operatorID *string
	if _, err := uuid.Parse(operator); err == nil {
		operatorID = &operator
	}
	return service.RequestMeta{RequestID: requestID(r.Context()), FleetOperatorID: operatorID}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrInvalidTransition):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, domain.ErrCapacityExceeded):
		status, code = http.StatusUnprocessableEntity, "capacity_exceeded"
	case errors.Is(err, domain.ErrExpired):
		status, code = http.StatusUnprocessableEntity, "expired"
	}
	writeJSON(w, status, map[string]string{"code": code, "message": err.Error(), "request_id": requestIDFromWriter(w)})
}

func requestIDFromWriter(w http.ResponseWriter) string { return w.Header().Get("X-Request-ID") }
