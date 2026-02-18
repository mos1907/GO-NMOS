<script>
  export let nmosNodes = [];
  export let selectedSenderNodeId = "";
  export let selectedReceiverNodeId = "";
  export let senderNodeSenders = [];
  export let receiverNodeReceivers = [];
  export let selectedPatchSender = null;
  export let selectedPatchReceiver = null;
  export let nmosIS05Base = "";
  export let nmosPatchStatus = "";
  export let nmosPatchError = "";
  export let nmosTakeBusy = false;
  export let senderFilterText = "";
  export let receiverFilterText = "";
  export let senderFormatFilter = "";
  export let receiverFormatFilter = "";
  export let showAddNodeModal = false;
  export let newNodeName = "";
  export let newNodeUrl = "";

  export let onReloadNodes;
  export let onOpenAddNode;
  export let onCancelAddNode;
  export let onConfirmAddNode;
  export let onChangeNewNodeName;
  export let onChangeNewNodeUrl;
  export let onSelectSenderNode;
  export let onSelectReceiverNode;
  export let onUpdateSenderFilterText;
  export let onUpdateReceiverFilterText;
  export let onUpdateSenderFormatFilter;
  export let onUpdateReceiverFormatFilter;
  export let onSelectPatchSender;
  export let onSelectPatchReceiver;
  export let onExecutePatchTake;
  export let isPatchTakeReady;

  export let filterSenders;
  export let filterReceivers;
</script>

