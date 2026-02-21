package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go-nmos/backend/internal/repository"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nmos_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nmos_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// Registry health metrics
	registryNodesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nmos_registry_nodes_total",
			Help: "Total number of NMOS nodes per site",
		},
		[]string{"site"},
	)

	registryDevicesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nmos_registry_devices_total",
			Help: "Total number of NMOS devices per site",
		},
		[]string{"site"},
	)

	registryFlowsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nmos_registry_flows_total",
			Help: "Total number of NMOS flows per site",
		},
		[]string{"site"},
	)

	registryHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nmos_registry_health",
			Help: "Registry health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"registry_url"},
	)

	// Routing operations metrics
	routingOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nmos_routing_operations_total",
			Help: "Total number of routing operations",
		},
		[]string{"operation", "status"},
	)

	routingOperationFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nmos_routing_operation_failures_total",
			Help: "Total number of routing operation failures",
		},
		[]string{"operation", "reason"},
	)

	// Automation job metrics
	automationJobsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nmos_automation_jobs_total",
			Help: "Total number of automation jobs",
		},
		[]string{"job_type", "status"},
	)

	automationJobRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "nmos_automation_job_runs_total",
			Help: "Total number of automation job runs",
		},
		[]string{"job_id", "status"},
	)

	automationJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "nmos_automation_job_duration_seconds",
			Help:    "Automation job execution duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
		},
		[]string{"job_id"},
	)
)

// MetricsMiddleware wraps HTTP handlers to collect request metrics
func (h *Handler) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.status)
		method := r.Method
		path := sanitizePath(r.URL.Path)

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	})
}

// MetricsHandler serves Prometheus metrics at /metrics
func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Update registry metrics before serving
	h.updateRegistryMetrics(r.Context())
	h.updateAutomationJobMetrics(r.Context())

	promhttp.Handler().ServeHTTP(w, r)
}

// updateRegistryMetrics updates registry health and node/device/flow counts per site
func (h *Handler) updateRegistryMetrics(ctx context.Context) {
	nodes, err := h.repo.ListNMOSNodes(ctx)
	if err != nil {
		return
	}

	// Group by site
	siteNodes := make(map[string]int)
	siteDevices := make(map[string]int)
	siteFlows := make(map[string]int)

	for _, node := range nodes {
		site := node.GetSiteTag()
		if site == "" {
			site = "unknown"
		}
		siteNodes[site]++

		devices, err := h.repo.ListNMOSDevices(ctx, node.ID)
		if err == nil {
			siteDevices[site] += len(devices)
		}
	}

	// Count flows by site
	flows, err := h.repo.ListNMOSFlows(ctx)
	if err == nil {
		for _, flow := range flows {
			site := flow.GetSiteTag()
			if site == "" {
				site = "unknown"
			}
			siteFlows[site]++
		}
	}

	// Update metrics
	for site, count := range siteNodes {
		registryNodesTotal.WithLabelValues(site).Set(float64(count))
	}
	for site, count := range siteDevices {
		registryDevicesTotal.WithLabelValues(site).Set(float64(count))
	}
	for site, count := range siteFlows {
		registryFlowsTotal.WithLabelValues(site).Set(float64(count))
	}

	// Registry health (simplified: if we have nodes, registry is healthy)
	if len(nodes) > 0 {
		registryHealth.WithLabelValues("internal").Set(1)
	} else {
		registryHealth.WithLabelValues("internal").Set(0)
	}
}

// updateAutomationJobMetrics updates automation job status metrics
func (h *Handler) updateAutomationJobMetrics(ctx context.Context) {
	jobs, err := h.repo.ListAutomationJobs(ctx)
	if err != nil {
		return
	}

	// Reset all job metrics
	automationJobsTotal.Reset()

	// Count jobs by type and status
	for _, job := range jobs {
		status := job.LastRunStatus
		if status == "" {
			status = "idle"
		}
		automationJobsTotal.WithLabelValues(job.JobType, status).Inc()
	}
}

// RecordRoutingOperation records a routing operation metric
func RecordRoutingOperation(operation, status string) {
	routingOperationsTotal.WithLabelValues(operation, status).Inc()
	if status == "failure" || status == "error" {
		routingOperationFailures.WithLabelValues(operation, "unknown").Inc()
	}
}

// RecordRoutingOperationFailure records a routing operation failure with reason
func RecordRoutingOperationFailure(operation, reason string) {
	routingOperationFailures.WithLabelValues(operation, reason).Inc()
	routingOperationsTotal.WithLabelValues(operation, "failure").Inc()
}

// RecordAutomationJobRun records an automation job execution
func RecordAutomationJobRun(jobID, status string, duration time.Duration) {
	automationJobRunsTotal.WithLabelValues(jobID, status).Inc()
	automationJobDuration.WithLabelValues(jobID).Observe(duration.Seconds())
}

// sanitizePath normalizes HTTP paths for metrics (removes IDs, etc.)
func sanitizePath(path string) string {
	return repository.SanitizePathForMetrics(path)
}
