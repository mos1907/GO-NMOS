package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

	// F.3: Generate incident hints based on detected issues
	hints := h.generateIncidentHints(r.Context(), dbOk, mqttOk, mqttEnabled, registryOk, registryStatus, registryCounts)

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
		"hints": hints,
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

// generateIncidentHints generates actionable hints based on detected issues (F.3)
func (h *Handler) generateIncidentHints(ctx context.Context, dbOk, mqttOk, mqttEnabled, registryOk bool, registryStatus string, registryCounts map[string]int) []map[string]any {
	hints := []map[string]any{}

	// Database issues
	if !dbOk {
		hints = append(hints, map[string]any{
			"severity": "critical",
			"component": "database",
			"title": "Database Connection Issue",
			"message": "The database connection is failing. Check database connectivity and credentials.",
			"suggestions": []string{
				"Verify database server is running",
				"Check database connection string in configuration",
				"Review database logs for errors",
				"Ensure network connectivity to database",
			},
		})
	}

	// MQTT issues
	if mqttEnabled && !mqttOk {
		suggestions := []string{
			"Verify MQTT broker is running at " + h.cfg.MQTTBrokerURL,
			"Check network connectivity to MQTT broker",
			"Review MQTT broker logs",
			"Verify MQTT credentials if authentication is required",
		}
		if strings.Contains(h.cfg.MQTTBrokerURL, "localhost") || strings.Contains(h.cfg.MQTTBrokerURL, "127.0.0.1") {
			suggestions = append(suggestions, "If backend runs in Docker: set MQTT_BROKER_URL=tcp://mqtt:1883 (compose) or tcp://host.docker.internal:1883 (broker on host)")
		}
		hints = append(hints, map[string]any{
			"severity":    "warning",
			"component":   "mqtt",
			"title":       "MQTT Broker Connection Issue",
			"message":     "MQTT is enabled but connection to broker is failing.",
			"suggestions": suggestions,
		})
	}

	// Registry empty
	if !registryOk && registryStatus == "empty" {
		hints = append(hints, map[string]any{
			"severity": "warning",
			"component": "registry",
			"title": "NMOS Registry Empty",
			"message": "No NMOS nodes, devices, or flows are registered.",
			"suggestions": []string{
				"Check if NMOS nodes are powered on and connected to the network",
				"Verify network connectivity between nodes and registry",
				"Review registry discovery configuration",
				"Check if nodes are configured to register with the correct registry URL",
				"Use 'Discover NMOS Nodes' to manually discover nodes",
			},
		})
	}

	// Registry error
	if !registryOk && registryStatus != "empty" {
		hints = append(hints, map[string]any{
			"severity": "error",
			"component": "registry",
			"title": "Registry Query Error",
			"message": "Failed to query NMOS registry data.",
			"suggestions": []string{
				"Check database connectivity",
				"Review application logs for registry query errors",
				"Verify registry data integrity",
			},
		})
	}

	// Check for PTP domain mismatches
	expectedPTPDomain, _ := h.repo.GetSetting(ctx, "system_ptp_domain")
	if expectedPTPDomain != "" {
		nodes, err := h.repo.ListNMOSNodes(ctx)
		if err == nil {
			mismatchCount := 0
			for _, node := range nodes {
				domain := node.GetNetworkDomain()
				if domain != "" && domain != expectedPTPDomain {
					mismatchCount++
				}
			}
			if mismatchCount > 0 {
				hints = append(hints, map[string]any{
					"severity": "critical",
					"component": "ptp",
					"title": fmt.Sprintf("PTP Domain Mismatch (%d nodes)", mismatchCount),
					"message": fmt.Sprintf("%d node(s) have PTP domain different from expected '%s'. This can cause timing synchronization issues.", mismatchCount, expectedPTPDomain),
					"suggestions": []string{
						"Review PTP domain configuration on affected nodes",
						"Verify PTP grandmaster configuration",
						"Check network PTP settings",
						"Use System Parameters Validation to see detailed mismatch information",
					},
				})
			}
		}
	}

	// Check for repeated connection errors
	audits, err := h.repo.ListRoutingPolicyAudits(ctx, 50)
	if err == nil {
		recentFailures := 0
		now := time.Now()
		for _, audit := range audits {
			if audit.Action == "violation" && now.Sub(audit.CreatedAt) < 5*time.Minute {
				recentFailures++
			}
		}
		if recentFailures > 10 {
			hints = append(hints, map[string]any{
				"severity": "warning",
				"component": "routing",
				"title": "Repeated Connection Errors",
				"message": fmt.Sprintf("Detected %d routing policy violations in the last 5 minutes.", recentFailures),
				"suggestions": []string{
					"Review routing policies for misconfigurations",
					"Check network connectivity between sources and destinations",
					"Verify flow configurations",
					"Review routing policy audit logs for patterns",
				},
			})
		}
	}

	return hints
}
