package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// CreateScheduledActivation creates a new scheduled activation (time-based IS-05 patch).
// POST /api/nmos/scheduled-activations
func (h *Handler) CreateScheduledActivation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FlowID      int64    `json:"flow_id"`
		ReceiverIDs []string `json:"receiver_ids"`
		IS05BaseURL string   `json:"is05_base_url"`
		SenderID    string   `json:"sender_id,omitempty"`
		ScheduledAt string   `json:"scheduled_at"` // ISO 8601 timestamp
		Mode        string   `json:"mode,omitempty"` // "immediate" | "safe_switch"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.FlowID == 0 || len(req.ReceiverIDs) == 0 || req.IS05BaseURL == "" || req.ScheduledAt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "flow_id, receiver_ids, is05_base_url, and scheduled_at are required"})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scheduled_at must be RFC3339 format"})
		return
	}

	if scheduledAt.Before(time.Now()) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scheduled_at must be in the future"})
		return
	}

	if req.Mode == "" {
		req.Mode = "immediate"
	}
	if req.Mode != "immediate" && req.Mode != "safe_switch" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mode must be 'immediate' or 'safe_switch'"})
		return
	}

	act := models.ScheduledActivation{
		FlowID:      req.FlowID,
		ReceiverIDs: req.ReceiverIDs,
		IS05BaseURL: req.IS05BaseURL,
		SenderID:    req.SenderID,
		ScheduledAt: scheduledAt,
		Status:      "pending",
		Mode:        req.Mode,
		Result:      json.RawMessage(`{}`),
	}

	if claims := r.Context().Value("claims"); claims != nil {
		if authClaims, ok := claims.(*AuthClaims); ok && authClaims.Username != "" {
			act.CreatedBy = authClaims.Username
		}
	}

	id, err := h.repo.CreateScheduledActivation(r.Context(), act)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create scheduled activation"})
		return
	}

	act.ID = id
	writeJSON(w, http.StatusOK, act)
}

// ListScheduledActivations lists scheduled activations.
// GET /api/nmos/scheduled-activations?status=pending&limit=50
func (h *Handler) ListScheduledActivations(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	acts, err := h.repo.ListScheduledActivations(r.Context(), status, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list scheduled activations"})
		return
	}

	writeJSON(w, http.StatusOK, acts)
}

// GetScheduledActivation gets a single scheduled activation by ID.
// GET /api/nmos/scheduled-activations/{id}
func (h *Handler) GetScheduledActivation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	act, err := h.repo.GetScheduledActivation(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "scheduled activation not found"})
		return
	}

	writeJSON(w, http.StatusOK, act)
}

// DeleteScheduledActivation cancels/deletes a scheduled activation.
// DELETE /api/nmos/scheduled-activations/{id}
func (h *Handler) DeleteScheduledActivation(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	act, err := h.repo.GetScheduledActivation(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "scheduled activation not found"})
		return
	}

	if act.Status != "pending" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "can only cancel pending activations"})
		return
	}

	if err := h.repo.UpdateScheduledActivation(r.Context(), id, map[string]any{"status": "cancelled"}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to cancel scheduled activation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
