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
