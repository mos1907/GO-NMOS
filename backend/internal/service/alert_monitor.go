package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-nmos/backend/internal/alerting"
	"go-nmos/backend/internal/repository"
)

// AlertMonitor periodically checks for critical conditions and sends alerts
type AlertMonitor struct {
	repo         repository.Repository
	alertManager *alerting.AlertManager
	lastChecks   map[string]time.Time
	errorCounts  map[string]int // Track repeated errors
}

func NewAlertMonitor(repo repository.Repository, alertManager *alerting.AlertManager) *AlertMonitor {
	return &AlertMonitor{
		repo:         repo,
		alertManager: alertManager,
		lastChecks:   make(map[string]time.Time),
		errorCounts:  make(map[string]int),
	}
}

// Start begins the alert monitoring loop
func (m *AlertMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	// Initial check
	m.checkConditions(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkConditions(ctx)
		}
	}
}

func (m *AlertMonitor) checkConditions(ctx context.Context) {
	// Check registry empty/down
	m.checkRegistry(ctx)

	// Check PTP mismatch
	m.checkPTPMismatch(ctx)

	// Check repeated connection errors
	m.checkConnectionErrors(ctx)
}

func (m *AlertMonitor) checkRegistry(ctx context.Context) {
	nodes, err := m.repo.ListNMOSNodes(ctx)
	if err != nil {
		m.recordError("registry_check", err)
		m.alertManager.SendError(
			"registry",
			"registry_check_failed",
			"Registry Check Failed",
			"Failed to query NMOS registry: "+err.Error(),
			map[string]any{"error": err.Error()},
		)
		return
	}

	m.clearError("registry_check")

	if len(nodes) == 0 {
		// Check if we've already alerted about empty registry
		lastAlert := m.lastChecks["registry_empty"]
		if lastAlert.IsZero() || time.Since(lastAlert) > 5*time.Minute {
			m.alertManager.SendWarning(
				"registry",
				"registry_empty",
				"NMOS Registry Empty",
				"The NMOS registry contains no nodes. This may indicate a discovery issue or all nodes are offline.",
				map[string]any{"node_count": 0},
			)
			m.lastChecks["registry_empty"] = time.Now()
		}
	} else {
		// Registry has nodes, clear empty alert
		delete(m.lastChecks, "registry_empty")
	}
}

func (m *AlertMonitor) checkPTPMismatch(ctx context.Context) {
	expectedPTPDomain, _ := m.repo.GetSetting(ctx, "system_ptp_domain")
	if expectedPTPDomain == "" {
		return // No PTP domain configured, skip check
	}

	nodes, err := m.repo.ListNMOSNodes(ctx)
	if err != nil {
		return
	}

	mismatchCount := 0
	mismatchNodes := []string{}

	for _, node := range nodes {
		domain := node.GetNetworkDomain()
		if domain != "" && domain != expectedPTPDomain {
			mismatchCount++
			mismatchNodes = append(mismatchNodes, node.Label+" ("+node.ID+")")
		}
	}

	if mismatchCount > 0 {
		lastAlert := m.lastChecks["ptp_mismatch"]
		if lastAlert.IsZero() || time.Since(lastAlert) > 10*time.Minute {
			m.alertManager.SendCritical(
				"ptp",
				"ptp_mismatch",
				"PTP Domain Mismatch Detected",
				fmt.Sprintf("%d node(s) have PTP domain mismatch. Expected: %s", mismatchCount, expectedPTPDomain),
				map[string]any{
					"expected_domain": expectedPTPDomain,
					"mismatch_count":  mismatchCount,
					"nodes":           mismatchNodes,
				},
			)
			m.lastChecks["ptp_mismatch"] = time.Now()
		}
	} else {
		delete(m.lastChecks, "ptp_mismatch")
	}
}

func (m *AlertMonitor) checkConnectionErrors(ctx context.Context) {
	// Check routing policy audit for repeated violations/failures
	// This gives us insight into connection issues
	audits, err := m.repo.ListRoutingPolicyAudits(ctx, 100)
	if err == nil {
		recentFailures := 0
		now := time.Now()
		for _, audit := range audits {
			if audit.Action == "violation" && now.Sub(audit.CreatedAt) < 5*time.Minute {
				recentFailures++
			}
		}

		if recentFailures > 10 {
			lastAlert := m.lastChecks["connection_errors"]
			if lastAlert.IsZero() || time.Since(lastAlert) > 5*time.Minute {
				m.alertManager.SendError(
					"connection",
					"repeated_connection_errors",
					"Repeated Connection Errors",
					fmt.Sprintf("Detected %d routing policy violations in the last 5 minutes. This may indicate connection issues.", recentFailures),
					map[string]any{"recent_failures": recentFailures},
				)
				m.lastChecks["connection_errors"] = time.Now()
			}
		} else {
			delete(m.lastChecks, "connection_errors")
		}
	}
}

func (m *AlertMonitor) recordError(key string, err error) {
	m.errorCounts[key]++
	if m.errorCounts[key] >= 3 {
		// After 3 consecutive errors, send alert
		log.Printf("AlertMonitor: %s has failed %d times", key, m.errorCounts[key])
	}
}

func (m *AlertMonitor) clearError(key string) {
	delete(m.errorCounts, key)
}
