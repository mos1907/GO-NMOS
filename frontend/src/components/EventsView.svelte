<script>
  import { api } from "../lib/api.js";

  let { token = "" } = $props();

  let events = $state([]);
  let loading = $state(false);
  let error = $state("");
  let filterSource = $state("");
  let filterSeverity = $state("");
  let filterSince = $state("");
  let limit = $state(100);

  const severities = ["", "info", "warning", "error", "critical"];

  async function loadEvents() {
    loading = true;
    error = "";
    try {
      const params = new URLSearchParams();
      if (filterSource) params.set("source", filterSource);
      if (filterSeverity) params.set("severity", filterSeverity);
      if (filterSince) params.set("since", new Date(filterSince).toISOString());
      params.set("limit", String(limit));
      const data = await api(`/events?${params.toString()}`, { token });
      events = data?.events ?? [];
    } catch (e) {
      error = e.message;
      events = [];
    } finally {
      loading = false;
    }
  }

  function severityClass(severity) {
    switch (severity) {
      case "critical":
      case "error":
        return "bg-red-900/50 text-red-200 border-red-700";
      case "warning":
        return "bg-amber-900/50 text-amber-200 border-amber-700";
      default:
        return "bg-gray-700/50 text-gray-200 border-gray-600";
    }
  }
</script>

<section class="mt-4 space-y-4">
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h2 class="text-lg font-semibold text-gray-100 mb-2">Events (IS-07 / Tally)</h2>
    <p class="text-sm text-gray-400 mb-4">
      View events from IS-07 sources, tally, and automation. Filter by source, severity, and time. Events can be correlated with flows, senders, and receivers.
    </p>

    <div class="flex flex-wrap items-center gap-2 mb-4">
      <input
        type="text"
        bind:value={filterSource}
        placeholder="Source URL or ID"
        class="w-48 px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500"
      />
      <select
        bind:value={filterSeverity}
        class="px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200"
      >
        {#each severities as s}
          <option value={s}>{s || "Any severity"}</option>
        {/each}
      </select>
      <input
        type="datetime-local"
        bind:value={filterSince}
        class="px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200"
      />
      <input
        type="number"
        bind:value={limit}
        min="1"
        max="500"
        class="w-20 px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200"
      />
      <span class="text-xs text-gray-500">limit</span>
      <button
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium disabled:opacity-60"
        disabled={loading}
        onclick={loadEvents}
      >
        {loading ? "Loading..." : "Load events"}
      </button>
    </div>

    {#if error}
      <p class="text-sm text-red-400 mb-4">{error}</p>
    {/if}

    <div class="overflow-x-auto border border-gray-800 rounded-lg">
      <table class="w-full text-sm">
        <thead class="bg-gray-800/80 text-gray-300 text-left">
          <tr>
            <th class="px-3 py-2 font-medium">Time</th>
            <th class="px-3 py-2 font-medium">Severity</th>
            <th class="px-3 py-2 font-medium">Source</th>
            <th class="px-3 py-2 font-medium">Message</th>
            <th class="px-3 py-2 font-medium">Flow / Sender / Receiver</th>
          </tr>
        </thead>
        <tbody>
          {#if events.length === 0 && !loading}
            <tr>
              <td colspan="5" class="px-3 py-6 text-center text-gray-500">No events. Use filters and Load, or POST events to /api/events.</td>
            </tr>
          {:else}
            {#each events as ev}
              <tr class="border-t border-gray-800 hover:bg-gray-800/40">
                <td class="px-3 py-2 text-gray-400 whitespace-nowrap">
                  {ev.created_at ? new Date(ev.created_at).toLocaleString() : "—"}
                </td>
                <td class="px-3 py-2">
                  <span class="inline-flex px-2 py-0.5 rounded text-xs font-medium border {severityClass(ev.severity)}">
                    {ev.severity || "info"}
                  </span>
                </td>
                <td class="px-3 py-2 text-gray-300 font-mono text-xs max-w-[120px] truncate" title={ev.source_url || ev.source_id}>
                  {ev.source_id || ev.source_url || "—"}
                </td>
                <td class="px-3 py-2 text-gray-200 max-w-[240px] truncate" title={ev.message}>{ev.message || "—"}</td>
                <td class="px-3 py-2 text-gray-400 text-xs">
                  {#if ev.flow_id || ev.sender_id || ev.receiver_id}
                    {#if ev.flow_id}<span class="block">flow: {ev.flow_id}</span>{/if}
                    {#if ev.sender_id}<span class="block">sender: {ev.sender_id}</span>{/if}
                    {#if ev.receiver_id}<span class="block">receiver: {ev.receiver_id}</span>{/if}
                    {#if ev.job_id}<span class="block">job: {ev.job_id}</span>{/if}
                  {:else}
                    —
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</section>
