package httpapi

import (
	"net/http"

	"drone-operations-control/internal/domain"
)

func (a *API) listFleetOperators(w http.ResponseWriter, r *http.Request) {
	items, total, err := a.service.ListFleetOperators(r.Context(), domain.FleetOperatorRole(r.URL.Query().Get("role")), 50, 0)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

type renameFleetOperatorRequest struct {
	Name string `json:"name"`
}

func (a *API) renameFleetOperator(w http.ResponseWriter, r *http.Request) {
	var in renameFleetOperatorRequest
	if !decode(w, r, &in) {
		return
	}
	if err := a.service.RenameFleetOperator(r.Context(), meta(r), r.PathValue("id"), in.Name); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "renamed"})
}
