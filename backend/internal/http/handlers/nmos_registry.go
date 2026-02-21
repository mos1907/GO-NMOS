package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
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
			errMsg := "could not discover IS-04 Query API versions"
			if err != nil {
				errMsg = fmt.Sprintf("%s: %v", errMsg, err)
			}
			errMsg += ". From Docker, use host.docker.internal instead of localhost if the registry runs on your host."
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": errMsg})
			return
		}
		sort.Strings(versions)
		version = strings.Trim(versions[len(versions)-1], "/")
	}

	nodesData, err := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/query/%s/nodes", root, version))
	if err != nil {
		errMsg := fmt.Sprintf("failed to read registry nodes: %v", err)
		errMsg += " From Docker, use host.docker.internal instead of localhost if the registry runs on your host."
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": errMsg})
		return
	}

	// When registry is at host.docker.internal, node URLs from registry often use localhost;
	// rewrite so backend (in Docker) can reach nodes on the host for patching.
	rewriteLocalhost := strings.Contains(root, "host.docker.internal")

	nodes := make([]nmosRegistryNodeCandidate, 0, len(nodesData))
	for _, item := range nodesData {
		id := asString(item["id"])
		label := fallback(asString(item["label"]), id)
		href := strings.TrimRight(asString(item["href"]), "/")
		baseURLFromItem := strings.TrimRight(asString(item["base_url"]), "/")

		baseURL := ""
		// First try to use base_url if provided directly
		if baseURLFromItem != "" {
			baseURL = baseURLFromItem
		} else if href != "" {
			// If href is like http://host/x-nmos/node/v1.3, derive base_url=http://host
			if nidx := strings.Index(href, "/x-nmos/node/"); nidx >= 0 {
				baseURL = strings.TrimRight(href[:nidx], "/")
			} else {
				baseURL = href
			}
		}

		if rewriteLocalhost {
			// Allow container to reach nodes on host (localhost -> host.docker.internal)
			if strings.Contains(baseURL, "localhost") {
				baseURL = strings.Replace(baseURL, "localhost", "host.docker.internal", 1)
			}
			if strings.Contains(baseURL, "127.0.0.1") {
				baseURL = strings.Replace(baseURL, "127.0.0.1", "host.docker.internal", 1)
			}
			if strings.Contains(href, "localhost") {
				href = strings.Replace(href, "localhost", "host.docker.internal", 1)
			}
			if strings.Contains(href, "127.0.0.1") {
				href = strings.Replace(href, "127.0.0.1", "host.docker.internal", 1)
			}
		}

		if id == "" && label == "" && href == "" && baseURL == "" {
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
	// D.3: Filter by site if provided
	if siteFilter := strings.TrimSpace(r.URL.Query().Get("site")); siteFilter != "" {
		filtered := make([]models.NMOSNode, 0)
		for _, node := range nodes {
			if node.GetSiteTag() == siteFilter {
				filtered = append(filtered, node)
			}
		}
		writeJSON(w, http.StatusOK, filtered)
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

// DeleteNMOSNodeHandler removes a node from the internal registry (and its devices/senders/receivers via CASCADE).
// DELETE /api/nmos/registry/nodes/:nodeId
func (h *Handler) DeleteNMOSNodeHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := strings.TrimSpace(chi.URLParam(r, "nodeId"))
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id is required"})
		return
	}
	if err := h.repo.DeleteNMOSNode(r.Context(), nodeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "node removed from registry"})
}

