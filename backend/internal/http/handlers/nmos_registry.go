package handlers

import (
	"net/http"
)

// NMOS registry read-only HTTP handlers.
// These expose the internal IS-04 style registry for UI and external tools.

func (h *Handler) ListNMOSNodesHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListNMOSNodes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS nodes"})
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (h *Handler) ListNMOSDevicesHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	devices, err := h.repo.ListNMOSDevices(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS devices"})
		return
	}
	writeJSON(w, http.StatusOK, devices)
}

func (h *Handler) ListNMOSFlowsHandler(w http.ResponseWriter, r *http.Request) {
	flows, err := h.repo.ListNMOSFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS flows"})
		return
	}
	writeJSON(w, http.StatusOK, flows)
}

func (h *Handler) ListNMOSSendersHandler(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	senders, err := h.repo.ListNMOSSenders(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS senders"})
		return
	}
	writeJSON(w, http.StatusOK, senders)
}

func (h *Handler) ListNMOSReceiversHandler(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	receivers, err := h.repo.ListNMOSReceivers(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS receivers"})
		return
	}
	writeJSON(w, http.StatusOK, receivers)
}
