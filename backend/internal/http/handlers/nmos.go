package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

type nmosDiscoverRequest struct {
	BaseURL string `json:"base_url"`
}

type nmosResource struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	FlowID       string `json:"flow_id"`
	ManifestHREF string `json:"manifest_href"`
	Format       string `json:"format,omitempty"`
}

func (h *Handler) DiscoverNMOS(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
	if baseURL == "" && r.Method == http.MethodPost {
		var req nmosDiscoverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			baseURL = strings.TrimSpace(req.BaseURL)
		}
	}
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}
	baseURL = strings.TrimRight(baseURL, "/")

	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover IS-04 versions"})
		return
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")

	sendersData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/senders", baseURL, version))
	receiversData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/receivers", baseURL, version))
	flowsData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/flows", baseURL, version))
	devicesData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/devices", baseURL, version))

	flowFormatByID := map[string]string{}
	for _, item := range flowsData {
		id, _ := item["id"].(string)
		format, _ := item["format"].(string)
		if id != "" {
			flowFormatByID[id] = format
		}
	}

	senders := make([]nmosResource, 0, len(sendersData))
	for _, s := range sendersData {
		flowID, _ := s["flow_id"].(string)
		senders = append(senders, nmosResource{
			ID:           asString(s["id"]),
			Label:        fallback(asString(s["label"]), asString(s["id"])),
			Description:  asString(s["description"]),
			FlowID:       flowID,
			ManifestHREF: asString(s["manifest_href"]),
			Format:       flowFormatByID[flowID],
		})
	}

	receivers := make([]nmosResource, 0, len(receiversData))
	for _, rec := range receiversData {
		receivers = append(receivers, nmosResource{
			ID:          asString(rec["id"]),
			Label:       fallback(asString(rec["label"]), asString(rec["id"])),
			Description: asString(rec["description"]),
			Format:      asString(rec["format"]),
		})
	}

	// Best-effort sync into internal NMOS registry (does not affect response)
	go h.syncNMOSRegistry(baseURL, version, devicesData, flowsData, sendersData, receiversData)

	writeJSON(w, http.StatusOK, map[string]any{
		"base_url":     baseURL,
		"is04_version": version,
		"senders":      senders,
		"receivers":    receivers,
		"counts": map[string]int{
			"senders":   len(senders),
			"receivers": len(receivers),
			"flows":     len(flowsData),
		},
	})
}

func (h *Handler) CheckFlowNMOS(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid flow id"})
		return
	}
	flow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}
	baseURL := strings.TrimSpace(r.URL.Query().Get("base_url"))
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}

	baseURL = strings.TrimRight(baseURL, "/")
	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "nmos node unavailable"})
		return
	}
	sort.Strings(versions)
	version := strings.Trim(versions[len(versions)-1], "/")
	sendersData, err := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/senders", baseURL, version))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "senders read failed"})
		return
	}

	matches := []map[string]any{}
	for _, s := range sendersData {
		if asString(s["flow_id"]) == flow.FlowID {
			matches = append(matches, map[string]any{
				"id":            asString(s["id"]),
				"label":         fallback(asString(s["label"]), asString(s["id"])),
				"manifest_href": asString(s["manifest_href"]),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"flow_id":       flow.FlowID,
		"display_name":  flow.DisplayName,
		"nmos_matches":  matches,
		"matched_count": len(matches),
		"is_match":      len(matches) > 0,
	})
}

// syncNMOSRegistry ingests the discovered NMOS resources into the internal registry tables.
// It is intentionally best-effort and runs in a separate goroutine from DiscoverNMOS.
func (h *Handler) syncNMOSRegistry(baseURL, version string, devicesData, flowsData, sendersData, receiversData []map[string]any) {
	ctx := context.Background()

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return
	}
	hostname := parsed.Hostname()

	// Build nodes from device.node_id references (we may not have explicit /self nodes here)
	nodesByID := map[string]models.NMOSNode{}
	for _, d := range devicesData {
		nodeID := asString(d["node_id"])
		if nodeID == "" {
			continue
		}
		if _, ok := nodesByID[nodeID]; !ok {
			nodesByID[nodeID] = models.NMOSNode{
				ID:         nodeID,
				Label:      nodeID,
				Hostname:   hostname,
				APIVersion: version,
			}
		}
	}

	// Upsert nodes
	for _, node := range nodesByID {
		_ = h.repo.UpsertNMOSNode(ctx, node)
	}

	// Upsert devices
	for _, d := range devicesData {
		devID := asString(d["id"])
		if devID == "" {
			continue
		}
		nodeID := asString(d["node_id"])
		dev := models.NMOSDevice{
			ID:          devID,
			NodeID:      nodeID,
			Label:       fallback(asString(d["label"]), devID),
			Description: asString(d["description"]),
			Type:        asString(d["type"]),
		}
		_ = h.repo.UpsertNMOSDevice(ctx, dev)
	}

	// Upsert flows
	for _, f := range flowsData {
		flowID := asString(f["id"])
		if flowID == "" {
			continue
		}
		flow := models.NMOSFlow{
			ID:          flowID,
			Label:       fallback(asString(f["label"]), flowID),
			Description: asString(f["description"]),
			Format:      asString(f["format"]),
			SourceID:    asString(f["source_id"]),
		}
		_ = h.repo.UpsertNMOSFlow(ctx, flow)
	}

	// Upsert senders
	for _, s := range sendersData {
		senderID := asString(s["id"])
		if senderID == "" {
			continue
		}
		sender := models.NMOSSender{
			ID:           senderID,
			Label:        fallback(asString(s["label"]), senderID),
			Description:  asString(s["description"]),
			FlowID:       asString(s["flow_id"]),
			Transport:    asString(s["transport"]),
			ManifestHREF: asString(s["manifest_href"]),
			DeviceID:     asString(s["device_id"]),
		}
		_ = h.repo.UpsertNMOSSender(ctx, sender)
	}

	// Upsert receivers
	for _, rsrc := range receiversData {
		recID := asString(rsrc["id"])
		if recID == "" {
			continue
		}
		rec := models.NMOSReceiver{
			ID:          recID,
			Label:       fallback(asString(rsrc["label"]), recID),
			Description: asString(rsrc["description"]),
			Format:      asString(rsrc["format"]),
			Transport:   asString(rsrc["transport"]),
			DeviceID:    asString(rsrc["device_id"]),
		}
		_ = h.repo.UpsertNMOSReceiver(ctx, rec)
	}
}

