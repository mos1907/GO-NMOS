package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SDNPing verifies reachability of the configured SDN controller (IS-06-style).
// URL resolution order:
//  1. Explicit URL in request body (if provided)
//  2. Stored setting "sdn_controller_url"
//  3. Static config h.cfg.SDNControllerURL
func (h *Handler) SDNPing(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL        string `json:"url"`
		TimeoutSec int    `json:"timeout_sec"`
	}

	// Best-effort decode; empty body is allowed.
	_ = json.NewDecoder(r.Body).Decode(&payload)

	raw := strings.TrimSpace(payload.URL)
	if raw == "" {
		if v, err := h.repo.GetSetting(r.Context(), "sdn_controller_url"); err == nil && strings.TrimSpace(v) != "" {
			raw = strings.TrimSpace(v)
		} else if strings.TrimSpace(h.cfg.SDNControllerURL) != "" {
			raw = strings.TrimSpace(h.cfg.SDNControllerURL)
		}
	}

	if raw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": "no SDN controller URL configured",
		})
		return
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": "invalid SDN controller URL",
		})
		return
	}

	timeout := time.Duration(payload.TimeoutSec)
	if timeout <= 0 || timeout > 30 {
		timeout = 5
	}

	client := &http.Client{Timeout: timeout * time.Second}
	start := time.Now()
	resp, err := client.Get(raw)
	elapsed := time.Since(start)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":       false,
			"error":    err.Error(),
			"url":      raw,
			"latency":  elapsed.String(),
			"status":   "unreachable",
			"checked":  time.Now().UTC().Format(time.RFC3339),
			"httpCode": 0,
		})
		return
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	status := "error"
	if ok {
		status = "ok"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       ok,
		"status":   status,
		"url":      raw,
		"httpCode": resp.StatusCode,
		"latency":  elapsed.String(),
		"checked":  time.Now().UTC().Format(time.RFC3339),
	})
}

// SDNTopology returns network topology (nodes and links). Stub implementation for B.4 / IS-06.
// If sdn_controller_url is set, could proxy to controller; for now returns demo data.
func (h *Handler) SDNTopology(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	// Stub: demo nodes and links for UI
	out := map[string]any{
		"nodes": []map[string]any{
			{"id": "switch-1", "label": "Core Switch 1", "type": "switch"},
			{"id": "switch-2", "label": "Core Switch 2", "type": "switch"},
			{"id": "node-a", "label": "Node A", "type": "device"},
			{"id": "node-b", "label": "Node B", "type": "device"},
		},
		"links": []map[string]any{
			{"id": "link-1", "source": "switch-1", "target": "switch-2"},
			{"id": "link-2", "source": "switch-1", "target": "node-a"},
			{"id": "link-3", "source": "switch-2", "target": "node-b"},
		},
	}
	writeJSON(w, http.StatusOK, out)
}

// SDNPaths returns available paths between source and destination. Stub for B.4.
// Query params: from, to (node ids). Returns list of path objects with id, name, from, to, link_ids.
func (h *Handler) SDNPaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	from := strings.TrimSpace(r.URL.Query().Get("from"))
	to := strings.TrimSpace(r.URL.Query().Get("to"))
	// Stub: return demo paths; if from/to provided, filter or use in label
	if from == "" {
		from = "node-a"
	}
	if to == "" {
		to = "node-b"
	}
	out := []map[string]any{
		{"id": "path-1", "name": "Path 1 (" + from + " → " + to + ")", "from": from, "to": to, "link_ids": []string{"link-2", "link-1", "link-3"}},
		{"id": "path-2", "name": "Path 2 (backup)", "from": from, "to": to, "link_ids": []string{"link-2", "link-1", "link-3"}},
	}
	writeJSON(w, http.StatusOK, map[string]any{"paths": out})
}

