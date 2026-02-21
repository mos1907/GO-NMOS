package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
	"go-nmos/backend/internal/repository"
	"go-nmos/backend/internal/sdp"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createFlowRequest struct {
	FlowID            string `json:"flow_id"`
	DisplayName       string `json:"display_name"`
	MulticastIP       string `json:"multicast_ip"`
	SourceIP          string `json:"source_ip"`
	Port              int    `json:"port"`
	FlowStatus        string `json:"flow_status"`
	Availability      string `json:"availability"`
	Locked            bool   `json:"locked"`
	Note              string `json:"note"`
	TransportProtocol string `json:"transport_protocol"`
	// ST2022-7 A/B paths
	SourceAddrA    string `json:"source_addr_a,omitempty"`
	SourcePortA    int    `json:"source_port_a,omitempty"`
	MulticastAddrA string `json:"multicast_addr_a,omitempty"`
	GroupPortA     int    `json:"group_port_a,omitempty"`
	SourceAddrB    string `json:"source_addr_b,omitempty"`
	SourcePortB    int    `json:"source_port_b,omitempty"`
	MulticastAddrB string `json:"multicast_addr_b,omitempty"`
	GroupPortB     int    `json:"group_port_b,omitempty"`
	// NMOS metadata
	NMOSNodeID          string `json:"nmos_node_id,omitempty"`
	NMOSFlowID          string `json:"nmos_flow_id,omitempty"`
	NMOSSenderID        string `json:"nmos_sender_id,omitempty"`
	NMOSDeviceID        string `json:"nmos_device_id,omitempty"`
	NMOSNodeLabel       string `json:"nmos_node_label,omitempty"`
	NMOSNodeDescription string `json:"nmos_node_description,omitempty"`
	NMOSIS04Host        string `json:"nmos_is04_host,omitempty"`
	NMOSIS04Port        int    `json:"nmos_is04_port,omitempty"`
	NMOSIS05Host        string `json:"nmos_is05_host,omitempty"`
	NMOSIS05Port        int    `json:"nmos_is05_port,omitempty"`
	NMOSIS04BaseURL     string `json:"nmos_is04_base_url,omitempty"`
	NMOSIS05BaseURL     string `json:"nmos_is05_base_url,omitempty"`
	NMOSIS04Version     string `json:"nmos_is04_version,omitempty"`
	NMOSIS05Version     string `json:"nmos_is05_version,omitempty"`
	NMOSLabel           string `json:"nmos_label,omitempty"`
	NMOSDescription     string `json:"nmos_description,omitempty"`
	ManagementURL       string `json:"management_url,omitempty"`
	// Media + source tracking
	MediaType       string `json:"media_type,omitempty"`
	ST2110Format    string `json:"st2110_format,omitempty"`
	RedundancyGroup string `json:"redundancy_group,omitempty"`
	DataSource      string `json:"data_source,omitempty"`
	RDSAddress      string `json:"rds_address,omitempty"`
	RDSAPIURL       string `json:"rds_api_url,omitempty"`
	RDSVersion      string `json:"rds_version,omitempty"`
	SDPURL          string `json:"sdp_url,omitempty"`
	SDPCache        string `json:"sdp_cache,omitempty"`
	Alias1          string `json:"alias_1,omitempty"`
	Alias2          string `json:"alias_2,omitempty"`
	Alias3          string `json:"alias_3,omitempty"`
	Alias4          string `json:"alias_4,omitempty"`
	Alias5          string `json:"alias_5,omitempty"`
	Alias6          string `json:"alias_6,omitempty"`
	Alias7          string `json:"alias_7,omitempty"`
	Alias8          string `json:"alias_8,omitempty"`
	UserField1      string `json:"user_field_1,omitempty"`
	UserField2      string `json:"user_field_2,omitempty"`
	UserField3      string `json:"user_field_3,omitempty"`
	UserField4      string `json:"user_field_4,omitempty"`
	UserField5      string `json:"user_field_5,omitempty"`
	UserField6      string `json:"user_field_6,omitempty"`
	UserField7      string `json:"user_field_7,omitempty"`
	UserField8      string `json:"user_field_8,omitempty"`
}

