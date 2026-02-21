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
	FormatSummary   string `json:"format_summary,omitempty"` // e.g. 1080i50, 1080p25, L24/48k
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
	// SDN / IS-06: optional path id from network controller
	SDNPathID string `json:"sdn_path_id,omitempty"`
	// Planner integration: bucket assignment
	BucketID *int64 `json:"bucket_id,omitempty"`
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

type AlternativeSuggestion struct {
	MulticastIP string `json:"multicast_ip"`
	Port        int    `json:"port"`
	Reason      string `json:"reason"` // e.g., "same_subnet_available", "different_port", "different_subnet"
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
	// NextRunAt is computed when returning jobs (not persisted)
	NextRunAt *time.Time `json:"next_run_at,omitempty"`
}

// Playbook represents a reusable operational workflow (E.1).
type Playbook struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Steps        json.RawMessage `json:"steps"`        // Array of action steps
	Parameters   json.RawMessage `json:"parameters"`   // Parameter definitions
	AllowedRoles []string        `json:"allowed_roles"` // E.3: Which roles can execute this playbook
	Enabled      bool            `json:"enabled"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PlaybookExecution represents a playbook execution run (E.1).
type PlaybookExecution struct {
	ID         int64           `json:"id"`
	PlaybookID string          `json:"playbook_id"`
	Parameters json.RawMessage `json:"parameters"` // Actual parameter values used
	Status     string          `json:"status"`     // "running" | "success" | "error"
	Result     json.RawMessage `json:"result,omitempty"`
	StartedAt  time.Time       `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// ScheduledPlaybookExecution represents a scheduled playbook execution (E.2).
