<script>
  import { api } from "../lib/api.js";

  let { token = "" } = $props();

  let flows = $state([]);
  let selectedFlowId = $state("");
  let chain = $state(null);
  let loadingFlows = $state(false);
  let loadingChain = $state(false);
  let error = $state("");

  async function loadFlows() {
    loadingFlows = true;
    error = "";
    try {
      const data = await api("/audio/flows?format=audio", { token });
      flows = data?.flows ?? [];
      if (flows.length && !selectedFlowId) selectedFlowId = flows[0].id ?? "";
    } catch (e) {
      error = e.message;
      flows = [];
    } finally {
      loadingFlows = false;
    }
  }

  async function loadChain() {
    if (!selectedFlowId) return;
    loadingChain = true;
    error = "";
    try {
      chain = await api(`/audio/chain?flow_id=${encodeURIComponent(selectedFlowId)}`, { token });
    } catch (e) {
      error = e.message;
      chain = null;
    } finally {
      loadingChain = false;
    }
  }

  $effect(() => {
    if (selectedFlowId && chain?.flow_id === selectedFlowId) return;
    if (selectedFlowId) loadChain();
  });
</script>

<section class="mt-4 space-y-4">
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h2 class="text-lg font-semibold text-gray-100 mb-2">Audio signal chain (C.2)</h2>
    <p class="text-sm text-gray-400 mb-4">
      Follow an audio (or any) flow across the plant: see which devices receive it and trace the programme path. Data comes from the registry and active receiver connections.
    </p>

    <div class="flex flex-wrap items-center gap-2 mb-4">
      <button
        class="px-4 py-2 rounded-md bg-gray-700 hover:bg-gray-600 text-gray-200 text-sm font-medium"
        disabled={loadingFlows}
        onclick={loadFlows}
      >
        {loadingFlows ? "Loading..." : "Load flows"}
      </button>
      {#if flows.length}
        <label class="text-sm text-gray-400">Flow:</label>
        <select
          class="px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-gray-200 text-sm min-w-[200px]"
          bind:value={selectedFlowId}
          onchange={loadChain}
        >
          <option value="">— Select —</option>
          {#each flows as f}
            <option value={f.id}>{f.label || f.id} ({f.format})</option>
          {/each}
        </select>
        <button
          class="px-3 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm"
          disabled={loadingChain}
          onclick={loadChain}
        >
          {loadingChain ? "Loading..." : "Show chain"}
        </button>
      {/if}
    </div>

    {#if error}
      <p class="text-sm text-red-400 mb-4">{error}</p>
    {/if}

    {#if chain && (chain.hops?.length || chain.flow_id)}
      <div class="border-t border-gray-800 pt-4">
        <h3 class="text-sm font-medium text-gray-200 mb-2">
          Chain for: {chain.flow_label || chain.flow_id} <span class="text-gray-500">({chain.format})</span>
        </h3>
        {#if chain.hops?.length}
          <ul class="space-y-2">
            {#each chain.hops as hop}
              <li class="flex flex-wrap items-center gap-2 text-sm bg-gray-800/50 rounded-lg px-3 py-2 border border-gray-700">
                <span class="font-mono text-gray-400">{hop.flow_label || hop.flow_id}</span>
                <span class="text-gray-500">→</span>
                <span class="text-gray-300">sender {hop.sender_label || hop.sender_id}</span>
                <span class="text-gray-500">→</span>
                <span class="text-amber-200/90">receiver {hop.receiver_label || hop.receiver_id}</span>
                <span class="text-gray-500">on</span>
                <span class="text-sky-200/90 font-medium">{hop.device_label || hop.device_id}</span>
              </li>
            {/each}
          </ul>
          {#if chain.next_flow_ids?.length}
            <p class="text-xs text-gray-500 mt-2">
              Downstream flows (from these devices): {chain.next_flow_ids.join(", ")}
            </p>
          {/if}
        {:else}
          <p class="text-sm text-gray-500">No active receiver connections for this flow. Connect receivers via IS-05 (Registry & Patch) to see the chain.</p>
        {/if}
      </div>
    {/if}
  </div>
</section>