type nmosApplyRequest struct {
	ConnectionURL string `json:"connection_url"`
	SenderID      string `json:"sender_id,omitempty"`
}

func (h *Handler) ApplyFlowNMOS(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt64(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid flow id"})
		return
	}
	flow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}
	var req nmosApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req.ConnectionURL = strings.TrimSpace(req.ConnectionURL)
	if req.ConnectionURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url is required"})
		return
	}
	parsedURL, err := url.Parse(req.ConnectionURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url must be a valid absolute URL"})
		return
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "connection_url must start with http:// or https://"})
		return
	}

	patchBody := map[string]any{
		"activation":    map[string]any{"mode": "activate_immediate"},
		"master_enable": true,
		"transport_params": []map[string]any{
			{
				"multicast_ip":     flow.MulticastIP,
				"source_ip":        flow.SourceIP,
				"destination_port": flow.Port,
				"rtp_enabled":      true,
			},
		},
	}
	if req.SenderID != "" {
		patchBody["sender_id"] = req.SenderID
	}

	client := &http.Client{Timeout: 8 * time.Second}
	payload, _ := json.Marshal(patchBody)
	httpReq, _ := http.NewRequest(http.MethodPatch, parsedURL.String(), strings.NewReader(string(payload)))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "patch request failed"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error":       "nmos apply returned non-2xx status",
			"status_code": resp.StatusCode,
			"body":        string(body),
			"request":     patchBody,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(body),
		"request":     patchBody,
	})
}

