<script>
  export let flows = [];
  export let flowLimit = 50;
  export let flowOffset = 0;
  export let flowTotal = 0;
  export let flowSortBy = "updated_at";
  export let flowSortOrder = "desc";

  export let canEdit = false;
  export let isAdmin = false;

  export let onApplyFlowSort;
  export let onPrevFlowPage;
  export let onNextFlowPage;
  export let onToggleFlowLock;
  export let onDeleteFlow;
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-black">Flows</h3>
      <p class="text-[11px] text-black/60">Filter and manage flow list</p>
    </div>
    <div class="flex flex-wrap items-center gap-2 text-xs">
      <label class="flex items-center gap-1 text-black/70">
        <span>Sort by</span>
        <select
          bind:value={flowSortBy}
          on:change={onApplyFlowSort}
          class="px-2 py-1 rounded-md border border-slate-300 bg-white text-xs"
        >
          <option value="updated_at">updated_at</option>
          <option value="created_at">created_at</option>
          <option value="display_name">display_name</option>
          <option value="flow_status">flow_status</option>
          <option value="multicast_ip">multicast_ip</option>
          <option value="source_ip">source_ip</option>
          <option value="port">port</option>
        </select>
      </label>
      <label class="flex items-center gap-1 text-black/70">
        <span>Order</span>
        <select
          bind:value={flowSortOrder}
          on:change={onApplyFlowSort}
          class="px-2 py-1 rounded-md border border-slate-300 bg-white text-xs"
        >
          <option value="desc">desc</option>
          <option value="asc">asc</option>
        </select>
      </label>
      <div class="flex items-center gap-1">
        <button
          class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
          on:click={onPrevFlowPage}
          disabled={flowOffset === 0}
        >
          Prev
        </button>
        <button
          class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
          on:click={onNextFlowPage}
          disabled={flowOffset + flowLimit >= flowTotal}
        >
          Next
        </button>
      </div>
      <span class="text-[11px] text-black/60">
        Showing {flowOffset + 1}-{Math.min(flowOffset + flowLimit, flowTotal)} / {flowTotal}
      </span>
    </div>
  </div>

  <div class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-slate-50">
        <tr>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Display Name</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Flow ID</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Multicast</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Source</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Port</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Status</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Availability</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Locked</th>
          {#if canEdit}
            <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Action</th>
          {/if}
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        {#each flows as flow}
          <tr class="hover:bg-slate-50/80">
            <td class="px-3 py-2 text-[13px] font-medium text-black">{flow.display_name}</td>
            <td class="px-3 py-2 text-black/70">{flow.flow_id}</td>
            <td class="px-3 py-2 text-black">{flow.multicast_ip}</td>
            <td class="px-3 py-2 text-black">{flow.source_ip}</td>
            <td class="px-3 py-2 text-black">{flow.port}</td>
            <td class="px-3 py-2 text-black/80">{flow.flow_status}</td>
            <td class="px-3 py-2 text-black/80">{flow.availability}</td>
            <td class="px-3 py-2 text-lg">{flow.locked ? "🔒" : "🔓"}</td>
            {#if canEdit}
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-1.5">
                  <button
                    class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-[11px] hover:bg-slate-50"
                    on:click={() => onToggleFlowLock?.(flow)}
                  >
                    {flow.locked ? "Unlock" : "Lock"}
                  </button>
                  {#if isAdmin}
                    <button
                      class="px-2.5 py-1 rounded-md border border-red-300 bg-red-50 text-[11px] text-red-700 hover:bg-red-100"
                      on:click={() => onDeleteFlow?.(flow)}
                    >
                      Delete
                    </button>
                  {/if}
                </div>
              </td>
            {/if}
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

