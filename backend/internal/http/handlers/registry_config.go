package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-nmos/backend/internal/models"
)

const registryConfigSettingKey = "nmos_registry_config"

// normalizeRegistryURL trims and strips trailing slash for consistent comparison.
func normalizeRegistryURL(u string) string {
	return strings.TrimRight(strings.TrimSpace(u), "/")
}

// GetRegistryConfig returns the configured external IS-04 registries.
// The data is stored as JSON in the settings table under nmos_registry_config.
func (h *Handler) GetRegistryConfig(w http.ResponseWriter, r *http.Request) {
	raw, err := h.repo.GetSetting(r.Context(), registryConfigSettingKey)
	if err != nil || raw == "" {
		// If not configured yet, return an empty list.
		writeJSON(w, http.StatusOK, []models.RegistryConfig{})
		return
	}

	var configs []models.RegistryConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		// On parse error, surface a 500 so the admin can fix the stored value.
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"error":   "invalid registry config in settings",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, configs)
}

// PutRegistryConfig replaces the configured registries with the provided list.
// Only minimally validates fields; deeper policy/role semantics can evolve later.
func (h *Handler) PutRegistryConfig(w http.ResponseWriter, r *http.Request) {
	var configs []models.RegistryConfig
	if err := json.NewDecoder(r.Body).Decode(&configs); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	// Basic validation: name and query_url must not be empty when enabled.
	for i, c := range configs {
		if c.Enabled {
			if c.Name == "" || c.QueryURL == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{
					"error":        "name and query_url are required for enabled registries",
					"invalid_index": i,
				})
				return
			}
		}
	}

	data, err := json.Marshal(configs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode config"})
		return
	}

	if err := h.repo.SetSetting(r.Context(), registryConfigSettingKey, string(data)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save config"})
		return
	}

	writeJSON(w, http.StatusOK, configs)
}

// RegistryStatsItem is returned by GetRegistryConfigStats for each configured registry.
type RegistryStatsItem struct {
	QueryURL  string `json:"query_url"`
	Name      string `json:"name,omitempty"`
	Nodes     int    `json:"nodes"`
	Devices   int    `json:"devices"`
	Senders   int    `json:"senders"`
	Receivers int    `json:"receivers"`
	Error     string `json:"error,omitempty"`
}

// GetRegistryConfigStats returns resource counts (nodes, devices, senders, receivers) for each configured registry.
// GET /api/registry/config/stats
func (h *Handler) GetRegistryConfigStats(w http.ResponseWriter, r *http.Request) {
	raw, err := h.repo.GetSetting(r.Context(), registryConfigSettingKey)
	if err != nil || raw == "" {
		writeJSON(w, http.StatusOK, []RegistryStatsItem{})
		return
	}
	var configs []models.RegistryConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid registry config"})
		return
	}
	result := make([]RegistryStatsItem, 0, len(configs))
	for _, c := range configs {
		item := RegistryStatsItem{QueryURL: c.QueryURL, Name: c.Name}
		nodes, devices, senders, receivers, err := h.getRegistryCounts(c.QueryURL)
		if err != nil {
			item.Error = err.Error()
		} else {
			item.Nodes = nodes
			item.Devices = devices
			item.Senders = senders
			item.Receivers = receivers
		}
		result = append(result, item)
	}
	writeJSON(w, http.StatusOK, result)
}

// RemoveRegistryConfigRequest is the body for RemoveRegistryConfig.
type RemoveRegistryConfigRequest struct {
	QueryURL string `json:"query_url"`
}

// RemoveRegistryConfig removes one registry from config and deletes all internal nodes that were
// discovered from that RDS (so Registry Patch sources/destinations from that RDS are removed).
// POST /api/registry/config/remove. Body: { "query_url": "http://..." }.
// Returns deleted_nodes count. Registry must be reachable to list node IDs.
func (h *Handler) RemoveRegistryConfig(w http.ResponseWriter, r *http.Request) {
	var req RemoveRegistryConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	queryURL := normalizeRegistryURL(req.QueryURL)
	if queryURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query_url is required"})
		return
	}

	nodeIDs, err := h.getRegistryNodeIDs(queryURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": "could not list nodes from registry (is it reachable?): " + err.Error(),
		})
		return
	}

	ctx := r.Context()
	deleted := 0
	for _, nodeID := range nodeIDs {
		if nodeID == "" {
			continue
		}
		if err := h.repo.DeleteNMOSNode(ctx, nodeID); err != nil {
			continue
		}
		deleted++
	}

	raw, err := h.repo.GetSetting(ctx, registryConfigSettingKey)
	if err != nil || raw == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"deleted_nodes": deleted,
			"message":       "Registry entry not in config or already removed.",
		})
		return
	}
	var configs []models.RegistryConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "invalid registry config"})
		return
	}
	before := len(configs)
	configs = filterRegistryConfigByQueryURL(configs, queryURL)
	if len(configs) < before {
		data, _ := json.Marshal(configs)
		_ = h.repo.SetSetting(ctx, registryConfigSettingKey, string(data))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"deleted_nodes": deleted,
		"message":       fmt.Sprintf("Registry removed. %d node(s) removed from Registry Patch.", deleted),
	})
}

func filterRegistryConfigByQueryURL(configs []models.RegistryConfig, queryURL string) []models.RegistryConfig {
	out := make([]models.RegistryConfig, 0, len(configs))
	for _, c := range configs {
		if normalizeRegistryURL(c.QueryURL) != queryURL {
			out = append(out, c)
		}
	}
	return out
}

