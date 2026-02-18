package handlers

import "net/http"

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	overallStatus := "ok"
	statusCode := http.StatusOK
	if err := h.repo.HealthCheck(r.Context()); err != nil {
		dbStatus = "error"
		overallStatus = "degraded"
		statusCode = http.StatusServiceUnavailable
	}
	writeJSON(w, statusCode, map[string]any{
		"status":  overallStatus,
		"service": "go-NMOS",
		"db":      dbStatus,
	})
}
