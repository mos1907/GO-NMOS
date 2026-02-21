package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// ListPlaybooks returns all playbooks (E.1).
// GET /playbooks
func (h *Handler) ListPlaybooks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	playbooks, err := h.repo.ListPlaybooks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list playbooks"})
		return
	}
	writeJSON(w, http.StatusOK, playbooks)
}

// GetPlaybook returns a single playbook (E.1).
// GET /playbooks/{id}
func (h *Handler) GetPlaybook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := chi.URLParam(r, "id")
	playbook, err := h.repo.GetPlaybook(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "playbook not found"})
		return
	}
	writeJSON(w, http.StatusOK, playbook)
}

// UpsertPlaybook creates or updates a playbook (E.1).
// PUT /playbooks/{id}
func (h *Handler) UpsertPlaybook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Name         string          `json:"name"`
		Description  string          `json:"description"`
		Steps        json.RawMessage `json:"steps"`
		Parameters   json.RawMessage `json:"parameters"`
		AllowedRoles []string        `json:"allowed_roles"` // E.3
		Enabled      bool            `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if len(req.Steps) == 0 {
		req.Steps = json.RawMessage("[]")
	}
	if len(req.Parameters) == 0 {
		req.Parameters = json.RawMessage("{}")
	}

	allowedRoles := req.AllowedRoles
	if len(allowedRoles) == 0 {
		allowedRoles = []string{"engineer", "admin"} // Default
	}
	playbook := models.Playbook{
		ID:           id,
		Name:         req.Name,
		Description:  req.Description,
		Steps:        req.Steps,
		Parameters:   req.Parameters,
		AllowedRoles: allowedRoles,
		Enabled:      req.Enabled,
	}
	if err := h.repo.UpsertPlaybook(r.Context(), playbook); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save playbook"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DeletePlaybook deletes a playbook (E.1).
// DELETE /playbooks/{id}
func (h *Handler) DeletePlaybook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.repo.DeletePlaybook(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete playbook"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ExecutePlaybook executes a playbook with provided parameters (E.1).
// POST /playbooks/{id}/execute
func (h *Handler) ExecutePlaybook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := chi.URLParam(r, "id")

	playbook, err := h.repo.GetPlaybook(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "playbook not found"})
		return
	}
	if !playbook.Enabled {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "playbook is disabled"})
		return
	}

	// E.3: Check if user's role is allowed to execute this playbook
	claims, _ := r.Context().Value(userContextKey).(*AuthClaims)
	userRole := "viewer" // Default
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
	// Admin can always execute
	if userRole == "admin" {
		allowed = true
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "your role is not allowed to execute this playbook"})
		return
	}

	var req struct {
		Parameters map[string]any `json:"parameters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	// Create execution record
	execParamsJSON, _ := json.Marshal(req.Parameters)
	exec := models.PlaybookExecution{
		PlaybookID: id,
		Parameters: json.RawMessage(execParamsJSON),
		Status:     "running",
		StartedAt:  time.Now(),
	}
	execID, err := h.repo.CreatePlaybookExecution(r.Context(), exec)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create execution"})
		return
	}

	// Execute steps
	var steps []map[string]any
	if err := json.Unmarshal(playbook.Steps, &steps); err != nil {
		h.repo.UpdatePlaybookExecution(r.Context(), execID, "error", json.RawMessage(`{"error":"invalid steps format"}`))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid playbook steps"})
		return
	}

	var executionResults []map[string]any
	var executionError error

	for i, step := range steps {
		action, _ := step["action"].(string)
		if action == "" {
			continue
		}

		// Replace template variables in step with actual parameter values
		stepJSON, _ := json.Marshal(step)
		stepStr := string(stepJSON)
		for key, value := range req.Parameters {
			placeholder := fmt.Sprintf("{{%s}}", key)
			valueStr := fmt.Sprintf("%v", value)
			stepStr = strings.ReplaceAll(stepStr, placeholder, valueStr)
		}
		var resolvedStep map[string]any
		if err := json.Unmarshal([]byte(stepStr), &resolvedStep); err != nil {
			executionError = fmt.Errorf("step %d: failed to resolve parameters", i+1)
			break
		}

		// Execute action
		stepResult := map[string]any{
			"step":   i + 1,
			"action": action,
		}

		switch action {
		case "connect_receiver":
			receiverID, _ := resolvedStep["receiver_id"].(string)
			senderID, _ := resolvedStep["sender_id"].(string)
			if receiverID == "" || senderID == "" {
				executionError = fmt.Errorf("step %d: receiver_id and sender_id required", i+1)
				break
			}
			conn := models.ReceiverConnection{
				ReceiverID: receiverID,
				SenderID:   senderID,
				State:      "active",
				Role:       "master",
			}
			if err := h.repo.UpsertReceiverConnection(r.Context(), conn); err != nil {
				executionError = fmt.Errorf("step %d: failed to connect receiver: %v", i+1, err)
				stepResult["error"] = err.Error()
			} else {
				stepResult["success"] = true
			}

		case "disconnect_receiver":
			receiverID, _ := resolvedStep["receiver_id"].(string)
			if receiverID == "" {
				executionError = fmt.Errorf("step %d: receiver_id required", i+1)
				break
			}
			// Delete active master connection
			if err := h.repo.DeleteReceiverConnection(r.Context(), receiverID, "active", "master"); err != nil {
				executionError = fmt.Errorf("step %d: failed to disconnect receiver: %v", i+1, err)
				stepResult["error"] = err.Error()
			} else {
				stepResult["success"] = true
			}

		default:
			executionError = fmt.Errorf("step %d: unknown action: %s", i+1, action)
			stepResult["error"] = executionError.Error()
		}

		executionResults = append(executionResults, stepResult)
		if executionError != nil {
			break
		}
	}

	// Update execution record
	resultJSON, _ := json.Marshal(map[string]any{
		"steps": executionResults,
		"error": func() string {
			if executionError != nil {
				return executionError.Error()
			}
			return ""
		}(),
	})
	status := "success"
	if executionError != nil {
		status = "error"
	}
	h.repo.UpdatePlaybookExecution(r.Context(), execID, status, json.RawMessage(resultJSON))

	if executionError != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"execution_id": execID,
			"error":        executionError.Error(),
			"steps":        executionResults,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"execution_id": execID,
		"status":       "success",
		"steps":        executionResults,
	})
}

// ListPlaybookExecutions returns execution history for a playbook (E.1).
// GET /playbooks/{id}/executions
func (h *Handler) ListPlaybookExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	id := chi.URLParam(r, "id")
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
		if limit > 100 {
			limit = 100
		}
	}
	executions, err := h.repo.ListPlaybookExecutions(r.Context(), id, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list executions"})
		return
	}
	writeJSON(w, http.StatusOK, executions)
}
