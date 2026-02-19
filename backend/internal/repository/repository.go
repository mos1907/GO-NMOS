package repository

import (
	"context"

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
	SearchFlows(ctx context.Context, query string, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error)
	CountSearchFlows(ctx context.Context, query string) (int, error)
	GetFlowSummary(ctx context.Context) (models.FlowSummary, error)
	ExportFlows(ctx context.Context) ([]models.Flow, error)
	ImportFlows(ctx context.Context, flows []models.Flow) (int, error)
	DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error)
	SaveCheckerResult(ctx context.Context, kind string, result []byte) error
	GetLatestCheckerResult(ctx context.Context, kind string) (*models.CheckerResult, error)
	CreateFlow(ctx context.Context, flow models.Flow) (int64, error)
	GetFlowByID(ctx context.Context, id int64) (*models.Flow, error)
	PatchFlow(ctx context.Context, id int64, updates map[string]any) error
	DeleteFlow(ctx context.Context, id int64) error

	ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error)
	GetAutomationJob(ctx context.Context, jobID string) (*models.AutomationJob, error)
	UpsertAutomationJob(ctx context.Context, job models.AutomationJob) error
	SetAutomationJobEnabled(ctx context.Context, jobID string, enabled bool) error
	UpdateAutomationJobRun(ctx context.Context, jobID, status string, result []byte) error

	ListRootBuckets(ctx context.Context) ([]models.AddressBucket, error)
	ListChildBuckets(ctx context.Context, parentID int64) ([]models.AddressBucket, error)
	CreateBucket(ctx context.Context, bucket models.AddressBucket) (int64, error)
	UpdateBucket(ctx context.Context, id int64, updates map[string]any) error
	DeleteBucket(ctx context.Context, id int64) error
	ExportBuckets(ctx context.Context) ([]models.AddressBucket, error)
	ImportBuckets(ctx context.Context, buckets []models.AddressBucket) (int, error)

	// NMOS registry (IS-04 oriented) — designed to scale to full controller
	ListNMOSNodes(ctx context.Context) ([]models.NMOSNode, error)
	ListNMOSDevices(ctx context.Context, nodeID string) ([]models.NMOSDevice, error)
	ListNMOSFlows(ctx context.Context) ([]models.NMOSFlow, error)
	ListNMOSSenders(ctx context.Context, deviceID string) ([]models.NMOSSender, error)
	ListNMOSReceivers(ctx context.Context, deviceID string) ([]models.NMOSReceiver, error)

	UpsertNMOSNode(ctx context.Context, node models.NMOSNode) error
	UpsertNMOSDevice(ctx context.Context, dev models.NMOSDevice) error
	UpsertNMOSFlow(ctx context.Context, flow models.NMOSFlow) error
	UpsertNMOSSender(ctx context.Context, sender models.NMOSSender) error
	UpsertNMOSReceiver(ctx context.Context, rec models.NMOSReceiver) error

	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
	HealthCheck(ctx context.Context) error
}
