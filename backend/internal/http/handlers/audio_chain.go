package handlers

import (
	"net/http"
	"strings"

	"go-nmos/backend/internal/models"
)

// AudioChainHop represents one step in an audio (or any) signal chain: flow → receiver on device.
type AudioChainHop struct {
	FlowID       string `json:"flow_id"`
	FlowLabel    string `json:"flow_label"`
	Format       string `json:"format"`
	SenderID     string `json:"sender_id"`
	SenderLabel  string `json:"sender_label"`
	ReceiverID   string `json:"receiver_id"`
	ReceiverLabel string `json:"receiver_label"`
	DeviceID     string `json:"device_id"`
	DeviceLabel  string `json:"device_label"`
}

// AudioChainResponse is the response for GET /audio/chain.
type AudioChainResponse struct {
	FlowID    string           `json:"flow_id"`
	FlowLabel string           `json:"flow_label"`
	Format    string           `json:"format"`
	Hops      []AudioChainHop  `json:"hops"`
	NextFlowIDs []string       `json:"next_flow_ids,omitempty"` // flows that appear downstream (for UI to optionally follow)
}

// GetAudioChain returns the downstream signal chain for a flow (C.2: audio programme across the plant).
// Query: flow_id (required). Traces flow → senders → connected receivers → devices → their output flows, up to max depth.
func (h *Handler) GetAudioChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	flowID := strings.TrimSpace(r.URL.Query().Get("flow_id"))
	if flowID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "flow_id is required"})
		return
	}

	ctx := r.Context()
	flows, err := h.repo.ListNMOSFlows(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list flows"})
		return
	}
	devices, err := h.repo.ListNMOSDevices(ctx, "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list devices"})
		return
	}
	senders, err := h.repo.ListNMOSSenders(ctx, "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list senders"})
		return
	}
	receivers, err := h.repo.ListNMOSReceivers(ctx, "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list receivers"})
		return
	}
	conns, err := h.repo.ListAllReceiverConnections(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
		return
	}

	deviceByID := make(map[string]*models.NMOSDevice)
	for i := range devices {
		deviceByID[devices[i].ID] = &devices[i]
	}
	flowByID := make(map[string]*models.NMOSFlow)
	for i := range flows {
		flowByID[flows[i].ID] = &flows[i]
	}
	sendersByFlow := make(map[string][]models.NMOSSender)
	for i := range senders {
		s := &senders[i]
		sendersByFlow[s.FlowID] = append(sendersByFlow[s.FlowID], *s)
	}
	receiverByID := make(map[string]*models.NMOSReceiver)
	for i := range receivers {
		r := &receivers[i]
		receiverByID[r.ID] = r
	}
	// receiver -> active connection (sender_id)
	receiverToSender := make(map[string]string)
	for _, c := range conns {
		if c.SenderID != "" {
			receiverToSender[c.ReceiverID] = c.SenderID
		}
	}
	// sender_id -> receiver IDs connected to it
	senderToReceivers := make(map[string][]string)
	for recID, senderID := range receiverToSender {
		senderToReceivers[senderID] = append(senderToReceivers[senderID], recID)
	}
	// device_id -> flow_ids of senders on that device (output flows)
	deviceOutputFlowIDs := make(map[string]map[string]struct{})
	for _, s := range senders {
		if s.DeviceID == "" || s.FlowID == "" {
			continue
		}
		if deviceOutputFlowIDs[s.DeviceID] == nil {
			deviceOutputFlowIDs[s.DeviceID] = make(map[string]struct{})
		}
		deviceOutputFlowIDs[s.DeviceID][s.FlowID] = struct{}{}
	}

	flow, ok := flowByID[flowID]
	if !ok {
		writeJSON(w, http.StatusOK, AudioChainResponse{
			FlowID:    flowID,
			FlowLabel: "",
			Format:    "",
			Hops:      nil,
		})
		return
	}

	hops := make([]AudioChainHop, 0)
	seenFlows := map[string]bool{flowID: true}
	currentFlowIDs := []string{flowID}
	const maxDepth = 10
	for depth := 0; depth < maxDepth && len(currentFlowIDs) > 0; depth++ {
		nextFlowIDs := make([]string, 0)
		for _, fid := range currentFlowIDs {
			f := flowByID[fid]
			if f == nil {
				continue
			}
			flowLabel := f.Label
			if flowLabel == "" {
				flowLabel = fid
			}
			for _, s := range sendersByFlow[fid] {
				sLabel := s.Label
				if sLabel == "" {
					sLabel = s.ID
				}
				for _, recID := range senderToReceivers[s.ID] {
					rec := receiverByID[recID]
					if rec == nil {
						continue
					}
					devID := rec.DeviceID
					devLabel := devID
					if dev := deviceByID[devID]; dev != nil && dev.Label != "" {
						devLabel = dev.Label
					}
					recLabel := rec.Label
					if recLabel == "" {
						recLabel = recID
					}
					hops = append(hops, AudioChainHop{
						FlowID:        fid,
						FlowLabel:     flowLabel,
						Format:        f.Format,
						SenderID:      s.ID,
						SenderLabel:   sLabel,
						ReceiverID:    recID,
						ReceiverLabel: recLabel,
						DeviceID:      devID,
						DeviceLabel:   devLabel,
					})
					// Next: flows produced by this device (its senders' flow_ids)
					for fid := range deviceOutputFlowIDs[devID] {
						if !seenFlows[fid] {
							seenFlows[fid] = true
							nextFlowIDs = append(nextFlowIDs, fid)
						}
					}
				}
			}
		}
		currentFlowIDs = nextFlowIDs
	}

	// Collect next_flow_ids (downstream flows) for UI
	nextFlowIDs := make([]string, 0)
	for fid := range seenFlows {
		if fid != flowID {
			nextFlowIDs = append(nextFlowIDs, fid)
		}
	}

	flowLabel := flow.Label
	if flowLabel == "" {
		flowLabel = flowID
	}
	writeJSON(w, http.StatusOK, AudioChainResponse{
		FlowID:       flowID,
		FlowLabel:    flowLabel,
		Format:       flow.Format,
		Hops:         hops,
		NextFlowIDs:  nextFlowIDs,
	})
}

// ListAudioFlows returns flows that are audio format (for C.2 chain starting point).
// GET /audio/flows - optional query format=audio to filter.
func (h *Handler) ListAudioFlows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	formatFilter := strings.TrimSpace(r.URL.Query().Get("format"))
	flows, err := h.repo.ListNMOSFlows(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list flows"})
		return
	}
	out := make([]map[string]string, 0)
	for _, f := range flows {
		if formatFilter != "" && !strings.EqualFold(f.Format, formatFilter) {
			continue
		}
		label := f.Label
		if label == "" {
			label = f.ID
		}
		out = append(out, map[string]string{
			"id":     f.ID,
			"label":  label,
			"format": f.Format,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"flows": out})
}