type ScheduledPlaybookExecution struct {
	ID          int64           `json:"id"`
	PlaybookID  string          `json:"playbook_id"`
	Parameters  json.RawMessage `json:"parameters"` // Parameter values for execution
	ScheduledAt time.Time       `json:"scheduled_at"`
	ExecutedAt  *time.Time      `json:"executed_at,omitempty"`
	Status      string          `json:"status"` // "pending" | "executed" | "failed" | "cancelled"
	ExecutionID *int64          `json:"execution_id,omitempty"` // Link to playbook_executions
	CreatedBy   string          `json:"created_by,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MaintenanceWindow represents a maintenance period (E.2).
type MaintenanceWindow struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	RoutingPolicyID *int64   `json:"routing_policy_id,omitempty"` // Policy to apply during maintenance
	Enabled         bool      `json:"enabled"`
	CreatedBy       string    `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

type BucketUsageStats struct {
	BucketID        int64   `json:"bucket_id"`
	TotalIPs        int     `json:"total_ips"`        // Total IPs in the bucket range
	UsedIPs         int     `json:"used_ips"`         // IPs currently used by flows
	AvailableIPs    int     `json:"available_ips"`   // Available IPs
	UsagePercentage float64 `json:"usage_percentage"` // Usage percentage (0-100)
	UsedFlowCount   int     `json:"used_flow_count"` // Number of flows using IPs in this bucket
}

// RegistryConfig represents a configured external IS-04 Query API ("registry").
// These are stored as a JSON array in the settings table under the
// "nmos_registry_config" key so that the controller and UI can share them.
type RegistryConfig struct {
	Name     string `json:"name"`      // Human-friendly label, e.g. "Core Registry"
	QueryURL string `json:"query_url"` // Root or /x-nmos/query URL
	Role     string `json:"role"`      // prod | lab | remote | other free-form tag
	Enabled  bool   `json:"enabled"`   // Whether this registry should be used in UI/workflows
}

// ReceiverConnection represents the current connection state of an NMOS receiver (IS-05).
// B.1: Tracks staged vs active, master vs backup flows, and change metadata.
type ReceiverConnection struct {
	ID         int64           `json:"id"`
	ReceiverID string          `json:"receiver_id"` // NMOS receiver ID (IS-04)
	State      string          `json:"state"`       // "staged" | "active"
	Role       string          `json:"role"`        // "master" | "backup"
	SenderID   string          `json:"sender_id"`   // NMOS sender ID (IS-04)
	FlowID     *int64          `json:"flow_id,omitempty"` // Optional link to internal flow
	ChangedAt  time.Time       `json:"changed_at"`
	ChangedBy  string          `json:"changed_by,omitempty"` // Username
	Metadata   json.RawMessage `json:"metadata,omitempty"`   // Extensible: transport_params, etc.
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ReceiverConnectionHistory represents a historical record of receiver connection changes.
// Used for audit trail and showing previous connections in UI.
type ReceiverConnectionHistory struct {
	ID         int64           `json:"id"`
	ReceiverID string          `json:"receiver_id"`
	State      string          `json:"state"` // "staged" | "active"
	Role       string          `json:"role"`  // "master" | "backup"
	SenderID   string          `json:"sender_id"`
	FlowID     *int64          `json:"flow_id,omitempty"`
	ChangedAt  time.Time       `json:"changed_at"`
	ChangedBy  string          `json:"changed_by,omitempty"`
	Action     string          `json:"action"` // "connect" | "disconnect" | "update"
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ScheduledActivation represents a time-based IS-05 patch/take operation (B.2).
// Scheduled activations are executed by a background scheduler service.
type ScheduledActivation struct {
	ID          int64           `json:"id"`
	FlowID      int64           `json:"flow_id"`      // Internal flow ID
	ReceiverIDs []string        `json:"receiver_ids"` // Array of NMOS receiver IDs
	IS05BaseURL string          `json:"is05_base_url"`
	SenderID    string          `json:"sender_id,omitempty"`
	ScheduledAt time.Time       `json:"scheduled_at"` // When to execute
	ExecutedAt  *time.Time      `json:"executed_at,omitempty"` // When executed (NULL = pending)
	Status      string          `json:"status"`       // "pending" | "executed" | "failed" | "cancelled"
	Mode        string          `json:"mode"`        // "immediate" | "safe_switch"
	CreatedBy   string          `json:"created_by,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"` // Bulk patch result
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// RoutingPolicy represents a routing policy rule (B.3).
// Policies define allowed/forbidden source-destination pairs, path requirements, and constraints.
type RoutingPolicy struct {
	ID                 int64           `json:"id"`
	Name               string          `json:"name"`
	PolicyType         string          `json:"policy_type"` // "allowed_pair" | "forbidden_pair" | "path_requirement" | "constraint"
	Enabled            bool            `json:"enabled"`
	SourcePattern      string          `json:"source_pattern,omitempty"`      // e.g. "sender:*", "flow:test-*"
	DestinationPattern string          `json:"destination_pattern,omitempty"` // e.g. "receiver:*", "device:TX-*"
	RequirePathA       bool            `json:"require_path_a,omitempty"`
	RequirePathB       bool            `json:"require_path_b,omitempty"`
	ConstraintField    string          `json:"constraint_field,omitempty"`    // e.g. "format", "site", "room"
	ConstraintValue    string          `json:"constraint_value,omitempty"`    // e.g. "test", "TX"
	ConstraintOperator string          `json:"constraint_operator,omitempty"` // "equals" | "contains" | "starts_with" | "ends_with"
	Description        string          `json:"description,omitempty"`
	Priority           int             `json:"priority"` // Lower = higher priority
	CreatedBy          string          `json:"created_by,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// RoutingPolicyAudit represents an audit log entry for policy checks (B.3).
type RoutingPolicyAudit struct {
	ID             int64           `json:"id"`
	PolicyID       *int64          `json:"policy_id,omitempty"`
	Action         string          `json:"action"` // "check" | "violation" | "override" | "allowed"
	SourceID       string          `json:"source_id,omitempty"`
	DestinationID  string          `json:"destination_id,omitempty"`
	FlowID         *int64          `json:"flow_id,omitempty"`
	ViolationReason string         `json:"violation_reason,omitempty"`
	OverriddenBy   string          `json:"overridden_by,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Event represents an event/tally record (C.3 IS-07 style). Stored for filtering and correlation with flows/senders/receivers.
type Event struct {
	ID         int64           `json:"id"`
	SourceURL  string          `json:"source_url,omitempty"`
	SourceID   string          `json:"source_id,omitempty"`
	Severity   string          `json:"severity"` // info | warning | error | critical
	Message    string          `json:"message"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	FlowID     string          `json:"flow_id,omitempty"`
	SenderID   string          `json:"sender_id,omitempty"`
	ReceiverID string          `json:"receiver_id,omitempty"`
	JobID      string          `json:"job_id,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
