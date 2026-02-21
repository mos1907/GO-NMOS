<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconRefresh } from "../lib/icons.js";
  import { api } from "../lib/api.js";

  let {
    checkerResult = null,
    nmosCheckerResult = null,
    onRunCollisionCheck,
    onRunNmosCheck = null,
    onFlowClick = null,
    token = "",
  } = $props();

  let showAlternativesModal = $state(false);
  let selectedCollision = $state(null); // {multicast_ip, port, flow_names}
  let alternatives = $state([]);
  let loadingAlternatives = $state(false);

  function handleFlowClick(flowName) {
    if (onFlowClick) {
      onFlowClick(flowName);
    }
  }

  async function showAlternatives(item) {
    selectedCollision = item;
    showAlternativesModal = true;
    loadingAlternatives = true;
    alternatives = [];
    
    try {
      const result = await api(`/checker/check?multicast_ip=${encodeURIComponent(item.multicast_ip)}&port=${item.port}`, { token });
      console.log("Alternatives API response:", result);
      if (result.alternatives && Array.isArray(result.alternatives)) {
        alternatives = result.alternatives;
      } else {
        alternatives = [];
      }
    } catch (e) {
      console.error("Failed to load alternatives:", e);
      alternatives = [];
    } finally {
      loadingAlternatives = false;
    }
  }

  function closeAlternativesModal() {
    showAlternativesModal = false;
    selectedCollision = null;
    alternatives = [];
  }
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Collision Checker</h3>
      <p class="text-[11px] text-gray-400">Detect multicast address and port collisions</p>
    </div>
    <div class="flex gap-2">
      <button
        onclick={onRunCollisionCheck}
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
      >
        {@html IconRefresh}
        Run Collision Check
      </button>
      {#if onRunNmosCheck}
        <button
          onclick={onRunNmosCheck}
          class="px-4 py-2 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
        >
          {@html IconRefresh}
          Run NMOS Check
        </button>
      {/if}
    </div>
  </div>

  {#if checkerResult}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-6">
      <div class="flex items-center gap-4 mb-4">
        <div class="flex items-center gap-2">
          <div
            class="w-3 h-3 rounded-full {(checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0) > 0
              ? 'bg-red-500'
              : 'bg-green-500'}"
          ></div>
          <span class="text-lg font-semibold text-gray-100">
            Total Collisions: {checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0}
          </span>
        </div>
      </div>

      {#if (checkerResult.result?.items || checkerResult.items || []).length > 0}
        <div class="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-800">
              <tr>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Multicast IP</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Port</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Count</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Flows</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each (checkerResult.result?.items || checkerResult.items || []) as item}
                <tr class="hover:bg-gray-800/70 transition-colors">
                  <td class="px-4 py-3 text-gray-100 font-medium">{item.multicast_ip}</td>
                  <td class="px-4 py-3 text-gray-300">{item.port}</td>
                  <td class="px-4 py-3">
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-red-900 text-red-200 border border-red-700">
                      {item.count}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-gray-300 text-[11px]">
                    {#each (item.flow_names || []) as flowName, idx}
                      {#if onFlowClick}
                        <button
                          onclick={() => handleFlowClick(flowName)}
                          class="text-blue-400 hover:text-blue-300 hover:underline transition-colors"
                        >
                          {flowName}
                        </button>
                      {:else}
                        <span>{flowName}</span>
                      {/if}
                      {#if idx < (item.flow_names || []).length - 1}
                        <span class="text-gray-500">, </span>
                      {/if}
                    {/each}
                  </td>
                  <td class="px-4 py-3">
                    <button
                      onclick={() => showAlternatives(item)}
                      class="px-2 py-1 text-[10px] bg-blue-600 hover:bg-blue-500 text-white rounded border border-blue-500 transition-colors"
                    >
                      Show Alternatives
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else if (checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0) === 0}
        <div class="text-center py-8">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-900/20 mb-4">
            <svg class="w-8 h-8 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <p class="text-gray-300 font-medium">No collisions detected</p>
          <p class="text-gray-400 text-sm mt-1">All flows have unique multicast addresses and ports</p>
        </div>
      {/if}
    </div>
  {:else}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-12">
      <EmptyState
        title="No check results"
        message="Click 'Run Collision Check' to analyze your flows for address and port conflicts."
        icon={IconRefresh}
      />
    </div>
  {/if}

  <!-- NMOS Difference Check -->
  <div class="mt-8 pt-6 border-t border-gray-800">
    <div class="flex items-center justify-between gap-3 flex-wrap mb-3">
      <div>
        <h3 class="text-sm font-semibold text-gray-100">NMOS Difference Check</h3>
        <p class="text-[11px] text-gray-400">Compare internal flows with NMOS registry (senders)</p>
      </div>
      {#if onRunNmosCheck}
        <button
          onclick={onRunNmosCheck}
          class="px-4 py-2 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
        >
          {@html IconRefresh}
          Run NMOS Check
        </button>
      {/if}
    </div>
    {#if nmosCheckerResult?.result != null}
      {@const res = nmosCheckerResult.result}
      {@const total = res.total_differences ?? res.total ?? 0}
      {@const items = res.items ?? []}
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-6">
        <div class="flex items-center gap-4 mb-4">
          <div
            class="w-3 h-3 rounded-full {total > 0 ? 'bg-amber-500' : 'bg-green-500'}"
          ></div>
          <span class="text-lg font-semibold text-gray-100">
            Differences: {total}
          </span>
          {#if res.checked_flows != null}
            <span class="text-sm text-gray-400">({res.checked_flows} flows checked)</span>
          {/if}
        </div>
        {#if nmosCheckerResult.created_at}
          <p class="text-xs text-gray-500 mb-3">Last run: {nmosCheckerResult.created_at}</p>
        {/if}
        {#if items.length > 0}
          <div class="rounded-lg border border-gray-800 overflow-hidden">
            <table class="min-w-full text-xs">
              <thead class="bg-gray-800">
                <tr>
                  <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Flow</th>
                  <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Issue</th>
                  <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Type</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-800">
                {#each items as item}
                  <tr class="hover:bg-gray-800/70 transition-colors">
                    <td class="px-4 py-3 text-gray-100 font-medium">{item.display_name ?? item.flow_id ?? "—"}</td>
                    <td class="px-4 py-3 text-gray-300">{item.issue ?? "—"}</td>
                    <td class="px-4 py-3 text-gray-400">{item.type ?? "—"}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else if total === 0}
          <p class="text-gray-400 text-sm">No differences; internal flows match NMOS registry.</p>
        {/if}
      </div>
    {:else}
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-8">
        <p class="text-gray-500 text-sm">Run NMOS Check to compare flows with the NMOS registry.</p>
      </div>
    {/if}
  </div>
</section>

<!-- Alternatives Modal -->
{#if showAlternativesModal && selectedCollision}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm"
    role="dialog"
    aria-modal="true"
    aria-labelledby="alternatives-title"
    tabindex="-1"
    onclick={closeAlternativesModal}
    onkeydown={(e) => e.key === "Escape" && closeAlternativesModal()}
  >
    <div
      class="w-full max-w-3xl rounded-xl border border-gray-800 bg-gray-900 p-6 max-h-[90vh] overflow-y-auto shadow-2xl"
      role="document"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between mb-6">
        <div>
          <h3 id="alternatives-title" class="text-lg font-semibold text-gray-100">
            Alternative IP/Port Suggestions
          </h3>
          <p class="text-sm text-gray-400 mt-1">
            For collision: <span class="font-mono text-gray-300">{selectedCollision.multicast_ip}:{selectedCollision.port}</span>
          </p>
          <p class="text-xs text-gray-500 mt-1">
            Conflicting flows: {selectedCollision.flow_names?.join(", ") || "N/A"}
          </p>
        </div>
        <button
          class="px-3 py-1.5 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm transition-colors"
          onclick={closeAlternativesModal}
        >
          Close
        </button>
      </div>

      {#if loadingAlternatives}
        <div class="text-center py-12">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500"></div>
          <p class="text-gray-400 text-sm mt-4">Loading alternatives...</p>
        </div>
      {:else if alternatives.length > 0}
        <div class="space-y-4">
          <p class="text-sm text-gray-300">
            💡 Found <span class="font-semibold text-orange-400">{alternatives.length}</span> alternative IP/Port combinations:
          </p>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {#each alternatives as alt, idx}
              <div class="px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg hover:border-orange-500 transition-colors">
                <div class="flex items-center justify-between mb-2">
                  <div class="font-mono text-gray-100 text-sm font-semibold">
                    {alt.multicast_ip}:{alt.port}
                  </div>
                  <span class="text-xs text-gray-500">#{idx + 1}</span>
                </div>
                <div class="text-gray-400 text-xs mt-2">
                  {#if alt.reason === 'same_subnet_available'}
                    <span class="inline-flex items-center gap-1">
                      <span class="w-2 h-2 rounded-full bg-green-500"></span>
                      Same subnet, available
                    </span>
                  {:else if alt.reason === 'different_port'}
                    <span class="inline-flex items-center gap-1">
                      <span class="w-2 h-2 rounded-full bg-blue-500"></span>
                      Different port
                    </span>
                  {:else if alt.reason === 'different_subnet'}
                    <span class="inline-flex items-center gap-1">
                      <span class="w-2 h-2 rounded-full bg-purple-500"></span>
                      Different subnet
                    </span>
                  {:else}
                    <span class="text-gray-500">{alt.reason || "Available"}</span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
          <div class="mt-4 p-3 bg-gray-800/50 rounded-lg border border-gray-700">
            <p class="text-xs text-gray-400">
              💡 <strong>Tip:</strong> These alternatives are suggested based on availability in your current flow configuration. 
              Choose one that doesn't conflict with your existing flows.
            </p>
          </div>
        </div>
      {:else}
        <div class="text-center py-12">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-800 mb-4">
            <svg class="w-8 h-8 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="text-gray-300 font-medium">No alternatives found</p>
          <p class="text-gray-400 text-sm mt-1">
            Could not find alternative IP/Port combinations for this collision.
          </p>
        </div>
      {/if}
    </div>
  </div>
{/if}
