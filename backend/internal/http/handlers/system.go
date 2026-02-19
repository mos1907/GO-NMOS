package handlers

import "net/http"

// SystemInfo returns basic system / timing parameters (IS-09-inspired).
func (h *Handler) SystemInfo(w http.ResponseWriter, r *http.Request) {
	get := func(key string) string {
		v, err := h.repo.GetSetting(r.Context(), key)
		if err != nil {
			return ""
		}
		return v
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ptp_domain":       get("system_ptp_domain"),
		"ptp_gmid":         get("system_ptp_gmid"),
		"expected_is04":    get("system_expected_is04"),
		"expected_is05":    get("system_expected_is05"),
		"api_base_url":     get("api_base_url"),
		"flow_lock_role":   get("flow_lock_role"),
		"anonymous_access": get("anonymous_access"),
		"hard_delete":      get("hard_delete_enabled"),
	})
}
