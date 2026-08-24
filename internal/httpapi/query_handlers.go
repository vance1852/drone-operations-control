package httpapi

import (
	"net/http"
	"strconv"

	"drone-operations-control/internal/domain"
	"drone-operations-control/internal/repository"
)

func (a *API) listMissions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := a.service.ListMissions(r.Context(), repository.MissionFilter{Status: domain.MissionStatus(r.URL.Query().Get("status")), Search: r.URL.Query().Get("search"), Limit: limit, Offset: offset})
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (a *API) listMissionDroneUnits(w http.ResponseWriter, r *http.Request) {
	items, err := a.service.ListMissionDroneUnits(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (a *API) auditHistory(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.service.AuditHistory(r.Context(), r.PathValue("object_type"), r.PathValue("object_id"), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}
