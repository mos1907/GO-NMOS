package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) CheckerCollisions(w http.ResponseWriter, r *http.Request) {
	collisions, err := h.repo.DetectCollisions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "collision check failed"})
		return
	}
	payload := map[string]any{
		"total_collisions": len(collisions),
		"items":            collisions,
	}
	raw, _ := json.Marshal(payload)
	_ = h.repo.SaveCheckerResult(r.Context(), "collisions", raw)
	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) CheckerLatest(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "collisions"
	}
	result, err := h.repo.GetLatestCheckerResult(r.Context(), kind)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusOK, map[string]any{
				"kind":       kind,
				"result":     map[string]any{},
				"created_at": nil,
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "latest check fetch failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
