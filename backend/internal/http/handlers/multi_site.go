package handlers

import (
	"net/http"

	"go-nmos/backend/internal/models"
)

// CrossSiteRouting represents a flow routing between different sites (D.3).
type CrossSiteRouting struct {
	FlowID        string `json:"flow_id"`
	FlowLabel     string `json:"flow_label"`
	SenderID      string `json:"sender_id"`
	SenderLabel   string `json:"sender_label"`
	ReceiverID    string `json:"receiver_id"`
	ReceiverLabel string `json:"receiver_label"`
	SourceSite    string `json:"source_site"`
	TargetSite    string `json:"target_site"`
	SourceDevice  string `json:"source_device"`
	TargetDevice  string `json:"target_device"`
}

// GetCrossSiteRoutings returns flows that route between different sites (D.3).
// GET /flows/cross-site
func (h *Handler) GetCrossSiteRoutings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	ctx := r.Context()

	// Get all flows, senders, receivers, devices
	flows, _ := h.repo.ListNMOSFlows(ctx)
	senders, _ := h.repo.ListNMOSSenders(ctx, "")
	receivers, _ := h.repo.ListNMOSReceivers(ctx, "")
	devices, _ := h.repo.ListNMOSDevices(ctx, "")
	connections, _ := h.repo.ListAllReceiverConnections(ctx)

	// Build maps for quick lookup
	flowMap := make(map[string]models.NMOSFlow)
	for _, f := range flows {
		flowMap[f.ID] = f
	}

	senderMap := make(map[string]models.NMOSSender)
	for _, s := range senders {
		senderMap[s.ID] = s
	}

	receiverMap := make(map[string]models.NMOSReceiver)
	for _, r := range receivers {
		receiverMap[r.ID] = r
	}

	deviceMap := make(map[string]models.NMOSDevice)
	for _, d := range devices {
		deviceMap[d.ID] = d
	}

	// Find cross-site routings
	crossSiteRoutings := make([]CrossSiteRouting, 0)

	for _, conn := range connections {
		if conn.State != "active" || conn.Role != "master" {
			continue
		}

		// Get sender and receiver
		sender, senderExists := senderMap[conn.SenderID]
		receiver, receiverExists := receiverMap[conn.ReceiverID]
		if !senderExists || !receiverExists {
			continue
		}

		// Get flow from sender (sender has flow_id)
		flow, flowExists := flowMap[sender.FlowID]
		if !flowExists {
			continue
		}

		// Get devices
		sourceDevice, sourceDeviceExists := deviceMap[sender.DeviceID]
		targetDevice, targetDeviceExists := deviceMap[receiver.DeviceID]
		if !sourceDeviceExists || !targetDeviceExists {
			continue
		}

		// Get site tags
		sourceSite := sourceDevice.GetSiteTag()
		targetSite := targetDevice.GetSiteTag()

		// Only include if sites are different and both are set
		if sourceSite != "" && targetSite != "" && sourceSite != targetSite {
			crossSiteRoutings = append(crossSiteRoutings, CrossSiteRouting{
				FlowID:        flow.ID,
				FlowLabel:     flow.Label,
				SenderID:      sender.ID,
				SenderLabel:   sender.Label,
				ReceiverID:    receiver.ID,
				ReceiverLabel: receiver.Label,
				SourceSite:    sourceSite,
				TargetSite:    targetSite,
				SourceDevice:  sourceDevice.Label,
				TargetDevice:  targetDevice.Label,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"routings": crossSiteRoutings,
		"count":    len(crossSiteRoutings),
	})
}
