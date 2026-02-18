<script>
  export let searchTerm = "";
  export let searchResults = [];
  export let searchLimit = 50;
  export let searchOffset = 0;
  export let searchTotal = 0;

  export let onRunSearch;
  export let onPrevSearchPage;
  export let onNextSearchPage;
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-black">Quick Search</h3>
      <p class="text-[11px] text-black/60">Quickly find flows by name, IP or flow ID</p>
    </div>
    <span class="text-[11px] text-black/60">
      {searchTotal > 0
        ? `${searchOffset + 1}-${Math.min(searchOffset + searchLimit, searchTotal)} / ${searchTotal}`
        : "0 result"}
    </span>
  </div>
  <div class="flex flex-wrap gap-2 items-center">
    <input
      bind:value={searchTerm}
      placeholder="Search by name/ip/flow id/note..."
      class="px-3 py-2 rounded-md border border-slate-300 bg-white text-sm min-w-[260px] flex-1"
    />
    <button
      class="px-3 py-2 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-black"
      on:click={onRunSearch}
    >
      Search
    </button>
    <button
      class="px-2.5 py-1.5 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
      on:click={onPrevSearchPage}
      disabled={searchOffset === 0}
    >
      Prev
    </button>
    <button
      class="px-2.5 py-1.5 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
      on:click={onNextSearchPage}
      disabled={searchOffset + searchLimit >= searchTotal}
    >
      Next
    </button>
  </div>
  <div class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-slate-50">
        <tr>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Display Name</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Flow ID</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Multicast</th>
          <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Port</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        {#each searchResults as flow}
          <tr class="hover:bg-slate-50/80">
            <td class="px-3 py-2 text-[13px] font-medium text-black">{flow.display_name}</td>
            <td class="px-3 py-2 text-black/70">{flow.flow_id}</td>
            <td class="px-3 py-2 text-black">{flow.multicast_ip}</td>
            <td class="px-3 py-2 text-black">{flow.port}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

