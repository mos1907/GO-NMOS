package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// GetReceiverConnections returns current connection state for a receiver.
// Query params: receiver_id (required), state (optional: staged|active), role (optional: master|backup).
func (h *Handler) GetReceiverConnections(w http.ResponseWriter, r *http.Request) {
	receiverID := r.URL.Query().Get("receiver_id")
	if receiverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id is required"})
		return
	}

	conns, err := h.repo.ListReceiverConnections(r.Context(), receiverID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
		return
	}

	writeJSON(w, http.StatusOK, conns)
}

// GetReceiverConnectionHistory returns connection history for a receiver (audit trail).
func (h *Handler) GetReceiverConnectionHistory(w http.ResponseWriter, r *http.Request) {
	receiverID := chi.URLParam(r, "receiver_id")
	if receiverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id is required"})
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	hist, err := h.repo.ListReceiverConnectionHistory(r.Context(), receiverID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list history"})
		return
	}

	writeJSON(w, http.StatusOK, hist)
}

// PutReceiverConnection updates or creates a receiver connection state.
// Body: receiver_id, state (staged|active), role (master|backup), sender_id, flow_id (optional), changed_by (optional).
func (h *Handler) PutReceiverConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ReceiverID string  `json:"receiver_id"`
		State      string  `json:"state"` // staged | active
		Role       string  `json:"role"`  // master | backup
		SenderID   string  `json:"sender_id"`
		FlowID     *int64  `json:"flow_id,omitempty"`
		ChangedBy  string  `json:"changed_by,omitempty"`
		Metadata   map[string]any `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.ReceiverID == "" || req.State == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id, state, and role are required"})
		return
	}

	if req.State != "staged" && req.State != "active" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "state must be 'staged' or 'active'"})
		return
	}

	if req.Role != "master" && req.Role != "backup" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be 'master' or 'backup'"})
		return
	}

	conn := models.ReceiverConnection{
		ReceiverID: req.ReceiverID,
		State:      req.State,
		Role:       req.Role,
		SenderID:   req.SenderID,
		FlowID:     req.FlowID,
		ChangedAt:  time.Now(),
		ChangedBy:  req.ChangedBy,
	}

	if req.Metadata != nil {
		data, _ := json.Marshal(req.Metadata)
		conn.Metadata = data
	}

	// Record history before updating
	hist := models.ReceiverConnectionHistory{
		ReceiverID: req.ReceiverID,
		State:      req.State,
		Role:       req.Role,
		SenderID:   req.SenderID,
		FlowID:     req.FlowID,
		ChangedAt:  time.Now(),
		ChangedBy:  req.ChangedBy,
		Action:     "connect",
	}
	if req.Metadata != nil {
		hist.Metadata = conn.Metadata
	}
	_ = h.repo.RecordReceiverConnectionHistory(r.Context(), hist)

	if err := h.repo.UpsertReceiverConnection(r.Context(), conn); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save connection"})
		return
	}

	writeJSON(w, http.StatusOK, conn)
}

// DeleteReceiverConnection removes a receiver connection (disconnect).
func (h *Handler) DeleteReceiverConnection(w http.ResponseWriter, r *http.Request) {
	receiverID := chi.URLParam(r, "receiver_id")
	state := r.URL.Query().Get("state")
	role := r.URL.Query().Get("role")

	if receiverID == "" || state == "" || role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id, state, and role are required"})
		return
	}

	// Record disconnect in history before deleting
	hist := models.ReceiverConnectionHistory{
		ReceiverID: receiverID,
		State:      state,
		Role:       role,
		ChangedAt:  time.Now(),
		Action:     "disconnect",
	}
	// Username can be extracted from JWT token in middleware if needed; for now leave empty
	_ = h.repo.RecordReceiverConnectionHistory(r.Context(), hist)

	if err := h.repo.DeleteReceiverConnection(r.Context(), receiverID, state, role); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete connection"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
