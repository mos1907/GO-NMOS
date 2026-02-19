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
	// WebSocket endpoint for NMOS registry events (no auth; informational only)
	r.Get("/ws/registry", h.RegistryEventsWS)

	r.Route("/api", func(api chi.Router) {
		api.Use(h.AuthMiddleware)
		api.Get("/me", h.Me)
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
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/{id}/nmos/check", h.CheckFlowNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Get("/flows/{id}/nmos/snapshot", h.GetFlowNMOSSnapShot)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/nmos/sync", h.SyncFlowFromNMOS)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/nmos/apply", h.ApplyFlowNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Post("/flows/{id}/is05/receiver-check", h.CheckIS05ReceiverState)
		api.With(requireRole("editor", "admin")).Post("/flows/{id}/fetch-sdp", h.FetchSDP)

		// Internal NMOS registry (IS-04 style) read-only APIs
		api.With(requireRole("viewer", "editor", "admin")).Post("/nmos/registry/discover-nodes", h.DiscoverNMOSRegistryNodes)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/health", h.NMOSRegistryHealth)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/nodes", h.ListNMOSNodesHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/devices", h.ListNMOSDevicesHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/flows", h.ListNMOSFlowsHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/senders", h.ListNMOSSendersHandler)
		api.With(requireRole("viewer", "editor", "admin")).Get("/nmos/registry/receivers", h.ListNMOSReceiversHandler)

		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/collisions", h.CheckerCollisions)
		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/nmos", h.CheckerNMOS)
		api.With(requireRole("viewer", "editor", "admin")).Get("/checker/latest", h.CheckerLatest)

		api.With(requireRole("editor", "admin")).Get("/automation/jobs", h.ListAutomationJobs)
		api.With(requireRole("editor", "admin")).Get("/automation/jobs/{job_id}", h.GetAutomationJob)
		api.With(requireRole("admin")).Put("/automation/jobs/{job_id}", h.PutAutomationJob)
		api.With(requireRole("admin")).Post("/automation/jobs/{job_id}/enable", h.EnableAutomationJob)
		api.With(requireRole("admin")).Post("/automation/jobs/{job_id}/disable", h.DisableAutomationJob)
		api.With(requireRole("viewer", "editor", "admin")).Get("/automation/summary", h.GetAutomationSummary)

		api.With(requireRole("viewer", "editor", "admin")).Get("/address-map", h.AddressMap)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/privileged", h.ListPrivilegedBuckets)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/{id}/children", h.ListBucketChildren)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/parent", h.CreateParentBucket)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/child", h.CreateChildBucket)
		api.With(requireRole("editor", "admin")).Patch("/address/buckets/{id}", h.PatchBucket)
		api.With(requireRole("admin")).Delete("/address/buckets/{id}", h.DeleteBucket)
		api.With(requireRole("viewer", "editor", "admin")).Get("/address/buckets/export", h.ExportBuckets)
		api.With(requireRole("editor", "admin")).Post("/address/buckets/import", h.ImportBuckets)
		api.With(requireRole("admin")).Get("/logs", h.Logs)
		api.With(requireRole("admin")).Get("/logs/download", h.DownloadLogs)
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
