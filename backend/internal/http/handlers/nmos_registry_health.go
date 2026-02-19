package handlers

import (
	"net/http"
	"time"
)

// NMOSRegistryHealth reports the status of the internal NMOS registry tables
// used by the UI (nodes/devices/flows/senders/receivers).
func (h *Handler) NMOSRegistryHealth(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListNMOSNodes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to list NMOS nodes"})
		return
	}
	devices, err := h.repo.ListNMOSDevices(r.Context(), "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to list NMOS devices"})
		return
	}
	flows, err := h.repo.ListNMOSFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to list NMOS flows"})
		return
	}
	senders, err := h.repo.ListNMOSSenders(r.Context(), "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to list NMOS senders"})
		return
	}
	receivers, err := h.repo.ListNMOSReceivers(r.Context(), "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to list NMOS receivers"})
		return
	}

	ok := len(nodes) > 0 || len(devices) > 0 || len(flows) > 0 || len(senders) > 0 || len(receivers) > 0
	status := "empty"
	if ok {
		status = "ok"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        ok,
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"counts": map[string]int{
			"nodes":     len(nodes),
			"devices":   len(devices),
			"flows":     len(flows),
			"senders":   len(senders),
			"receivers": len(receivers),
		},
	})
}
