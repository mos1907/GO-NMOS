package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// CreateScheduledPlaybookExecution schedules a playbook execution (E.2).
// POST /playbooks/{id}/schedule
func (h *Handler) CreateScheduledPlaybookExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	playbookID := chi.URLParam(r, "id")

	var req struct {
		Parameters  map[string]any `json:"parameters"`
		ScheduledAt string          `json:"scheduled_at"` // ISO 8601 timestamp
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.ScheduledAt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "scheduled_at is required"})
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

	// Verify playbook exists and check permissions (E.3)
	playbook, err := h.repo.GetPlaybook(r.Context(), playbookID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "playbook not found"})
		return
	}
	if !playbook.Enabled {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "playbook is disabled"})
		return
	}

	// E.3: Check if user's role is allowed to schedule this playbook
	claims, _ := r.Context().Value(userContextKey).(*AuthClaims)
	userRole := "viewer"
	if claims != nil {
		userRole = claims.Role
	}
	allowed := false
	for _, role := range playbook.AllowedRoles {
		if role == userRole {
			allowed = true
			break
		}
	}
	if userRole == "admin" {
		allowed = true
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "your role is not allowed to schedule this playbook"})
		return
	}

	paramsJSON, _ := json.Marshal(req.Parameters)
	exec := models.ScheduledPlaybookExecution{
		PlaybookID:  playbookID,
		Parameters:  json.RawMessage(paramsJSON),
		ScheduledAt: scheduledAt,
		Status:      "pending",
		Result:      json.RawMessage(`{}`),
	}
	if claims != nil {
		exec.CreatedBy = claims.Username
	}

	id, err := h.repo.CreateScheduledPlaybookExecution(r.Context(), exec)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create scheduled execution"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":           id,
		"playbook_id":  playbookID,
		"scheduled_at": scheduledAt,
		"status":       "pending",
	})
}

// ListScheduledPlaybookExecutions lists scheduled playbook executions (E.2).
// GET /playbooks/{id}/schedule or GET /schedule/playbooks
func (h *Handler) ListScheduledPlaybookExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	playbookID := chi.URLParam(r, "id")
	if playbookID == "" {
		playbookID = r.URL.Query().Get("playbook_id")
	}
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	executions, err := h.repo.ListScheduledPlaybookExecutions(r.Context(), playbookID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list scheduled executions"})
		return
	}
	writeJSON(w, http.StatusOK, executions)
}

// DeleteScheduledPlaybookExecution cancels a scheduled playbook execution (E.2).
// DELETE /schedule/playbooks/{id}
func (h *Handler) DeleteScheduledPlaybookExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.repo.UpdateScheduledPlaybookExecution(r.Context(), id, map[string]any{"status": "cancelled"}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to cancel scheduled execution"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Maintenance Windows

// ListMaintenanceWindows lists maintenance windows (E.2).
// GET /maintenance/windows
func (h *Handler) ListMaintenanceWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var startTime, endTime *time.Time
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}
	windows, err := h.repo.ListMaintenanceWindows(r.Context(), startTime, endTime)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list maintenance windows"})
		return
	}
	writeJSON(w, http.StatusOK, windows)
}

// CreateMaintenanceWindow creates a maintenance window (E.2).
// POST /maintenance/windows
func (h *Handler) CreateMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Name            string  `json:"name"`
		Description     string  `json:"description"`
		StartTime       string  `json:"start_time"` // RFC3339
		EndTime         string  `json:"end_time"`   // RFC3339
		RoutingPolicyID *int64  `json:"routing_policy_id,omitempty"`
		Enabled         bool    `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" || req.StartTime == "" || req.EndTime == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, start_time, and end_time are required"})
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "start_time must be RFC3339 format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "end_time must be RFC3339 format"})
		return
	}
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "end_time must be after start_time"})
		return
	}

	window := models.MaintenanceWindow{
		Name:            req.Name,
		Description:     req.Description,
		StartTime:       startTime,
		EndTime:         endTime,
		RoutingPolicyID: req.RoutingPolicyID,
		Enabled:         req.Enabled,
	}
	claims, _ := r.Context().Value(userContextKey).(*AuthClaims)
	if claims != nil {
		window.CreatedBy = claims.Username
	}

	id, err := h.repo.CreateMaintenanceWindow(r.Context(), window)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create maintenance window"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"name":       window.Name,
		"start_time": window.StartTime,
		"end_time":   window.EndTime,
	})
}

// UpdateMaintenanceWindow updates a maintenance window (E.2).
// PUT /maintenance/windows/{id}
func (h *Handler) UpdateMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	updates := make(map[string]any)
	for key, value := range req {
		if key == "start_time" || key == "end_time" {
			if str, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339, str); err == nil {
					updates[key] = t
				}
			}
		} else if key == "name" || key == "description" || key == "enabled" || key == "routing_policy_id" {
			updates[key] = value
		}
	}
	if err := h.repo.UpdateMaintenanceWindow(r.Context(), id, updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update maintenance window"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DeleteMaintenanceWindow deletes a maintenance window (E.2).
// DELETE /maintenance/windows/{id}
func (h *Handler) DeleteMaintenanceWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.repo.DeleteMaintenanceWindow(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete maintenance window"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GetActiveMaintenanceWindows returns currently active maintenance windows (E.2).
// GET /maintenance/windows/active
func (h *Handler) GetActiveMaintenanceWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	at := time.Now()
	if atStr := r.URL.Query().Get("at"); atStr != "" {
		if t, err := time.Parse(time.RFC3339, atStr); err == nil {
			at = t
		}
	}
	windows, err := h.repo.GetActiveMaintenanceWindows(r.Context(), at)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get active maintenance windows"})
		return
	}
	writeJSON(w, http.StatusOK, windows)
}
