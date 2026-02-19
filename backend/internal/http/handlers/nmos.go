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
	"go-nmos/backend/internal/sdp"

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

type flowSnapshot struct {
	// ST2022-7 A/B (best-effort; may be empty if IS-05 transport params not available)
	SourceAddrA    string `json:"source_addr_a,omitempty"`
	SourcePortA    int    `json:"source_port_a,omitempty"`
	MulticastAddrA string `json:"multicast_addr_a,omitempty"`
	GroupPortA     int    `json:"group_port_a,omitempty"`
	SourceAddrB    string `json:"source_addr_b,omitempty"`
	SourcePortB    int    `json:"source_port_b,omitempty"`
	MulticastAddrB string `json:"multicast_addr_b,omitempty"`
	GroupPortB     int    `json:"group_port_b,omitempty"`

	// NMOS
	NMOSNodeID          string `json:"nmos_node_id,omitempty"`
	NMOSFlowID          string `json:"nmos_flow_id,omitempty"`
	NMOSSenderID        string `json:"nmos_sender_id,omitempty"`
	NMOSDeviceID        string `json:"nmos_device_id,omitempty"`
	NMOSNodeLabel       string `json:"nmos_node_label,omitempty"`
	NMOSNodeDescription string `json:"nmos_node_description,omitempty"`
	NMOSIS04BaseURL     string `json:"nmos_is04_base_url,omitempty"`
	NMOSIS05BaseURL     string `json:"nmos_is05_base_url,omitempty"`
	NMOSIS04Version     string `json:"nmos_is04_version,omitempty"`
	NMOSIS05Version     string `json:"nmos_is05_version,omitempty"`

	// SDP / media
	SDPURL            string `json:"sdp_url,omitempty"`
	SDPCache          string `json:"sdp_cache,omitempty"`
	MediaType         string `json:"media_type,omitempty"`
	RedundancyGroup   string `json:"redundancy_group,omitempty"`
	TransportProtocol string `json:"transport_protocol,omitempty"`
	ST2110Format      string `json:"st2110_format,omitempty"`

	// Raw NMOS payloads for debugging
	RawNode   map[string]any `json:"raw_node,omitempty"`
	RawFlow   map[string]any `json:"raw_flow,omitempty"`
	RawSender map[string]any `json:"raw_sender,omitempty"`
	RawIS05   map[string]any `json:"raw_is05,omitempty"`
}

// GetFlowNMOSSnapShot returns a best-effort NMOS snapshot for a flow.
// It prefers the flow's stored NMOS base URLs/versions, but can also accept overrides via query params.
func (h *Handler) GetFlowNMOSSnapShot(w http.ResponseWriter, r *http.Request) {
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

	is04Base := strings.TrimSpace(r.URL.Query().Get("is04_base_url"))
	if is04Base == "" {
		is04Base = strings.TrimSpace(flow.NMOSIS04BaseURL)
	}
	is05Base := strings.TrimSpace(r.URL.Query().Get("is05_base_url"))
	if is05Base == "" {
		is05Base = strings.TrimSpace(flow.NMOSIS05BaseURL)
	}
	timeoutSec := 6
	if t := strings.TrimSpace(r.URL.Query().Get("timeout")); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 && n <= 30 {
			timeoutSec = n
		}
	}
	if is04Base == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is04_base_url is required (or store nmos_is04_base_url on the flow)"})
		return
	}
	is04Base = strings.TrimRight(is04Base, "/")
	if is05Base != "" {
		is05Base = strings.TrimRight(is05Base, "/")
	}

	snap, err := h.buildNMOSSnapShot(*flow, is04Base, is05Base, timeoutSec)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"flow_id":  flow.FlowID,
		"flow_db":  flow,
		"snapshot": snap,
	})
}

