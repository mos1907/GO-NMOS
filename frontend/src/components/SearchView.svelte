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
      <h3 class="text-sm font-semibold text-gray-100">Quick Search</h3>
      <p class="text-[11px] text-gray-400">Quickly find flows by name, IP or flow ID</p>
    </div>
    <span class="text-[11px] text-gray-400">
      {searchTotal > 0
        ? `${searchOffset + 1}-${Math.min(searchOffset + searchLimit, searchTotal)} / ${searchTotal}`
        : "0 result"}
    </span>
  </div>
  <div class="flex flex-wrap gap-2 items-center">
    <input
      bind:value={searchTerm}
      placeholder="Search by name/ip/flow id/note..."
      class="px-3 py-2 rounded-md border border-gray-700 bg-gray-900 text-sm min-w-[260px] flex-1 text-gray-100 placeholder:text-gray-500 focus:outline-none focus:border-orange-500"
    />
    <button
      class="px-3 py-2 rounded-md bg-orange-600 text-white text-xs font-semibold hover:bg-orange-500 transition-colors"
      on:click={onRunSearch}
    >
      Search
    </button>
    <button
      class="px-2.5 py-1.5 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40"
      on:click={onPrevSearchPage}
      disabled={searchOffset === 0}
    >
      Prev
    </button>
    <button
      class="px-2.5 py-1.5 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40"
      on:click={onNextSearchPage}
      disabled={searchOffset + searchLimit >= searchTotal}
    >
      Next
    </button>
  </div>
  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-gray-800">
        <tr>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Display Name</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Flow ID</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Multicast</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Port</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#each searchResults as flow}
          <tr class="hover:bg-gray-800/70">
            <td class="px-3 py-2 text-[13px] font-medium text-gray-100">{flow.display_name}</td>
            <td class="px-3 py-2 text-gray-300">{flow.flow_id}</td>
            <td class="px-3 py-2 text-gray-300">{flow.multicast_ip}</td>
            <td class="px-3 py-2 text-gray-300">{flow.port}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