func (h *Handler) ListFlows(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	sortBy := strings.TrimSpace(r.URL.Query().Get("sort_by"))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sort_order"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	flowStatus := strings.TrimSpace(r.URL.Query().Get("flow_status"))
	availability := strings.TrimSpace(r.URL.Query().Get("availability"))
	dataSource := strings.TrimSpace(r.URL.Query().Get("data_source"))
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	filters := repository.FlowListFilters{
		Q:            q,
		FlowStatus:   flowStatus,
		Availability: availability,
		DataSource:   dataSource,
	}
	useFilters := filters.Q != "" || filters.FlowStatus != "" || filters.Availability != "" || filters.DataSource != ""

	var total int
	var flows []models.Flow
	var err error
	if useFilters {
		total, err = h.repo.CountFlowsFiltered(r.Context(), filters)
		if err == nil {
			flows, err = h.repo.ListFlowsFiltered(r.Context(), filters, limit, offset, sortBy, sortOrder)
		}
	} else {
		total, err = h.repo.CountFlows(r.Context())
		if err == nil {
			flows, err = h.repo.ListFlows(r.Context(), limit, offset, sortBy, sortOrder)
		}
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list flows failed"})
		return
	}
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
	writeJSON(w, http.StatusOK, flows)
}

func (h *Handler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	var req createFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "display_name is required"})
		return
	}
	// If flow_id is provided and a flow with that flow_id already exists, return it (avoid duplicates)
	if req.FlowID != "" {
		existing, err := h.repo.GetFlowByFlowID(r.Context(), req.FlowID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup failed"})
			return
		}
		if existing != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"id": existing.ID, "flow_id": existing.FlowID, "display_name": existing.DisplayName,
			})
			return
		}
	}
	// If a flow with the same display_name already exists, return it and do not create a duplicate
	existingByName, err := h.repo.GetFlowByDisplayName(r.Context(), req.DisplayName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lookup failed"})
		return
	}
	if existingByName != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"id":             existingByName.ID,
			"flow_id":        existingByName.FlowID,
			"display_name":   existingByName.DisplayName,
			"already_exists": true,
			"message":        "A flow with this name already exists. Using existing flow.",
		})
		return
	}
	if req.FlowID == "" {
		req.FlowID = uuid.NewString()
	}
	if req.TransportProtocol == "" {
		req.TransportProtocol = "RTP/UDP"
	}
	if req.FlowStatus == "" {
		req.FlowStatus = "active"
	}
	if req.Availability == "" {
		req.Availability = "available"
	}

	flow := models.Flow{
		FlowID:              req.FlowID,
		DisplayName:         req.DisplayName,
		MulticastIP:         req.MulticastIP,
		SourceIP:            req.SourceIP,
		Port:                req.Port,
		FlowStatus:          req.FlowStatus,
		Availability:        req.Availability,
		Locked:              req.Locked,
		Note:                req.Note,
		TransportProto:      req.TransportProtocol,
		SourceAddrA:         req.SourceAddrA,
		SourcePortA:         req.SourcePortA,
		MulticastAddrA:      req.MulticastAddrA,
		GroupPortA:          req.GroupPortA,
		SourceAddrB:         req.SourceAddrB,
		SourcePortB:         req.SourcePortB,
		MulticastAddrB:      req.MulticastAddrB,
		GroupPortB:          req.GroupPortB,
		NMOSNodeID:          req.NMOSNodeID,
		NMOSFlowID:          req.NMOSFlowID,
		NMOSSenderID:        req.NMOSSenderID,
		NMOSDeviceID:        req.NMOSDeviceID,
		NMOSNodeLabel:       req.NMOSNodeLabel,
		NMOSNodeDescription: req.NMOSNodeDescription,
		NMOSIS04Host:        req.NMOSIS04Host,
		NMOSIS04Port:        req.NMOSIS04Port,
		NMOSIS05Host:        req.NMOSIS05Host,
		NMOSIS05Port:        req.NMOSIS05Port,
		NMOSIS04BaseURL:     req.NMOSIS04BaseURL,
		NMOSIS05BaseURL:     req.NMOSIS05BaseURL,
		NMOSIS04Version:     req.NMOSIS04Version,
		NMOSIS05Version:     req.NMOSIS05Version,
		NMOSLabel:           req.NMOSLabel,
		NMOSDescription:     req.NMOSDescription,
		ManagementURL:       req.ManagementURL,
		MediaType:           req.MediaType,
		ST2110Format:        req.ST2110Format,
		RedundancyGroup:     req.RedundancyGroup,
		DataSource:          req.DataSource,
		RDSAddress:          req.RDSAddress,
		RDSAPIURL:           req.RDSAPIURL,
		RDSVersion:          req.RDSVersion,
		SDPURL:              req.SDPURL,
		SDPCache:            req.SDPCache,
		Alias1:              req.Alias1,
		Alias2:              req.Alias2,
		Alias3:              req.Alias3,
		Alias4:              req.Alias4,
		Alias5:              req.Alias5,
		Alias6:              req.Alias6,
		Alias7:              req.Alias7,
		Alias8:              req.Alias8,
		UserField1:          req.UserField1,
		UserField2:          req.UserField2,
		UserField3:          req.UserField3,
		UserField4:          req.UserField4,
		UserField5:          req.UserField5,
		UserField6:          req.UserField6,
		UserField7:          req.UserField7,
		UserField8:          req.UserField8,
	}

	id, err := h.repo.CreateFlow(r.Context(), flow)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create flow failed"})
		return
	}

	// Auto-fetch SDP when sdp_url/manifest URL is provided (e.g. flow created from registry) so multicast_ip, source_ip, port are filled
	if req.SDPURL != "" {
		updates, fetchErr := fetchSDPFromURL(req.SDPURL)
		if fetchErr == nil && len(updates) > 0 {
			_ = h.repo.PatchFlow(r.Context(), id, updates)
		}
	}

	// Publish MQTT event
	if h.mqtt != nil {
		createdFlow, _ := h.repo.GetFlowByID(r.Context(), id)
		if createdFlow != nil {
			flowMap := flowToMap(*createdFlow)
			h.mqtt.PublishFlowEvent("created", createdFlow.FlowID, flowMap, nil)
		} else {
			flowMap := flowToMap(flow)
			h.mqtt.PublishFlowEvent("created", flow.FlowID, flowMap, nil)
		}
	}

	createdFlow, _ := h.repo.GetFlowByID(r.Context(), id)
	if createdFlow != nil {
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "flow_id": flow.FlowID, "flow": createdFlow})
	} else {
		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "flow_id": flow.FlowID})
	}
}

