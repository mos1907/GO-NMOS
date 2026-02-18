package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]string{}
	keys := []string{"api_base_url", "anonymous_access", "flow_lock_role", "hard_delete_enabled"}
	for _, key := range keys {
		value, err := h.repo.GetSetting(r.Context(), key)
		if err != nil {
			continue
		}
		settings[key] = value
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) PatchSetting(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.repo.SetSetting(r.Context(), key, req.Value); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
