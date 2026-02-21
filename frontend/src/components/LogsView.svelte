<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconDownload } from "../lib/icons.js";
  import { api } from "../lib/api.js";

  let {
    logsKind = $bindable("api"),
    logsLines = [],
    onLoadLogs,
    token = "",
  } = $props();

  let loading = $state(false);
  let downloading = $state(false);
  let previousLogsKind = $state(logsKind);

  let levelFilter = $state("");
  let componentFilter = $state("");
  let correlationIdFilter = $state("");
  let requestIdFilter = $state("");
  let siteFilter = $state("");
  let textFilter = $state("");
  let parsedLogs = $state([]);
  let filteredLogs = $state([]);

  // Watch for logsKind changes and reload logs
  $effect(() => {
    if (logsKind && logsKind !== previousLogsKind && onLoadLogs) {
      previousLogsKind = logsKind;
      onLoadLogs();
    }
  });

  async function handleRefresh() {
    loading = true;
    try {
      await onLoadLogs?.();
    } finally {
      loading = false;
    }
  }

  async function handleDownload() {
    if (!token) {
      alert("Authentication token is required for download");
      return;
    }

    downloading = true;
    try {
      const protocol = typeof window !== "undefined" ? window.location.protocol : "http:";
      const hostname = typeof window !== "undefined" ? window.location.hostname : "localhost";
      const url = `${protocol}//${hostname}:9090/api/logs/download?kind=${logsKind}`;

      const response = await fetch(url, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Download failed: ${response.statusText}`);
      }

      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = downloadUrl;
      link.download = `${logsKind}.log`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      console.error("Download error:", error);
      alert(`Failed to download logs: ${error.message}`);
    } finally {
      downloading = false;
    }
  }

  function parseLogLine(line) {
    try {
      const obj = JSON.parse(line);
      return {
        ts: obj.ts || null,
        level: obj.level || "",
        kind: obj.kind || "",
        component: obj.component || "",
        message: obj.message || "",
        method: obj.method || "",
        path: obj.path || "",
        status: obj.status || "",
        duration_ms: obj.duration_ms || "",
        user: obj.user || "",
        request_id: obj.request_id || "",
        correlation_id: obj.correlation_id || "",
        site: obj.site || "",
        remote_ip: obj.remote_ip || "",
        raw: line,
      };
    } catch (e) {
      return {
        ts: null,
        level: "",
        kind: logsKind,
        component: "",
        message: line,
        method: "",
        path: "",
        status: "",
        duration_ms: "",
        user: "",
        request_id: "",
        correlation_id: "",
        site: "",
        remote_ip: "",
        raw: line,
      };
    }
  }

  function formatTimestamp(ts) {
    if (!ts) return "";
    try {
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return ts;
      return d.toLocaleString();
    } catch (e) {
      return ts;
    }
  }

  $effect(() => {
    parsedLogs = logsLines.map(parseLogLine);
  });

  $effect(() => {
    filteredLogs = parsedLogs.filter((entry) => {
      if (levelFilter && entry.level !== levelFilter) return false;
      if (componentFilter && entry.component !== componentFilter) return false;
      if (correlationIdFilter && !entry.correlation_id.includes(correlationIdFilter)) return false;
      if (requestIdFilter && !entry.request_id.includes(requestIdFilter)) return false;
      if (siteFilter && !entry.site.includes(siteFilter)) return false;
      if (textFilter) {
        const haystack = (
          entry.raw ||
          `${entry.message} ${entry.path} ${entry.user} ${entry.request_id} ${entry.correlation_id}`
        ).toLowerCase();
        if (!haystack.includes(textFilter.toLowerCase())) return false;
      }
      return true;
    });
  });
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Logs</h3>
      <p class="text-[11px] text-gray-400">
        View and download structured application logs with filters
      </p>
    </div>
    <div class="flex items-center gap-2">
      <select
        bind:value={logsKind}
        class="px-3 py-1.5 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-xs focus:outline-none focus:border-orange-500 transition-colors"
      >
        <option value="api">API Logs</option>
        <option value="audit">Audit Logs</option>
      </select>
      <button
        onclick={handleRefresh}
        disabled={loading}
        class="px-3 py-1.5 rounded-md bg-gray-800 border border-gray-700 text-gray-200 text-xs font-medium hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {loading ? "Loading..." : "Refresh"}
      </button>
      <button
        onclick={handleDownload}
        disabled={downloading || !token}
        class="px-3 py-1.5 rounded-md bg-orange-600 hover:bg-orange-500 disabled:opacity-50 disabled:cursor-not-allowed text-white text-xs font-medium transition-colors flex items-center gap-1.5"
      >
        {@html IconDownload}
        {downloading ? "Downloading..." : "Download"}
      </button>
    </div>
  </div>

  <div class="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 text-[11px] text-gray-200">
    <div class="flex flex-col gap-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Severity</label>
      <select
        bind:value={levelFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
      >
        <option value="">All</option>
        <option value="info">info</option>
        <option value="warn">warn</option>
        <option value="error">error</option>
      </select>
    </div>
    <div class="flex flex-col gap-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Component</label>
      <input
        bind:value={componentFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
        placeholder="e.g. flows, playbooks, registry"
      />
    </div>
    <div class="flex flex-col gap-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Correlation ID</label>
      <input
        bind:value={correlationIdFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
        placeholder="Search by correlation ID"
      />
    </div>
    <div class="flex flex-col gap-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Request ID</label>
      <input
        bind:value={requestIdFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
        placeholder="Search by request ID"
      />
    </div>
    <div class="flex flex-col gap-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Registry / Site</label>
      <input
        bind:value={siteFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
        placeholder="Site or registry identifier"
      />
    </div>
    <div class="flex flex-col gap-1 sm:col-span-2 lg:col-span-1">
      <label class="text-[10px] uppercase tracking-wide text-gray-500">Text search</label>
      <input
        bind:value={textFilter}
        class="px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-200 text-[11px] focus:outline-none focus:border-orange-500 transition-colors"
        placeholder="Search in message, path, user..."
      />
    </div>
  </div>

  <div class="mt-3 rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    {#if filteredLogs.length === 0}
      <div class="p-12">
        <EmptyState
          title="No logs available"
          message="Logs will appear here once the application starts generating log entries."
        />
      </div>
    {:else}
      <div class="p-4 bg-gray-950 border-b border-gray-800">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-400">Log Type:</span>
            <span class="text-xs font-medium text-gray-200 uppercase">{logsKind}</span>
          </div>
          <span class="text-xs text-gray-500">
            {filteredLogs.length} of {logsLines.length} lines
          </span>
        </div>
      </div>
      <div class="overflow-auto max-h-[600px]">
        <table class="min-w-full text-[11px] text-left text-gray-200">
          <thead class="bg-gray-950 border-b border-gray-800 sticky top-0 z-10">
            <tr>
              <th class="px-3 py-2 font-medium text-gray-400">Time</th>
              <th class="px-3 py-2 font-medium text-gray-400">Level</th>
              <th class="px-3 py-2 font-medium text-gray-400">Component</th>
              <th class="px-3 py-2 font-medium text-gray-400">User</th>
              <th class="px-3 py-2 font-medium text-gray-400">Request / Corr ID</th>
              <th class="px-3 py-2 font-medium text-gray-400">Site</th>
              <th class="px-3 py-2 font-medium text-gray-400">Method</th>
              <th class="px-3 py-2 font-medium text-gray-400">Path</th>
              <th class="px-3 py-2 font-medium text-gray-400">Status</th>
              <th class="px-3 py-2 font-medium text-gray-400">Message</th>
            </tr>
          </thead>
          <tbody>
            {#each filteredLogs as entry (entry.raw)}
              <tr class="border-b border-gray-800/80 hover:bg-gray-950/70">
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {formatTimestamp(entry.ts)}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap">
                  {#if entry.level === "error"}
                    <span class="px-1.5 py-0.5 rounded bg-red-900/60 text-red-200 border border-red-700/60">
                      {entry.level || "info"}
                    </span>
                  {:else if entry.level === "warn"}
                    <span class="px-1.5 py-0.5 rounded bg-yellow-900/60 text-yellow-200 border border-yellow-700/60">
                      {entry.level}
                    </span>
                  {:else}
                    <span class="px-1.5 py-0.5 rounded bg-gray-800 text-gray-200 border border-gray-700">
                      {entry.level || "info"}
                    </span>
                  {/if}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {entry.component || "–"}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {entry.user || "–"}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap">
                  <div class="flex flex-col gap-0.5">
                    <span class="text-gray-300 truncate max-w-[160px]" title={entry.request_id}>
                      {entry.request_id || "–"}
                    </span>
                    <span class="text-[10px] text-gray-500 truncate max-w-[160px]" title={entry.correlation_id}>
                      {entry.correlation_id || ""}
                    </span>
                  </div>
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {entry.site || "–"}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {entry.method || "–"}
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300" title={entry.path}>
                  <span class="truncate max-w-[200px] inline-block">{entry.path || "–"}</span>
                </td>
                <td class="px-3 py-1.5 whitespace-nowrap text-gray-300">
                  {entry.status || "–"}
                </td>
                <td class="px-3 py-1.5 text-gray-200">
                  <div class="max-w-[320px] whitespace-pre-wrap break-words">
                    {entry.message || entry.raw}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</section>

