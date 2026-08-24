package httpapi

import (
	"net/http"
	"time"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

type assignmentRequest struct {
	MissionID       string    `json:"mission_id"`
	DroneUnitID     string    `json:"drone_id"`
	FleetOperatorID string    `json:"operator_id"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
}

func (a *API) createAssignment(w http.ResponseWriter, r *http.Request) {
	var in assignmentRequest
	if !decode(w, r, &in) {
		return
	}
	operator, err := a.service.LoadFleetOperator(r.Context(), in.FleetOperatorID)
	if err != nil {
		writeError(w, err)
		return
	}
	assignment := repository.NewAssignment(in.MissionID, in.DroneUnitID, in.FleetOperatorID, in.StartsAt, in.EndsAt)
	if err := a.service.AssignDroneUnit(r.Context(), meta(r), assignment, operator); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, assignment)
}

type assignmentAdvanceRequest struct {
	Status  string `json:"status"`
	Version int64  `json:"version"`
}

func (a *API) advanceAssignment(w http.ResponseWriter, r *http.Request) {
	var in assignmentAdvanceRequest
	if !decode(w, r, &in) {
		return
	}
	if in.Status != "active" && in.Status != "completed" && in.Status != "cancelled" {
		writeError(w, domain.ErrConflict)
		return
	}
	if err := a.service.AdvanceAssignment(r.Context(), meta(r), r.PathValue("id"), in.Status, in.Version); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": in.Status})
}
