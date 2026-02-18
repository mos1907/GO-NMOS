<script>
  export let nmosBaseUrl = "";
  export let nmosResult = null;
  export let nmosIS05Base = "";
  export let flows = [];
  export let selectedNMOSFlow = null;
  export let selectedNMOSReceiver = null;
  export let nmosTakeBusy = false;

  export let onBaseUrlChange;
  export let onDiscoverNMOS;
  export let onIS05BaseChange;
  export let onSelectFlow;
  export let onSelectReceiver;
  export let onExecuteTake;
  export let isTakeReady;
</script>

<div class="space-y-4">
  <div class="flex flex-wrap gap-3 items-end">
    <div class="flex flex-col gap-1">
      <label class="text-sm font-medium text-slate-300">NMOS Node Base URL</label>
      <input
        value={nmosBaseUrl}
        on:input={(e) => onBaseUrlChange?.(e.target.value)}
        placeholder="http://192.168.x.x:port"
        class="px-3 py-2 rounded-md bg-slate-900 border border-slate-700 text-sm min-w-[320px]"
      />
    </div>
    <button
      class="px-4 py-2 rounded-md bg-svelte hover:bg-orange-500 text-sm font-semibold text-white"
      on:click={onDiscoverNMOS}
    >
      Discover
    </button>
  </div>

  {#if nmosResult}
    <div class="grid md:grid-cols-3 gap-4">
      <div class="rounded-xl bg-slate-900/60 border border-slate-800 p-4 space-y-2">
        <p class="text-xs text-slate-400">IS-04 Version</p>
        <p class="text-lg font-semibold">{nmosResult.is04_version}</p>
        <p class="text-xs text-slate-400 break-all">Base: {nmosResult.base_url}</p>
      </div>
      <div class="rounded-xl bg-slate-900/60 border border-slate-800 p-4">
        <p class="text-xs text-slate-400 mb-1">Counts</p>
        <p class="text-sm">
          Senders: {nmosResult.counts?.senders} | Receivers: {nmosResult.counts?.receivers} | Flows:
          {nmosResult.counts?.flows}
        </p>
      </div>
      <div class="rounded-xl bg-slate-900/60 border border-slate-800 p-4 space-y-2">
        <label class="text-xs text-slate-400">IS-05 Base URL</label>
        <input
          value={nmosIS05Base}
          on:input={(e) => onIS05BaseChange?.(e.target.value)}
          class="w-full px-3 py-2 rounded-md bg-slate-950 border border-slate-700 text-xs"
        />
        <p class="text-[11px] text-slate-500">Typically: base_url + /x-nmos/connection/&lt;version&gt;</p>
      </div>
    </div>

    <div class="grid md:grid-cols-[3fr_1fr_3fr] gap-6 mt-4 items-stretch">
      <!-- Local flows (sources) -->
      <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-base font-semibold text-black">Sources (Local Flows)</h4>
          <span class="text-sm text-black/70">{flows.length} flows</span>
        </div>
        <div class="overflow-auto max-h-72 divide-y divide-svelte/20">
          {#each flows as f}
            <button
              type="button"
              class="w-full text-left px-3 py-2 text-sm hover:bg-svelte/20 flex justify-between gap-2 {selectedNMOSFlow && selectedNMOSFlow.id === f.id
                ? 'bg-svelte/30 border-l-4 border-svelte'
                : ''}"
              on:click={() => onSelectFlow?.(f)}
            >
              <span class="truncate text-black font-medium">{f.display_name}</span>
              <span class="text-[12px] text-black/60 truncate">{f.multicast_ip}:{f.port}</span>
            </button>
          {/each}
        </div>
      </div>

      <!-- TAKE button -->
      <div class="flex flex-col items-center justify-center gap-4">
        <button
          class="w-40 h-40 rounded-2xl bg-gradient-to-br from-svelte to-orange-500 text-black font-bold text-xl shadow-[0_0_50px_rgba(255,62,0,0.7)] flex flex-col items-center justify-center gap-2 disabled:opacity-30 disabled:shadow-none disabled:cursor-not-allowed hover:scale-105 active:scale-100 transition"
          on:click={onExecuteTake}
          disabled={!isTakeReady?.()}
        >
          <svg class="w-9 h-9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polyline points="9 18 15 12 9 6" />
          </svg>
          <span>{nmosTakeBusy ? "TAKING..." : "TAKE"}</span>
        </button>
        <div class="flex items-center gap-2 text-sm text-black px-4 py-2 rounded-full bg-svelte/20 border border-svelte/40">
          <span class="inline-flex h-3 w-3 rounded-full {isTakeReady?.() ? 'bg-svelte' : 'bg-black/30'}"></span>
          <span class="font-medium">
            {#if !selectedNMOSFlow && !selectedNMOSReceiver}
              Select flow and receiver
            {:else if !selectedNMOSFlow}
              Select a source flow
            {:else if !selectedNMOSReceiver}
              Select a receiver
            {:else if !nmosIS05Base}
              IS-05 base URL required
            {:else}
              Ready
            {/if}
          </span>
        </div>
      </div>

      <!-- NMOS receivers (destinations) -->
      <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-base font-semibold text-black">Destinations (NMOS Receivers)</h4>
          <span class="text-sm text-black/70">{(nmosResult.receivers || []).length} receivers</span>
        </div>
        <div class="overflow-auto max-h-72 divide-y divide-svelte/20">
          {#each nmosResult.receivers || [] as r}
            <button
              type="button"
              class="w-full text-left px-3 py-2 text-sm hover:bg-svelte/20 flex justify-between gap-2 {selectedNMOSReceiver && selectedNMOSReceiver.id === r.id
                ? 'bg-svelte/30 border-l-4 border-svelte'
                : ''}"
              on:click={() => onSelectReceiver?.(r)}
            >
              <span class="truncate text-black font-medium">{r.label}</span>
              <span class="text-[12px] text-black/60 uppercase truncate">{r.format}</span>
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>

