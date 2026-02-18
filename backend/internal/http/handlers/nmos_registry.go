package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// NMOS registry read-only HTTP handlers.
// These expose the internal IS-04 style registry for UI and external tools.

type nmosRegistryDiscoverNodesRequest struct {
	QueryURL string `json:"query_url"`
}

type nmosRegistryNodeCandidate struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Href    string `json:"href,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
}

// DiscoverNMOSRegistryNodes discovers available NMOS Nodes from an IS-04 Query API ("registry") URL.
// This mirrors the "Connect RDS" flow in tools like NMOS Simple BCC.
//
// It accepts either:
// - root registry URL: http(s)://host:port
// - versioned query API URL: http(s)://host:port/x-nmos/query/<ver>
func (h *Handler) DiscoverNMOSRegistryNodes(w http.ResponseWriter, r *http.Request) {
	var req nmosRegistryDiscoverNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	raw := strings.TrimSpace(req.QueryURL)
	if raw == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query_url is required"})
		return
	}
	raw = strings.TrimRight(raw, "/")

	root := raw
	version := ""

	// If the user provided a versioned Query API URL, extract root + version.
	if idx := strings.Index(raw, "/x-nmos/query/"); idx >= 0 {
		root = strings.TrimRight(raw[:idx], "/")
		rest := strings.Trim(raw[idx+len("/x-nmos/query/"):], "/")
		if rest != "" {
			version = strings.Split(rest, "/")[0]
		}
	}

	// Discover versions only when version isn't provided.
	if version == "" {
		versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/query/", root))
		if err != nil || len(versions) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover IS-04 Query API versions"})
			return
		}
		sort.Strings(versions)
		version = strings.Trim(versions[len(versions)-1], "/")
	}

	nodesData, err := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/query/%s/nodes", root, version))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to read registry nodes"})
		return
	}

	nodes := make([]nmosRegistryNodeCandidate, 0, len(nodesData))
	for _, item := range nodesData {
		id := asString(item["id"])
		label := fallback(asString(item["label"]), id)
		href := strings.TrimRight(asString(item["href"]), "/")

		baseURL := ""
		// If href is like http://host/x-nmos/node/v1.3, derive base_url=http://host
		if href != "" {
			if nidx := strings.Index(href, "/x-nmos/node/"); nidx >= 0 {
				baseURL = strings.TrimRight(href[:nidx], "/")
			} else {
				baseURL = href
			}
		}

		if id == "" && label == "" && href == "" {
			continue
		}

		nodes = append(nodes, nmosRegistryNodeCandidate{
			ID:      id,
			Label:   label,
			Href:    href,
			BaseURL: baseURL,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query_root":     root,
		"is04_query_ver": version,
		"nodes":          nodes,
		"count":          len(nodes),
	})
}

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
