package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GetReceiversActive returns IS-05 active state for the given receivers (BCC-style connection status).
// GET /api/nmos/receivers-active?is05_base=...&receiver_ids=id1,id2,id3
func (h *Handler) GetReceiversActive(w http.ResponseWriter, r *http.Request) {
	is05Base := strings.TrimSpace(r.URL.Query().Get("is05_base"))
	if is05Base == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is05_base is required"})
		return
	}
	is05Base = strings.TrimRight(is05Base, "/")

	receiverIDsParam := strings.TrimSpace(r.URL.Query().Get("receiver_ids"))
	if receiverIDsParam == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	receiverIDs := strings.Split(receiverIDsParam, ",")
	for i := range receiverIDs {
		receiverIDs[i] = strings.TrimSpace(receiverIDs[i])
	}
	if len(receiverIDs) == 0 {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	client := &http.Client{Timeout: 6 * time.Second}

	type flowInfo struct {
		ID          int64  `json:"id,omitempty"`
		DisplayName string `json:"display_name,omitempty"`
		MulticastIP string `json:"multicast_ip,omitempty"`
		SourceIP    string `json:"source_ip,omitempty"`
		Port        int    `json:"port,omitempty"`
	}
	type activeItem struct {
		ReceiverID   string    `json:"receiver_id"`
		SenderID     string    `json:"sender_id,omitempty"`
		MasterEnable *bool     `json:"master_enable,omitempty"`
		Flow         *flowInfo `json:"flow,omitempty"` // connected flow (from receiver_connections) so UI can show "flow is active"
		Error        string    `json:"error,omitempty"`
	}

	results := make([]activeItem, len(receiverIDs))
	var wg sync.WaitGroup
	for i, recID := range receiverIDs {
		if recID == "" {
			continue
		}
		wg.Add(1)
		go func(idx int, receiverID string) {
			defer wg.Done()
			activeURL := fmt.Sprintf("%s/single/receivers/%s/active", is05Base, receiverID)
			obj, err := h.fetchJSONMapWithClient(client, activeURL)
			item := activeItem{ReceiverID: receiverID}
			if err != nil {
				item.Error = err.Error()
				results[idx] = item
				return
			}
			if obj != nil {
				if s, ok := obj["sender_id"].(string); ok {
					item.SenderID = s
				}
				if b, ok := obj["master_enable"].(bool); ok {
					item.MasterEnable = &b
				}
			}
			results[idx] = item
		}(i, recID)
	}
	wg.Wait()

	// Enrich with flow info from our DB (receiver_connections) so UI can show "Flow: X (multicast:port)"
	ctx := r.Context()
	for i := range results {
		if results[i].ReceiverID == "" {
			continue
		}
		conn, err := h.repo.GetReceiverConnection(ctx, results[i].ReceiverID, "staged", "master")
		if err != nil || conn == nil || conn.FlowID == nil {
			conn, _ = h.repo.GetReceiverConnection(ctx, results[i].ReceiverID, "active", "master")
		}
		if conn != nil && conn.FlowID != nil {
			flow, err := h.repo.GetFlowByID(ctx, *conn.FlowID)
			if err == nil && flow != nil {
				results[i].Flow = &flowInfo{
					ID:          flow.ID,
					DisplayName: flow.DisplayName,
					MulticastIP: flow.MulticastIP,
					SourceIP:    flow.SourceIP,
					Port:        flow.Port,
				}
			}
		}
	}

	// Filter out empty receiver_ids
	out := make([]activeItem, 0, len(results))
	for _, item := range results {
		if item.ReceiverID != "" {
			out = append(out, item)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ReceiverDisableRequest is the body for POST /api/nmos/receiver-disable (BCC-style un-TAKE).
type ReceiverDisableRequest struct {
	IS05Base   string `json:"is05_base"`
	ReceiverID string `json:"receiver_id"`
}

// ReceiverDisable sends IS-05 PATCH .../staged with master_enable: false and activate_immediate (un-TAKE).
// POST /api/nmos/receiver-disable
func (h *Handler) ReceiverDisable(w http.ResponseWriter, r *http.Request) {
	var req ReceiverDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	is05Base := strings.TrimRight(strings.TrimSpace(req.IS05Base), "/")
	receiverID := strings.TrimSpace(req.ReceiverID)
	if is05Base == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is05_base is required"})
		return
	}
	if receiverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id is required"})
		return
	}

	stagedURL := fmt.Sprintf("%s/single/receivers/%s/staged", is05Base, receiverID)
	patchBody := map[string]any{
		"master_enable": false,
		"activation":    map[string]any{"mode": "activate_immediate"},
	}
	payload, _ := json.Marshal(patchBody)

	client := &http.Client{Timeout: 8 * time.Second}
	httpReq, err := http.NewRequest(http.MethodPatch, stagedURL, strings.NewReader(string(payload)))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "patch request failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error":       "receiver disable returned non-2xx",
			"status_code": resp.StatusCode,
			"body":        string(body),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"receiver_id":  receiverID,
		"status_code":  resp.StatusCode,
	})
}
