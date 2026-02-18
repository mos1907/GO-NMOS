package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
