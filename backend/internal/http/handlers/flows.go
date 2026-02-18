package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go-nmos/backend/internal/models"

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
	Alias1            string `json:"alias_1,omitempty"`
	Alias2            string `json:"alias_2,omitempty"`
	Alias3            string `json:"alias_3,omitempty"`
	Alias4            string `json:"alias_4,omitempty"`
	Alias5            string `json:"alias_5,omitempty"`
	Alias6            string `json:"alias_6,omitempty"`
	Alias7            string `json:"alias_7,omitempty"`
	Alias8            string `json:"alias_8,omitempty"`
	UserField1        string `json:"user_field_1,omitempty"`
	UserField2        string `json:"user_field_2,omitempty"`
	UserField3        string `json:"user_field_3,omitempty"`
	UserField4        string `json:"user_field_4,omitempty"`
	UserField5        string `json:"user_field_5,omitempty"`
	UserField6        string `json:"user_field_6,omitempty"`
	UserField7        string `json:"user_field_7,omitempty"`
	UserField8        string `json:"user_field_8,omitempty"`
}

func (h *Handler) ListFlows(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.repo.CountFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "count flows failed"})
		return
	}

	flows, err := h.repo.ListFlows(r.Context(), limit, offset, sortBy, sortOrder)
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
		FlowID:         req.FlowID,
		DisplayName:    req.DisplayName,
		MulticastIP:    req.MulticastIP,
		SourceIP:       req.SourceIP,
		Port:           req.Port,
		FlowStatus:     req.FlowStatus,
		Availability:   req.Availability,
		Locked:         req.Locked,
		Note:           req.Note,
		TransportProto: req.TransportProtocol,
		Alias1:         req.Alias1,
		Alias2:         req.Alias2,
		Alias3:         req.Alias3,
		Alias4:         req.Alias4,
		Alias5:         req.Alias5,
		Alias6:         req.Alias6,
		Alias7:         req.Alias7,
		Alias8:         req.Alias8,
		UserField1:     req.UserField1,
		UserField2:     req.UserField2,
		UserField3:     req.UserField3,
		UserField4:     req.UserField4,
		UserField5:     req.UserField5,
		UserField6:     req.UserField6,
		UserField7:     req.UserField7,
		UserField8:     req.UserField8,
	}

	id, err := h.repo.CreateFlow(r.Context(), flow)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create flow failed"})
		return
	}

	// Publish MQTT event
	if h.mqtt != nil {
		flowMap := flowToMap(flow)
		h.mqtt.PublishFlowEvent("created", flow.FlowID, flowMap, nil)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "flow_id": flow.FlowID})
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
