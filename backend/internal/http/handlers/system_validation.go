package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"go-nmos/backend/internal/models"
)

// SystemValidationIssue represents a validation mismatch (D.2).
type SystemValidationIssue struct {
	Type        string `json:"type"`         // "registry" | "node"
	ResourceID  string `json:"resource_id"`  // registry query URL or node ID
	ResourceLabel string `json:"resource_label,omitempty"`
	Field       string `json:"field"`        // "is04_version" | "is05_version" | "ptp_domain"
	Expected    string `json:"expected"`     // Expected value
	Actual      string `json:"actual"`      // Actual value from registry/node
	Severity    string `json:"severity"`    // "error" | "warning"
	Message     string `json:"message"`
}

// ValidateSystemParameters checks registry and node information against configured system parameters (D.2).
// GET /system/validation - compares expected IS-04/05 versions and PTP domain with actual registry/node values.
func (h *Handler) ValidateSystemParameters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	ctx := r.Context()
	issues := make([]SystemValidationIssue, 0)

	// Get expected values
	expectedIS04, _ := h.repo.GetSetting(ctx, "system_expected_is04")
	expectedIS05, _ := h.repo.GetSetting(ctx, "system_expected_is05")
	expectedPTPDomain, _ := h.repo.GetSetting(ctx, "system_ptp_domain")

	// Check registries (from registry_compat logic)
	registryConfigsJSON, _ := h.repo.GetSetting(ctx, "nmos_registry_config")
	if registryConfigsJSON != "" {
		var configs []models.RegistryConfig
		if err := json.Unmarshal([]byte(registryConfigsJSON), &configs); err == nil {
			for _, reg := range configs {
				if !reg.Enabled {
					continue
				}
				// Try to discover IS-04 version from registry
				versions, err := h.fetchJSONList(strings.TrimRight(reg.QueryURL, "/") + "/x-nmos/query/")
				if err == nil && len(versions) > 0 {
					sort.Strings(versions)
					actualVer := strings.Trim(versions[len(versions)-1], "/")
					if expectedIS04 != "" && actualVer != expectedIS04 {
						issues = append(issues, SystemValidationIssue{
							Type:         "registry",
							ResourceID:   reg.QueryURL,
							ResourceLabel: reg.Name,
							Field:        "is04_version",
							Expected:     expectedIS04,
							Actual:       actualVer,
							Severity:     "warning",
							Message:      "Registry IS-04 version mismatch",
						})
					}
				}
			}
		}
	}

	// Check nodes (from internal registry)
	nodes, _ := h.repo.ListNMOSNodes(ctx)
	for _, node := range nodes {
		// Check IS-04 version
		if expectedIS04 != "" && node.APIVersion != "" && node.APIVersion != expectedIS04 {
			issues = append(issues, SystemValidationIssue{
				Type:         "node",
				ResourceID:   node.ID,
				ResourceLabel: node.Label,
				Field:        "is04_version",
				Expected:     expectedIS04,
				Actual:       node.APIVersion,
				Severity:     "warning",
				Message:      "Node IS-04 version mismatch",
			})
		}
		// Check PTP domain (if stored in meta)
		if expectedPTPDomain != "" {
			domain := node.GetNetworkDomain()
			if domain != "" && domain != expectedPTPDomain {
				issues = append(issues, SystemValidationIssue{
					Type:         "node",
					ResourceID:   node.ID,
					ResourceLabel: node.Label,
					Field:        "ptp_domain",
					Expected:     expectedPTPDomain,
					Actual:       domain,
					Severity:     "error",
					Message:      "Node PTP domain mismatch",
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"issues": issues,
		"count":  len(issues),
		"expected": map[string]string{
			"is04_version": expectedIS04,
			"is05_version": expectedIS05,
			"ptp_domain":   expectedPTPDomain,
		},
	})
}
