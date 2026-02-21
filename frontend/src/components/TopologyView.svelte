<script>
  /**
   * System Topology View – read-only node → device → sender/receiver hierarchy,
   * grouped by site/room. Use Registry & Patch for patch operations.
   */
  let {
    registryNodes = [],
    registryDevices = [],
    registrySenders = [],
    registryReceivers = [],
    token = ""
  } = $props();

  let safeNodes = $derived(registryNodes || []);
  let safeDevices = $derived(registryDevices || []);
  let safeSenders = $derived(registrySenders || []);
  let safeReceivers = $derived(registryReceivers || []);

  let groupBy = $state("site"); // "site" | "room" | "none"

  const groups = $derived.by(() => {
    const list = safeNodes;
    if (groupBy === "none") return [{ key: "", label: "All nodes", nodes: list }];
    const map = new Map();
    for (const n of list) {
      const site = n.tags?.site?.[0] || "";
      const room = n.tags?.room?.[0] || "";
      const key = groupBy === "site" ? (site || "_ungrouped") : groupBy === "room" ? (room || "_ungrouped") : "";
      const label = key === "_ungrouped" ? "Ungrouped" : key || "—";
      if (!map.has(key)) map.set(key, { key, label, nodes: [] });
      map.get(key).nodes.push(n);
    }
    return [...map.values()].sort((a, b) =>
      a.label === "Ungrouped" ? 1 : b.label === "Ungrouped" ? -1 : a.label.localeCompare(b.label)
    );
  });

  function devicesForNode(nodeId) {
    return safeDevices.filter((d) => d.node_id === nodeId);
  }

  function sendersForDevice(deviceId) {
    return safeSenders.filter((s) => s.device_id === deviceId);
  }

  function receiversForDevice(deviceId) {
    return safeReceivers.filter((r) => r.device_id === deviceId);
  }

  function deviceTypeShort(type) {
    if (!type) return "—";
    return type.split(":").pop() || type;
  }
</script>

<section class="mt-4 space-y-4">
  <header class="flex flex-wrap items-center justify-between gap-3">
    <div>
      <h3 class="text-sm font-semibold text-slate-50">System topology</h3>
      <p class="text-[11px] text-slate-400">
        Nodes, devices and endpoints by site/room. Read-only; use Registry & Patch to patch.
      </p>
    </div>
    <div class="flex items-center gap-2">
      <label for="topology-group-by" class="text-[11px] text-slate-400">Group by</label>
      <select
        id="topology-group-by"
        class="px-2 py-1 rounded-md bg-slate-900 border border-slate-700 text-[11px] text-slate-200"
        bind:value={groupBy}
      >
        <option value="none">None</option>
        <option value="site">Site</option>
        <option value="room">Room</option>
      </select>
    </div>
  </header>

  {#if safeNodes.length === 0 && safeDevices.length === 0}
    <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-8 text-center">
      <p class="text-slate-400">Registry empty. Run discovery/sync from Registry & Patch or NMOS tab.</p>
    </div>
  {:else}
    <div class="space-y-6">
      {#each groups as { key, label, nodes }}
        <div class="rounded-xl border border-slate-800 bg-slate-950/50 overflow-hidden">
          {#if groupBy !== "none"}
            <div class="px-4 py-2 bg-slate-800/80 border-b border-slate-700">
              <span class="text-xs font-semibold text-slate-300 uppercase tracking-wide">{label}</span>
              <span class="ml-2 text-[11px] text-slate-500">{nodes.length} node(s)</span>
            </div>
          {/if}
          <div class="p-4 flex flex-wrap gap-4">
            {#each nodes as node}
              {@const devs = devicesForNode(node.id)}
              <div class="rounded-lg border border-slate-700 bg-slate-900/80 p-3 min-w-[200px] max-w-[280px]">
                <div class="flex items-center justify-between gap-2 mb-2">
                  <span class="text-sm font-medium text-slate-100 truncate" title={node.label || node.id}>
                    {node.label || node.id}
                  </span>
                  <span class="text-[10px] text-slate-500 shrink-0">node</span>
                </div>
                {#if node.hostname}
                  <p class="text-[10px] text-slate-500 truncate mb-2">{node.hostname}</p>
                {/if}
                <div class="space-y-2">
                  {#each devs as dev}
                    {@const numSenders = sendersForDevice(dev.id).length}
                    {@const numReceivers = receiversForDevice(dev.id).length}
                    <div class="rounded border border-slate-700/80 bg-slate-800/50 p-2">
                      <div class="text-xs font-medium text-slate-200 truncate" title={dev.label || dev.id}>
                        {dev.label || dev.id}
                      </div>
                      <div class="text-[10px] text-slate-500 mt-0.5">{deviceTypeShort(dev.type)}</div>
                      <div class="flex gap-2 mt-1.5">
                        <span class="px-1.5 py-0.5 rounded bg-emerald-900/50 text-emerald-200 text-[10px]">
                          {numSenders} sender
                        </span>
                        <span class="px-1.5 py-0.5 rounded bg-violet-900/50 text-violet-200 text-[10px]">
                          {numReceivers} receiver
                        </span>
                      </div>
                    </div>
                  {/each}
                  {#if devs.length === 0}
                    <p class="text-[11px] text-slate-500 italic">No devices</p>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <div class="rounded-lg border border-slate-800 bg-slate-900/50 px-4 py-3 flex flex-wrap items-center gap-4 text-[11px] text-slate-400">
      <span>Nodes: <strong class="text-slate-200">{safeNodes.length}</strong></span>
      <span>Devices: <strong class="text-slate-200">{safeDevices.length}</strong></span>
      <span>Senders: <strong class="text-slate-200">{safeSenders.length}</strong></span>
      <span>Receivers: <strong class="text-slate-200">{safeReceivers.length}</strong></span>
    </div>
  {/if}
</section>
