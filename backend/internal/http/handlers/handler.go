package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-nmos/backend/internal/config"
	"go-nmos/backend/internal/mqtt"
	"go-nmos/backend/internal/repository"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	cfg  config.Config
	repo repository.Repository
	mqtt *mqtt.Client

	registryMu    sync.Mutex
	registryConns map[*wsConn]struct{}
}

type AuthClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewHandler(cfg config.Config, repo repository.Repository, mqttClient *mqtt.Client) *Handler {
	return &Handler{
		cfg:           cfg,
		repo:          repo,
		mqtt:          mqttClient,
		registryConns: make(map[*wsConn]struct{}),
	}
}

// RealtimeConfig exposes MQTT/WebSocket configuration for the frontend.
// This mirrors mmam-docker's get_frontend_config behaviour.
func (h *Handler) RealtimeConfig(w http.ResponseWriter, r *http.Request) {
	wsURL := ""
	if h.cfg.MQTTEnabled {
		scheme := "ws"
		if h.cfg.HTTPSEnabled || r.TLS != nil {
			scheme = "wss"
		}
		host := r.Host
		if idx := strings.Index(host, ":"); idx >= 0 {
			host = host[:idx]
		}
		port := h.cfg.MQTTWSPort
		if port == "" {
			port = "9001"
		}
		wsURL = fmt.Sprintf("%s://%s:%s", scheme, host, port)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"mqtt_enabled": h.cfg.MQTTEnabled,
		"ws_url":       wsURL,
		"topic_prefix": h.cfg.MQTTTopicPrefix,
		"broker_url":   h.cfg.MQTTBrokerURL,
	})
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(60 * time.Second))
	r.Use(h.RateLimitMiddleware)
	r.Use(h.RequestLogMiddleware)
	r.Use(h.MetricsMiddleware)

	allowedOrigin := h.cfg.CORSOrigin
	origins := []string{allowedOrigin}
	if allowedOrigin == "*" {
		origins = []string{"*"}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", h.Health)
	r.Post("/api/login", h.Login)
	// Prometheus metrics endpoint (no auth required for scraping)
	r.Get("/metrics", h.MetricsHandler)
	// WebSocket endpoint for NMOS registry events (no auth; informational only)
	r.Get("/ws/registry", h.RegistryEventsWS)

	// IS-04 Query API (Registry only, per AMWA spec): external NMOS clients (e.g. BCC) discover
	// senders, receivers, flows, devices, nodes at /x-nmos/query/<ver>/... (no auth).
	r.Get("/x-nmos/query", h.QueryCompatVersions)
	r.Get("/x-nmos/query/", h.QueryCompatVersions)
	r.Route("/x-nmos/query/{version}", func(q chi.Router) {
		q.Get("/", h.QueryCompatVersionRoot)
		q.Get("/nodes", h.QueryCompatNodes)
		q.Get("/devices", h.QueryCompatDevices)
		q.Get("/flows", h.QueryCompatFlows)
		q.Get("/senders", h.QueryCompatSenders)
		q.Get("/receivers", h.QueryCompatReceivers)
		q.Get("/health", h.QueryCompatHealth)
	})

	r.Route("/api", func(api chi.Router) {
		api.Use(h.AuthMiddleware)
		api.Get("/me", h.Me)
		api.With(requireRole("viewer", "editor", "admin")).Get("/system", h.SystemInfo)
		api.With(requireRole("viewer", "editor", "admin")).Get("/system/validation", h.ValidateSystemParameters)
		api.With(requireRole("viewer", "editor", "admin")).Get("/health/detail", h.HealthDetail)
		api.With(requireRole("viewer", "editor", "admin")).Post("/health/check-node", h.CheckNMOSNode)
		api.With(requireRole("viewer", "editor", "admin")).Post("/sdn/ping", h.SDNPing)
		api.With(requireRole("viewer", "editor", "admin")).Get("/sdn/topology", h.SDNTopology)
		api.With(requireRole("viewer", "editor", "admin")).Get("/sdn/paths", h.SDNPaths)

		// C.1: IS-08 Audio Channel Mapping (proxy to device)
		api.With(requireRole("viewer", "editor", "admin")).Get("/is08/io", h.IS08GetIO)
		api.With(requireRole("viewer", "editor", "admin")).Get("/is08/map/active", h.IS08GetMapActive)
		api.With(requireRole("editor", "admin")).Post("/is08/map/activations", h.IS08PostMapActivations)

		// C.2: Audio signal chain visibility
		api.With(requireRole("viewer", "editor", "admin")).Get("/audio/flows", h.ListAudioFlows)
		api.With(requireRole("viewer", "editor", "admin")).Get("/audio/chain", h.GetAudioChain)

		// C.3: Events (IS-07 / tally)
		api.With(requireRole("viewer", "editor", "admin")).Get("/events", h.ListEvents)
		api.With(requireRole("editor", "admin")).Post("/events", h.CreateEvent)
		api.With(requireRole("viewer", "editor", "admin")).Get("/events/is07/sources", h.GetIS07Sources)
		api.With(requireRole("viewer", "editor", "admin")).Get("/realtime/config", h.RealtimeConfig)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows", h.ListFlows)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/summary", h.FlowSummary)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/search", h.SearchFlows)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/export", h.ExportFlows)
		api.With(requireRole("editor", "admin")).Post("/flows", h.CreateFlow)
		api.With(requireRole("editor", "admin")).Post("/flows/import", h.ImportFlows)
		api.With(requireRole("editor", "admin")).Patch("/flows/{id}", h.PatchFlow)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/lock", h.SetFlowLock)
		api.With(requireRole("admin")).Delete("/flows/{id}", h.DeleteFlow)
		api.With(requireRole("admin")).Delete("/flows/{id}/hard", h.HardDeleteFlow)

		api.With(requireRole("viewer", "editor", "admin")).Get("/settings", h.GetSettings)
		api.With(requireRole("admin")).Patch("/settings/{key}", h.PatchSetting)

		api.With(requireRole("editor", "admin")).Get("/users", h.ListUsers)
		api.With(requireRole("admin")).Post("/users", h.CreateUser)
		api.With(requireRole("admin")).Patch("/users/{username}", h.UpdateUser)
		api.With(requireRole("admin")).Delete("/users/{username}", h.DeleteUser)

		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/discover", h.DiscoverNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Post("/nmos/discover", h.DiscoverNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Post("/nmos/detect-is05", h.DetectIS05Endpoint)
		api.With(requireRole("viewer", "editor", "admin")).Post("/nmos/detect-is04-from-rds", h.DetectIS04FromRDS)
		api.With(requireRole("admin")).Post("/nmos/explore-ports", h.ExplorePorts)
		api.With(requireRole("editor", "admin")).Post("/nmos/register-node", h.RegisterNodeFromPortScan)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/{id}/nmos/check", h.CheckFlowNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/{id}/nmos/snapshot", h.GetFlowNMOSSnapShot)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/nmos/sync", h.SyncFlowFromNMOS)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/nmos/apply", h.ApplyFlowNMOS)
		api.With(requireRole("editor", "admin")).Post("/nmos/bulk-patch", h.BulkPatch)
		api.With(requireRole("viewer", "editor", "admin")).Post("/flows/{id}/is05/receiver-check", h.CheckIS05ReceiverState)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/fetch-sdp", h.FetchSDP)

		// Internal NMOS registry (IS-04 style) read-only APIs
		api.With(requireRole("viewer", "editor", "admin")).Post("/nmos/registry/discover-nodes", h.DiscoverNMOSRegistryNodes)
		api.With(requireRole("editor", "admin")).Post("/nmos/registry/sync", h.SyncRegistry)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/health", h.NMOSRegistryHealth)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/nodes", h.ListNMOSNodesHandler)
		api.With(requireRole("editor", "admin")).Delete("/nmos/registry/nodes/{nodeId}", h.DeleteNMOSNodeHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/devices", h.ListNMOSDevicesHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/flows", h.ListNMOSFlowsHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/senders", h.ListNMOSSendersHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/receivers", h.ListNMOSReceiversHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/sites-summary", h.GetNMOSSitesRoomsSummary)

		// D.3: Multi-site views
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/cross-site", h.GetCrossSiteRoutings)

		// E.1: Operational playbooks
		api.With(requireRole("viewer", "editor", "admin")).Get("/playbooks", h.ListPlaybooks)
		api.With(requireRole("viewer", "editor", "admin")).Get("/playbooks/{id}", h.GetPlaybook)
		api.With(requireRole("editor", "admin")).Put("/playbooks/{id}", h.UpsertPlaybook)
		api.With(requireRole("admin")).Delete("/playbooks/{id}", h.DeletePlaybook)
		api.With(requireRole("editor", "admin")).Post("/playbooks/{id}/execute", h.ExecutePlaybook)
		api.With(requireRole("viewer", "editor", "admin")).Get("/playbooks/{id}/executions", h.ListPlaybookExecutions)

		// E.2: Scheduling & maintenance windows
		api.With(requireRole("editor", "admin")).Post("/playbooks/{id}/schedule", h.CreateScheduledPlaybookExecution)
		api.With(requireRole("viewer", "editor", "admin")).Get("/playbooks/{id}/schedule", h.ListScheduledPlaybookExecutions)
		api.With(requireRole("viewer", "editor", "admin")).Get("/schedule/playbooks", h.ListScheduledPlaybookExecutions)
		api.With(requireRole("editor", "admin")).Delete("/schedule/playbooks/{id}", h.DeleteScheduledPlaybookExecution)

		api.With(requireRole("viewer", "editor", "admin")).Get("/maintenance/windows", h.ListMaintenanceWindows)
		api.With(requireRole("viewer", "editor", "admin")).Get("/maintenance/windows/active", h.GetActiveMaintenanceWindows)
		api.With(requireRole("editor", "admin")).Post("/maintenance/windows", h.CreateMaintenanceWindow)
		api.With(requireRole("editor", "admin")).Put("/maintenance/windows/{id}", h.UpdateMaintenanceWindow)
		api.With(requireRole("admin")).Delete("/maintenance/windows/{id}", h.DeleteMaintenanceWindow)

		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/collisions", h.CheckerCollisions)
		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/nmos", h.CheckerNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/latest", h.CheckerLatest)
		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/check", h.CheckCollision)
		api.With(requireRole("editor", "admin")).Post("/checker/run", h.CheckerRun)

		api.With(requireRole("editor", "admin")).Get("/automation/jobs", h.ListAutomationJobs)
		api.With(requireRole("editor", "admin")).Get("/automation/jobs/{job_id}", h.GetAutomationJob)
		api.With(requireRole("admin")).Put("/automation/jobs/{job_id}", h.PutAutomationJob)
		api.With(requireRole("admin")).Post("/automation/jobs/{job_id}/enable", h.EnableAutomationJob)
		api.With(requireRole("admin")).Post("/automation/jobs/{job_id}/disable", h.DisableAutomationJob)
		api.With(requireRole("viewer", "editor", "admin")).Get("/automation/summary", h.GetAutomationSummary)

		api.With(requireRole("viewer", "editor", "admin")).Get("/address-map", h.AddressMap)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address-map/subnet/analysis", h.GetSubnetDetailedAnalysis)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/privileged", h.ListPrivilegedBuckets)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/all", h.ListAllBuckets)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/{id}/children", h.ListBucketChildren)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/{id}/usage", h.GetBucketUsageStats)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/{id}/flows", h.GetBucketFlows)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/parent", h.CreateParentBucket)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/child", h.CreateChildBucket)
		api.With(requireRole("editor", "admin")).Patch("/address/buckets/{id}", h.PatchBucket)
		api.With(requireRole("admin")).Delete("/address/buckets/{id}", h.DeleteBucket)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/export", h.ExportBuckets)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/import", h.ImportBuckets)
		api.With(requireRole("admin")).Get("/logs", h.Logs)
		api.With(requireRole("admin")).Get("/logs/download", h.DownloadLogs)

		// F.2: Metrics & dashboards
		api.With(requireRole("viewer", "editor", "admin")).Get("/metrics/summary", h.GetMetricsSummary)

		// NMOS registry configuration (multi-registry support – A.1)
		api.With(requireRole("viewer", "editor", "admin")).Get("/registry/config", h.GetRegistryConfig)
		api.With(requireRole("viewer", "editor", "admin")).Get("/registry/config/stats", h.GetRegistryConfigStats)
		api.With(requireRole("admin")).Put("/registry/config", h.PutRegistryConfig)
		api.With(requireRole("admin")).Post("/registry/config/remove", h.RemoveRegistryConfig)

		// NMOS registry compatibility matrix (A.2)
		api.With(requireRole("viewer", "editor", "admin")).Get("/registry/compat", h.RegistryCompatibilityMatrix)

		// NMOS snapshot & conformance (A.5)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/snapshot", h.ExportNMOSSnapshot)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/conformance", h.CheckNMOSConformance)

		// B.1: Receiver connection state (IS-05)
		api.With(requireRole("viewer", "editor", "admin")).Get("/receiver/connections", h.GetReceiverConnections)
		api.With(requireRole("viewer", "editor", "admin")).Get("/receiver/{receiver_id}/history", h.GetReceiverConnectionHistory)
		api.With(requireRole("editor", "admin")).Put("/receiver/connection", h.PutReceiverConnection)
		api.With(requireRole("editor", "admin")).Delete("/receiver/{receiver_id}/connection", h.DeleteReceiverConnection)

		// BCC-style: IS-05 active state and receiver disable (un-TAKE)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/receivers-active", h.GetReceiversActive)
		api.With(requireRole("editor", "admin")).Post("/nmos/receiver-disable", h.ReceiverDisable)

		// B.2: Scheduled activations (time-based IS-05 patches)
		// Note: Specific routes ({id}) must come before generic routes to avoid route conflicts
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/scheduled-activations/{id}", h.GetScheduledActivation)
		api.With(requireRole("editor", "admin")).Delete("/nmos/scheduled-activations/{id}", h.DeleteScheduledActivation)
		api.With(requireRole("editor", "admin")).Post("/nmos/scheduled-activations", h.CreateScheduledActivation)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/scheduled-activations", h.ListScheduledActivations)

		// B.3: Routing policies
		api.With(requireRole("viewer", "editor", "admin")).Post("/routing/check", h.CheckRoutingPolicy)
		api.With(requireRole("viewer", "editor", "admin")).Get("/routing/policies/audits", h.ListRoutingPolicyAudits)
		api.With(requireRole("viewer", "editor", "admin")).Get("/routing/policies/{id}", h.GetRoutingPolicy)
		api.With(requireRole("editor", "admin")).Put("/routing/policies/{id}", h.UpdateRoutingPolicy)
		api.With(requireRole("editor", "admin")).Delete("/routing/policies/{id}", h.DeleteRoutingPolicy)
		api.With(requireRole("editor", "admin")).Post("/routing/policies", h.CreateRoutingPolicy)
		api.With(requireRole("viewer", "editor", "admin")).Get("/routing/policies", h.ListRoutingPolicies)

		// System backup & restore
		api.With(requireRole("viewer", "editor", "admin")).Get("/system/backup", h.ExportSystemBackup)
		api.With(requireRole("admin")).Post("/system/restore", h.ImportSystemBackup)

		// G.3: Interoperability & test harness
		api.With(requireRole("viewer", "editor", "admin")).Get("/interop/targets", h.ListInteropTargets)
		api.With(requireRole("editor", "admin")).Post("/interop/test", h.RunInteropTests)
	})

	return r
}

func (h *Handler) createToken(username, role string) (string, error) {
	now := time.Now()
	claims := AuthClaims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-nmos",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(12 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func parseBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}
