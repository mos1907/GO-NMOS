package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"go-nmos/backend/internal/models"
)

// RegistryCompatibility describes IS-04 compatibility for a single configured registry.
type RegistryCompatibility struct {
	Name               string   `json:"name"`
	Role               string   `json:"role"`
	QueryURL           string   `json:"query_url"`
	Enabled            bool     `json:"enabled"`
	DiscoveredVersions []string `json:"discovered_versions"`
	ExpectedIS04       string   `json:"expected_is04"`
	Status             string   `json:"status"`           // ok | warning | unsupported | error
	Error              string   `json:"error,omitempty"`  // present when Status == "error"
	ChosenQueryVersion string   `json:"chosen_query_ver"` // version we would pick for /x-nmos/query
}

// RegistryCompatibilityMatrix returns a simple IS-04 compatibility matrix
// for all configured registries in nmos_registry_config.
//
// It discovers supported IS-04 Query API versions and compares them to
// system_expected_is04 (if set), returning a status per registry.
func (h *Handler) RegistryCompatibilityMatrix(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawCfg, err := h.repo.GetSetting(ctx, registryConfigSettingKey)
	if err != nil {
		writeJSON(w, http.StatusOK, []RegistryCompatibility{})
		return
	}

	var configs []models.RegistryConfig
	if strings.TrimSpace(rawCfg) != "" {
		if err := json.Unmarshal([]byte(rawCfg), &configs); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":   "invalid registry config in settings",
				"details": err.Error(),
			})
			return
		}
	}

	expectedIS04, _ := h.repo.GetSetting(ctx, "system_expected_is04")

	var out []RegistryCompatibility

	for _, cfg := range configs {
		rc := RegistryCompatibility{
			Name:         cfg.Name,
			Role:         cfg.Role,
			QueryURL:     cfg.QueryURL,
			Enabled:      cfg.Enabled,
			ExpectedIS04: expectedIS04,
		}

		trimmed := strings.TrimSpace(cfg.QueryURL)
		if trimmed == "" {
			rc.Status = "unsupported"
			rc.Error = "empty query_url"
			out = append(out, rc)
			continue
		}

		root := strings.TrimRight(trimmed, "/")
		// Allow both root and /x-nmos/query/<ver> forms, similar to DiscoverNMOSRegistryNodes.
		if idx := strings.Index(root, "/x-nmos/query/"); idx >= 0 {
			root = strings.TrimRight(root[:idx], "/")
		}

		versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/query/", root))
		if err != nil || len(versions) == 0 {
			rc.Status = "error"
			if err != nil {
				rc.Error = err.Error()
			} else {
				rc.Error = "no versions discovered"
			}
			out = append(out, rc)
			continue
		}

		// Normalise versions: trim slashes.
		for i, v := range versions {
			versions[i] = strings.Trim(v, "/")
		}
		sort.Strings(versions)
		rc.DiscoveredVersions = versions
		rc.ChosenQueryVersion = versions[len(versions)-1]

		// Compute status based on expectedIS04 (if set).
		if strings.TrimSpace(expectedIS04) == "" {
			// No expectation configured: treat as ok if we discovered anything.
			rc.Status = "ok"
		} else {
			expected := strings.Trim(expectedIS04, "/")
			foundExact := false
			for _, v := range versions {
				if v == expected {
					foundExact = true
					break
				}
			}
			if foundExact {
				rc.Status = "ok"
			} else if len(versions) > 0 {
				// We discovered some versions but not the expected one.
				rc.Status = "warning"
			} else {
				rc.Status = "unsupported"
			}
		}

		out = append(out, rc)
	}

	writeJSON(w, http.StatusOK, out)
}