func (h *Handler) ListNMOSDevicesHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	devices, err := h.repo.ListNMOSDevices(r.Context(), nodeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list NMOS devices"})
		return
	}
	// D.3: Filter by site if provided
	if siteFilter := strings.TrimSpace(r.URL.Query().Get("site")); siteFilter != "" {
		filtered := make([]models.NMOSDevice, 0)
		for _, device := range devices {
			if device.GetSiteTag() == siteFilter {
				filtered = append(filtered, device)
			}
		}
		writeJSON(w, http.StatusOK, filtered)
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
	// D.3: Filter by site if provided
	if siteFilter := strings.TrimSpace(r.URL.Query().Get("site")); siteFilter != "" {
		filtered := make([]models.NMOSFlow, 0)
		for _, flow := range flows {
			if flow.GetSiteTag() == siteFilter {
				filtered = append(filtered, flow)
			}
		}
		writeJSON(w, http.StatusOK, filtered)
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

// GetNMOSSitesRoomsSummary returns a summary of sites and rooms from the registry (D.1).
// Counts nodes, devices, flows, senders, receivers grouped by site and room tags.
func (h *Handler) GetNMOSSitesRoomsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	nodes, _ := h.repo.ListNMOSNodes(ctx)
	devices, _ := h.repo.ListNMOSDevices(ctx, "")
	flows, _ := h.repo.ListNMOSFlows(ctx)
	senders, _ := h.repo.ListNMOSSenders(ctx, "")
	receivers, _ := h.repo.ListNMOSReceivers(ctx, "")

	// Count by site
	siteCounts := make(map[string]map[string]int) // site -> {nodes, devices, flows, senders, receivers}
	roomCounts := make(map[string]map[string]int) // room -> {nodes, devices, flows, senders, receivers}
	domains := make(map[string]bool)

	for _, n := range nodes {
		site := n.GetSiteTag()
		room := n.GetRoomTag()
		domain := n.GetNetworkDomain()
		if site != "" {
			if siteCounts[site] == nil {
				siteCounts[site] = make(map[string]int)
			}
			siteCounts[site]["nodes"]++
		}
		if room != "" {
			if roomCounts[room] == nil {
				roomCounts[room] = make(map[string]int)
			}
			roomCounts[room]["nodes"]++
		}
		if domain != "" {
			domains[domain] = true
		}
	}
	for _, d := range devices {
		site := d.GetSiteTag()
		room := d.GetRoomTag()
		if site != "" {
			if siteCounts[site] == nil {
				siteCounts[site] = make(map[string]int)
			}
			siteCounts[site]["devices"]++
		}
		if room != "" {
			if roomCounts[room] == nil {
				roomCounts[room] = make(map[string]int)
			}
			roomCounts[room]["devices"]++
		}
	}
	for _, f := range flows {
		site := f.GetSiteTag()
		room := f.GetRoomTag()
		if site != "" {
			if siteCounts[site] == nil {
				siteCounts[site] = make(map[string]int)
			}
			siteCounts[site]["flows"]++
		}
		if room != "" {
			if roomCounts[room] == nil {
				roomCounts[room] = make(map[string]int)
			}
			roomCounts[room]["flows"]++
		}
	}
	for _, s := range senders {
		// Senders inherit from device - find device
		for _, d := range devices {
			if d.ID == s.DeviceID {
				site := d.GetSiteTag()
				room := d.GetRoomTag()
				if site != "" {
					if siteCounts[site] == nil {
						siteCounts[site] = make(map[string]int)
					}
					siteCounts[site]["senders"]++
				}
				if room != "" {
					if roomCounts[room] == nil {
						roomCounts[room] = make(map[string]int)
					}
					roomCounts[room]["senders"]++
				}
				break
			}
		}
	}
	for _, rec := range receivers {
		// Receivers inherit from device
		for _, d := range devices {
			if d.ID == rec.DeviceID {
				site := d.GetSiteTag()
				room := d.GetRoomTag()
				if site != "" {
					if siteCounts[site] == nil {
						siteCounts[site] = make(map[string]int)
					}
					siteCounts[site]["receivers"]++
				}
				if room != "" {
					if roomCounts[room] == nil {
						roomCounts[room] = make(map[string]int)
					}
					roomCounts[room]["receivers"]++
				}
				break
			}
		}
	}

	// Convert to arrays for JSON
	sites := make([]map[string]any, 0, len(siteCounts))
	for site, counts := range siteCounts {
		sites = append(sites, map[string]any{
			"site":   site,
			"counts": counts,
		})
	}
	sort.Slice(sites, func(i, j int) bool {
		return sites[i]["site"].(string) < sites[j]["site"].(string)
	})

	rooms := make([]map[string]any, 0, len(roomCounts))
	for room, counts := range roomCounts {
		rooms = append(rooms, map[string]any{
			"room":   room,
			"counts": counts,
		})
	}
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i]["room"].(string) < rooms[j]["room"].(string)
	})

	domainList := make([]string, 0, len(domains))
	for d := range domains {
		domainList = append(domainList, d)
	}
	sort.Strings(domainList)

	writeJSON(w, http.StatusOK, map[string]any{
		"sites":   sites,
		"rooms":   rooms,
		"domains": domainList,
	})
}
