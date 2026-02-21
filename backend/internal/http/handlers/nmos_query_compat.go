package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-nmos/backend/internal/models"
)

// IS-04 Query API (AMWA spec: Query API is provided by the Registry only, not by nodes).
// We expose internal registry data at /x-nmos/query/<ver>/... so external NMOS clients (e.g. BCC)
// can discover senders, receivers, flows, devices, nodes via the standard Registry endpoint
// without using our auth-only /api/nmos/registry/... endpoints.

const queryCompatVersion = "v1.3"

// supportedQueryVersions is returned for GET /x-nmos/query/
var supportedQueryVersions = []string{"v1.0", "v1.1", "v1.2", "v1.3"}

// QueryCompatVersions handles GET /x-nmos/query/ and GET /x-nmos/query
func (h *Handler) QueryCompatVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(supportedQueryVersions)
}

// QueryCompatVersionRoot handles GET /x-nmos/query/v1.3/
func (h *Handler) QueryCompatVersionRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]string{queryCompatVersion})
}

// QueryCompatNodes handles GET /x-nmos/query/v1.3/nodes
func (h *Handler) QueryCompatNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListNMOSNodes(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list nodes"})
		return
	}
	// IS-04 Query: each node can have id, label, description, href, api (versions), ...
	baseURL := requestBaseURL(r)
	out := make([]map[string]any, 0, len(nodes))
	for _, n := range nodes {
		href := baseURL + "/x-nmos/node/" + queryCompatVersion
		if n.APIVersion != "" {
			href = baseURL + "/x-nmos/node/" + strings.Trim(n.APIVersion, "/")
		}
		item := map[string]any{
			"id":          n.ID,
			"label":       n.Label,
			"description": n.Description,
			"href":        href,
		}
		if n.Hostname != "" {
			item["hostname"] = n.Hostname
		}
		if len(n.Tags) > 0 {
			item["tags"] = rawToAny(n.Tags)
		}
		if len(n.Meta) > 0 {
			item["meta"] = rawToAny(n.Meta)
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, out)
}

// QueryCompatDevices handles GET /x-nmos/query/v1.3/devices
func (h *Handler) QueryCompatDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.repo.ListNMOSDevices(r.Context(), "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list devices"})
		return
	}
	out := nmosDevicesToQueryFormat(devices)
	writeJSON(w, http.StatusOK, out)
}

// QueryCompatFlows handles GET /x-nmos/query/v1.3/flows
func (h *Handler) QueryCompatFlows(w http.ResponseWriter, r *http.Request) {
	flows, err := h.repo.ListNMOSFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list flows"})
		return
	}
	out := nmosFlowsToQueryFormat(flows)
	writeJSON(w, http.StatusOK, out)
}

// QueryCompatSenders handles GET /x-nmos/query/v1.3/senders (fixes 404 for BCC and other NMOS clients)
func (h *Handler) QueryCompatSenders(w http.ResponseWriter, r *http.Request) {
	deviceID := strings.TrimSpace(r.URL.Query().Get("device_id"))
	senders, err := h.repo.ListNMOSSenders(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list senders"})
		return
	}
	out := nmosSendersToQueryFormat(senders)
	writeJSON(w, http.StatusOK, out)
}

// QueryCompatReceivers handles GET /x-nmos/query/v1.3/receivers
func (h *Handler) QueryCompatReceivers(w http.ResponseWriter, r *http.Request) {
	deviceID := strings.TrimSpace(r.URL.Query().Get("device_id"))
	receivers, err := h.repo.ListNMOSReceivers(r.Context(), deviceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list receivers"})
		return
	}
	out := nmosReceiversToQueryFormat(receivers)
	writeJSON(w, http.StatusOK, out)
}

// QueryCompatHealth handles GET /x-nmos/query/v1.3/health (optional, for IS-04 health)
func (h *Handler) QueryCompatHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"health": "ok"})
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func rawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	_ = json.Unmarshal(raw, &v)
	return v
}

func nmosDevicesToQueryFormat(devices []models.NMOSDevice) []map[string]any {
	out := make([]map[string]any, 0, len(devices))
	for _, d := range devices {
		item := map[string]any{
			"id":          d.ID,
			"label":       d.Label,
			"description": d.Description,
			"node_id":     d.NodeID,
			"type":        d.Type,
		}
		if len(d.Tags) > 0 {
			item["tags"] = rawToAny(d.Tags)
		}
		if len(d.Meta) > 0 {
			item["meta"] = rawToAny(d.Meta)
		}
		out = append(out, item)
	}
	return out
}

func nmosFlowsToQueryFormat(flows []models.NMOSFlow) []map[string]any {
	out := make([]map[string]any, 0, len(flows))
	for _, f := range flows {
		item := map[string]any{
			"id":          f.ID,
			"label":       f.Label,
			"description": f.Description,
			"format":      f.Format,
			"source_id":   f.SourceID,
		}
		if len(f.Tags) > 0 {
			item["tags"] = rawToAny(f.Tags)
		}
		if len(f.Meta) > 0 {
			item["meta"] = rawToAny(f.Meta)
		}
		out = append(out, item)
	}
	return out
}

func nmosSendersToQueryFormat(senders []models.NMOSSender) []map[string]any {
	out := make([]map[string]any, 0, len(senders))
	for _, s := range senders {
		item := map[string]any{
			"id":            s.ID,
			"label":         s.Label,
			"description":   s.Description,
			"flow_id":       s.FlowID,
			"transport":     s.Transport,
			"manifest_href": s.ManifestHREF,
			"device_id":     s.DeviceID,
		}
		if len(s.Tags) > 0 {
			item["tags"] = rawToAny(s.Tags)
		}
		if len(s.Meta) > 0 {
			item["meta"] = rawToAny(s.Meta)
		}
		out = append(out, item)
	}
	return out
}

func nmosReceiversToQueryFormat(receivers []models.NMOSReceiver) []map[string]any {
	out := make([]map[string]any, 0, len(receivers))
	for _, r := range receivers {
		item := map[string]any{
			"id":          r.ID,
			"label":       r.Label,
			"description": r.Description,
			"format":      r.Format,
			"transport":   r.Transport,
			"device_id":   r.DeviceID,
		}
		if len(r.Tags) > 0 {
			item["tags"] = rawToAny(r.Tags)
		}
		if len(r.Meta) > 0 {
			item["meta"] = rawToAny(r.Meta)
		}
		out = append(out, item)
	}
	return out
}
