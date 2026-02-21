package models

import "encoding/json"

// NMOSNode models an IS-04 Node resource.
// In a full registry this would correspond to /x-nmos/node/<ver>/self and related metadata.
type NMOSNode struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Hostname    string          `json:"hostname"`
	APIVersion  string          `json:"api_version"` // e.g. "v1.3"
	Tags        json.RawMessage `json:"tags,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"` // extensible metadata for future specs (IS-09, monitoring, etc.)
}

// GetSiteTag extracts the "site" tag value from Tags JSONB (e.g. "CampusA").
func (n *NMOSNode) GetSiteTag() string {
	return extractTagValue(n.Tags, "site")
}

// GetRoomTag extracts the "room" tag value from Tags JSONB (e.g. "Studio1").
func (n *NMOSNode) GetRoomTag() string {
	return extractTagValue(n.Tags, "room")
}

// GetNetworkDomain extracts network domain/VLAN info from Meta JSONB.
func (n *NMOSNode) GetNetworkDomain() string {
	return extractMetaValue(n.Meta, "network_domain")
}

// NMOSDevice models an IS-04 Device resource.
type NMOSDevice struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	NodeID      string          `json:"node_id"`
	Type        string          `json:"type"` // e.g. "urn:x-nmos:device:generic"
	Tags        json.RawMessage `json:"tags,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// GetSiteTag extracts the "site" tag value from Tags JSONB.
func (d *NMOSDevice) GetSiteTag() string {
	return extractTagValue(d.Tags, "site")
}

// GetRoomTag extracts the "room" tag value from Tags JSONB.
func (d *NMOSDevice) GetRoomTag() string {
	return extractTagValue(d.Tags, "room")
}

// GetCapabilityHints extracts capability hints from Meta JSONB (e.g. max_resolutions, audio_layouts).
func (d *NMOSDevice) GetCapabilityHints() map[string]any {
	return extractMetaObject(d.Meta, "capabilities")
}

// NMOSFlow models an IS-04 Flow resource.
type NMOSFlow struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Format      string          `json:"format"`    // e.g. "video", "audio", "data"
	SourceID    string          `json:"source_id"` // IS-04 source_id
	Tags        json.RawMessage `json:"tags,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// GetSiteTag extracts the "site" tag value from Tags JSONB.
func (f *NMOSFlow) GetSiteTag() string {
	return extractTagValue(f.Tags, "site")
}

// GetRoomTag extracts the "room" tag value from Tags JSONB.
func (f *NMOSFlow) GetRoomTag() string {
	return extractTagValue(f.Tags, "room")
}

// NMOSSender models an IS-04 Sender resource.
type NMOSSender struct {
	ID           string          `json:"id"`
	Label        string          `json:"label"`
	Description  string          `json:"description"`
	FlowID       string          `json:"flow_id"`
	Transport    string          `json:"transport"`     // e.g. "urn:x-nmos:transport:rtp"
	ManifestHREF string          `json:"manifest_href"` // SDP or equivalent
	DeviceID     string          `json:"device_id"`
	Tags         json.RawMessage `json:"tags,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
}

// NMOSReceiver models an IS-04 Receiver resource.
type NMOSReceiver struct {
	ID          string          `json:"id"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Format      string          `json:"format"`    // video / audio / data / mux
	Transport   string          `json:"transport"` // urn:x-nmos:transport:rtp etc.
	DeviceID    string          `json:"device_id"`
	Tags        json.RawMessage `json:"tags,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// GetSiteTag extracts the "site" tag value from Tags JSONB.
func (r *NMOSReceiver) GetSiteTag() string {
	return extractTagValue(r.Tags, "site")
}

// GetRoomTag extracts the "room" tag value from Tags JSONB.
func (r *NMOSReceiver) GetRoomTag() string {
	return extractTagValue(r.Tags, "room")
}

// Helper functions for extracting tag/meta values from JSONB fields.

func extractTagValue(tags json.RawMessage, key string) string {
	if len(tags) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(tags, &m); err != nil {
		return ""
	}
	if arr, ok := m[key].([]any); ok && len(arr) > 0 {
		if s, ok := arr[0].(string); ok {
			return s
		}
	}
	return ""
}

func extractMetaValue(meta json.RawMessage, key string) string {
	if len(meta) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func extractMetaObject(meta json.RawMessage, key string) map[string]any {
	if len(meta) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if obj, ok := v.(map[string]any); ok {
			return obj
		}
	}
	return nil
}
