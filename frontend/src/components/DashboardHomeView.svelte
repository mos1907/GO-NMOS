<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconPlus } from "../lib/icons.js";

  let {
    summary = { total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 },
    flows = [],
    flowTotal = 0,
    onCreateFlow = null,
    systemInfo = null,
    realtimeEvents = [],
    registryEvents = [],
    registryHealth = null,
    automationSummary = null,
    registryConfigs = [],
    registryCompat = [],
    sitesRoomsSummary = null,
    // Diagnostics / Health panel
    healthDetail = null,
    healthLoading = false,
    healthError = "",
    lastHealthLoadedAt = "",
    onRunHealthDetail = null,
    // Diagnostics: Check Node at URL
    nodeCheckUrl = "",
    nodeCheckLoading = false,
    nodeCheckError = "",
    nodeCheckResult = null,
    onNodeUrlChange = null,
    onRunNodeCheck = null,
  } = $props();

  // Safe derived: UI won't break if parent sends null or $state proxy
  // In Svelte 5, $state proxies are array-like (length, map, filter work)
  // Use flows directly in $derived(); use || [] for null check
  const safeFlows = $derived(flows || []);
</script>

<section class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-9 gap-3 mb-4">
  <div class="rounded-xl border border-gray-800 bg-gray-900 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-gray-400 font-medium uppercase tracking-wide">Total</p>
    <p class="mt-1 text-2xl font-semibold text-gray-100">{summary.total}</p>
  </div>
  <div class="rounded-xl border border-emerald-700 bg-emerald-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-emerald-300 font-medium uppercase tracking-wide">Active</p>
    <p class="mt-1 text-2xl font-semibold text-emerald-100">{summary.active}</p>
  </div>
  <div class="rounded-xl border border-amber-700 bg-amber-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-amber-300 font-medium uppercase tracking-wide">Locked</p>
    <p class="mt-1 text-2xl font-semibold text-amber-100">{summary.locked}</p>
  </div>
  <div class="rounded-xl border border-gray-800 bg-gray-900 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-gray-400 font-medium uppercase tracking-wide">Unused</p>
    <p class="mt-1 text-2xl font-semibold text-gray-100">{summary.unused}</p>
  </div>
  <div class="rounded-xl border border-sky-700 bg-sky-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-sky-300 font-medium uppercase tracking-wide">Maintenance</p>
    <p class="mt-1 text-2xl font-semibold text-sky-100">{summary.maintenance}</p>
  </div>
  <div class="rounded-xl border border-indigo-700 bg-indigo-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-indigo-300 font-medium uppercase tracking-wide">System / Timing</p>
    {#if systemInfo}
      <div class="mt-1 text-[11px] text-indigo-100 space-y-0.5">
        <div>PTP Domain: <span class="font-semibold">{systemInfo.ptp_domain || "-"}</span></div>
        <div>GMID: <span class="font-semibold break-all">{systemInfo.ptp_gmid || "-"}</span></div>
        <div>IS-04: <span class="font-semibold">{systemInfo.expected_is04 || "-"}</span></div>
        <div>IS-05: <span class="font-semibold">{systemInfo.expected_is05 || "-"}</span></div>
      </div>
    {:else}
      <p class="mt-1 text-[11px] text-indigo-100/80">No system parameters configured.</p>
    {/if}
  </div>
  <div class="rounded-xl border border-purple-700 bg-purple-950 px-3 py-3 shadow-sm md:col-span-2">
    <p class="text-[11px] text-purple-300 font-medium uppercase tracking-wide">Automation / Checks</p>
    {#if automationSummary}
      <div class="mt-1 flex items-center justify-between text-[11px] text-purple-100">
        <div class="space-y-0.5">
          <div>Total jobs: <span class="font-semibold">{automationSummary.total_jobs}</span></div>
          <div>Enabled: <span class="font-semibold">{automationSummary.enabled_jobs}</span></div>
        </div>
        <div class="space-y-0.5 text-right">
          <div>Collisions: <span class="font-semibold">{automationSummary.collision_count}</span></div>
          <div>NMOS diffs: <span class="font-semibold">{automationSummary.nmos_difference_count}</span></div>
        </div>
      </div>
      {#if automationSummary.last_updated}
        <p class="mt-1 text-[10px] text-purple-200/80">
          Last check: {new Date(automationSummary.last_updated).toLocaleString()}
        </p>
      {/if}
    {:else}
      <p class="mt-1 text-[11px] text-purple-100/80">No automation/check results yet.</p>
    {/if}
  </div>
  <div class="rounded-xl border border-teal-700 bg-teal-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-teal-300 font-medium uppercase tracking-wide">Registries</p>
    {#if registryConfigs && registryConfigs.length > 0}
      <div class="mt-1 text-[11px] text-teal-100 space-y-0.5">
        <div>
          Total: <span class="font-semibold">{registryConfigs.length}</span>
        </div>
        <div>
          Enabled:
          <span class="font-semibold">
            {registryConfigs.filter((r) => r.enabled).length}
          </span>
        </div>
      </div>
    {:else}
      <p class="mt-1 text-[11px] text-teal-100/80">No registries configured.</p>
    {/if}
  </div>
  <div class="rounded-xl border border-violet-700 bg-violet-950 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-violet-300 font-medium uppercase tracking-wide">Sites & Rooms</p>
    {#if sitesRoomsSummary}
      <div class="mt-1 text-[11px] text-violet-100 space-y-0.5">
        <div>
          Sites: <span class="font-semibold">{sitesRoomsSummary.sites?.length || 0}</span>
        </div>
        <div>
          Rooms: <span class="font-semibold">{sitesRoomsSummary.rooms?.length || 0}</span>
        </div>
        {#if sitesRoomsSummary.domains?.length}
          <div>
            Domains: <span class="font-semibold">{sitesRoomsSummary.domains.length}</span>
          </div>
        {/if}
      </div>
    {:else}
      <p class="mt-1 text-[11px] text-violet-100/80">No site/room tags found.</p>
    {/if}
  </div>
</section>

<section class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
  <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Latest Flows</h3>
      <p class="text-[11px] text-gray-400 mt-0.5">Summary of recently added/updated flows</p>
    </div>
    <span class="text-[11px] text-gray-300 bg-slate-900 px-2 py-0.5 rounded-full border border-gray-700">
      Showing {Math.min(safeFlows.length, 12)} of {flowTotal || 0}
    </span>
  </div>
  <div class="overflow-x-auto">
    <table class="min-w-full text-xs">
      <thead class="bg-gray-800">
        <tr>
          <th class="text-left px-4 py-2 border-b border-gray-800 font-medium text-gray-200">Display Name</th>
          <th class="text-left px-4 py-2 border-b border-gray-800 font-medium text-gray-200">Flow ID</th>
          <th class="text-left px-4 py-2 border-b border-gray-800 font-medium text-gray-200">Multicast</th>
          <th class="text-left px-4 py-2 border-b border-gray-800 font-medium text-gray-200">Port</th>
          <th class="text-left px-4 py-2 border-b border-gray-800 font-medium text-gray-200">Status</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#if safeFlows.length === 0}
          <tr>
            <td colspan="5" class="px-6 py-12">
              <EmptyState
                title="No flows yet"
                message="Create your first flow to get started with NMOS management."
                actionLabel={onCreateFlow ? "Create Flow" : ""}
                onAction={onCreateFlow}
                icon={IconPlus}
              />
            </td>
          </tr>
        {:else}
          {#each safeFlows.slice(0, 12) as flow}
            <tr class="hover:bg-gray-800/70 transition-colors">
              <td class="px-4 py-2 text-gray-100 truncate text-[13px] font-medium">{flow.display_name}</td>
              <td class="px-4 py-2 text-gray-300 truncate">{flow.flow_id}</td>
              <td class="px-4 py-2 text-gray-300">{flow.multicast_ip}</td>
              <td class="px-4 py-2 text-gray-300">{flow.port}</td>
              <td class="px-4 py-2">
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium transition-colors {flow.flow_status === 'active'
                    ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                    : flow.flow_status === 'maintenance'
                      ? 'bg-amber-900 text-amber-200 border border-amber-700'
                      : 'bg-slate-800 text-slate-200 border border-slate-700'}"
                >
                  {flow.flow_status}
                </span>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<section class="mt-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div>
          <h3 class="text-sm font-semibold text-gray-100">IS-04 Compatibility Matrix</h3>
          <p class="text-[11px] text-gray-400 mt-0.5">Configured registries vs expected IS-04</p>
        </div>
      </div>
      <div class="max-h-40 overflow-y-auto text-[11px]">
        {#if !registryCompat || registryCompat.length === 0}
          <div class="px-4 py-3 text-gray-500">No registry compatibility data yet.</div>
        {:else}
          <table class="min-w-full">
            <thead>
              <tr class="bg-gray-800/60">
                <th class="px-3 py-2 text-left font-medium text-gray-200">Name</th>
                <th class="px-3 py-2 text-left font-medium text-gray-200">Role</th>
                <th class="px-3 py-2 text-left font-medium text-gray-200">Query ver</th>
                <th class="px-3 py-2 text-left font-medium text-gray-200">Expected</th>
                <th class="px-3 py-2 text-left font-medium text-gray-200">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each registryCompat as reg}
                <tr>
                  <td class="px-3 py-1.5 text-gray-100 truncate">
                    {reg.name || "(unnamed)"}
                  </td>
                  <td class="px-3 py-1.5 text-gray-300">
                    {reg.role || "-"}
                  </td>
                  <td class="px-3 py-1.5 text-gray-300">
                    {reg.chosen_query_ver || "-"}
                  </td>
                  <td class="px-3 py-1.5 text-gray-300">
                    {reg.expected_is04 || "-"}
                  </td>
                  <td class="px-3 py-1.5">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide
                        {reg.status === 'ok'
                          ? 'bg-emerald-950 text-emerald-200 border border-emerald-700'
                          : reg.status === 'warning'
                            ? 'bg-amber-950 text-amber-200 border border-amber-700'
                            : reg.status === 'unsupported'
                              ? 'bg-slate-900 text-slate-200 border border-slate-700'
                              : 'bg-red-950 text-red-200 border border-red-700'}"
                      title={reg.error}
                    >
                      {reg.status}
                    </span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    </div>

    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div>
          <h3 class="text-sm font-semibold text-gray-100">Realtime Flow Events</h3>
          <p class="text-[11px] text-gray-400 mt-0.5">MQTT feed (last 20)</p>
        </div>
      </div>
      <div class="max-h-40 overflow-y-auto text-[11px]">
        {#if !realtimeEvents || realtimeEvents.length === 0}
          <div class="px-4 py-3 text-gray-500">No events yet</div>
        {:else}
          <ul class="divide-y divide-gray-800">
            {#each realtimeEvents.slice(0, 20) as ev}
              <li class="px-4 py-2 flex items-start gap-2">
                <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-semibold
                  {ev.event === 'created'
                    ? 'bg-emerald-900 text-emerald-200'
                    : ev.event === 'updated'
                      ? 'bg-sky-900 text-sky-200'
                      : ev.event === 'deleted'
                        ? 'bg-red-900 text-red-200'
                        : 'bg-slate-800 text-slate-200'}">
                  {ev.event}
                </span>
                <div class="flex-1 space-y-0.5">
                  <div class="text-gray-100 truncate">
                    {ev.flow?.display_name || ev.flow_id}
                  </div>
                  <div class="text-gray-500 truncate">
                    {ev.flow?.multicast_ip}:{ev.flow?.port}
                  </div>
                  <div class="text-gray-600">
                    {ev.timestamp}
                  </div>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>

    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div>
          <h3 class="text-sm font-semibold text-gray-100">Registry Events</h3>
          <p class="text-[11px] text-gray-400 mt-0.5">NMOS registry sync (last 20)</p>
        </div>
        {#if registryHealth}
          <span
            class="text-[10px] px-2 py-0.5 rounded-full border font-semibold uppercase tracking-wide
              {registryHealth.ok
                ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                : 'bg-amber-950 text-amber-200 border-amber-700'}"
            title={"nodes=" + registryHealth.counts?.nodes + ", devices=" + registryHealth.counts?.devices}
          >
            {registryHealth.ok ? "OK" : "EMPTY"}
          </span>
        {/if}
      </div>
      <div class="max-h-40 overflow-y-auto text-[11px]">
        {#if !registryEvents || registryEvents.length === 0}
          <div class="px-4 py-3 text-gray-500">No registry events yet</div>
        {:else}
          <ul class="divide-y divide-gray-800">
            {#each registryEvents.slice(0, 20) as ev}
              <li class="px-4 py-2 flex items-start gap-2">
                <span class="px-1.5 py-0.5 rounded text-[10px] uppercase font-semibold bg-indigo-900 text-indigo-200">
                  {ev.kind}
                </span>
                <div class="flex-1 space-y-0.5">
                  <div class="text-gray-100 truncate">
                    {ev.resource}
                  </div>
                  <div class="text-gray-500 truncate">
                    {#if ev.info}
                      {ev.info.base_url} · nodes:{ev.info.nodes} flows:{ev.info.flows} senders:{ev.info.senders} receivers:{ev.info.receivers}
                    {/if}
                  </div>
                  <div class="text-gray-600">
                    {ev.timestamp}
                  </div>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>

    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
        <div>
          <h3 class="text-sm font-semibold text-gray-100">Diagnostics / Quick Check</h3>
          <p class="text-[11px] text-gray-400 mt-0.5">One-click backend health snapshot</p>
        </div>
        <button
          class="text-[11px] px-2 py-1 rounded-md border border-orange-700 bg-orange-900 text-orange-100 hover:bg-orange-800 disabled:opacity-60 disabled:cursor-not-allowed"
          disabled={healthLoading}
          on:click={onRunHealthDetail}
        >
          {healthLoading ? "Running..." : "Run check"}
        </button>
      </div>
      <div class="px-4 py-3 text-[11px] space-y-3">
        {#if healthError}
          <div class="text-red-400">Error: {healthError}</div>
        {:else if !healthDetail}
          <div class="text-gray-500">No diagnostics run yet.</div>
        {:else}
          <div class="flex items-center justify-between">
            <span class="text-gray-300">Overall</span>
            <span
              class="px-2 py-0.5 rounded-full border text-[10px] font-semibold uppercase tracking-wide
                {healthDetail.status === 'ok'
                  ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                  : 'bg-amber-950 text-amber-200 border-amber-700'}"
            >
              {healthDetail.status}
            </span>
          </div>
          <div class="mt-2 grid grid-cols-1 gap-2">
            <div class="flex items-center justify-between">
              <div class="text-gray-300">Database</div>
              {#if healthDetail.components?.db}
                <span
                  class="px-2 py-0.5 rounded-full border text-[10px] font-semibold uppercase tracking-wide
                    {healthDetail.components.db.ok
                      ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                      : 'bg-red-950 text-red-200 border-red-700'}"
                  title={healthDetail.components.db.error}
                >
                  {healthDetail.components.db.ok ? "OK" : "ERROR"}
                </span>
              {/if}
            </div>
            <div class="flex items-center justify-between">
              <div class="text-gray-300">MQTT</div>
              {#if healthDetail.components?.mqtt}
                <span
                  class="px-2 py-0.5 rounded-full border text-[10px] font-semibold uppercase tracking-wide
                    {healthDetail.components.mqtt.enabled
                      ? (healthDetail.components.mqtt.ok
                          ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                          : 'bg-red-950 text-red-200 border-red-700')
                      : 'bg-slate-900 text-slate-200 border-slate-700'}"
                  title={healthDetail.components.mqtt.broker_url}
                >
                  {#if !healthDetail.components.mqtt.enabled}
                    DISABLED
                  {:else if healthDetail.components.mqtt.ok}
                    OK
                  {:else}
                    ERROR
                  {/if}
                </span>
              {/if}
            </div>
            <div class="flex items-center justify-between">
              <div class="text-gray-300">Registry</div>
              {#if healthDetail.components?.registry}
                <span
                  class="px-2 py-0.5 rounded-full border text-[10px] font-semibold uppercase tracking-wide
                    {healthDetail.components.registry.ok
                      ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                      : 'bg-amber-950 text-amber-200 border-amber-700'}"
                  title={"nodes=" + (healthDetail.components.registry.counts?.nodes || 0)}
                >
                  {healthDetail.components.registry.ok ? "OK" : "EMPTY"}
                </span>
              {/if}
            </div>
          </div>
          {#if lastHealthLoadedAt}
            <p class="mt-2 text-[10px] text-gray-500">
              Last run: {new Date(lastHealthLoadedAt).toLocaleString()}
            </p>
          {/if}

          <!-- F.3: Incident Hints -->
          {#if healthDetail?.hints && healthDetail.hints.length > 0}
            <div class="mt-4 pt-3 border-t border-gray-800">
              <p class="text-[10px] text-gray-500 mb-2 uppercase font-semibold">What to Check Next</p>
              <div class="space-y-2">
                {#each healthDetail.hints as hint}
                  <div class="rounded-lg border p-2.5 {
                    hint.severity === 'critical' 
                      ? 'border-red-800 bg-red-950/30' 
                      : hint.severity === 'error'
                      ? 'border-orange-800 bg-orange-950/30'
                      : 'border-yellow-800 bg-yellow-950/30'
                  }">
                    <div class="flex items-start justify-between gap-2 mb-1.5">
                      <div class="flex-1">
                        <div class="flex items-center gap-1.5 mb-1">
                          <span class="text-[10px] font-semibold uppercase tracking-wide {
                            hint.severity === 'critical'
                              ? 'text-red-300'
                              : hint.severity === 'error'
                              ? 'text-orange-300'
                              : 'text-yellow-300'
                          }">
                            {hint.severity}
                          </span>
                          <span class="text-[10px] text-gray-400">·</span>
                          <span class="text-[10px] text-gray-400">{hint.component}</span>
                        </div>
                        <p class="text-xs font-medium text-gray-200 mb-0.5">{hint.title}</p>
                        <p class="text-[11px] text-gray-400">{hint.message}</p>
                      </div>
                    </div>
                    {#if hint.suggestions && hint.suggestions.length > 0}
                      <div class="mt-2 pt-2 border-t border-gray-800/50">
                        <p class="text-[10px] text-gray-500 mb-1.5 uppercase">Suggestions:</p>
                        <ul class="space-y-1">
                          {#each hint.suggestions as suggestion}
                            <li class="text-[11px] text-gray-300 flex items-start gap-1.5">
                              <span class="text-gray-500 mt-0.5">•</span>
                              <span>{suggestion}</span>
                            </li>
                          {/each}
                        </ul>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {/if}

        <div class="border-t border-gray-800 pt-3 mt-2 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-gray-300">Check Node at URL</span>
          </div>
          <div class="flex gap-2">
            <input
              type="text"
              class="flex-1 px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-100 placeholder-gray-500 text-[11px] focus:outline-none focus:border-orange-500"
              placeholder="http://node-host:port"
              value={nodeCheckUrl}
              on:input={(e) => onNodeUrlChange && onNodeUrlChange(e.target.value)}
            />
            <button
              class="text-[11px] px-2 py-1 rounded-md border border-sky-700 bg-sky-900 text-sky-100 hover:bg-sky-800 disabled:opacity-60 disabled:cursor-not-allowed"
              disabled={nodeCheckLoading}
              on:click={onRunNodeCheck}
            >
              {nodeCheckLoading ? "Checking..." : "Check"}
            </button>
          </div>
          {#if nodeCheckError}
            <div class="text-[11px] text-red-400">{nodeCheckError}</div>
          {:else if nodeCheckResult}
            <div class="text-[11px] text-gray-300 flex items-center justify-between">
              <span>
                {nodeCheckResult.base_url || nodeCheckResult.target} →
                {nodeCheckResult.httpCode} ({nodeCheckResult.status})
              </span>
              <span class="text-[10px] text-gray-400 ml-2">
                {nodeCheckResult.latency}
              </span>
            </div>
          {/if}
        </div>
      </div>
    </div>
</section>
