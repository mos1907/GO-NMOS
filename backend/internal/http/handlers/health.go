package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Health is a very small, fast liveness probe (used e.g. by Docker/Ingress).
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	overallStatus := "ok"
	statusCode := http.StatusOK
	if err := h.repo.HealthCheck(r.Context()); err != nil {
		dbStatus = "error"
		overallStatus = "degraded"
		statusCode = http.StatusServiceUnavailable
	}
	writeJSON(w, statusCode, map[string]any{
		"status":  overallStatus,
		"service": "go-NMOS",
		"db":      dbStatus,
	})
}

// HealthDetail exposes a richer, human-friendly health summary for the UI.
// It is intentionally a bit more expensive than /api/health and should be used
// for on-demand diagnostics rather than tight liveness checks.
func (h *Handler) HealthDetail(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC().Format(time.RFC3339)

	// DB status
	dbOk := true
	dbError := ""
	if err := h.repo.HealthCheck(r.Context()); err != nil {
		dbOk = false
		dbError = err.Error()
	}

	// MQTT status (we treat "disabled" as not-an-error)
	mqttEnabled := h.cfg.MQTTEnabled
	mqttOk := !mqttEnabled || (h.mqtt != nil)
	mqttStatus := "disabled"
	if mqttEnabled {
		if mqttOk {
			mqttStatus = "ok"
		} else {
			mqttStatus = "error"
		}
	}

	// Internal NMOS registry view (re-use same logic as NMOSRegistryHealth)
	registryStatus := "empty"
	registryOk := false
	registryErr := ""
	registryCounts := map[string]int{}
	if nodes, err := h.repo.ListNMOSNodes(r.Context()); err != nil {
		registryErr = "failed to list NMOS nodes"
	} else if devices, err := h.repo.ListNMOSDevices(r.Context(), ""); err != nil {
		registryErr = "failed to list NMOS devices"
	} else if flows, err := h.repo.ListNMOSFlows(r.Context()); err != nil {
		registryErr = "failed to list NMOS flows"
	} else if senders, err := h.repo.ListNMOSSenders(r.Context(), ""); err != nil {
		registryErr = "failed to list NMOS senders"
	} else if receivers, err := h.repo.ListNMOSReceivers(r.Context(), ""); err != nil {
		registryErr = "failed to list NMOS receivers"
	} else {
		registryCounts = map[string]int{
			"nodes":     len(nodes),
			"devices":   len(devices),
			"flows":     len(flows),
			"senders":   len(senders),
			"receivers": len(receivers),
		}
		registryOk = len(nodes) > 0 || len(devices) > 0 || len(flows) > 0 || len(senders) > 0 || len(receivers) > 0
		if registryOk {
			registryStatus = "ok"
		}
	}

	overallStatus := "ok"
	if !dbOk || (mqttEnabled && !mqttOk) || (!registryOk && registryErr != "") {
		overallStatus = "degraded"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":    overallStatus,
		"timestamp": now,
		"service":   "go-NMOS",
		"components": map[string]any{
			"db": map[string]any{
				"ok":      dbOk,
				"error":   dbError,
				"checked": now,
			},
			"mqtt": map[string]any{
				"ok":         mqttOk,
				"enabled":    mqttEnabled,
				"status":     mqttStatus,
				"broker_url": h.cfg.MQTTBrokerURL,
				"checked":    now,
			},
			"registry": map[string]any{
				"ok":      registryOk,
				"status":  registryStatus,
				"error":   registryErr,
				"counts":  registryCounts,
				"checked": now,
			},
		},
	})
}

// CheckNMOSNode performs a quick HTTP check against a given NMOS Node base URL.
// It is intended for operator use from the Diagnostics panel.
func (h *Handler) CheckNMOSNode(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL        string `json:"url"`
		TimeoutSec int    `json:"timeout_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": "invalid JSON body",
		})
		return
	}
	raw := strings.TrimSpace(payload.URL)
	if raw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": "url is required",
		})
		return
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"ok":    false,
			"error": "invalid URL",
		})
		return
	}

	timeout := time.Duration(payload.TimeoutSec)
	if timeout <= 0 || timeout > 30 {
		timeout = 5
	}

	// Try NMOS Node self endpoint if path looks like base URL, otherwise use as-is.
	target := raw
	if !strings.Contains(parsed.Path, "/x-nmos/") {
		base := strings.TrimRight(raw, "/")
		// Default to IS-04 v1.3 self, which is commonly implemented.
		target = base + "/x-nmos/node/v1.3/self"
	}

	client := &http.Client{Timeout: timeout * time.Second}
	start := time.Now()
	resp, err := client.Get(target)
	elapsed := time.Since(start)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":       false,
			"error":    err.Error(),
			"url":      target,
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
		"base_url": raw,
		"target":   target,
		"httpCode": resp.StatusCode,
		"latency":  elapsed.String(),
		"checked":  time.Now().UTC().Format(time.RFC3339),
	})
}
