package repository

import (
	"context"
	"encoding/json"
	"time"

	"go-nmos/backend/internal/models"
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user models.User) error
	UpdateUser(ctx context.Context, username string, updates map[string]any) error
	DeleteUser(ctx context.Context, username string) error

	ListFlows(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error)
	CountFlows(ctx context.Context) (int, error)
	ListFlowsFiltered(ctx context.Context, filters FlowListFilters, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error)
	CountFlowsFiltered(ctx context.Context, filters FlowListFilters) (int, error)
	SearchFlows(ctx context.Context, query string, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error)
	CountSearchFlows(ctx context.Context, query string) (int, error)
	GetFlowSummary(ctx context.Context) (models.FlowSummary, error)
	ExportFlows(ctx context.Context) ([]models.Flow, error)
	ImportFlows(ctx context.Context, flows []models.Flow) (int, error)
	DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error)
	GetAlternativeSuggestions(ctx context.Context, multicastIP string, port int, excludeFlowID *int64) ([]models.AlternativeSuggestion, error)
	SaveCheckerResult(ctx context.Context, kind string, result []byte) error
	GetLatestCheckerResult(ctx context.Context, kind string) (*models.CheckerResult, error)
	CreateFlow(ctx context.Context, flow models.Flow) (int64, error)
	GetFlowByID(ctx context.Context, id int64) (*models.Flow, error)
	GetFlowByFlowID(ctx context.Context, flowID string) (*models.Flow, error)
	GetFlowByDisplayName(ctx context.Context, displayName string) (*models.Flow, error)
	PatchFlow(ctx context.Context, id int64, updates map[string]any) error
	DeleteFlow(ctx context.Context, id int64) error

	ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error)
	GetAutomationJob(ctx context.Context, jobID string) (*models.AutomationJob, error)
	UpsertAutomationJob(ctx context.Context, job models.AutomationJob) error
	SetAutomationJobEnabled(ctx context.Context, jobID string, enabled bool) error
	UpdateAutomationJobRun(ctx context.Context, jobID, status string, result []byte) error

	// E.1: Operational playbooks
	ListPlaybooks(ctx context.Context) ([]models.Playbook, error)
	GetPlaybook(ctx context.Context, id string) (*models.Playbook, error)
	UpsertPlaybook(ctx context.Context, playbook models.Playbook) error
	DeletePlaybook(ctx context.Context, id string) error
	CreatePlaybookExecution(ctx context.Context, exec models.PlaybookExecution) (int64, error)
	UpdatePlaybookExecution(ctx context.Context, execID int64, status string, result json.RawMessage) error
	ListPlaybookExecutions(ctx context.Context, playbookID string, limit int) ([]models.PlaybookExecution, error)

	// E.2: Scheduling & maintenance windows
	CreateScheduledPlaybookExecution(ctx context.Context, exec models.ScheduledPlaybookExecution) (int64, error)
	ListPendingScheduledPlaybookExecutions(ctx context.Context, before time.Time) ([]models.ScheduledPlaybookExecution, error)
	UpdateScheduledPlaybookExecution(ctx context.Context, id int64, updates map[string]any) error
	GetScheduledPlaybookExecution(ctx context.Context, id int64) (*models.ScheduledPlaybookExecution, error)
	ListScheduledPlaybookExecutions(ctx context.Context, playbookID string, limit int) ([]models.ScheduledPlaybookExecution, error)
	DeleteScheduledPlaybookExecution(ctx context.Context, id int64) error

	ListMaintenanceWindows(ctx context.Context, startTime, endTime *time.Time) ([]models.MaintenanceWindow, error)
	GetMaintenanceWindow(ctx context.Context, id int64) (*models.MaintenanceWindow, error)
	CreateMaintenanceWindow(ctx context.Context, window models.MaintenanceWindow) (int64, error)
	UpdateMaintenanceWindow(ctx context.Context, id int64, updates map[string]any) error
	DeleteMaintenanceWindow(ctx context.Context, id int64) error
	GetActiveMaintenanceWindows(ctx context.Context, at time.Time) ([]models.MaintenanceWindow, error)

	ListRootBuckets(ctx context.Context) ([]models.AddressBucket, error)
	ListChildBuckets(ctx context.Context, parentID int64) ([]models.AddressBucket, error)
	ListAllBuckets(ctx context.Context) ([]models.AddressBucket, error)
	CreateBucket(ctx context.Context, bucket models.AddressBucket) (int64, error)
	UpdateBucket(ctx context.Context, id int64, updates map[string]any) error
	DeleteBucket(ctx context.Context, id int64) error
	ExportBuckets(ctx context.Context) ([]models.AddressBucket, error)
	ImportBuckets(ctx context.Context, buckets []models.AddressBucket) (int, error)
	GetBucketUsageStats(ctx context.Context, bucketID int64) (*models.BucketUsageStats, error)

	// NMOS registry (IS-04 oriented) — designed to scale to full controller
	ListNMOSNodes(ctx context.Context) ([]models.NMOSNode, error)
	ListNMOSDevices(ctx context.Context, nodeID string) ([]models.NMOSDevice, error)
	ListNMOSFlows(ctx context.Context) ([]models.NMOSFlow, error)
	ListNMOSSenders(ctx context.Context, deviceID string) ([]models.NMOSSender, error)
	ListNMOSReceivers(ctx context.Context, deviceID string) ([]models.NMOSReceiver, error)

	UpsertNMOSNode(ctx context.Context, node models.NMOSNode) error
	DeleteNMOSNode(ctx context.Context, nodeID string) error
	UpsertNMOSDevice(ctx context.Context, dev models.NMOSDevice) error
	UpsertNMOSFlow(ctx context.Context, flow models.NMOSFlow) error
	UpsertNMOSSender(ctx context.Context, sender models.NMOSSender) error
	UpsertNMOSReceiver(ctx context.Context, rec models.NMOSReceiver) error

	// B.1: Receiver connection state (IS-05)
	GetReceiverConnection(ctx context.Context, receiverID, state, role string) (*models.ReceiverConnection, error)
	ListReceiverConnections(ctx context.Context, receiverID string) ([]models.ReceiverConnection, error)
	ListAllReceiverConnections(ctx context.Context) ([]models.ReceiverConnection, error)
	UpsertReceiverConnection(ctx context.Context, conn models.ReceiverConnection) error
	DeleteReceiverConnection(ctx context.Context, receiverID, state, role string) error
	ListReceiverConnectionHistory(ctx context.Context, receiverID string, limit int) ([]models.ReceiverConnectionHistory, error)
	RecordReceiverConnectionHistory(ctx context.Context, hist models.ReceiverConnectionHistory) error

	// B.2: Scheduled activations (time-based IS-05 patches)
	CreateScheduledActivation(ctx context.Context, act models.ScheduledActivation) (int64, error)
	ListScheduledActivations(ctx context.Context, status string, limit int) ([]models.ScheduledActivation, error)
	GetScheduledActivation(ctx context.Context, id int64) (*models.ScheduledActivation, error)
	UpdateScheduledActivation(ctx context.Context, id int64, updates map[string]any) error
	DeleteScheduledActivation(ctx context.Context, id int64) error
	ListPendingScheduledActivations(ctx context.Context, before time.Time) ([]models.ScheduledActivation, error)

	// B.3: Routing policies
	CreateRoutingPolicy(ctx context.Context, policy models.RoutingPolicy) (int64, error)
	ListRoutingPolicies(ctx context.Context, enabledOnly bool) ([]models.RoutingPolicy, error)
	GetRoutingPolicy(ctx context.Context, id int64) (*models.RoutingPolicy, error)
	UpdateRoutingPolicy(ctx context.Context, id int64, updates map[string]any) error
	DeleteRoutingPolicy(ctx context.Context, id int64) error
	RecordRoutingPolicyAudit(ctx context.Context, audit models.RoutingPolicyAudit) error
	ListRoutingPolicyAudits(ctx context.Context, limit int) ([]models.RoutingPolicyAudit, error)

	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
	HealthCheck(ctx context.Context) error

	// C.3: Events (IS-07 / tally)
	InsertEvent(ctx context.Context, e models.Event) (int64, error)
	ListEvents(ctx context.Context, source, severity string, since *time.Time, limit int) ([]models.Event, error)
}
