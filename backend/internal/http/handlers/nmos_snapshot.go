package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// NMOSSnapshot represents a portable snapshot of the internal registry state.
type NMOSSnapshot struct {
	Timestamp   string                `json:"timestamp"`
	Version     string                `json:"version"` // snapshot format version
	Nodes       []models.NMOSNode     `json:"nodes"`
	Devices     []models.NMOSDevice   `json:"devices"`
	Flows       []models.NMOSFlow     `json:"flows"`
	Senders     []models.NMOSSender   `json:"senders"`
	Receivers   []models.NMOSReceiver `json:"receivers"`
	Summary     SnapshotSummary       `json:"summary"`
}

// SnapshotSummary provides counts and metadata about the snapshot.
type SnapshotSummary struct {
	TotalNodes     int `json:"total_nodes"`
	TotalDevices   int `json:"total_devices"`
	TotalFlows     int `json:"total_flows"`
	TotalSenders   int `json:"total_senders"`
	TotalReceivers int `json:"total_receivers"`
}

// ExportNMOSSnapshot exports the current internal registry state as a portable JSON snapshot.
// This can be used for backup, migration, or interop testing.
func (h *Handler) ExportNMOSSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nodes, err := h.repo.ListNMOSNodes(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list nodes"})
		return
	}

	devices, err := h.repo.ListNMOSDevices(ctx, "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list devices"})
		return
	}

	flows, err := h.repo.ListNMOSFlows(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list flows"})
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

	snapshot := NMOSSnapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0",
		Nodes:     nodes,
		Devices:   devices,
		Flows:     flows,
		Senders:   senders,
		Receivers: receivers,
		Summary: SnapshotSummary{
			TotalNodes:     len(nodes),
			TotalDevices:   len(devices),
			TotalFlows:     len(flows),
			TotalSenders:   len(senders),
			TotalReceivers: len(receivers),
		},
	}

	writeJSON(w, http.StatusOK, snapshot)
}

// ConformanceIssue represents a single conformance check failure.
type ConformanceIssue struct {
	ResourceType  string `json:"resource_type"`  // "node", "device", "flow", "sender", "receiver"
	ResourceID    string `json:"resource_id"`
	ResourceLabel string `json:"resource_label,omitempty"`
	Field         string `json:"field"`   // field name or "tags.site", "tags.room", etc.
	Issue         string `json:"issue"`   // "missing", "invalid_format", "invalid_value", "recommendation"
	Severity      string `json:"severity"` // "error" (blocking) or "warning" (recommendation)
	Message       string `json:"message"`
}

// ConformanceReport contains conformance check results.
type ConformanceReport struct {
	Timestamp   string             `json:"timestamp"`
	TotalIssues int                `json:"total_issues"`
	Issues      []ConformanceIssue `json:"issues"`
	Summary     map[string]int     `json:"summary"` // counts by resource_type
}

