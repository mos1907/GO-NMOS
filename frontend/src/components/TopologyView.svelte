<script>
  let {
    registryNodes = [],
    registryDevices = [],
    registrySenders = [],
    registryReceivers = [],
    selectedRegistryNodeId = "",
    selectedRegistryDeviceId = "",
    onSelectNode,
    onSelectDevice,
    isPatchTakeReady,
    selectedPatchSender = null,
    selectedPatchReceiver = null,
    nmosIS05Base = "",
    nmosTakeBusy = false,
    onSelectPatchSender,
    onSelectPatchReceiver,
    onExecutePatchTake,
  } = $props();

  // Ensure arrays are never null (reactive)
  let safeRegistryNodes = $derived(registryNodes || []);
  let safeRegistryDevices = $derived(registryDevices || []);
  let safeRegistrySenders = $derived(registrySenders || []);
  let safeRegistryReceivers = $derived(registryReceivers || []);
</script>

<section class="mt-4 space-y-4">
  <header class="flex flex-wrap items-center justify-between gap-3">
    <div>
      <h3 class="text-sm font-semibold text-slate-50">System Topology</h3>
      <p class="text-[11px] text-slate-400">
        Visualize NMOS nodes, devices, senders and receivers in a router-friendly view.
      </p>
    </div>
    <div class="text-[11px] text-slate-400 space-y-0.5 text-right">
      <p>
        Registry: Nodes {safeRegistryNodes.length} · Devices {safeRegistryDevices.length}
      </p>
      <p>
        Endpoints: Senders {safeRegistrySenders.length} · Receivers {safeRegistryReceivers.length}
      </p>
    </div>
  </header>

  {#if safeRegistryNodes.length === 0 && safeRegistrySenders.length === 0 && safeRegistryReceivers.length === 0}
    <div class="rounded-xl border border-gray-800 bg-gray-900 p-8">
      <div class="flex flex-col items-center justify-center text-center">
        <svg class="w-16 h-16 text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" />
        </svg>
        <h3 class="text-lg font-semibold text-gray-200 mb-2">Registry is empty</h3>
        <p class="text-sm text-gray-400 max-w-md">
          Run an NMOS discovery from the <span class="font-semibold text-gray-300">NMOS</span> tab to ingest nodes and endpoints into the internal registry.
        </p>
      </div>
    </div>
  {:else}
    <div class="grid md:grid-cols-[2fr_3fr_2fr] gap-4 items-start">
      <!-- Nodes & Devices column -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 space-y-3">
        <div class="flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Nodes</h4>
          <span class="text-[11px] text-slate-400">{safeRegistryNodes.length} nodes</span>
        </div>
        <div class="space-y-1 max-h-40 overflow-auto pr-1">
          {#each safeRegistryNodes as node}
            <button
              type="button"
              class="w-full text-left px-3 py-1.5 rounded-lg border text-[11px] {selectedRegistryNodeId === node.id
                ? 'border-svelte bg-slate-900 text-slate-50'
                : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-600'}"
              on:click={() => onSelectNode?.(node.id)}
            >
              <span class="font-medium truncate">{node.label || node.id}</span>
              <span class="block text-[10px] text-slate-400 truncate">{node.hostname}</span>
            </button>
          {/each}
        </div>

        <div class="mt-4 flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Devices</h4>
          <span class="text-[11px] text-slate-400">
            {#if selectedRegistryNodeId}
              {safeRegistryDevices.filter((d) => d.node_id === selectedRegistryNodeId).length} of {safeRegistryDevices.length}
            {:else}
              {safeRegistryDevices.length} devices
            {/if}
          </span>
        </div>
        <div class="space-y-1 max-h-48 overflow-auto pr-1">
          {#each safeRegistryDevices.filter((d) => !selectedRegistryNodeId || d.node_id === selectedRegistryNodeId) as dev}
            <button
              type="button"
              class="w-full text-left px-3 py-1.5 rounded-lg border text-[11px] {selectedRegistryDeviceId === dev.id
                ? 'border-svelte bg-slate-900 text-slate-50'
                : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-600'}"
              on:click={() => onSelectDevice?.(dev.id)}
            >
              <span class="font-medium truncate">{dev.label || dev.id}</span>
              <span class="block text-[10px] text-slate-400 truncate">{dev.type}</span>
            </button>
          {/each}
        </div>
      </div>

      <!-- Center "patch map" / senders -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/80 p-4 space-y-4">
        <div class="flex items-center justify-between gap-2">
          <div>
            <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Sources (Senders)</h4>
            <p class="text-[11px] text-slate-400">
              Choose a sender from the selected device to patch.
            </p>
          </div>
          <div class="flex items-center gap-2 text-[11px] text-slate-400">
            <span class="inline-flex h-2 w-2 rounded-full {isPatchTakeReady?.() ? 'bg-emerald-400' : 'bg-slate-600'}"></span>
            <span>{isPatchTakeReady?.() ? "Ready to patch" : "Select source & destination"}</span>
          </div>
        </div>

        <div class="space-y-1 max-h-56 overflow-auto pr-1 text-[11px]">
          {#if safeRegistrySenders.length === 0}
            <p class="text-slate-500 italic">No senders in registry.</p>
          {:else}
            {#each safeRegistrySenders.filter((s) => !selectedRegistryDeviceId || s.device_id === selectedRegistryDeviceId) as s}
              <button
                type="button"
                class="w-full text-left px-3 py-2 rounded-lg border border-slate-800 bg-slate-900/60 hover:border-svelte/70 hover:bg-slate-900 flex flex-col gap-0.5"
                on:click={() => onSelectPatchSender?.(s)}
              >
                <span class="text-[13px] font-medium text-slate-50 truncate">{s.label}</span>
                <span class="text-[11px] text-slate-400 truncate">{s.flow_id}</span>
                <span class="text-[10px] text-slate-500 truncate uppercase">{s.transport}</span>
              </button>
            {/each}
          {/if}
        </div>

        <div class="border-t border-slate-800 pt-3 text-[11px] text-slate-400 space-y-1">
          <p class="font-semibold text-slate-200">Selected patch</p>
          <p class="truncate">
            Source:
            {#if selectedPatchSender}
              {selectedPatchSender.label}
            {:else}
              none
            {/if}
          </p>
          <p class="truncate">
            Destination:
            {#if selectedPatchReceiver}
              {selectedPatchReceiver.label}
            {:else}
              none
            {/if}
          </p>
          <p class="truncate">IS-05 base: {nmosIS05Base || "not set"}</p>
        </div>
      </div>

      <!-- Destinations column (receivers) -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Destinations (Receivers)</h4>
            <p class="text-[11px] text-slate-400">Choose a receiver to patch to.</p>
          </div>
          <span class="text-[11px] text-slate-400">{safeRegistryReceivers.length} receivers</span>
        </div>
        <div class="space-y-1 max-h-72 overflow-auto pr-1 text-[11px]">
          {#if safeRegistryReceivers.length === 0}
            <p class="text-slate-500 italic">No receivers in registry.</p>
          {:else}
            {#each safeRegistryReceivers.filter((r) => !selectedRegistryDeviceId || r.device_id === selectedRegistryDeviceId) as r}
              <button
                type="button"
                class="w-full text-left px-3 py-2 rounded-lg border border-slate-800 bg-slate-900/60 hover:border-svelte/70 hover:bg-slate-900 flex flex-col gap-0.5"
                on:click={() => onSelectPatchReceiver?.(r)}
              >
                <span class="text-[13px] font-medium text-slate-50 truncate">{r.label}</span>
                <span class="text-[11px] text-slate-400 truncate">{r.description}</span>
                <span class="text-[10px] text-slate-500 truncate uppercase">{r.format} · {r.transport}</span>
              </button>
            {/each}
          {/if}
        </div>

        <div class="pt-2 flex justify-end">
          <button
            class="px-3 py-2 rounded-lg bg-svelte text-slate-950 text-xs font-semibold disabled:opacity-40"
            disabled={!isPatchTakeReady?.()}
            on:click={onExecutePatchTake}
          >
            {nmosTakeBusy ? "TAKING..." : "TAKE PATCH"}
          </button>
        </div>
      </div>
    </div>
  {/if}
</section>