func (h *Handler) FlowSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetFlowSummary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "summary failed"})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handler) SearchFlows(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "q is required"})
		return
	}
	limit := 50
	offset := 0
	sortBy := strings.TrimSpace(r.URL.Query().Get("sort_by"))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sort_order"))
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	total, err := h.repo.CountSearchFlows(r.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search count failed"})
		return
	}
	flows, err := h.repo.SearchFlows(r.Context(), query, limit, offset, sortBy, sortOrder)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
		return
	}
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	w.Header().Set("X-Limit", strconv.Itoa(limit))
	w.Header().Set("X-Offset", strconv.Itoa(offset))
	writeJSON(w, http.StatusOK, flows)
}

func (h *Handler) ExportFlows(w http.ResponseWriter, r *http.Request) {
	flows, err := h.repo.ExportFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "export failed"})
		return
	}
	writeJSON(w, http.StatusOK, flows)
}

func (h *Handler) ImportFlows(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	var payload []models.Flow
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json array"})
		return
	}
	count, err := h.repo.ImportFlows(r.Context(), payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "import failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"imported": count})
}

func (h *Handler) PatchFlow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	// Read payload first
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	// Check if flow is locked
	currentFlow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}

	// Check if this is only a lock update
	onlyLockUpdate := len(payload) == 1 && payload["locked"] != nil

	// If flow is locked and we're trying to update non-lock fields, reject
	if currentFlow.Locked && !onlyLockUpdate {
		writeJSON(w, http.StatusLocked, map[string]string{"error": "flow is locked"})
		return
	}

	oldFlow := currentFlow
	if err := h.repo.PatchFlow(r.Context(), id, payload); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}

	// Publish MQTT event
	if h.mqtt != nil {
		updatedFlow, _ := h.repo.GetFlowByID(r.Context(), id)
		if updatedFlow != nil {
			diff := computeFlowDiff(oldFlow, updatedFlow)
			flowMap := flowToMap(*updatedFlow)
			h.mqtt.PublishFlowEvent("updated", updatedFlow.FlowID, flowMap, diff)
		}
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeleteFlow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	// Get flow before deletion for MQTT event
	flow, _ := h.repo.GetFlowByID(r.Context(), id)

	if err := h.repo.DeleteFlow(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete failed"})
		return
	}

	// Publish MQTT event
	if h.mqtt != nil && flow != nil {
		flowMap := flowToMap(*flow)
		h.mqtt.PublishFlowEvent("deleted", flow.FlowID, flowMap, nil)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// HardDeleteFlow permanently deletes a flow (admin only).
// This bypasses any soft-delete mechanisms if they exist.
func (h *Handler) HardDeleteFlow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	// Check if hard delete is enabled
	hardDeleteEnabled, _ := h.repo.GetSetting(r.Context(), "hard_delete_enabled")
	if hardDeleteEnabled != "true" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "hard delete is disabled"})
		return
	}

	// Get flow before deletion for MQTT event
	flow, _ := h.repo.GetFlowByID(r.Context(), id)

	// Hard delete: direct database deletion
	if err := h.repo.DeleteFlow(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hard delete failed"})
		return
	}

	// Publish MQTT event
	if h.mqtt != nil && flow != nil {
		flowMap := flowToMap(*flow)
		h.mqtt.PublishFlowEvent("hard_deleted", flow.FlowID, flowMap, nil)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true, "hard_deleted": true})
}

