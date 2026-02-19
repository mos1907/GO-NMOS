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
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Logs</h3>
      <p class="text-[11px] text-gray-400">View and download application logs</p>
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

  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    {#if logsLines.length === 0}
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
          <span class="text-xs text-gray-500">{logsLines.length} lines</span>
        </div>
      </div>
      <div class="overflow-auto max-h-[600px]">
        <pre class="p-4 text-xs font-mono text-green-200 bg-gray-950 leading-relaxed whitespace-pre-wrap break-words">{logsLines.join("\n")}</pre>
      </div>
    {/if}
  </div>
</section>

