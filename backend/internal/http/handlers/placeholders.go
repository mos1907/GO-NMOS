package handlers

import "net/http"

func (h *Handler) Placeholder(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "this endpoint is reserved for upcoming production modules (NMOS discovery, scheduler, logs, address planner)",
	})
}
