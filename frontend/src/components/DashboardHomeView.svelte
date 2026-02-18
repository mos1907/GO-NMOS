<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconPlus } from "../lib/icons.js";

  export let summary = { total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 };
  export let flows = [];
  export let flowTotal = 0;
  export let onCreateFlow = null;
</script>

<section class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 mb-4">
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
</section>

<section class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm">
  <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Latest Flows</h3>
      <p class="text-[11px] text-gray-400 mt-0.5">Son eklenen/ güncellenen akışların özeti</p>
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