func (h *Handler) SetFlowLock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var payload struct {
		Locked bool `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	currentFlow, err := h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}

	if currentFlow.Locked == payload.Locked {
		writeJSON(w, http.StatusOK, map[string]any{"flow_id": currentFlow.FlowID, "locked": payload.Locked})
		return
	}

	if err := h.repo.PatchFlow(r.Context(), id, map[string]any{"locked": payload.Locked}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}

	// Publish MQTT event
	if h.mqtt != nil {
		updatedFlow, _ := h.repo.GetFlowByID(r.Context(), id)
		if updatedFlow != nil {
			diff := map[string]interface{}{"locked": payload.Locked}
			flowMap := flowToMap(*updatedFlow)
			h.mqtt.PublishFlowEvent("updated", updatedFlow.FlowID, flowMap, diff)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"flow_id": currentFlow.FlowID, "locked": payload.Locked})
}

// fetchSDPFromURL fetches SDP from manifestURL (rewriting localhost for Docker), parses it and returns
// a map of flow field updates. Original manifestURL is stored as sdp_url; fetch uses localhost→host.docker.internal.
func fetchSDPFromURL(manifestURL string) (map[string]any, error) {
	manifestURL = strings.TrimSpace(manifestURL)
	if manifestURL == "" {
		return nil, nil
	}
	if !strings.HasPrefix(manifestURL, "http://") && !strings.HasPrefix(manifestURL, "https://") {
		return nil, nil
	}
	fetchURL := manifestURL
	if strings.Contains(fetchURL, "localhost") {
		fetchURL = strings.Replace(fetchURL, "localhost", "host.docker.internal", 1)
	}
	if strings.Contains(fetchURL, "127.0.0.1") {
		fetchURL = strings.Replace(fetchURL, "127.0.0.1", "host.docker.internal", 1)
	}
	// virtualtest_go camera-node is in a different compose; backend cannot resolve camera-node hostname
	if strings.Contains(fetchURL, "camera-node") {
		fetchURL = strings.Replace(fetchURL, "camera-node", "host.docker.internal", 1)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fetchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SDP fetch returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sdpText := strings.TrimSpace(string(body))
	parsed := sdp.ParseSDP(sdpText)
	updates := map[string]any{
		"sdp_url":   manifestURL,
		"sdp_cache": sdpText,
	}
	if parsed.MediaType != "" {
		updates["media_type"] = parsed.MediaType
	}
	if parsed.RedundancyGroup != "" {
		updates["redundancy_group"] = parsed.RedundancyGroup
	}
	if parsed.MulticastAddrA != "" {
		updates["multicast_addr_a"] = parsed.MulticastAddrA
	}
	if parsed.SourceAddrA != "" {
		updates["source_addr_a"] = parsed.SourceAddrA
	}
	if parsed.GroupPortA > 0 {
		updates["group_port_a"] = parsed.GroupPortA
	}
	if parsed.MulticastAddrA != "" {
		updates["multicast_ip"] = parsed.MulticastAddrA
	}
	if parsed.SourceAddrA != "" {
		updates["source_ip"] = parsed.SourceAddrA
	}
	if parsed.GroupPortA > 0 {
		updates["port"] = parsed.GroupPortA
	}
	return updates, nil
}

// FetchSDP fetches SDP content from manifest_url and updates the flow.
func (h *Handler) FetchSDP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var payload struct {
		ManifestURL string `json:"manifest_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	manifestURL := strings.TrimSpace(payload.ManifestURL)
	if manifestURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "manifest_url is required"})
		return
	}
	if !strings.HasPrefix(manifestURL, "http://") && !strings.HasPrefix(manifestURL, "https://") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "manifest_url must be http or https"})
		return
	}

	_, err = h.repo.GetFlowByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}

	updates, err := fetchSDPFromURL(manifestURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch SDP: " + err.Error()})
		return
	}
	if len(updates) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "no SDP data from URL"})
		return
	}
	if err := h.repo.PatchFlow(r.Context(), id, updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update failed"})
		return
	}
	updatedFlow, _ := h.repo.GetFlowByID(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"sdp_url": manifestURL,
		"parsed":  sdp.ParseSDP(updatedFlow.SDPCache),
		"flow":    updatedFlow,
	})
}

