package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// CreateRoutingPolicy creates a new routing policy.
// POST /api/routing/policies
func (h *Handler) CreateRoutingPolicy(w http.ResponseWriter, r *http.Request) {
	var policy models.RoutingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	policy.Name = strings.TrimSpace(policy.Name)
	if policy.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if policy.PolicyType == "" {
		policy.PolicyType = "forbidden_pair"
	}
	validTypes := map[string]bool{"allowed_pair": true, "forbidden_pair": true, "path_requirement": true, "constraint": true}
	if !validTypes[policy.PolicyType] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "policy_type must be allowed_pair, forbidden_pair, path_requirement, or constraint"})
		return
	}
	if policy.Priority == 0 {
		policy.Priority = 100
	}
	policy.Enabled = true

	if claims := r.Context().Value("claims"); claims != nil {
		if authClaims, ok := claims.(*AuthClaims); ok && authClaims.Username != "" {
			policy.CreatedBy = authClaims.Username
		}
	}

	id, err := h.repo.CreateRoutingPolicy(r.Context(), policy)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create policy"})
		return
	}
	policy.ID = id
	writeJSON(w, http.StatusOK, policy)
}

// ListRoutingPolicies lists routing policies.
// GET /api/routing/policies?enabled_only=true
func (h *Handler) ListRoutingPolicies(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"
	policies, err := h.repo.ListRoutingPolicies(r.Context(), enabledOnly)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list policies"})
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

// GetRoutingPolicy gets a single policy by ID.
// GET /api/routing/policies/{id}
func (h *Handler) GetRoutingPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	policy, err := h.repo.GetRoutingPolicy(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "policy not found"})
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

// UpdateRoutingPolicy updates a routing policy.
// PUT /api/routing/policies/{id}
func (h *Handler) UpdateRoutingPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	allowed := map[string]bool{"name": true, "policy_type": true, "enabled": true, "source_pattern": true, "destination_pattern": true, "require_path_a": true, "require_path_b": true, "constraint_field": true, "constraint_value": true, "constraint_operator": true, "description": true, "priority": true}
	filtered := make(map[string]any)
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}
	if err := h.repo.UpdateRoutingPolicy(r.Context(), id, filtered); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update policy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DeleteRoutingPolicy deletes a routing policy.
// DELETE /api/routing/policies/{id}
func (h *Handler) DeleteRoutingPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.repo.DeleteRoutingPolicy(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete policy"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// CheckRoutingPolicyRequest is the request body for POST /api/routing/check.
type CheckRoutingPolicyRequest struct {
	SenderID    string `json:"sender_id"`
	ReceiverID  string `json:"receiver_id"`
	FlowID      int64  `json:"flow_id,omitempty"`
	FlowLabel   string `json:"flow_label,omitempty"`   // For constraint checks (e.g. "test" in label)
	SenderLabel string `json:"sender_label,omitempty"` // For constraint checks
}

// PolicyViolation represents a single policy violation.
type PolicyViolation struct {
	PolicyID   int64  `json:"policy_id"`
	PolicyName string `json:"policy_name"`
	Reason     string `json:"reason"`
}

// CheckRoutingPolicyResponse is the response for POST /api/routing/check.
type CheckRoutingPolicyResponse struct {
	Allowed    bool              `json:"allowed"`
	Violations []PolicyViolation `json:"violations"`
}

// CheckRoutingPolicy checks if a sender→receiver connection is allowed by routing policies.
// POST /api/routing/check
func (h *Handler) CheckRoutingPolicy(w http.ResponseWriter, r *http.Request) {
	var req CheckRoutingPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.SenderID == "" || req.ReceiverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sender_id and receiver_id are required"})
		return
	}

	violations := h.checkPoliciesForConnection(r.Context(), req.SenderID, req.ReceiverID, req.FlowID, req.FlowLabel, req.SenderLabel)

	resp := CheckRoutingPolicyResponse{
		Allowed:    len(violations) == 0,
		Violations: violations,
	}

	// Record audit
	action := "allowed"
	if len(violations) > 0 {
		action = "violation"
	}
	audit := models.RoutingPolicyAudit{
		Action:        action,
		SourceID:      req.SenderID,
		DestinationID: req.ReceiverID,
		FlowID:        &req.FlowID,
	}
	if len(violations) > 0 {
		audit.ViolationReason = violations[0].Reason
		audit.PolicyID = &violations[0].PolicyID
	}
	_ = h.repo.RecordRoutingPolicyAudit(r.Context(), audit)

	// Record metrics
	if len(violations) > 0 {
		RecordRoutingOperationFailure("check", violations[0].Reason)
	} else {
		RecordRoutingOperation("check", "success")
	}

	writeJSON(w, http.StatusOK, resp)
}

// CheckPoliciesForConnection checks policies for a sender→receiver connection.
// Returns violations; if empty, connection is allowed.
func (h *Handler) checkPoliciesForConnection(ctx context.Context, senderID, receiverID string, flowID int64, flowLabel, senderLabel string) []PolicyViolation {
	policies, err := h.repo.ListRoutingPolicies(ctx, true)
	if err != nil {
		return nil
	}
	var violations []PolicyViolation
	for _, p := range policies {
		if !p.Enabled {
			continue
		}
		switch p.PolicyType {
		case "forbidden_pair":
			if matchPattern(p.SourcePattern, senderID) && matchPattern(p.DestinationPattern, receiverID) {
				violations = append(violations, PolicyViolation{PolicyID: p.ID, PolicyName: p.Name, Reason: "forbidden source-destination pair"})
			}
		case "constraint":
			if p.ConstraintField != "" && p.ConstraintValue != "" {
				val := ""
				switch p.ConstraintField {
				case "flow_label", "label":
					val = flowLabel
				case "sender_label":
					val = senderLabel
				}
				if matchConstraint(val, p.ConstraintValue, p.ConstraintOperator) {
					violations = append(violations, PolicyViolation{PolicyID: p.ID, PolicyName: p.Name, Reason: "constraint violated: " + p.ConstraintField + " " + p.ConstraintOperator + " " + p.ConstraintValue})
				}
			}
		}
	}
	return violations
}

func matchPattern(pattern, id string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, ":*")
		return strings.HasPrefix(id, prefix) || id == prefix
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(id, prefix)
	}
	return id == pattern
}

func matchConstraint(val, expected, op string) bool {
	switch op {
	case "equals":
		return val == expected
	case "contains":
		return strings.Contains(strings.ToLower(val), strings.ToLower(expected))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(val), strings.ToLower(expected))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(val), strings.ToLower(expected))
	default:
		return val == expected
	}
}

// ListRoutingPolicyAudits lists policy audit log entries.
// GET /api/routing/policies/audits?limit=100
func (h *Handler) ListRoutingPolicyAudits(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	audits, err := h.repo.ListRoutingPolicyAudits(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list audits"})
		return
	}
	writeJSON(w, http.StatusOK, audits)
}