func (h *Handler) fetchJSONList(url string) ([]string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var versions []string
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (h *Handler) fetchJSONArray(url string) ([]map[string]any, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}

func parsePathInt64(v string) (int64, error) {
	return strconv.ParseInt(v, 10, 64)
}

// ExplorePortsRequest represents a port exploration request
type ExplorePortsRequest struct {
	Host        string `json:"host"`        // IP or hostname
	Ports       []int  `json:"ports"`       // List of ports to scan
	PortRange   string `json:"port_range"`  // Range like "8080-8090"
	Concurrency int    `json:"concurrency"` // Max concurrent scans (default: 10)
	Timeout     int    `json:"timeout"`     // Timeout per port in seconds (default: 3)
}

// PortScanResult represents a single port scan result
type PortScanResult struct {
	Port        int    `json:"port"`
	IsNMOS      bool   `json:"is_nmos"`
	IsIS04      bool   `json:"is_is04"`
	IsIS05      bool   `json:"is_is05"`
	IS04Version string `json:"is04_version,omitempty"`
	IS05Version string `json:"is05_version,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
	Error       string `json:"error,omitempty"`
	Probability int    `json:"probability"` // 0-100 score
}

// ExplorePorts scans multiple ports on a host for NMOS endpoints
func (h *Handler) ExplorePorts(w http.ResponseWriter, r *http.Request) {
	var req ExplorePortsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	host := strings.TrimSpace(req.Host)
	if host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "host is required"})
		return
	}

	// Parse port range if provided
	ports := req.Ports
	if req.PortRange != "" {
		rangePorts, err := parsePortRange(req.PortRange)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid port_range format (use 'start-end')"})
			return
		}
		ports = append(ports, rangePorts...)
	}

	if len(ports) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ports or port_range is required"})
		return
	}

	// Validate and set defaults
	concurrency := req.Concurrency
	if concurrency <= 0 || concurrency > 50 {
		concurrency = 10 // Default and max limit
	}

	timeout := req.Timeout
	if timeout <= 0 || timeout > 30 {
		timeout = 3 // Default 3 seconds, max 30
	}

	// Check if host is local (security warning)
	isLocal := isLocalHost(host)

	// Scan ports concurrently
	results := scanPortsConcurrently(host, ports, concurrency, timeout)

	writeJSON(w, http.StatusOK, map[string]any{
		"host":        host,
		"is_local":    isLocal,
		"total_ports": len(ports),
		"scanned":     len(results),
		"found_nmos":  countNMOS(results),
		"concurrency": concurrency,
		"timeout_sec": timeout,
		"results":     results,
	})
}

func parsePortRange(rangeStr string) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format")
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("invalid port numbers")
	}
	if start > end || start < 1 || end > 65535 {
		return nil, fmt.Errorf("invalid port range")
	}
	ports := make([]int, 0, end-start+1)
	for p := start; p <= end; p++ {
		ports = append(ports, p)
	}
	return ports, nil
}

func isLocalHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	localHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
	for _, lh := range localHosts {
		if host == lh {
			return true
		}
	}
	// Check if it's a private IP range
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.") {
		return true
	}
	return false
}

func scanPortsConcurrently(host string, ports []int, concurrency, timeoutSec int) []PortScanResult {
	sem := make(chan struct{}, concurrency)
	results := make([]PortScanResult, len(ports))
	var wg sync.WaitGroup

	for i, port := range ports {
		wg.Add(1)
		go func(idx int, p int) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			result := scanPort(host, p, timeoutSec)
			results[idx] = result
		}(i, port)
	}

	wg.Wait()
	return results
}

func scanPort(host string, port int, timeoutSec int) PortScanResult {
	result := PortScanResult{
		Port:        port,
		IsNMOS:      false,
		IsIS04:      false,
		IsIS05:      false,
		Probability: 0,
	}

	// Common NMOS ports get higher probability
	commonPorts := map[int]int{
		8080: 80,
		8081: 70,
		8082: 60,
		8083: 50,
		8084: 40,
		80:   30,
		443:  30,
		8443: 30,
	}
	if score, ok := commonPorts[port]; ok {
		result.Probability = score
	}

	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	// Try IS-04 Node API
	is04Versions, err := fetchJSONListWithClient(client, fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err == nil && len(is04Versions) > 0 {
		sort.Strings(is04Versions)
		version := strings.Trim(is04Versions[len(is04Versions)-1], "/")
		result.IsIS04 = true
		result.IsNMOS = true
		result.IS04Version = version
		result.BaseURL = baseURL
		result.Probability = 100
		return result
	}

	// Try IS-05 Connection API
	is05Versions, err := fetchJSONListWithClient(client, fmt.Sprintf("%s/x-nmos/connection/", baseURL))
	if err == nil && len(is05Versions) > 0 {
		sort.Strings(is05Versions)
		version := strings.Trim(is05Versions[len(is05Versions)-1], "/")
		result.IsIS05 = true
		result.IsNMOS = true
		result.IS05Version = version
		result.BaseURL = baseURL
		result.Probability = 90
		return result
	}

	// Try IS-04 Query API (Registry)
	queryVersions, err := fetchJSONListWithClient(client, fmt.Sprintf("%s/x-nmos/query/", baseURL))
	if err == nil && len(queryVersions) > 0 {
		sort.Strings(queryVersions)
		version := strings.Trim(queryVersions[len(queryVersions)-1], "/")
		result.IsIS04 = true
		result.IsNMOS = true
		result.IS04Version = version
		result.BaseURL = baseURL
		result.Probability = 95
		return result
	}

	if err != nil {
		result.Error = err.Error()
	}

	return result
}

func fetchJSONListWithClient(client *http.Client, url string) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var versions []string
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func countNMOS(results []PortScanResult) int {
	count := 0
	for _, r := range results {
		if r.IsNMOS {
			count++
		}
	}
	return count
}

// DetectIS05Endpoint automatically detects IS-05 connection API endpoint from IS-04 base URL.
// Tries common patterns: /x-nmos/connection/<version>
func (h *Handler) DetectIS05Endpoint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BaseURL string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	baseURL := strings.TrimSpace(req.BaseURL)
	if baseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "base_url is required"})
		return
	}
	baseURL = strings.TrimRight(baseURL, "/")

	// First, get IS-04 version to use same version for IS-05
	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", baseURL))
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover IS-04 versions"})
		return
	}
	sort.Strings(versions)
	is04Version := strings.Trim(versions[len(versions)-1], "/")

	// Try IS-05 connection API with same version
	is05Base := fmt.Sprintf("%s/x-nmos/connection/%s", baseURL, is04Version)

	// Test if IS-05 endpoint exists by checking /single/receivers
	client := &http.Client{Timeout: 5 * time.Second}
	testURL := fmt.Sprintf("%s/single/receivers", is05Base)
	resp, err := client.Get(testURL)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 || resp.StatusCode == 404 {
			// 200 = exists, 404 = endpoint exists but no receivers (still valid)
			writeJSON(w, http.StatusOK, map[string]any{
				"base_url":      baseURL,
				"is04_version":  is04Version,
				"is05_base_url": is05Base,
				"detected":      true,
			})
			return
		}
	}

	// If standard pattern failed, try to discover IS-05 versions
	is05Versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/connection/", baseURL))
	if err == nil && len(is05Versions) > 0 {
		sort.Strings(is05Versions)
		is05Version := strings.Trim(is05Versions[len(is05Versions)-1], "/")
		detectedURL := fmt.Sprintf("%s/x-nmos/connection/%s", baseURL, is05Version)
		writeJSON(w, http.StatusOK, map[string]any{
			"base_url":      baseURL,
			"is04_version":  is04Version,
			"is05_base_url": detectedURL,
			"is05_version":  is05Version,
			"detected":      true,
		})
		return
	}

	// Fallback: return best guess
	writeJSON(w, http.StatusOK, map[string]any{
		"base_url":      baseURL,
		"is04_version":  is04Version,
		"is05_base_url": is05Base,
		"detected":      false,
		"note":          "IS-05 endpoint not verified; using standard pattern",
	})
}

// DetectIS04FromRDS detects IS-04 Node API endpoint from RDS (Registry) Query API URL.
func (h *Handler) DetectIS04FromRDS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RDSQueryURL string `json:"rds_query_url"`
		NodeID      string `json:"node_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	rdsURL := strings.TrimSpace(req.RDSQueryURL)
	if rdsURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "rds_query_url is required"})
		return
	}
	rdsURL = strings.TrimRight(rdsURL, "/")

	// Extract root and version from RDS URL
	root := rdsURL
	version := ""
	if idx := strings.Index(rdsURL, "/x-nmos/query/"); idx >= 0 {
		root = strings.TrimRight(rdsURL[:idx], "/")
		rest := strings.Trim(rdsURL[idx+len("/x-nmos/query/"):], "/")
		if rest != "" {
			version = strings.Split(rest, "/")[0]
		}
	}
	if version == "" {
		versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/query/", root))
		if err != nil || len(versions) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover Query API versions"})
			return
		}
		sort.Strings(versions)
		version = strings.Trim(versions[len(versions)-1], "/")
	}

	// Get nodes from registry
	nodesData, err := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/query/%s/nodes", root, version))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to read registry nodes"})
		return
	}

	// If node_id provided, find that specific node
	if req.NodeID != "" {
		for _, node := range nodesData {
			if asString(node["id"]) == req.NodeID {
				href := strings.TrimRight(asString(node["href"]), "/")
				if href != "" {
					// Extract base URL from href (e.g., http://host/x-nmos/node/v1.3 -> http://host)
					var baseURL string
					if nidx := strings.Index(href, "/x-nmos/node/"); nidx >= 0 {
						baseURL = strings.TrimRight(href[:nidx], "/")
					} else {
						baseURL = href
					}
					writeJSON(w, http.StatusOK, map[string]any{
						"node_id":       req.NodeID,
						"is04_base_url": baseURL,
						"href":          href,
						"detected":      true,
					})
					return
				}
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found in registry"})
		return
	}

	// Return all nodes with their IS-04 base URLs
	results := make([]map[string]any, 0, len(nodesData))
	for _, node := range nodesData {
		id := asString(node["id"])
		href := strings.TrimRight(asString(node["href"]), "/")
		if href == "" {
			continue
		}
		var baseURL string
		if nidx := strings.Index(href, "/x-nmos/node/"); nidx >= 0 {
			baseURL = strings.TrimRight(href[:nidx], "/")
		} else {
			baseURL = href
		}
		results = append(results, map[string]any{
			"node_id":       id,
			"label":         fallback(asString(node["label"]), id),
			"is04_base_url": baseURL,
			"href":          href,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"rds_query_url": rdsURL,
		"query_version": version,
		"nodes":         results,
		"count":         len(results),
	})
}
