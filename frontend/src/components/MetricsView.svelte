<script>
  import { api } from "../lib/api.js";
  import EmptyState from "./EmptyState.svelte";

  let { token = "" } = $props();

  // Get Prometheus metrics URL (same base as API but /metrics endpoint)
  const getPrometheusURL = () => {
    if (typeof window === "undefined") return "/metrics";
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;
    const port = "9090"; // Backend port
    return `${protocol}//${hostname}:${port}/metrics`;
  };

  let loading = $state(false);
  let error = $state("");
  let metrics = $state(null);
  let refreshInterval = $state(null);

  async function loadMetrics() {
    loading = true;
    error = "";
    try {
      metrics = await api("/metrics/summary", { token });
    } catch (e) {
      error = e.message;
      metrics = null;
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadMetrics();
    // Auto-refresh every 30 seconds
    refreshInterval = setInterval(() => {
      loadMetrics();
    }, 30000);
    return () => {
      if (refreshInterval) {
        clearInterval(refreshInterval);
      }
    };
  });

  function formatNumber(num) {
    if (num == null) return "0";
    return new Intl.NumberFormat().format(num);
  }
</script>

<section class="mt-4 space-y-4">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Metrics & Monitoring</h3>
      <p class="text-[11px] text-gray-400">
        Real-time system metrics and performance indicators
      </p>
    </div>
    <button
      onclick={loadMetrics}
      disabled={loading}
      class="px-3 py-1.5 rounded-md bg-gray-800 border border-gray-700 text-gray-200 text-xs font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
    >
      {loading ? "Loading..." : "Refresh"}
    </button>
  </div>

  {#if error}
    <div class="p-4 rounded-xl border border-red-800 bg-red-900/20 text-red-200 text-sm">
      Error loading metrics: {error}
    </div>
  {:else if loading && !metrics}
    <div class="p-12">
      <EmptyState title="Loading metrics..." message="Fetching metrics data..." />
    </div>
  {:else if !metrics}
    <div class="p-12">
      <EmptyState title="No metrics available" message="Metrics will appear here once data is collected." />
    </div>
  {:else}
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <!-- HTTP Requests -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-xs font-semibold text-gray-300 mb-3 uppercase tracking-wide">HTTP Requests</h4>
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Total Requests</span>
            <span class="text-sm font-medium text-gray-200">{formatNumber(metrics.http_requests?.total)}</span>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">By Method</p>
            {#each Object.entries(metrics.http_requests?.by_method || {}) as [method, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{method}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
          <div class="mt-3 pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">By Status</p>
            {#each Object.entries(metrics.http_requests?.by_status || {}) as [status, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{status}</span>
                <span class="text-xs font-medium {status.startsWith('2') ? 'text-green-300' : status.startsWith('4') || status.startsWith('5') ? 'text-red-300' : 'text-gray-300'}">
                  {formatNumber(count)}
                </span>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Registry -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-xs font-semibold text-gray-300 mb-3 uppercase tracking-wide">NMOS Registry</h4>
        <div class="space-y-3">
          <div>
            <p class="text-[10px] text-gray-500 mb-2 uppercase">Nodes by Site</p>
            {#each Object.entries(metrics.registry?.nodes_by_site || {}) as [site, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{site || "unknown"}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
          <div class="pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">Devices by Site</p>
            {#each Object.entries(metrics.registry?.devices_by_site || {}) as [site, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{site || "unknown"}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
          <div class="pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">Flows by Site</p>
            {#each Object.entries(metrics.registry?.flows_by_site || {}) as [site, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{site || "unknown"}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
          <div class="pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">Health Status</p>
            {#each Object.entries(metrics.registry?.health_by_registry || {}) as [registry, health]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{registry}</span>
                <span class="text-xs font-medium {health === 1 ? 'text-green-300' : 'text-red-300'}">
                  {health === 1 ? "Healthy" : "Unhealthy"}
                </span>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Routing Operations -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-xs font-semibold text-gray-300 mb-3 uppercase tracking-wide">Routing Operations</h4>
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Total Operations</span>
            <span class="text-sm font-medium text-gray-200">{formatNumber(metrics.routing?.operations_total)}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Total Failures</span>
            <span class="text-sm font-medium text-red-300">{formatNumber(metrics.routing?.failures_total)}</span>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">By Operation</p>
            {#each Object.entries(metrics.routing?.by_operation || {}) as [operation, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{operation}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Automation Jobs -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-xs font-semibold text-gray-300 mb-3 uppercase tracking-wide">Automation Jobs</h4>
        <div class="space-y-2">
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Total Jobs</span>
            <span class="text-sm font-medium text-gray-200">{formatNumber(metrics.automation?.jobs_total)}</span>
          </div>
          <div class="flex justify-between items-center">
            <span class="text-xs text-gray-400">Total Runs</span>
            <span class="text-sm font-medium text-gray-200">{formatNumber(metrics.automation?.runs_total)}</span>
          </div>
          <div class="mt-3 pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">By Job Type</p>
            {#each Object.entries(metrics.automation?.by_job_type || {}) as [jobType, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{jobType}</span>
                <span class="text-xs font-medium text-gray-300">{formatNumber(count)}</span>
              </div>
            {/each}
          </div>
          <div class="mt-3 pt-3 border-t border-gray-800">
            <p class="text-[10px] text-gray-500 mb-2 uppercase">By Status</p>
            {#each Object.entries(metrics.automation?.by_status || {}) as [status, count]}
              <div class="flex justify-between items-center mb-1">
                <span class="text-xs text-gray-400">{status}</span>
                <span class="text-xs font-medium {status === 'success' ? 'text-green-300' : status === 'error' ? 'text-red-300' : 'text-gray-300'}">
                  {formatNumber(count)}
                </span>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>

    <div class="mt-4 p-4 rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
      <p class="text-xs text-gray-400">
        Metrics are automatically refreshed every 30 seconds. For advanced monitoring and historical data, use Prometheus and Grafana.
        <a href={getPrometheusURL()} target="_blank" class="text-orange-500 hover:text-orange-400 underline ml-1">
          View Prometheus endpoint →
        </a>
      </p>
    </div>
  {/if}
</section>
