package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsSummaryResponse represents a summary of key metrics for UI display
type MetricsSummaryResponse struct {
	HTTPRequests struct {
		Total    int64             `json:"total"`
		ByMethod map[string]int64  `json:"by_method"`
		ByStatus map[string]int64  `json:"by_status"`
		Latency  map[string]float64 `json:"latency_p95"` // p95 latency by path
	} `json:"http_requests"`
	Registry struct {
		Nodes   map[string]int64 `json:"nodes_by_site"`
		Devices map[string]int64 `json:"devices_by_site"`
		Flows   map[string]int64 `json:"flows_by_site"`
		Health  map[string]int64 `json:"health_by_registry"`
	} `json:"registry"`
	Routing struct {
		OperationsTotal int64            `json:"operations_total"`
		FailuresTotal   int64            `json:"failures_total"`
		ByOperation     map[string]int64 `json:"by_operation"`
	} `json:"routing"`
	Automation struct {
		JobsTotal    int64            `json:"jobs_total"`
		RunsTotal    int64            `json:"runs_total"`
		ByJobType    map[string]int64 `json:"by_job_type"`
		ByStatus     map[string]int64 `json:"by_status"`
	} `json:"automation"`
}

// GetMetricsSummary returns a JSON summary of key metrics for UI display
// GET /api/metrics/summary
func (h *Handler) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	// Update registry metrics before serving
	h.updateRegistryMetrics(r.Context())
	h.updateAutomationJobMetrics(r.Context())

	summary := MetricsSummaryResponse{}
	summary.HTTPRequests.ByMethod = make(map[string]int64)
	summary.HTTPRequests.ByStatus = make(map[string]int64)
	summary.HTTPRequests.Latency = make(map[string]float64)
	summary.Registry.Nodes = make(map[string]int64)
	summary.Registry.Devices = make(map[string]int64)
	summary.Registry.Flows = make(map[string]int64)
	summary.Registry.Health = make(map[string]int64)
	summary.Routing.ByOperation = make(map[string]int64)
	summary.Automation.ByJobType = make(map[string]int64)
	summary.Automation.ByStatus = make(map[string]int64)

	// Collect HTTP request metrics
	collectCounterMetric(httpRequestsTotal, func(labels map[string]string, value float64) {
		method := labels["method"]
		status := labels["status"]
		summary.HTTPRequests.Total += int64(value)
		summary.HTTPRequests.ByMethod[method] += int64(value)
		summary.HTTPRequests.ByStatus[status] += int64(value)
	})

	// Collect registry metrics
	collectGaugeMetric(registryNodesTotal, func(labels map[string]string, value float64) {
		site := labels["site"]
		if site == "" {
			site = "unknown"
		}
		summary.Registry.Nodes[site] = int64(value)
	})
	collectGaugeMetric(registryDevicesTotal, func(labels map[string]string, value float64) {
		site := labels["site"]
		if site == "" {
			site = "unknown"
		}
		summary.Registry.Devices[site] = int64(value)
	})
	collectGaugeMetric(registryFlowsTotal, func(labels map[string]string, value float64) {
		site := labels["site"]
		if site == "" {
			site = "unknown"
		}
		summary.Registry.Flows[site] = int64(value)
	})
	collectGaugeMetric(registryHealth, func(labels map[string]string, value float64) {
		registryURL := labels["registry_url"]
		if registryURL == "" {
			registryURL = "internal"
		}
		summary.Registry.Health[registryURL] = int64(value)
	})

	// Collect routing metrics
	collectCounterMetric(routingOperationsTotal, func(labels map[string]string, value float64) {
		operation := labels["operation"]
		summary.Routing.OperationsTotal += int64(value)
		summary.Routing.ByOperation[operation] += int64(value)
	})
	collectCounterMetric(routingOperationFailures, func(labels map[string]string, value float64) {
		summary.Routing.FailuresTotal += int64(value)
	})

	// Collect automation metrics
	collectGaugeMetric(automationJobsTotal, func(labels map[string]string, value float64) {
		jobType := labels["job_type"]
		status := labels["status"]
		summary.Automation.JobsTotal += int64(value)
		summary.Automation.ByJobType[jobType] += int64(value)
		summary.Automation.ByStatus[status] += int64(value)
	})
	collectCounterMetric(automationJobRunsTotal, func(labels map[string]string, value float64) {
		summary.Automation.RunsTotal += int64(value)
	})

	writeJSON(w, http.StatusOK, summary)
}

// Helper functions to collect Prometheus metrics using default registry
func collectCounterMetric(metric *prometheus.CounterVec, fn func(map[string]string, float64)) {
	registry := prometheus.DefaultGatherer
	mfs, err := registry.Gather()
	if err != nil {
		return
	}
	metricName := "nmos_http_requests_total"
	if metric == routingOperationsTotal {
		metricName = "nmos_routing_operations_total"
	} else if metric == routingOperationFailures {
		metricName = "nmos_routing_operation_failures_total"
	} else if metric == automationJobRunsTotal {
		metricName = "nmos_automation_job_runs_total"
	}
	for _, mf := range mfs {
		if mf.GetName() != metricName {
			continue
		}
		for _, m := range mf.Metric {
			labels := make(map[string]string)
			for _, labelPair := range m.Label {
				labels[*labelPair.Name] = *labelPair.Value
			}
			var value float64
			if m.Counter != nil {
				value = *m.Counter.Value
			}
			fn(labels, value)
		}
	}
}

func collectGaugeMetric(metric *prometheus.GaugeVec, fn func(map[string]string, float64)) {
	registry := prometheus.DefaultGatherer
	mfs, err := registry.Gather()
	if err != nil {
		return
	}
	var metricName string
	if metric == registryNodesTotal {
		metricName = "nmos_registry_nodes_total"
	} else if metric == registryDevicesTotal {
		metricName = "nmos_registry_devices_total"
	} else if metric == registryFlowsTotal {
		metricName = "nmos_registry_flows_total"
	} else if metric == registryHealth {
		metricName = "nmos_registry_health"
	} else if metric == automationJobsTotal {
		metricName = "nmos_automation_jobs_total"
	}
	for _, mf := range mfs {
		if mf.GetName() != metricName {
			continue
		}
		for _, m := range mf.Metric {
			labels := make(map[string]string)
			for _, labelPair := range m.Label {
				labels[*labelPair.Name] = *labelPair.Value
			}
			var value float64
			if m.Gauge != nil {
				value = *m.Gauge.Value
			}
			fn(labels, value)
		}
	}
}
