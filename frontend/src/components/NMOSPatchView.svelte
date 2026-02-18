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

  // Registry (RDS) connect flow (optional)
  export let showConnectRDSModal = false;
  export let registryQueryUrl = "";
  export let registryDiscovering = false;
  export let registryNodes = [];
  export let registrySelectedIds = [];
  export let registryError = "";

  export let onReloadNodes;
  export let onOpenAddNode;
  export let onCancelAddNode;
  export let onConfirmAddNode;
  export let onChangeNewNodeName;
  export let onChangeNewNodeUrl;
  export let onOpenRDS;
  export let onCloseRDS;
  export let onChangeRegistryQueryUrl;
  export let onDiscoverRegistryNodes;
  export let onToggleRegistryNode;
  export let onSelectAllRegistryNodes;
  export let onAddSelectedRegistryNodes;
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

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-wrap items-center justify-between gap-4">
    <div>
      <h2 class="text-xl font-bold text-white mb-1">NMOS Patch Panel</h2>
      <p class="text-sm text-gray-400">
        Connect senders to receivers using IS-04/IS-05
      </p>
    </div>
    <div class="flex items-center gap-2">
      <button
        class="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium transition-colors border border-gray-700"
        on:click={onReloadNodes}
      >
        Reload Nodes
      </button>
      <button
        class="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium transition-colors border border-gray-700"
        on:click={() => onOpenRDS?.()}
      >
        Connect RDS
      </button>
      <button
        class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold transition-colors"
        on:click={onOpenAddNode}
      >
        Add Node
      </button>
    </div>
  </div>

  <!-- Status bar -->
  <div class="flex flex-wrap items-center gap-3">
    <div class="flex items-center gap-2 px-4 py-2 rounded-lg bg-gray-800 border border-gray-700">
      <span
        class="w-2 h-2 rounded-full {isPatchTakeReady?.()
          ? 'bg-green-500'
          : 'bg-gray-500'}"
      ></span>
      <span class="text-sm text-gray-300">
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
      <div class="px-4 py-2 rounded-lg bg-red-900/50 border border-red-800 text-sm text-red-300">
        {nmosPatchError}
      </div>
    {/if}
    {#if nmosPatchStatus}
      <div class="px-4 py-2 rounded-lg bg-green-900/50 border border-green-800 text-sm text-green-300">
        {nmosPatchStatus}
      </div>
    {/if}
  </div>

  <!-- Main grid -->
  <div class="grid md:grid-cols-[1fr_auto_1fr] gap-6">
    <!-- Sources panel -->
    <div class="bg-gray-900 border border-gray-800 rounded-lg p-4 flex flex-col min-h-[500px]">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-base font-semibold text-white mb-1">Sources</h3>
          <p class="text-xs text-gray-400">IS-04 senders</p>
        </div>
        <select
          class="px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 focus:outline-none focus:border-orange-600"
          value={selectedSenderNodeId}
          on:change={(e) => onSelectSenderNode?.(e.target.value)}
        >
          <option value="">Select node…</option>
          {#each nmosNodes as node}
            <option value={node.id}>{node.name}</option>
          {/each}
        </select>
      </div>

      <div class="flex flex-wrap gap-2 mb-4">
        <input
          value={senderFilterText}
          class="flex-1 min-w-[150px] px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 placeholder:text-gray-500 focus:outline-none focus:border-orange-600"
          placeholder="Search sources..."
          on:input={(e) => onUpdateSenderFilterText?.(e.target.value)}
        />
        <select
          value={senderFormatFilter}
          class="px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 focus:outline-none focus:border-orange-600"
          on:change={(e) => onUpdateSenderFormatFilter?.(e.target.value)}
        >
          <option value="">All Formats</option>
          <option value="video">Video</option>
          <option value="audio">Audio</option>
          <option value="data">Data</option>
          <option value="mux">Mux</option>
        </select>
      </div>

      <div class="flex-1 overflow-auto space-y-1">
        {#if senderNodeSenders.length === 0}
          <div class="flex flex-col items-center justify-center py-12 text-center">
            <svg class="w-12 h-12 text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <p class="text-sm font-medium text-gray-400 mb-1">No sources available</p>
            <p class="text-xs text-gray-500">Select a node to load sources</p>
          </div>
        {:else}
          {#each (filterSenders ? filterSenders(senderNodeSenders) : senderNodeSenders) as s}
            <button
              type="button"
              class="w-full text-left px-3 py-2 rounded bg-gray-800 hover:bg-gray-700 border transition-all duration-150 {selectedPatchSender && selectedPatchSender.id === s.id
                ? 'border-orange-600 bg-orange-900/20 shadow-md shadow-orange-500/20'
                : 'border-gray-700'}"
              on:click={() => onSelectPatchSender?.(s)}
            >
              <div class="flex items-center justify-between">
                <div class="min-w-0">
                  <div class="text-sm font-medium text-white truncate">{s.label}</div>
                  <div class="text-xs text-gray-400 truncate">{s.flow_id}</div>
                </div>
                <span class="text-xs text-gray-500 uppercase ml-2">
                  {(s.format || "").split(":").pop()}
                </span>
              </div>
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <!-- TAKE button -->
    <div class="flex flex-col items-center justify-center gap-4">
      <button
        class="w-32 h-32 rounded-xl bg-gradient-to-br from-green-500 to-green-600 hover:from-green-400 hover:to-green-500 text-black font-bold text-lg flex flex-col items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed transition-all shadow-lg shadow-green-500/30 border-2 border-green-400"
        on:click={onExecutePatchTake}
        disabled={!isPatchTakeReady?.()}
      >
        <svg class="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
          <polyline points="9 18 15 12 9 6" />
        </svg>
        <span class="text-sm uppercase tracking-wide">
          {nmosTakeBusy ? "PATCHING..." : "TAKE"}
        </span>
      </button>
      <div class="text-xs text-gray-400 text-center max-w-[120px]">
        {#if !isPatchTakeReady?.()}
          Select endpoints
        {:else}
          Ready to patch
        {/if}
      </div>
    </div>

    <!-- Destinations panel -->
    <div class="bg-gray-900 border border-gray-800 rounded-lg p-4 flex flex-col min-h-[500px]">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-base font-semibold text-white mb-1">Destinations</h3>
          <p class="text-xs text-gray-400">IS-04 receivers</p>
        </div>
        <select
          class="px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 focus:outline-none focus:border-orange-600"
          value={selectedReceiverNodeId}
          on:change={(e) => onSelectReceiverNode?.(e.target.value)}
        >
          <option value="">Select node…</option>
          {#each nmosNodes as node}
            <option value={node.id}>{node.name}</option>
          {/each}
        </select>
      </div>

      <div class="flex flex-wrap gap-2 mb-4">
        <input
          value={receiverFilterText}
          class="flex-1 min-w-[150px] px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 placeholder:text-gray-500 focus:outline-none focus:border-orange-600"
          placeholder="Search destinations..."
          on:input={(e) => onUpdateReceiverFilterText?.(e.target.value)}
        />
        <select
          value={receiverFormatFilter}
          class="px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 focus:outline-none focus:border-orange-600"
          on:change={(e) => onUpdateReceiverFormatFilter?.(e.target.value)}
        >
          <option value="">All Formats</option>
          <option value="video">Video</option>
          <option value="audio">Audio</option>
          <option value="data">Data</option>
          <option value="mux">Mux</option>
        </select>
      </div>

      <div class="flex-1 overflow-auto space-y-1">
        {#if receiverNodeReceivers.length === 0}
          <div class="flex flex-col items-center justify-center py-12 text-center">
            <svg class="w-12 h-12 text-gray-600 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <p class="text-sm font-medium text-gray-400 mb-1">No destinations available</p>
            <p class="text-xs text-gray-500">Select a node to load receivers</p>
          </div>
        {:else}
          {#each (filterReceivers ? filterReceivers(receiverNodeReceivers) : receiverNodeReceivers) as r}
            <button
              type="button"
              class="w-full text-left px-3 py-2 rounded bg-gray-800 hover:bg-gray-700 border transition-colors {selectedPatchReceiver && selectedPatchReceiver.id === r.id
                ? 'border-orange-600 bg-orange-900/20'
                : 'border-gray-700'}"
              on:click={() => onSelectPatchReceiver?.(r)}
            >
              <div class="flex items-center justify-between">
                <div class="min-w-0">
                  <div class="text-sm font-medium text-white truncate">{r.label}</div>
                  <div class="text-xs text-gray-400 truncate">{r.description}</div>
                </div>
                <span class="text-xs text-gray-500 uppercase ml-2">
                  {(r.format || "").split(":").pop()}
                </span>
              </div>
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </div>

  <!-- Add Node Modal -->
  {#if showAddNodeModal}
    <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-gray-900 border border-gray-800 rounded-lg p-6 w-full max-w-md space-y-4">
        <h3 class="text-lg font-semibold text-white">Add NMOS Node</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1" for="new-node-name">Node Name</label>
            <input
              value={newNodeName}
              id="new-node-name"
              class="w-full px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 placeholder:text-gray-500 focus:outline-none focus:border-orange-600"
              placeholder="e.g. Camera Router"
              on:input={(e) => onChangeNewNodeName?.(e.target.value)}
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1" for="new-node-url">IS-04 URL</label>
            <input
              value={newNodeUrl}
              id="new-node-url"
              class="w-full px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 placeholder:text-gray-500 focus:outline-none focus:border-orange-600"
              placeholder="http://192.168.x.x:port"
              on:input={(e) => onChangeNewNodeUrl?.(e.target.value)}
            />
            <p class="mt-1 text-xs text-gray-500">
              NMOS Node IS-04 URL. IS-05 endpoint is derived automatically.
            </p>
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button
            class="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium transition-colors border border-gray-700"
            on:click={onCancelAddNode}
          >
            Cancel
          </button>
          <button
            class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold transition-colors"
            on:click={onConfirmAddNode}
          >
            Add Node
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Connect RDS Modal -->
  {#if showConnectRDSModal}
    <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-gray-900 border border-gray-800 rounded-lg p-6 w-full max-w-2xl space-y-4">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-lg font-semibold text-white mb-1">Connect to NMOS Registry (RDS)</h3>
            <p class="text-sm text-gray-400">
              Enter the IS-04 Query API base URL, then discover available nodes.
            </p>
          </div>
          <button
            type="button"
            class="px-2 py-1 rounded text-gray-400 hover:text-white hover:bg-gray-800"
            on:click={() => onCloseRDS?.()}
          >
            ✕
          </button>
        </div>

        <div class="flex flex-wrap gap-2">
          <input
            value={registryQueryUrl}
            class="flex-1 min-w-[260px] px-3 py-2 rounded bg-gray-800 border border-gray-700 text-sm text-gray-200 placeholder:text-gray-500 focus:outline-none focus:border-orange-600"
            placeholder="Registry Query API URL"
            on:input={(e) => onChangeRegistryQueryUrl?.(e.target.value)}
          />
          <button
            class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            disabled={!registryQueryUrl?.trim() || registryDiscovering}
            on:click={() => onDiscoverRegistryNodes?.()}
          >
            {registryDiscovering ? "Discovering..." : "Discover Nodes"}
          </button>
        </div>

        {#if registryError}
          <div class="px-4 py-2 rounded-lg bg-red-900/50 border border-red-800 text-sm text-red-300">
            {registryError}
          </div>
        {/if}

        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-400">
            Available Nodes: <span class="font-semibold text-gray-200">{registryNodes?.length || 0}</span>
            {#if registryNodes?.length}
              <span class="mx-2 text-gray-600">•</span>
              Selected: <span class="font-semibold text-gray-200">{registrySelectedIds?.length || 0}</span>
            {/if}
          </div>
          {#if registryNodes?.length}
            <button
              type="button"
              class="px-3 py-1.5 rounded bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors border border-gray-700"
              on:click={() => onSelectAllRegistryNodes?.()}
            >
              Select All
            </button>
          {/if}
        </div>

        <div class="max-h-[320px] overflow-auto rounded border border-gray-800">
          {#if !registryNodes?.length}
            <div class="p-4 text-sm text-gray-500 text-center">No nodes discovered yet.</div>
          {:else}
            <div class="divide-y divide-gray-800">
              {#each registryNodes as n}
                <label class="flex items-start gap-3 p-3 hover:bg-gray-800 cursor-pointer">
                  <input
                    type="checkbox"
                    class="mt-1"
                    checked={registrySelectedIds?.includes?.(n.id)}
                    on:change={() => onToggleRegistryNode?.(n.id)}
                  />
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium text-white truncate">{n.label || n.id}</div>
                    <div class="text-xs text-gray-400 truncate">{n.base_url || n.href || "—"}</div>
                  </div>
                </label>
              {/each}
            </div>
          {/if}
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button
            class="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium transition-colors border border-gray-700"
            on:click={() => onCloseRDS?.()}
          >
            Cancel
          </button>
          <button
            class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            disabled={!registrySelectedIds?.length}
            on:click={() => onAddSelectedRegistryNodes?.()}
          >
            Add Selected Nodes
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
