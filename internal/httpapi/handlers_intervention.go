package httpapi

import (
	"net/http"

	"drone-operations-control/internal/repository"
)

func (a *API) openIntervention(w http.ResponseWriter, r *http.Request) {
	var in repository.InterventionInput
	if !decode(w, r, &in) {
		return
	}
	id, err := a.service.OpenIntervention(r.Context(), meta(r), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) closeIntervention(w http.ResponseWriter, r *http.Request) {
	if err := a.service.CloseIntervention(r.Context(), meta(r), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (a *API) startIntervention(w http.ResponseWriter, r *http.Request) {
	if err := a.service.MarkInterventionInProgress(r.Context(), meta(r), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "in_progress"})
}