func flowToMap(f models.Flow) map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = f.ID
	m["flow_id"] = f.FlowID
	m["display_name"] = f.DisplayName
	m["multicast_ip"] = f.MulticastIP
	m["source_ip"] = f.SourceIP
	m["port"] = f.Port
	m["flow_status"] = f.FlowStatus
	m["availability"] = f.Availability
	m["locked"] = f.Locked
	m["note"] = f.Note
	m["transport_protocol"] = f.TransportProto
	m["source_addr_a"] = f.SourceAddrA
	m["source_port_a"] = f.SourcePortA
	m["multicast_addr_a"] = f.MulticastAddrA
	m["group_port_a"] = f.GroupPortA
	m["source_addr_b"] = f.SourceAddrB
	m["source_port_b"] = f.SourcePortB
	m["multicast_addr_b"] = f.MulticastAddrB
	m["group_port_b"] = f.GroupPortB
	m["nmos_node_id"] = f.NMOSNodeID
	m["nmos_flow_id"] = f.NMOSFlowID
	m["nmos_sender_id"] = f.NMOSSenderID
	m["nmos_device_id"] = f.NMOSDeviceID
	m["nmos_node_label"] = f.NMOSNodeLabel
	m["nmos_node_description"] = f.NMOSNodeDescription
	m["nmos_is04_host"] = f.NMOSIS04Host
	m["nmos_is04_port"] = f.NMOSIS04Port
	m["nmos_is05_host"] = f.NMOSIS05Host
	m["nmos_is05_port"] = f.NMOSIS05Port
	m["nmos_is04_base_url"] = f.NMOSIS04BaseURL
	m["nmos_is05_base_url"] = f.NMOSIS05BaseURL
	m["nmos_is04_version"] = f.NMOSIS04Version
	m["nmos_is05_version"] = f.NMOSIS05Version
	m["nmos_label"] = f.NMOSLabel
	m["nmos_description"] = f.NMOSDescription
	m["management_url"] = f.ManagementURL
	m["media_type"] = f.MediaType
	m["st2110_format"] = f.ST2110Format
	m["format_summary"] = f.FormatSummary
	m["redundancy_group"] = f.RedundancyGroup
	m["data_source"] = f.DataSource
	m["rds_address"] = f.RDSAddress
	m["rds_api_url"] = f.RDSAPIURL
	m["rds_version"] = f.RDSVersion
	m["sdp_url"] = f.SDPURL
	m["sdp_cache"] = f.SDPCache
	m["updated_at"] = f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	if f.LastSeen != nil {
		m["last_seen"] = f.LastSeen.Format("2006-01-02T15:04:05Z07:00")
	}
	return m
}

func computeFlowDiff(old, new *models.Flow) map[string]interface{} {
	diff := make(map[string]interface{})
	if old.DisplayName != new.DisplayName {
		diff["display_name"] = new.DisplayName
	}
	if old.MulticastIP != new.MulticastIP {
		diff["multicast_ip"] = new.MulticastIP
	}
	if old.SourceIP != new.SourceIP {
		diff["source_ip"] = new.SourceIP
	}
	if old.Port != new.Port {
		diff["port"] = new.Port
	}
	if old.FlowStatus != new.FlowStatus {
		diff["flow_status"] = new.FlowStatus
	}
	if old.Availability != new.Availability {
		diff["availability"] = new.Availability
	}
	if old.Locked != new.Locked {
		diff["locked"] = new.Locked
	}
	if old.Note != new.Note {
		diff["note"] = new.Note
	}
	if old.TransportProto != new.TransportProto {
		diff["transport_protocol"] = new.TransportProto
	}
	return diff
}