type syncFromNMOSRequest struct {
	Fields      []string `json:"fields"`
	Is04BaseURL string   `json:"is04_base_url,omitempty"`
	Is05BaseURL string   `json:"is05_base_url,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
}

// SyncFlowFromNMOS pulls NMOS/SDP/IS-05 transport params and updates the flow record for selected fields.
func (h *Handler) SyncFlowFromNMOS(w http.ResponseWriter, r *http.Request) {
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
	if flow.Locked {
		writeJSON(w, http.StatusLocked, map[string]string{"error": "flow is locked. Unlock before syncing from NMOS."})
		return
	}

	var req syncFromNMOSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if len(req.Fields) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "fields is required"})
		return
	}
	timeoutSec := req.Timeout
	if timeoutSec <= 0 || timeoutSec > 30 {
		timeoutSec = 6
	}

	is04Base := strings.TrimSpace(req.Is04BaseURL)
	if is04Base == "" {
		is04Base = strings.TrimSpace(flow.NMOSIS04BaseURL)
	}
	is05Base := strings.TrimSpace(req.Is05BaseURL)
	if is05Base == "" {
		is05Base = strings.TrimSpace(flow.NMOSIS05BaseURL)
	}
	if is04Base == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is04_base_url is required (or store nmos_is04_base_url on the flow)"})
		return
	}
	is04Base = strings.TrimRight(is04Base, "/")
	if is05Base != "" {
		is05Base = strings.TrimRight(is05Base, "/")
	}

	snap, err := h.buildNMOSSnapShot(*flow, is04Base, is05Base, timeoutSec)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	allowed := map[string]func(map[string]any, flowSnapshot){
		"source_addr_a":         func(u map[string]any, s flowSnapshot) { u["source_addr_a"] = s.SourceAddrA },
		"source_port_a":         func(u map[string]any, s flowSnapshot) { u["source_port_a"] = s.SourcePortA },
		"multicast_addr_a":      func(u map[string]any, s flowSnapshot) { u["multicast_addr_a"] = s.MulticastAddrA },
		"group_port_a":          func(u map[string]any, s flowSnapshot) { u["group_port_a"] = s.GroupPortA },
		"source_addr_b":         func(u map[string]any, s flowSnapshot) { u["source_addr_b"] = s.SourceAddrB },
		"source_port_b":         func(u map[string]any, s flowSnapshot) { u["source_port_b"] = s.SourcePortB },
		"multicast_addr_b":      func(u map[string]any, s flowSnapshot) { u["multicast_addr_b"] = s.MulticastAddrB },
		"group_port_b":          func(u map[string]any, s flowSnapshot) { u["group_port_b"] = s.GroupPortB },
		"nmos_node_id":          func(u map[string]any, s flowSnapshot) { u["nmos_node_id"] = s.NMOSNodeID },
		"nmos_flow_id":          func(u map[string]any, s flowSnapshot) { u["nmos_flow_id"] = s.NMOSFlowID },
		"nmos_sender_id":        func(u map[string]any, s flowSnapshot) { u["nmos_sender_id"] = s.NMOSSenderID },
		"nmos_device_id":        func(u map[string]any, s flowSnapshot) { u["nmos_device_id"] = s.NMOSDeviceID },
		"nmos_node_label":       func(u map[string]any, s flowSnapshot) { u["nmos_node_label"] = s.NMOSNodeLabel },
		"nmos_node_description": func(u map[string]any, s flowSnapshot) { u["nmos_node_description"] = s.NMOSNodeDescription },
		"nmos_is04_base_url":    func(u map[string]any, s flowSnapshot) { u["nmos_is04_base_url"] = s.NMOSIS04BaseURL },
		"nmos_is05_base_url":    func(u map[string]any, s flowSnapshot) { u["nmos_is05_base_url"] = s.NMOSIS05BaseURL },
		"nmos_is04_version":     func(u map[string]any, s flowSnapshot) { u["nmos_is04_version"] = s.NMOSIS04Version },
		"nmos_is05_version":     func(u map[string]any, s flowSnapshot) { u["nmos_is05_version"] = s.NMOSIS05Version },
		"sdp_url":               func(u map[string]any, s flowSnapshot) { u["sdp_url"] = s.SDPURL },
		"sdp_cache":             func(u map[string]any, s flowSnapshot) { u["sdp_cache"] = s.SDPCache },
		"media_type":            func(u map[string]any, s flowSnapshot) { u["media_type"] = s.MediaType },
		"redundancy_group":      func(u map[string]any, s flowSnapshot) { u["redundancy_group"] = s.RedundancyGroup },
		"transport_protocol":    func(u map[string]any, s flowSnapshot) { u["transport_protocol"] = s.TransportProtocol },
		"st2110_format":         func(u map[string]any, s flowSnapshot) { u["st2110_format"] = s.ST2110Format },
		"data_source":           func(u map[string]any, s flowSnapshot) { u["data_source"] = "nmos" },
	}

	updates := map[string]any{}
	applied := []string{}
	for _, f := range req.Fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if apply, ok := allowed[f]; ok {
			apply(updates, snap)
			applied = append(applied, f)
		}
	}
	if len(updates) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no valid fields to sync"})
		return
	}

	if err := h.repo.PatchFlow(r.Context(), id, updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}
	updatedFlow, _ := h.repo.GetFlowByID(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"applied_fields": applied,
		"snapshot":       snap,
		"flow":           updatedFlow,
	})
}

func (h *Handler) buildNMOSSnapShot(flow models.Flow, is04Base, is05Base string, timeoutSec int) (flowSnapshot, error) {
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	versions, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/node/", is04Base))
	if err != nil || len(versions) == 0 {
		return flowSnapshot{}, fmt.Errorf("could not discover IS-04 versions")
	}
	sort.Strings(versions)
	is04Ver := strings.Trim(versions[len(versions)-1], "/")

	selfObj, _ := h.fetchJSONMapWithClient(client, fmt.Sprintf("%s/x-nmos/node/%s/self", is04Base, is04Ver))
	nodeID := asString(selfObj["id"])
	nodeLabel := asString(selfObj["label"])
	nodeDesc := asString(selfObj["description"])

	nmosFlowID := strings.TrimSpace(flow.NMOSFlowID)
	if nmosFlowID == "" {
		nmosFlowID = flow.FlowID
	}
	flowObj, _ := h.fetchJSONMapWithClient(client, fmt.Sprintf("%s/x-nmos/node/%s/flows/%s", is04Base, is04Ver, nmosFlowID))

	sendersData, _ := h.fetchJSONArray(fmt.Sprintf("%s/x-nmos/node/%s/senders", is04Base, is04Ver))
	var senderObj map[string]any
	for _, s := range sendersData {
		if asString(s["flow_id"]) == nmosFlowID {
			senderObj = s
			break
		}
	}
	senderID := asString(senderObj["id"])
	deviceID := asString(senderObj["device_id"])
	manifest := asString(senderObj["manifest_href"])
	transport := asString(senderObj["transport"])

	sdpText := ""
	parsed := sdp.ParsedDetails{}
	if manifest != "" {
		resp, err := client.Get(manifest)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				sdpText = strings.TrimSpace(string(b))
				parsed = sdp.ParseSDP(sdpText)
			}
		}
	}

	// Best-effort IS-05 transport params (path A/B)
	var is05Obj map[string]any
	var tpA, tpB map[string]any
	if is05Base != "" && senderID != "" {
		// Discover versions
		connVers, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/connection/", is05Base))
		connVer := ""
		if err == nil && len(connVers) > 0 {
			sort.Strings(connVers)
			connVer = strings.Trim(connVers[len(connVers)-1], "/")
		}
		if connVer != "" {
			paths := []string{
				fmt.Sprintf("%s/x-nmos/connection/%s/single/senders/%s/active", is05Base, connVer, senderID),
				fmt.Sprintf("%s/x-nmos/connection/%s/single/senders/%s/staged", is05Base, connVer, senderID),
			}
			for _, p := range paths {
				obj, err := h.fetchJSONMapWithClient(client, p)
				if err == nil && obj != nil {
					is05Obj = obj
					break
				}
			}
		}
	}
	if tps, ok := is05Obj["transport_params"].([]any); ok && len(tps) > 0 {
		if m, ok := tps[0].(map[string]any); ok {
			tpA = m
		}
		if len(tps) > 1 {
			if m, ok := tps[1].(map[string]any); ok {
				tpB = m
			}
		}
	}
	pickStr := func(m map[string]any, key string) string {
		if m == nil {
			return ""
		}
		v, _ := m[key].(string)
		return v
	}
	pickInt := func(m map[string]any, key string) int {
		if m == nil {
			return 0
		}
		switch v := m[key].(type) {
		case float64:
			return int(v)
		case int:
			return v
		default:
			return 0
		}
	}

	snap := flowSnapshot{
		SourceAddrA:         fallback(pickStr(tpA, "source_ip"), parsed.SourceAddrA),
		SourcePortA:         pickInt(tpA, "source_port"),
		MulticastAddrA:      fallback(pickStr(tpA, "destination_ip"), parsed.MulticastAddrA),
		GroupPortA:          pickInt(tpA, "destination_port"),
		SourceAddrB:         pickStr(tpB, "source_ip"),
		SourcePortB:         pickInt(tpB, "source_port"),
		MulticastAddrB:      pickStr(tpB, "destination_ip"),
		GroupPortB:          pickInt(tpB, "destination_port"),
		NMOSNodeID:          nodeID,
		NMOSFlowID:          nmosFlowID,
		NMOSSenderID:        senderID,
		NMOSDeviceID:        deviceID,
		NMOSNodeLabel:       nodeLabel,
		NMOSNodeDescription: nodeDesc,
		NMOSIS04BaseURL:     is04Base,
		NMOSIS05BaseURL:     is05Base,
		NMOSIS04Version:     is04Ver,
		SDPURL:              manifest,
		SDPCache:            sdpText,
		MediaType:           parsed.MediaType,
		RedundancyGroup:     parsed.RedundancyGroup,
		TransportProtocol:   fallback(transport, flow.TransportProto),
		ST2110Format:        asString(flowObj["format"]),
		RawNode:             selfObj,
		RawFlow:             flowObj,
		RawSender:           senderObj,
		RawIS05:             is05Obj,
	}
	return snap, nil
}

func (h *Handler) fetchJSONMapWithClient(client *http.Client, url string) (map[string]any, error) {
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
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, err
	}
	return obj, nil
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

	// Broadcast a high-level "sync completed" registry event over WebSocket (best-effort).
	h.publishRegistryEvent(RegistryEvent{
		Kind:     "sync",
		Resource: "nmos_registry",
		Info: map[string]any{
			"base_url":  baseURL,
			"version":   version,
			"nodes":     len(nodesByID),
			"devices":   len(devicesData),
			"flows":     len(flowsData),
			"senders":   len(sendersData),
			"receivers": len(receiversData),
		},
	})
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
