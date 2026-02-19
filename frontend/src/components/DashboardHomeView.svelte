<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconPlus } from "../lib/icons.js";

  export let summary = { total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 };
  export let flows = [];
  export let flowTotal = 0;
  export let onCreateFlow = null;
  export let realtimeEvents = [];
  export let registryEvents = [];
  export let registryHealth = null;
  export let automationSummary = null;
</script>

<section class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3 mb-4">
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
</section>

<section class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
  <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Latest Flows</h3>
      <p class="text-[11px] text-gray-400 mt-0.5">Summary of recently added/updated flows</p>
    </div>
    <span class="text-[11px] text-gray-300 bg-slate-900 px-2 py-0.5 rounded-full border border-gray-700">
      Showing {Math.min(flows.length, 12)} of {flowTotal}
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
        {#if flows.length === 0}
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
          {#each flows.slice(0, 12) as flow}
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

<section class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-3">
  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm md:col-span-2">
    <!-- existing Latest Flows table kept as-is above -->
  </div>
  <div class="space-y-3">
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
  </div>
</section>
