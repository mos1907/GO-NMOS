package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// BulkPatchRequest is the request body for POST /api/nmos/bulk-patch.
type BulkPatchRequest struct {
	FlowID         int64    `json:"flow_id"`
	ReceiverIDs    []string `json:"receiver_ids"`
	IS05BaseURL    string   `json:"is05_base_url"`
	SenderID       string   `json:"sender_id,omitempty"`
	Mode           string   `json:"mode,omitempty"`           // "immediate" | "safe_switch"
	OverridePolicy bool     `json:"override_policy,omitempty"` // B.3: If true, proceed despite policy violations
}

// BulkPatchResult is the result for a single receiver in a bulk patch.
type BulkPatchResult struct {
	ReceiverID string `json:"receiver_id"`
	OK         bool   `json:"ok"`
	Error      string `json:"error,omitempty"`
}

// BulkPatchResponse is the response for POST /api/nmos/bulk-patch.
type BulkPatchResponse struct {
	Results   []BulkPatchResult `json:"results"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	Mode      string            `json:"mode"`
}

// formatOrderForSafeSwitch defines the order for safe switching: audio first, then video, then data, mux.
var formatOrderForSafeSwitch = map[string]int{
	"audio": 0,
	"video": 1,
	"data":  2,
	"mux":   3,
}

// BulkPatch applies a flow to multiple receivers in one request.
// Mode "safe_switch" orders receivers by format (audio first, then video) to avoid audio/video sync issues.
func (h *Handler) BulkPatch(w http.ResponseWriter, r *http.Request) {
	var req BulkPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req.IS05BaseURL = strings.TrimSpace(strings.TrimSuffix(req.IS05BaseURL, "/"))
	req.Mode = strings.TrimSpace(strings.ToLower(req.Mode))
	if req.Mode == "" {
		req.Mode = "immediate"
	}
	if req.Mode != "immediate" && req.Mode != "safe_switch" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mode must be 'immediate' or 'safe_switch'"})
		return
	}
	if len(req.ReceiverIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_ids is required and must not be empty"})
		return
	}
	if req.IS05BaseURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is05_base_url is required"})
		return
	}

	flow, err := h.repo.GetFlowByID(r.Context(), req.FlowID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "flow not found"})
		return
	}

	// B.3: Policy check for each receiver (check first receiver as representative; full check would iterate all)
	if req.SenderID != "" && len(req.ReceiverIDs) > 0 && !req.OverridePolicy {
		for _, recID := range req.ReceiverIDs {
			violations := h.checkPoliciesForConnection(r.Context(), req.SenderID, recID, req.FlowID, flow.DisplayName, "")
			if len(violations) > 0 {
				writeJSON(w, http.StatusForbidden, map[string]any{
					"error":       "routing policy violation",
					"receiver_id": recID,
					"violations":  violations,
					"override":    "send override_policy: true to proceed",
				})
				return
			}
		}
	}
	if req.SenderID != "" && len(req.ReceiverIDs) > 0 && req.OverridePolicy {
		for _, recID := range req.ReceiverIDs {
			violations := h.checkPoliciesForConnection(r.Context(), req.SenderID, recID, req.FlowID, flow.DisplayName, "")
			for _, v := range violations {
				audit := models.RoutingPolicyAudit{
					PolicyID:        &v.PolicyID,
					Action:          "override",
					SourceID:        req.SenderID,
					DestinationID:   recID,
					FlowID:          &req.FlowID,
					ViolationReason: v.Reason,
				}
				if claims := r.Context().Value("claims"); claims != nil {
					if authClaims, ok := claims.(*AuthClaims); ok && authClaims.Username != "" {
						audit.OverriddenBy = authClaims.Username
					}
				}
				_ = h.repo.RecordRoutingPolicyAudit(r.Context(), audit)
			}
		}
	}

	// Order receivers for safe_switch: audio first, then video, then data, mux
	receiverIDs := req.ReceiverIDs
	if req.Mode == "safe_switch" {
		allRecvs, _ := h.repo.ListNMOSReceivers(r.Context(), "")
		formatByID := make(map[string]string)
		for _, rec := range allRecvs {
			formatByID[rec.ID] = strings.ToLower(rec.Format)
		}
		sort.Slice(receiverIDs, func(i, j int) bool {
			ordI, okI := formatOrderForSafeSwitch[formatByID[receiverIDs[i]]]
			ordJ, okJ := formatOrderForSafeSwitch[formatByID[receiverIDs[j]]]
			if !okI {
				ordI = 99
			}
			if !okJ {
				ordJ = 99
			}
			return ordI < ordJ
		})
	}

	results := make([]BulkPatchResult, 0, len(receiverIDs))
	client := &http.Client{Timeout: 8 * time.Second}

	for _, receiverID := range receiverIDs {
		connectionURL := req.IS05BaseURL + "/single/receivers/" + receiverID + "/staged"
		parsedURL, err := url.Parse(connectionURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			results = append(results, BulkPatchResult{ReceiverID: receiverID, OK: false, Error: "invalid connection URL"})
			continue
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

		payload, _ := json.Marshal(patchBody)
		httpReq, _ := http.NewRequest(http.MethodPatch, connectionURL, strings.NewReader(string(payload)))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		if err != nil {
			results = append(results, BulkPatchResult{ReceiverID: receiverID, OK: false, Error: "patch request failed: " + err.Error()})
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			results = append(results, BulkPatchResult{ReceiverID: receiverID, OK: false, Error: "non-2xx status: " + strconv.Itoa(resp.StatusCode)})
			continue
		}

		// B.1: Record connection state
		conn := models.ReceiverConnection{
			ReceiverID: receiverID,
			State:      "staged",
			Role:       "master",
			SenderID:   req.SenderID,
			FlowID:     &req.FlowID,
			ChangedAt:  time.Now(),
		}
		if claims := r.Context().Value("claims"); claims != nil {
			if authClaims, ok := claims.(*AuthClaims); ok && authClaims.Username != "" {
				conn.ChangedBy = authClaims.Username
			}
		}
		_ = h.repo.UpsertReceiverConnection(r.Context(), conn)
		hist := models.ReceiverConnectionHistory{
			ReceiverID: receiverID,
			State:      "staged",
			Role:       "master",
			SenderID:   req.SenderID,
			FlowID:     &req.FlowID,
			ChangedAt:  time.Now(),
			Action:     "connect",
			ChangedBy:  conn.ChangedBy,
		}
		_ = h.repo.RecordReceiverConnectionHistory(r.Context(), hist)

		results = append(results, BulkPatchResult{ReceiverID: receiverID, OK: true})
	}

	success, failed := 0, 0
	for _, res := range results {
		if res.OK {
			success++
		} else {
			failed++
		}
	}

	writeJSON(w, http.StatusOK, BulkPatchResponse{
		Results: results,
		Success: success,
		Failed:  failed,
		Mode:    req.Mode,
	})
}
