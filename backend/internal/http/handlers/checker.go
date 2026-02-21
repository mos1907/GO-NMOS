package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// CheckerRun runs both collision and NMOS checks, saves results, and returns a summary.
// POST /api/checker/run (optional body: { "nmos_timeout": 5 })
func (h *Handler) CheckerRun(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NmosTimeout int `json:"nmos_timeout"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.NmosTimeout <= 0 || body.NmosTimeout > 30 {
		body.NmosTimeout = 5
	}

	// Run collision check
	collisions, err := h.repo.DetectCollisions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "collision check failed"})
		return
	}
	collPayload := map[string]any{
		"total_collisions": len(collisions),
		"items":            collisions,
	}
	collRaw, _ := json.Marshal(collPayload)
	_ = h.repo.SaveCheckerResult(r.Context(), "collisions", collRaw)

	// Run NMOS check (same logic as CheckerNMOS)
	flows, err := h.repo.ListFlows(r.Context(), 10000, 0, "updated_at", "desc")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"collisions": collPayload,
			"nmos":       map[string]any{"error": "failed to list flows"},
		})
		return
	}
	senders, _ := h.repo.ListNMOSSenders(r.Context(), "")
	differences := []map[string]any{}
	for _, flow := range flows {
		if flow.FlowID == "" {
			continue
		}
		found := false
		for _, sender := range senders {
			if sender.FlowID == flow.FlowID {
				found = true
				break
			}
		}
		if !found {
			differences = append(differences, map[string]any{
				"flow_id":      flow.FlowID,
				"display_name": flow.DisplayName,
				"issue":        "flow_id not found in NMOS registry",
				"type":         "missing_sender",
			})
		}
	}
	nmosPayload := map[string]any{
		"total_differences": len(differences),
		"items":             differences,
		"checked_flows":    len(flows),
		"timeout_seconds":  body.NmosTimeout,
	}
	nmosRaw, _ := json.Marshal(nmosPayload)
	_ = h.repo.SaveCheckerResult(r.Context(), "nmos", nmosRaw)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"collisions": collPayload,
		"nmos":       nmosPayload,
	})
}

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

// CheckerNMOS checks for differences between flows and NMOS nodes.
// This compares flow metadata with actual NMOS sender/receiver states.
func (h *Handler) CheckerNMOS(w http.ResponseWriter, r *http.Request) {
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 5
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 && t <= 30 {
			timeout = t
		}
	}

	// Get all flows
	flows, err := h.repo.ListFlows(r.Context(), 10000, 0, "updated_at", "desc")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list flows"})
		return
	}

	differences := []map[string]any{}
	for _, flow := range flows {
		// For each flow, check if it matches any NMOS sender
		// This is a simplified check - in production, you'd query actual NMOS nodes
		// For now, we check if flow_id exists in our internal NMOS registry
		if flow.FlowID == "" {
			continue
		}

		// Check internal NMOS registry for matching senders
		senders, err := h.repo.ListNMOSSenders(r.Context(), "")
		if err == nil {
			found := false
			for _, sender := range senders {
				if sender.FlowID == flow.FlowID {
					found = true
					break
				}
			}
			if !found {
				differences = append(differences, map[string]any{
					"flow_id":      flow.FlowID,
					"display_name": flow.DisplayName,
					"issue":        "flow_id not found in NMOS registry",
					"type":         "missing_sender",
				})
			}
		}
	}

	payload := map[string]any{
		"total_differences": len(differences),
		"items":             differences,
		"checked_flows":     len(flows),
		"timeout_seconds":   timeout,
	}
	raw, _ := json.Marshal(payload)
	_ = h.repo.SaveCheckerResult(r.Context(), "nmos", raw)
	writeJSON(w, http.StatusOK, payload)
}

// CheckCollision checks if a specific IP:Port combination would cause a collision
// GET /api/checker/check?multicast_ip=239.0.0.1&port=5004&exclude_flow_id=123
func (h *Handler) CheckCollision(w http.ResponseWriter, r *http.Request) {
	multicastIP := r.URL.Query().Get("multicast_ip")
	portStr := r.URL.Query().Get("port")
	excludeFlowIDStr := r.URL.Query().Get("exclude_flow_id")

	if multicastIP == "" || portStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "multicast_ip and port are required"})
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid port"})
		return
	}

	var excludeFlowID int64
	if excludeFlowIDStr != "" {
		if id, err := strconv.ParseInt(excludeFlowIDStr, 10, 64); err == nil {
			excludeFlowID = id
		}
	}

	// Check for collisions
	collisions, err := h.repo.DetectCollisions(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "collision check failed"})
		return
	}

	// Check if the provided IP:Port combination exists in collisions
	hasCollision := false
	var conflictingFlows []string
	for _, collision := range collisions {
		if collision.MulticastIP == multicastIP && collision.Port == port {
			hasCollision = true
			conflictingFlows = collision.FlowNames
			break
		}
	}

	// If excluding a flow ID (e.g., editing existing flow), filter it out
	if excludeFlowID > 0 && hasCollision {
		// Get flow by ID to check if it's one of the conflicting flows
		flow, err := h.repo.GetFlowByID(r.Context(), excludeFlowID)
		if err == nil && flow != nil {
			// Remove the excluded flow from conflicting flows list
			filtered := []string{}
			for _, name := range conflictingFlows {
				if name != flow.DisplayName {
					filtered = append(filtered, name)
				}
			}
			conflictingFlows = filtered
			// If no other flows conflict, it's not a collision
			if len(filtered) == 0 {
				hasCollision = false
			}
		}
	}

	response := map[string]any{
		"has_collision":     hasCollision,
		"multicast_ip":      multicastIP,
		"port":              port,
		"conflicting_flows": conflictingFlows,
		"conflict_count":    len(conflictingFlows),
	}

	// Always provide alternative suggestions (even if no collision, user might want alternatives)
	var excludeFlowIDPtr *int64
	if excludeFlowIDStr != "" {
		if id, err := strconv.ParseInt(excludeFlowIDStr, 10, 64); err == nil {
			excludeFlowIDPtr = &id
		}
	}
	suggestions, err := h.repo.GetAlternativeSuggestions(r.Context(), multicastIP, port, excludeFlowIDPtr)
	if err == nil && len(suggestions) > 0 {
		response["alternatives"] = suggestions
	}

	writeJSON(w, http.StatusOK, response)
}
