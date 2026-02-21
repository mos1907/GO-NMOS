package handlers

import (
	"encoding/json"
	"net/http"

	"go-nmos/backend/internal/interop"
)

// RunInteropTests runs interoperability tests against a target device or registry
// POST /api/interop/test
func (h *Handler) RunInteropTests(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target interop.TestTarget `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.Target.BaseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "target.base_url is required"})
		return
	}

	suite := interop.NewTestSuite()
	ctx := r.Context()

	var results []interop.TestResult
	if req.Target.Type == "registry" {
		results = suite.RunRegistryTests(ctx, req.Target)
	} else {
		results = suite.RunNodeTests(ctx, req.Target)
	}

	// Calculate summary
	passed := 0
	failed := 0
	warnings := 0
	skipped := 0
	for _, result := range results {
		switch result.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "warning":
			warnings++
		case "skip":
			skipped++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"target": req.Target,
		"results": results,
		"summary": map[string]int{
			"total":   len(results),
			"passed":  passed,
			"failed":  failed,
			"warnings": warnings,
			"skipped": skipped,
		},
	})
}

// ListInteropTargets returns a list of known test targets (reference devices/registries)
// GET /api/interop/targets
func (h *Handler) ListInteropTargets(w http.ResponseWriter, r *http.Request) {
	// Reference targets - these can be extended with database-stored targets
	targets := []interop.TestTarget{
		{
			Name:        "NMOS Test Suite (Reference)",
			Type:        "registry",
			BaseURL:     "https://nmos-test.amwa.tv",
			Vendor:      "AMWA",
			Description: "AMWA NMOS Test Suite reference registry",
		},
		{
			Name:        "NMOS Test Suite Node",
			Type:        "node",
			BaseURL:     "https://nmos-test.amwa.tv",
			Vendor:      "AMWA",
			Description: "AMWA NMOS Test Suite reference node",
		},
		// Add more reference targets as needed
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"targets": targets,
	})
}