// CheckNMOSConformance performs basic conformance checks on the registry state.
// Checks for required fields, valid formats, and expected tags (site/room).
func (h *Handler) CheckNMOSConformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var issues []ConformanceIssue

	// Valid flow format: video, audio, data or urn:x-nmos:format:*
	validFlowFormat := regexp.MustCompile(`^(video|audio|data|urn:x-nmos:format:[a-z0-9-]+)$`)

	nodes, _ := h.repo.ListNMOSNodes(ctx)
	for _, n := range nodes {
		if n.ID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "node", ResourceID: n.ID, ResourceLabel: n.Label,
				Field: "id", Issue: "missing", Severity: "error",
				Message: "Node ID is required",
			})
		}
		if n.Label == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "node", ResourceID: n.ID,
				Field: "label", Issue: "missing", Severity: "warning",
				Message: "Node label is recommended",
			})
		}
		if n.APIVersion == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "node", ResourceID: n.ID, ResourceLabel: n.Label,
				Field: "api_version", Issue: "missing", Severity: "warning",
				Message: "API version should be set (e.g. v1.3)",
			})
		}
		// A.5: Recommend site/room tags for topology grouping
		if site := n.GetSiteTag(); site == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "node", ResourceID: n.ID, ResourceLabel: n.Label,
				Field: "tags.site", Issue: "recommendation", Severity: "warning",
				Message: "Site tag recommended for topology grouping",
			})
		}
		if room := n.GetRoomTag(); room == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "node", ResourceID: n.ID, ResourceLabel: n.Label,
				Field: "tags.room", Issue: "recommendation", Severity: "warning",
				Message: "Room tag recommended for topology grouping",
			})
		}
	}

	devices, _ := h.repo.ListNMOSDevices(ctx, "")
	for _, d := range devices {
		if d.ID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "device", ResourceID: d.ID, ResourceLabel: d.Label,
				Field: "id", Issue: "missing", Severity: "error",
				Message: "Device ID is required",
			})
		}
		if d.NodeID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "device", ResourceID: d.ID, ResourceLabel: d.Label,
				Field: "node_id", Issue: "missing", Severity: "error",
				Message: "Device must reference a node",
			})
		}
		if d.Type == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "device", ResourceID: d.ID, ResourceLabel: d.Label,
				Field: "type", Issue: "missing", Severity: "warning",
				Message: "Device type should be set",
			})
		}
		if site := d.GetSiteTag(); site == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "device", ResourceID: d.ID, ResourceLabel: d.Label,
				Field: "tags.site", Issue: "recommendation", Severity: "warning",
				Message: "Site tag recommended for topology grouping",
			})
		}
		if room := d.GetRoomTag(); room == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "device", ResourceID: d.ID, ResourceLabel: d.Label,
				Field: "tags.room", Issue: "recommendation", Severity: "warning",
				Message: "Room tag recommended for topology grouping",
			})
		}
	}

	flows, _ := h.repo.ListNMOSFlows(ctx)
	for _, f := range flows {
		if f.ID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "flow", ResourceID: f.ID, ResourceLabel: f.Label,
				Field: "id", Issue: "missing", Severity: "error",
				Message: "Flow ID is required",
			})
		}
		if f.Format == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "flow", ResourceID: f.ID, ResourceLabel: f.Label,
				Field: "format", Issue: "missing", Severity: "warning",
				Message: "Flow format should be set (video/audio/data)",
			})
		} else if !validFlowFormat.MatchString(strings.TrimSpace(f.Format)) {
			issues = append(issues, ConformanceIssue{
				ResourceType: "flow", ResourceID: f.ID, ResourceLabel: f.Label,
				Field: "format", Issue: "invalid_format", Severity: "warning",
				Message: "Flow format should be video, audio, data, or urn:x-nmos:format:*",
			})
		}
	}

	senders, _ := h.repo.ListNMOSSenders(ctx, "")
	for _, s := range senders {
		if s.ID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "sender", ResourceID: s.ID, ResourceLabel: s.Label,
				Field: "id", Issue: "missing", Severity: "error",
				Message: "Sender ID is required",
			})
		}
		if s.DeviceID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "sender", ResourceID: s.ID, ResourceLabel: s.Label,
				Field: "device_id", Issue: "missing", Severity: "error",
				Message: "Sender must reference a device",
			})
		}
		if s.FlowID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "sender", ResourceID: s.ID, ResourceLabel: s.Label,
				Field: "flow_id", Issue: "missing", Severity: "error",
				Message: "Sender must reference a flow",
			})
		}
	}

	receivers, _ := h.repo.ListNMOSReceivers(ctx, "")
	for _, rec := range receivers {
		if rec.ID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "receiver", ResourceID: rec.ID, ResourceLabel: rec.Label,
				Field: "id", Issue: "missing", Severity: "error",
				Message: "Receiver ID is required",
			})
		}
		if rec.DeviceID == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "receiver", ResourceID: rec.ID, ResourceLabel: rec.Label,
				Field: "device_id", Issue: "missing", Severity: "error",
				Message: "Receiver must reference a device",
			})
		}
		if rec.Format == "" {
			issues = append(issues, ConformanceIssue{
				ResourceType: "receiver", ResourceID: rec.ID, ResourceLabel: rec.Label,
				Field: "format", Issue: "missing", Severity: "warning",
				Message: "Receiver format should be set",
			})
		}
	}

	summary := make(map[string]int)
	for _, issue := range issues {
		summary[issue.ResourceType]++
	}

	report := ConformanceReport{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		TotalIssues: len(issues),
		Issues:      issues,
		Summary:     summary,
	}

	writeJSON(w, http.StatusOK, report)
}
