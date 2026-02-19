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

