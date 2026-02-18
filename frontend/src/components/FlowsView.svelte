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
      <h3 class="text-sm font-semibold text-gray-100">Flows</h3>
      <p class="text-[11px] text-gray-400">Filter and manage flow list</p>
    </div>
    <div class="flex flex-wrap items-center gap-2 text-xs">
      <label class="flex items-center gap-1 text-gray-300">
        <span>Sort by</span>
        <select
          bind:value={flowSortBy}
          on:change={onApplyFlowSort}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500"
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
      <label class="flex items-center gap-1 text-gray-300">
        <span>Order</span>
        <select
          bind:value={flowSortOrder}
          on:change={onApplyFlowSort}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500"
        >
          <option value="desc">desc</option>
          <option value="asc">asc</option>
        </select>
      </label>
      <div class="flex items-center gap-1">
        <button
          class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40"
          on:click={onPrevFlowPage}
          disabled={flowOffset === 0}
        >
          Prev
        </button>
        <button
          class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40"
          on:click={onNextFlowPage}
          disabled={flowOffset + flowLimit >= flowTotal}
        >
          Next
        </button>
      </div>
      <span class="text-[11px] text-gray-400">
        Showing {flowOffset + 1}-{Math.min(flowOffset + flowLimit, flowTotal)} / {flowTotal}
      </span>
    </div>
  </div>

  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-gray-800">
        <tr>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Display Name</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Flow ID</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Multicast</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Source</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Port</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Status</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Availability</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Locked</th>
          {#if canEdit}
            <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Action</th>
          {/if}
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#each flows as flow}
          <tr class="hover:bg-gray-800/70">
            <td class="px-3 py-2 text-[13px] font-medium text-gray-100">{flow.display_name}</td>
            <td class="px-3 py-2 text-gray-300">{flow.flow_id}</td>
            <td class="px-3 py-2 text-gray-300">{flow.multicast_ip}</td>
            <td class="px-3 py-2 text-gray-300">{flow.source_ip}</td>
            <td class="px-3 py-2 text-gray-300">{flow.port}</td>
            <td class="px-3 py-2 text-gray-200">{flow.flow_status}</td>
            <td class="px-3 py-2 text-gray-200">{flow.availability}</td>
            <td class="px-3 py-2 text-lg">{flow.locked ? "🔒" : "🔓"}</td>
            {#if canEdit}
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-1.5">
                  <button
                    class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-[11px] text-gray-200 hover:bg-gray-800"
                    on:click={() => onToggleFlowLock?.(flow)}
                  >
                    {flow.locked ? "Unlock" : "Lock"}
                  </button>
                  {#if isAdmin}
                    <button
                      class="px-2.5 py-1 rounded-md border border-red-800 bg-red-900/60 text-[11px] text-red-200 hover:bg-red-900"
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