<div class="space-y-4">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <h3 class="text-black font-semibold text-lg">NMOS Patch</h3>
    </div>
    <div class="flex items-center gap-2">
      <button
        class="px-3 py-1.5 rounded-md text-sm bg-nmos-bg hover:bg-svelte/20 border border-svelte/40 text-black font-medium"
        on:click={onReloadNodes}
      >
        Reload Nodes
      </button>
      <button
        class="px-3 py-1.5 rounded-md text-sm bg-svelte hover:bg-orange-500 text-black font-semibold"
        on:click={onOpenAddNode}
      >
        Add Node
      </button>
    </div>
  </div>

  <div class="flex flex-wrap items-center gap-2 text-sm">
    <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-svelte/20 border border-svelte/40">
      <span class="inline-flex h-2.5 w-2.5 rounded-full {isPatchTakeReady?.() ? 'bg-svelte' : 'bg-black/30'}"></span>
      <span class="text-black font-medium">
        {#if !selectedPatchSender && !selectedPatchReceiver}
          Select source and destination
        {:else if !selectedPatchSender}
          Select a source
        {:else if !selectedPatchReceiver}
          Select a destination
        {:else if !nmosIS05Base}
          IS-05 base URL missing
        {:else}
          Ready to patch
        {/if}
      </span>
    </div>
    {#if nmosPatchError}
      <span class="text-sm text-red-600 font-medium">Error: {nmosPatchError}</span>
    {/if}
    {#if nmosPatchStatus}
      <span class="text-sm text-svelte font-medium">{nmosPatchStatus}</span>
    {/if}
  </div>

  <div class="grid md:grid-cols-[3fr_1fr_3fr] gap-6 items-stretch">
    <!-- Source panel -->
    <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col min-h-[500px]">
      <div class="flex items-center justify-between mb-3 gap-2">
        <div>
          <h4 class="text-base font-semibold text-black">Sources</h4>
          <p class="text-[12px] text-black/70">Source selection</p>
        </div>
        <select
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
          value={selectedSenderNodeId}
          on:change={(e) => onSelectSenderNode?.(e.target.value)}
        >
          <option value="">Select node…</option>
          {#each nmosNodes as node}
            <option value={node.id}>{node.name}</option>
          {/each}
        </select>
      </div>

      <div class="flex flex-wrap gap-2 mb-3 text-sm">
        <input
          value={senderFilterText}
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 flex-1 min-w-[150px] text-black"
          placeholder="Search sources..."
          on:input={(e) => onUpdateSenderFilterText?.(e.target.value)}
        />
        <select
          value={senderFormatFilter}
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-black"
          on:change={(e) => onUpdateSenderFormatFilter?.(e.target.value)}
        >
          <option value="">All Formats</option>
          <option value="video">Video</option>
          <option value="audio">Audio</option>
          <option value="data">Data</option>
          <option value="mux">Mux</option>
        </select>
      </div>

      <div class="overflow-auto flex-1 divide-y divide-svelte/20 text-sm">
        {#if senderNodeSenders.length === 0}
          <div class="px-2 py-4 text-center text-black/60 text-sm">
            No sources. Add a node and load sources via IS-04.
          </div>
        {:else}
          {#each (filterSenders ? filterSenders(senderNodeSenders) : senderNodeSenders) as s}
            <button
              type="button"
              class="w-full text-left px-3 py-2 hover:bg-svelte/20 flex justify-between gap-2 {selectedPatchSender && selectedPatchSender.id === s.id
                ? 'bg-svelte/30 border-l-4 border-svelte'
                : ''}"
              on:click={() => onSelectPatchSender?.(s)}
            >
              <span class="truncate text-black font-medium">{s.label}</span>
              <span class="text-[12px] text-black/60 truncate">{s.flow_id}</span>
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Center TAKE button -->
    <div class="flex flex-col items-center justify-center gap-4">
      <div class="rounded-2xl border border-slate-300 bg-white shadow-lg shadow-orange-500/20 p-3">
        <button
          class="w-40 h-40 rounded-xl bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-svelte-soft font-bold text-xl flex flex-col items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed hover:scale-105 active:scale-100 transition transform"
          on:click={onExecutePatchTake}
          disabled={!isPatchTakeReady?.()}
        >
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-svelte to-orange-400 flex items-center justify-center shadow-inner">
            <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </div>
          <span class="tracking-wide">{nmosTakeBusy ? "TAKING..." : "TAKE"}</span>
          <span class="text-[11px] font-medium text-slate-300">
            {#if !isPatchTakeReady?.()}
              Select endpoints to enable
            {:else}
              Ready to patch
            {/if}
          </span>
        </button>
      </div>
      <div class="flex items-center gap-2 text-sm text-black px-4 py-2 rounded-full bg-svelte/20 border border-svelte/40">
        <span class="inline-flex h-3 w-3 rounded-full {isPatchTakeReady?.() ? 'bg-svelte' : 'bg-black/30'}"></span>
        <span class="font-medium">
          {#if !selectedPatchSender && !selectedPatchReceiver}
            Select source and destination
          {:else if !selectedPatchSender}
            Select a source
          {:else if !selectedPatchReceiver}
            Select a destination
          {:else if !nmosIS05Base}
            IS-05 base URL missing
          {:else}
            Ready
          {/if}
        </span>
      </div>
    </div>

    <!-- Destinations panel -->
    <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col min-h-[500px]">
      <div class="flex items-center justify-between mb-3 gap-2">
        <div>
          <h4 class="text-base font-semibold text-black">Destinations</h4>
          <p class="text-[12px] text-black/70">Destination selection</p>
        </div>
        <select
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
          value={selectedReceiverNodeId}
          on:change={(e) => onSelectReceiverNode?.(e.target.value)}
        >
          <option value="">Select node…</option>
          {#each nmosNodes as node}
            <option value={node.id}>{node.name}</option>
          {/each}
        </select>
      </div>

      <div class="flex flex-wrap gap-2 mb-3 text-sm">
        <input
          value={receiverFilterText}
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 flex-1 min-w-[150px] text-black"
          placeholder="Search destinations..."
          on:input={(e) => onUpdateReceiverFilterText?.(e.target.value)}
        />
        <select
          value={receiverFormatFilter}
          class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-black"
          on:change={(e) => onUpdateReceiverFormatFilter?.(e.target.value)}
        >
          <option value="">All Formats</option>
          <option value="video">Video</option>
          <option value="audio">Audio</option>
          <option value="data">Data</option>
          <option value="mux">Mux</option>
        </select>
      </div>

      <div class="overflow-auto flex-1 divide-y divide-svelte/20 text-sm">
        {#if receiverNodeReceivers.length === 0}
          <div class="px-2 py-4 text-center text-black/60 text-sm">
            No destinations. Add a node and load destinations via IS-04.
          </div>
        {:else}
          {#each (filterReceivers ? filterReceivers(receiverNodeReceivers) : receiverNodeReceivers) as r}
            <button
              type="button"
              class="w-full text-left px-3 py-2 hover:bg-svelte/20 flex justify-between gap-2 {selectedPatchReceiver && selectedPatchReceiver.id === r.id
                ? 'bg-svelte/30 border-l-4 border-svelte'
                : ''}"
              on:click={() => onSelectPatchReceiver?.(r)}
            >
              <span class="truncate text-black font-medium">{r.label}</span>
              <span class="text-[12px] text-black/60 uppercase truncate">{r.format}</span>
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </div>

  {#if showAddNodeModal}
    <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div class="bg-nmos-bg border border-svelte/60 rounded-xl p-4 w-full max-w-md space-y-3">
        <h4 class="text-base font-semibold text-black">Add NMOS Node</h4>
        <div class="space-y-2">
          <div class="flex flex-col gap-1">
            <label class="text-sm text-black/80 font-medium">Node Name</label>
            <input
              value={newNodeName}
              class="px-3 py-2 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
              placeholder="e.g. Camera Router"
              on:input={(e) => onChangeNewNodeName?.(e.target.value)}
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-sm text-black/80 font-medium">IS-04 URL</label>
            <input
              value={newNodeUrl}
              class="px-3 py-2 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
              placeholder="http://192.168.x.x:port"
              on:input={(e) => onChangeNewNodeUrl?.(e.target.value)}
            />
            <p class="text-xs text-black/60">
              NMOS Node IS-04 URL. IS-05 endpoint is derived automatically.
            </p>
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button
            class="px-3 py-1.5 rounded-md text-sm bg-nmos-bg border border-svelte/40 text-black font-medium hover:bg-svelte/10"
            on:click={onCancelAddNode}
          >
            Cancel
          </button>
          <button
            class="px-3 py-1.5 rounded-md text-sm bg-svelte hover:bg-orange-500 text-black font-semibold"
            on:click={onConfirmAddNode}
          >
            Add Node
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

