package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-nmos/backend/internal/models"
)

// SystemBackup represents a complete system backup (settings, policies, registry config, flows).
type SystemBackup struct {
	Version       string                      `json:"version"`
	Timestamp     string                      `json:"timestamp"`
	Settings      map[string]string           `json:"settings"`
	RegistryConfig []models.RegistryConfig    `json:"registry_config"`
	RoutingPolicies []models.RoutingPolicy    `json:"routing_policies"`
	Flows         []models.Flow               `json:"flows"`
}

// SystemRestoreRequest represents a restore request.
type SystemRestoreRequest struct {
	Backup        SystemBackup `json:"backup"`
	RestoreFlows  bool         `json:"restore_flows,omitempty"`
	RestoreSettings bool       `json:"restore_settings,omitempty"`
	RestorePolicies bool       `json:"restore_policies,omitempty"`
	RestoreRegistryConfig bool `json:"restore_registry_config,omitempty"`
}

// ExportSystemBackup exports a complete system backup.
// GET /api/system/backup
func (h *Handler) ExportSystemBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	backup := SystemBackup{
		Version:   "1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Settings:  make(map[string]string),
	}

	// Export settings
	settingsKeys := []string{"api_base_url", "anonymous_access", "flow_lock_role", "hard_delete_enabled", "system_ptp_domain", "system_expected_is04", "system_expected_is05", "sdn_controller_url"}
	for _, key := range settingsKeys {
		if val, err := h.repo.GetSetting(ctx, key); err == nil {
			backup.Settings[key] = val
		}
	}

	// Export registry config
	if regConfig, err := h.repo.GetSetting(ctx, "nmos_registry_config"); err == nil && regConfig != "" {
		var configs []models.RegistryConfig
		if err := json.Unmarshal([]byte(regConfig), &configs); err == nil {
			backup.RegistryConfig = configs
		}
	}

	// Export routing policies
	if policies, err := h.repo.ListRoutingPolicies(ctx, false); err == nil {
		backup.RoutingPolicies = policies
	}

	// Export flows
	if flows, err := h.repo.ExportFlows(ctx); err == nil {
		backup.Flows = flows
	}

	writeJSON(w, http.StatusOK, backup)
}

// ImportSystemBackup imports a system backup.
// POST /api/system/restore
func (h *Handler) ImportSystemBackup(w http.ResponseWriter, r *http.Request) {
	var req SystemRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	ctx := r.Context()
	restored := map[string]int{
		"settings":          0,
		"registry_config":   0,
		"routing_policies":  0,
		"flows":             0,
	}

	// Restore settings
	if req.RestoreSettings {
		for key, value := range req.Backup.Settings {
			if err := h.repo.SetSetting(ctx, key, value); err == nil {
				restored["settings"]++
			}
		}
	}

	// Restore registry config
	if req.RestoreRegistryConfig && len(req.Backup.RegistryConfig) > 0 {
		if data, err := json.Marshal(req.Backup.RegistryConfig); err == nil {
			if err := h.repo.SetSetting(ctx, "nmos_registry_config", string(data)); err == nil {
				restored["registry_config"] = len(req.Backup.RegistryConfig)
			}
		}
	}

	// Restore routing policies
	if req.RestorePolicies && len(req.Backup.RoutingPolicies) > 0 {
		for _, policy := range req.Backup.RoutingPolicies {
			// Delete existing policy if exists, then create new
			if policy.ID > 0 {
				_ = h.repo.DeleteRoutingPolicy(ctx, policy.ID)
			}
			policy.ID = 0 // Reset ID for new creation
			if _, err := h.repo.CreateRoutingPolicy(ctx, policy); err == nil {
				restored["routing_policies"]++
			}
		}
	}

	// Restore flows
	if req.RestoreFlows && len(req.Backup.Flows) > 0 {
		if imported, err := h.repo.ImportFlows(ctx, req.Backup.Flows); err == nil {
			restored["flows"] = imported
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"restored": restored,
	})
}
