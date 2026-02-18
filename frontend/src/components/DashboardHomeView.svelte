<script>
  export let summary = { total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 };
  export let flows = [];
  export let flowTotal = 0;
</script>

<section class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 mb-4">
  <div class="rounded-xl border border-slate-200 bg-white px-3 py-3 shadow-sm">
    <p class="text-[11px] text-black/60 font-medium uppercase tracking-wide">Total</p>
    <p class="mt-1 text-2xl font-semibold text-black">{summary.total}</p>
  </div>
  <div class="rounded-xl border border-slate-200 bg-emerald-50 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-emerald-800 font-medium uppercase tracking-wide">Active</p>
    <p class="mt-1 text-2xl font-semibold text-emerald-900">{summary.active}</p>
  </div>
  <div class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-amber-800 font-medium uppercase tracking-wide">Locked</p>
    <p class="mt-1 text-2xl font-semibold text-amber-900">{summary.locked}</p>
  </div>
  <div class="rounded-xl border border-slate-200 bg-white px-3 py-3 shadow-sm">
    <p class="text-[11px] text-black/60 font-medium uppercase tracking-wide">Unused</p>
    <p class="mt-1 text-2xl font-semibold text-black">{summary.unused}</p>
  </div>
  <div class="rounded-xl border border-sky-200 bg-sky-50 px-3 py-3 shadow-sm">
    <p class="text-[11px] text-sky-800 font-medium uppercase tracking-wide">Maintenance</p>
    <p class="mt-1 text-2xl font-semibold text-sky-900">{summary.maintenance}</p>
  </div>
</section>

<section class="rounded-xl border border-slate-200 bg-white shadow-sm">
  <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100">
    <div>
      <h3 class="text-sm font-semibold text-black">Latest Flows</h3>
      <p class="text-[11px] text-black/60 mt-0.5">Son eklenen/ güncellenen akışların özeti</p>
    </div>
    <span class="text-[11px] text-black/70 bg-nmos-bg px-2 py-0.5 rounded-full border border-slate-200">
      Showing {Math.min(flows.length, 12)} of {flowTotal}
    </span>
  </div>
  <div class="overflow-x-auto">
    <table class="min-w-full text-xs">
      <thead class="bg-slate-50">
        <tr>
          <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Display Name</th>
          <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Flow ID</th>
          <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Multicast</th>
          <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Port</th>
          <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Status</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        {#each flows.slice(0, 12) as flow}
          <tr class="hover:bg-slate-50/80">
            <td class="px-4 py-2 text-black truncate text-[13px] font-medium">{flow.display_name}</td>
            <td class="px-4 py-2 text-black/70 truncate">{flow.flow_id}</td>
            <td class="px-4 py-2 text-black">{flow.multicast_ip}</td>
            <td class="px-4 py-2 text-black">{flow.port}</td>
            <td class="px-4 py-2">
              <span
                class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium {flow.flow_status === 'active'
                  ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
                  : flow.flow_status === 'maintenance'
                    ? 'bg-amber-50 text-amber-800 border border-amber-200'
                    : 'bg-slate-50 text-slate-700 border border-slate-200'}"
              >
                {flow.flow_status}
              </span>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

