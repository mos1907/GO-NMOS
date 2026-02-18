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
