package models

import (
	"encoding/json"
	"time"
)

type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Flow struct {
	ID             int64      `json:"id"`
	FlowID         string     `json:"flow_id"`
	DisplayName    string     `json:"display_name"`
	MulticastIP    string     `json:"multicast_ip"`
	SourceIP       string     `json:"source_ip"`
	Port           int        `json:"port"`
	FlowStatus     string     `json:"flow_status"`
	Availability   string     `json:"availability"`
	Locked         bool       `json:"locked"`
	Note           string     `json:"note"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	TransportProto string     `json:"transport_protocol"`
	// ST2022-7 A/B paths (more detailed than the legacy single-path fields above)
	SourceAddrA    string `json:"source_addr_a,omitempty"`
	SourcePortA    int    `json:"source_port_a,omitempty"`
	MulticastAddrA string `json:"multicast_addr_a,omitempty"`
	GroupPortA     int    `json:"group_port_a,omitempty"`
	SourceAddrB    string `json:"source_addr_b,omitempty"`
	SourcePortB    int    `json:"source_port_b,omitempty"`
	MulticastAddrB string `json:"multicast_addr_b,omitempty"`
	GroupPortB     int    `json:"group_port_b,omitempty"`
	// NMOS metadata (stored on the flow record for easier syncing / auditing)
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
	// Media info
	MediaType       string `json:"media_type,omitempty"`
	ST2110Format    string `json:"st2110_format,omitempty"`
	RedundancyGroup string `json:"redundancy_group,omitempty"`
	// Data source tracking (manual/nmos/rds)
	DataSource string `json:"data_source,omitempty"`
	RDSAddress string `json:"rds_address,omitempty"`
	RDSAPIURL  string `json:"rds_api_url,omitempty"`
	RDSVersion string `json:"rds_version,omitempty"`
	// SDP (Session Description Protocol) - ST 2110 / NMOS manifest
	SDPURL   string `json:"sdp_url,omitempty"`
	SDPCache string `json:"sdp_cache,omitempty"`
	// Alias fields for text message sharing hub functionality
	Alias1 string `json:"alias_1,omitempty"`
	Alias2 string `json:"alias_2,omitempty"`
	Alias3 string `json:"alias_3,omitempty"`
	Alias4 string `json:"alias_4,omitempty"`
	Alias5 string `json:"alias_5,omitempty"`
	Alias6 string `json:"alias_6,omitempty"`
	Alias7 string `json:"alias_7,omitempty"`
	Alias8 string `json:"alias_8,omitempty"`
	// User-defined fields for custom metadata
	UserField1 string `json:"user_field_1,omitempty"`
	UserField2 string `json:"user_field_2,omitempty"`
	UserField3 string `json:"user_field_3,omitempty"`
	UserField4 string `json:"user_field_4,omitempty"`
	UserField5 string `json:"user_field_5,omitempty"`
	UserField6 string `json:"user_field_6,omitempty"`
	UserField7 string `json:"user_field_7,omitempty"`
	UserField8 string `json:"user_field_8,omitempty"`
}

type FlowSummary struct {
	Total       int `json:"total"`
	Active      int `json:"active"`
	Locked      int `json:"locked"`
	Unused      int `json:"unused"`
	Maintenance int `json:"maintenance"`
}

type CollisionGroup struct {
	MulticastIP string   `json:"multicast_ip"`
	Port        int      `json:"port"`
	Count       int      `json:"count"`
	FlowNames   []string `json:"flow_names"`
}

type CheckerResult struct {
	Kind      string          `json:"kind"`
	Result    json.RawMessage `json:"result"`
	CreatedAt time.Time       `json:"created_at"`
}

type AutomationJob struct {
	JobID         string          `json:"job_id"`
	JobType       string          `json:"job_type"`
	Enabled       bool            `json:"enabled"`
	ScheduleType  string          `json:"schedule_type"`
	ScheduleValue string          `json:"schedule_value"`
	LastRunAt     *time.Time      `json:"last_run_at,omitempty"`
	LastRunStatus string          `json:"last_run_status,omitempty"`
	LastRunResult json.RawMessage `json:"last_run_result,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type AddressBucket struct {
	ID          int64           `json:"id"`
	ParentID    *int64          `json:"parent_id,omitempty"`
	BucketType  string          `json:"bucket_type"` // drive | parent | child
	Name        string          `json:"name"`
	CIDR        string          `json:"cidr"`
	StartIP     string          `json:"start_ip"`
	EndIP       string          `json:"end_ip"`
	Color       string          `json:"color"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
