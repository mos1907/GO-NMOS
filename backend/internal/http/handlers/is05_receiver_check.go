package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type is05ReceiverCheckRequest struct {
	IS05BaseURL string `json:"is05_base_url"`
	ReceiverID  string `json:"receiver_id"`
	TimeoutSec  int    `json:"timeout_sec,omitempty"`
}

// CheckIS05ReceiverState reads IS-05 active/staged transport params for a receiver and compares them to the flow record.
func (h *Handler) CheckIS05ReceiverState(w http.ResponseWriter, r *http.Request) {
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

	var req is05ReceiverCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	is05Base := strings.TrimRight(strings.TrimSpace(req.IS05BaseURL), "/")
	receiverID := strings.TrimSpace(req.ReceiverID)
	if is05Base == "" {
		is05Base = strings.TrimRight(strings.TrimSpace(flow.NMOSIS05BaseURL), "/")
	}
	if is05Base == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "is05_base_url is required (or store nmos_is05_base_url on the flow)"})
		return
	}
	if receiverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "receiver_id is required"})
		return
	}

	timeoutSec := req.TimeoutSec
	if timeoutSec <= 0 || timeoutSec > 30 {
		timeoutSec = 6
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	// Discover IS-05 versions
	vers, err := h.fetchJSONList(fmt.Sprintf("%s/x-nmos/connection/", is05Base))
	if err != nil || len(vers) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not discover IS-05 versions"})
		return
	}
	sort.Strings(vers)
	connVer := strings.Trim(vers[len(vers)-1], "/")

	activeURL := fmt.Sprintf("%s/x-nmos/connection/%s/single/receivers/%s/active", is05Base, connVer, receiverID)
	stagedURL := fmt.Sprintf("%s/x-nmos/connection/%s/single/receivers/%s/staged", is05Base, connVer, receiverID)

	activeObj, activeErr := h.fetchJSONMapWithClient(client, activeURL)
	stagedObj, stagedErr := h.fetchJSONMapWithClient(client, stagedURL)
	if activeErr != nil && stagedErr != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to read IS-05 active and staged state"})
		return
	}

	pickTP := func(obj map[string]any, idx int) map[string]any {
		if obj == nil {
			return nil
		}
		tps, ok := obj["transport_params"].([]any)
		if !ok || len(tps) <= idx {
			return nil
		}
		m, _ := tps[idx].(map[string]any)
		return m
	}
	tpA := func(m map[string]any) map[string]any {
		if m == nil {
			return nil
		}
		out := map[string]any{
			"destination_ip":   m["destination_ip"],
			"destination_port": m["destination_port"],
			"source_ip":        m["source_ip"],
			"source_port":      m["source_port"],
		}
		return out
	}

	activeA := tpA(pickTP(activeObj, 0))
	activeB := tpA(pickTP(activeObj, 1))
	stagedA := tpA(pickTP(stagedObj, 0))
	stagedB := tpA(pickTP(stagedObj, 1))

	// Compare (best-effort): use group_port_a/multicast_addr_a when present, otherwise legacy multicast_ip/port.
	wantDestIP := flow.MulticastAddrA
	if wantDestIP == "" {
		wantDestIP = flow.MulticastIP
	}
	wantDestPort := flow.GroupPortA
	if wantDestPort == 0 {
		wantDestPort = flow.Port
	}
	wantSrcIP := flow.SourceAddrA
	if wantSrcIP == "" {
		wantSrcIP = flow.SourceIP
	}

	matches := func(tp map[string]any) bool {
		if tp == nil {
			return false
		}
		destIP, _ := tp["destination_ip"].(string)
		srcIP, _ := tp["source_ip"].(string)
		destPort := 0
		switch v := tp["destination_port"].(type) {
		case float64:
			destPort = int(v)
		case int:
			destPort = v
		}
		if wantDestIP != "" && destIP != "" && wantDestIP != destIP {
			return false
		}
		if wantSrcIP != "" && srcIP != "" && wantSrcIP != srcIP {
			return false
		}
		if wantDestPort > 0 && destPort > 0 && wantDestPort != destPort {
			return false
		}
		return true
	}

	resp := map[string]any{
		"ok": true,
		"flow": map[string]any{
			"id":               flow.ID,
			"flow_id":          flow.FlowID,
			"display_name":     flow.DisplayName,
			"multicast_ip":     flow.MulticastIP,
			"source_ip":        flow.SourceIP,
			"port":             flow.Port,
			"multicast_addr_a": flow.MulticastAddrA,
			"group_port_a":     flow.GroupPortA,
			"source_addr_a":    flow.SourceAddrA,
		},
		"is05": map[string]any{
			"base_url":    is05Base,
			"version":     connVer,
			"receiver_id": receiverID,
		},
		"active": map[string]any{
			"transport_params_a": activeA,
			"transport_params_b": activeB,
			"matches_a":          matches(activeA),
			"matches_b":          matches(activeB),
		},
		"staged": map[string]any{
			"transport_params_a": stagedA,
			"transport_params_b": stagedB,
			"matches_a":          matches(stagedA),
			"matches_b":          matches(stagedB),
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
