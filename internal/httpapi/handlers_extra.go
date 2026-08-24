package httpapi

import (
	"net/http"
	"strconv"

	"drone-operations-control/internal/domain"
)

type operatorRequest struct {
	Name string                   `json:"name"`
	Role domain.FleetOperatorRole `json:"role"`
}

func (a *API) createFleetOperator(w http.ResponseWriter, r *http.Request) {
	var in operatorRequest
	if !decode(w, r, &in) {
		return
	}
	operator, err := a.service.RegisterFleetOperator(r.Context(), meta(r), in.Name, in.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, operator)
}

func (a *API) missionProgress(w http.ResponseWriter, r *http.Request) {
	progress, err := a.service.MissionProgress(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (a *API) missionBatchVersion(w http.ResponseWriter, r *http.Request) (int64, bool) {
	version, err := strconv.ParseInt(r.URL.Query().Get("version"), 10, 64)
	if err != nil || version < 1 {
		writeError(w, domain.ErrConflict)
		return 0, false
	}
	return version, true
}

func (a *API) startDroneMissionBatch(w http.ResponseWriter, r *http.Request) {
	a.changeDroneMissionBatch(w, r, "start")
}
func (a *API) completeDroneMissionBatch(w http.ResponseWriter, r *http.Request) {
	a.changeDroneMissionBatch(w, r, "complete")
}
func (a *API) cancelDroneMissionBatch(w http.ResponseWriter, r *http.Request) {
	a.changeDroneMissionBatch(w, r, "cancel")
}

func (a *API) changeDroneMissionBatch(w http.ResponseWriter, r *http.Request, action string) {
	version, ok := a.missionBatchVersion(w, r)
	if !ok {
		return
	}
	var err error
	switch action {
	case "start":
		err = a.service.StartDroneMissionBatch(r.Context(), meta(r), r.PathValue("id"), version)
	case "complete":
		err = a.service.CompleteDroneMissionBatch(r.Context(), meta(r), r.PathValue("id"), version)
	case "cancel":
		err = a.service.CancelDroneMissionBatch(r.Context(), meta(r), r.PathValue("id"), version)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": action})
}
