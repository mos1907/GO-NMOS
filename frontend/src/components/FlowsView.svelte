<script>
  import EmptyState from "./EmptyState.svelte";
  import NewFlowView from "./NewFlowView.svelte";
  import { IconPlus } from "../lib/icons.js";

  let {
    flows = [],
    flowLimit = 50,
    flowOffset = 0,
    flowTotal = 0,
    flowSortBy = "updated_at",
    flowSortOrder = "desc",
    canEdit = false,
    isAdmin = false,
    onApplyFlowSort,
    onPrevFlowPage,
    onNextFlowPage,
    onToggleFlowLock,
    onDeleteFlow,
    onCreateFlow = null,
    onEditFlow = null,
    onUpdateFlow = null,
    onCheckFlow = null,
    onFetchSDP = null,
    onSyncFromNMOS = null,
    onCheckIS05Receiver = null,
    newFlow = null,
    editingFlow = null,
  } = $props();

  let showNewFlowModal = $state(false);
  let showDetailModal = $state(false);
  let detailFlow = $state(null);
  let showCheckModal = $state(false);
  let checkFlow = $state(null);
  let checkResult = $state(null);
  let checking = $state(false);
  let checkBaseUrl = $state("");
  
  // Lock/Unlock info modal
  let showLockInfoModal = $state(false);
  let lockInfoFlow = $state(null);
  let lockInfoMessage = $state("");

  // Fetch SDP
  let fetchSdpManifestUrl = $state("");
  let fetchSdpLoading = $state(false);
  let fetchSdpError = $state("");

  // NMOS -> DB sync (pull)
  let nmosSyncIs04 = $state("");
  let nmosSyncIs05 = $state("");
  let nmosSyncLoading = $state(false);
  let nmosSyncError = $state("");

  // IS-05 receiver state check
  let is05CheckBaseUrl = $state("");
  let is05CheckReceiverId = $state("");
  let is05CheckLoading = $state(false);
  let is05CheckError = $state("");
  let is05CheckResult = $state(null);

  $effect(() => {
    // Pre-fill from flow record when detail modal opens
    if (!detailFlow) return;
    if (!nmosSyncIs04 && detailFlow.nmos_is04_base_url) nmosSyncIs04 = detailFlow.nmos_is04_base_url;
    if (!nmosSyncIs05 && detailFlow.nmos_is05_base_url) nmosSyncIs05 = detailFlow.nmos_is05_base_url;
    if (!is05CheckBaseUrl && detailFlow.nmos_is05_base_url) is05CheckBaseUrl = detailFlow.nmos_is05_base_url;
  });

  // Derived transport/format helpers for nicer IS-05 / ST 2110 badges
  const formatShortMap = {
    "urn:x-nmos:format:video": "ST 2110-20 (video)",
    "urn:x-nmos:format:audio": "ST 2110-30 (audio)",
    "urn:x-nmos:format:data": "ST 2110-40 (data)",
  };

  // Filters
  let statusFilter = $state("");
  let availabilityFilter = $state("");
  let dataSourceFilter = $state("");

  function handleOpenModal() {
    showNewFlowModal = true;
  }

  function handleCloseModal() {
    showNewFlowModal = false;
  }

  function handleCreateFlow() {
    onCreateFlow?.();
    handleCloseModal();
  }

  function handleEditFlow(flow) {
    if (!flow) return;
    // Call parent's openEditFlowModal to update newFlow and editingFlow state
    onEditFlow?.(flow);
    // Open modal after a short delay to ensure state is updated
    setTimeout(() => {
      showNewFlowModal = true;
    }, 100);
  }
  
  // Watch for editingFlow changes to auto-open modal when editing starts
  let lastEditingFlowId = $state(null);
  $effect(() => {
    if (editingFlow && editingFlow.id !== lastEditingFlowId && newFlow && newFlow.display_name) {
      // Auto-open modal when editingFlow is set and newFlow is populated
      lastEditingFlowId = editingFlow.id;
      showNewFlowModal = true;
    }
  });

  function handleUpdateFlow() {
    // This will be called from NewFlowView
    // The actual update is handled in DashboardPage
  }

  function openDetailModal(flow) {
    detailFlow = flow;
    showDetailModal = true;
  }

  function closeDetailModal() {
    detailFlow = null;
    showDetailModal = false;
  }

  function openCheckModal(flow) {
    checkFlow = flow;
    checkBaseUrl = "";
    checkResult = null;
    showCheckModal = true;
  }

  function closeCheckModal() {
    checkFlow = null;
    checkBaseUrl = "";
    checkResult = null;
    showCheckModal = false;
  }

  async function handleCheckFlow() {
    if (!checkFlow || !checkBaseUrl.trim()) return;
    checking = true;
    checkResult = null;
    try {
      const result = await onCheckFlow?.(checkFlow.id, checkBaseUrl.trim());
      checkResult = result;
    } catch (e) {
      checkResult = { error: e.message };
    } finally {
      checking = false;
    }
  }

  // Filter flows
  let filteredFlows = $derived.by(() => {
    let filtered = flows;
    if (statusFilter) {
      filtered = filtered.filter((f) => f.flow_status === statusFilter);
    }
    if (availabilityFilter) {
      filtered = filtered.filter((f) => f.availability === availabilityFilter);
    }
    if (dataSourceFilter) {
      filtered = filtered.filter((f) => (f.data_source || "manual") === dataSourceFilter);
    }
    return filtered;
  });
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Flows</h3>
      <p class="text-[11px] text-gray-400">Filter and manage flow list</p>
    </div>
    <div class="flex flex-wrap items-center gap-2 text-xs">
      {#if canEdit}
        <button
          onclick={handleOpenModal}
          class="px-3 py-1.5 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-xs font-semibold transition-colors flex items-center gap-1.5"
        >
          {@html IconPlus}
          Add Flow
        </button>
      {/if}
      <label class="flex items-center gap-1 text-gray-300">
        <span>Sort by</span>
        <select
          bind:value={flowSortBy}
          onchange={onApplyFlowSort}
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
          onchange={onApplyFlowSort}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
        >
          <option value="desc">desc</option>
          <option value="asc">asc</option>
        </select>
      </label>
      <label class="flex items-center gap-1 text-gray-300">
        <span>Status</span>
        <select
          bind:value={statusFilter}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
        >
          <option value="">All</option>
          <option value="active">active</option>
          <option value="unused">unused</option>
          <option value="maintenance">maintenance</option>
        </select>
      </label>
      <label class="flex items-center gap-1 text-gray-300">
        <span>Availability</span>
        <select
          bind:value={availabilityFilter}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
        >
          <option value="">All</option>
          <option value="available">available</option>
          <option value="lost">lost</option>
          <option value="maintenance">maintenance</option>
        </select>
      </label>
      <label class="flex items-center gap-1 text-gray-300">
        <span>Source</span>
        <select
          bind:value={dataSourceFilter}
          class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
        >
          <option value="">All</option>
          <option value="manual">manual</option>
          <option value="nmos">nmos</option>
          <option value="rds">rds</option>
        </select>
      </label>
      <div class="flex items-center gap-1">
        <button
          class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40 transition-colors"
          onclick={onPrevFlowPage}
          disabled={flowOffset === 0}
        >
          Prev
        </button>
        <button
          class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-900 text-xs text-gray-200 hover:bg-gray-800 disabled:opacity-40 transition-colors"
          onclick={onNextFlowPage}
          disabled={flowOffset + flowLimit >= flowTotal}
        >
          Next
        </button>
      </div>
      <span class="text-[11px] text-gray-400">
        Showing {flowOffset + 1}-{Math.min(flowOffset + flowLimit, flowTotal)} / {flowTotal}
        {#if statusFilter || availabilityFilter || dataSourceFilter}
          <span class="text-orange-400">({filteredFlows?.length || 0} filtered)</span>
        {/if}
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
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Transport</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Updated</th>
          <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Locked</th>
          {#if canEdit}
            <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200 min-w-[200px]">Action</th>
          {/if}
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#if flows.length === 0}
          <tr>
            <td colspan={canEdit ? 11 : 10} class="px-6 py-12">
              <EmptyState
                title="No flows found"
                message="Get started by creating your first flow or importing flows from a backup."
                actionLabel={canEdit ? "Create Flow" : ""}
                onAction={canEdit ? onCreateFlow : null}
                icon={IconPlus}
              />
            </td>
          </tr>
        {:else if (statusFilter || availabilityFilter || dataSourceFilter) && filteredFlows.length === 0}
          <tr>
            <td colspan={canEdit ? 11 : 10} class="px-6 py-12 text-center text-gray-400 text-sm">
              No flows match the selected filters
            </td>
          </tr>
        {:else}
          {#each (statusFilter || availabilityFilter || dataSourceFilter ? filteredFlows : flows) as flow}
            <tr class="hover:bg-gray-800/70 transition-colors cursor-pointer" onclick={() => openDetailModal(flow)}>
              <td class="px-3 py-2 text-[13px] font-medium text-gray-100">{flow.display_name}</td>
              <td class="px-3 py-2 text-gray-300">{flow.flow_id}</td>
              <td class="px-3 py-2 text-gray-300">{flow.multicast_ip}</td>
              <td class="px-3 py-2 text-gray-300">{flow.source_ip}</td>
              <td class="px-3 py-2 text-gray-300">{flow.port}</td>
              <td class="px-3 py-2">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium {flow.flow_status === 'active'
                    ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                    : flow.flow_status === 'maintenance'
                      ? 'bg-amber-900 text-amber-200 border border-amber-700'
                      : 'bg-slate-800 text-slate-200 border border-slate-700'}"
                >
                  {flow.flow_status}
                </span>
              </td>
              <td class="px-3 py-2">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium {flow.availability === 'available'
                    ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                    : flow.availability === 'lost'
                      ? 'bg-red-900 text-red-200 border border-red-700'
                      : 'bg-amber-900 text-amber-200 border border-amber-700'}"
                >
                  {flow.availability}
                </span>
              </td>
              <td class="px-3 py-2 text-gray-300 text-[11px]">{flow.transport_protocol || "-"}</td>
              <td class="px-3 py-2 text-gray-400 text-[11px]">
                {flow.updated_at ? new Date(flow.updated_at).toLocaleDateString() : "-"}
              </td>
              <td class="px-3 py-2">
                {#if flow.locked}
                  <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                {:else}
                  <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                  </svg>
                {/if}
              </td>
              {#if canEdit}
                <td class="px-3 py-2" onclick={(e) => e.stopPropagation()}>
                  <div class="flex flex-nowrap gap-1 items-center">
                    <button
                      class="px-2 py-1 rounded-md border border-blue-700 bg-blue-900/60 text-[10px] text-blue-200 hover:bg-blue-900 transition-colors whitespace-nowrap shrink-0"
                      onclick={() => openCheckModal(flow)}
                      title="Check NMOS"
                    >
                      Check
                    </button>
                    <button
                      class="px-2 py-1 rounded-md border border-gray-700 bg-gray-900 text-[10px] text-gray-200 hover:bg-gray-800 transition-colors whitespace-nowrap shrink-0"
                      onclick={async () => {
                        try {
                          const result = await onToggleFlowLock?.(flow);
                          if (result && result.flow) {
                            lockInfoFlow = result.flow;
                            lockInfoMessage = result.locked 
                              ? `Flow "${result.flow.display_name}" has been locked successfully.`
                              : `Flow "${result.flow.display_name}" has been unlocked successfully.`;
                            showLockInfoModal = true;
                          }
                        } catch (e) {
                          console.error("Error toggling flow lock:", e);
                        }
                      }}
                      title={flow.locked ? "Unlock flow" : "Lock flow"}
                    >
                      {flow.locked ? "Unlock" : "Lock"}
                    </button>
                    {#if isAdmin}
                      <button
                        class="px-2 py-1 rounded-md border border-red-800 bg-red-900/60 text-[10px] text-red-200 hover:bg-red-900 transition-colors whitespace-nowrap shrink-0"
                        onclick={() => onDeleteFlow?.(flow)}
                        title="Delete flow"
                      >
                        Delete
                      </button>
                    {/if}
                  </div>
                </td>
              {/if}
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<!-- New Flow / Edit Flow Modal -->
{#if newFlow}
  <NewFlowView
    {newFlow}
    editingFlow={editingFlow}
    isOpen={showNewFlowModal}
    onCreateFlow={handleCreateFlow}
    onUpdateFlow={onUpdateFlow}
    onClose={handleCloseModal}
  />
{/if}

<!-- Flow Detail Modal -->
{#if showDetailModal && detailFlow}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in"
    onclick={closeDetailModal}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 class="text-xl font-bold text-gray-100">Flow Details: {detailFlow.display_name}</h2>
        <button
          onclick={closeDetailModal}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-6 space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-xs text-gray-400 mb-1">Display Name</p>
            <p class="text-sm font-medium text-gray-100">{detailFlow.display_name}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Flow ID</p>
            <p class="text-sm text-gray-300 font-mono">{detailFlow.flow_id}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Multicast IP</p>
            <p class="text-sm text-gray-300">{detailFlow.multicast_ip}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Source IP</p>
            <p class="text-sm text-gray-300">{detailFlow.source_ip}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Port</p>
            <p class="text-sm text-gray-300">{detailFlow.port}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Transport Protocol</p>
            <p class="text-sm text-gray-300">{detailFlow.transport_protocol || "-"}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Status</p>
            <span
              class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {detailFlow.flow_status === 'active'
                ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                : detailFlow.flow_status === 'maintenance'
                  ? 'bg-amber-900 text-amber-200 border border-amber-700'
                  : 'bg-slate-800 text-slate-200 border border-slate-700'}"
            >
              {detailFlow.flow_status}
            </span>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Availability</p>
            <span
              class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {detailFlow.availability === 'available'
                ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                : detailFlow.availability === 'lost'
                  ? 'bg-red-900 text-red-200 border border-red-700'
                  : 'bg-amber-900 text-amber-200 border border-amber-700'}"
            >
              {detailFlow.availability}
            </span>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Locked</p>
            <p class="text-sm text-gray-300">{detailFlow.locked ? "Yes" : "No"}</p>
          </div>
          <div>
            <p class="text-xs text-gray-400 mb-1">Updated At</p>
            <p class="text-sm text-gray-300">{detailFlow.updated_at ? new Date(detailFlow.updated_at).toLocaleString() : "-"}</p>
          </div>
        </div>

        {#if detailFlow.note}
          <div>
            <p class="text-xs text-gray-400 mb-1">Note</p>
            <p class="text-sm text-gray-300 bg-gray-800 p-3 rounded-md">{detailFlow.note}</p>
          </div>
        {/if}

        {#if detailFlow.source_addr_a || detailFlow.multicast_addr_a || detailFlow.group_port_a || detailFlow.source_addr_b || detailFlow.multicast_addr_b || detailFlow.group_port_b || detailFlow.nmos_node_id || detailFlow.nmos_flow_id || detailFlow.nmos_sender_id || detailFlow.nmos_device_id || detailFlow.media_type || detailFlow.redundancy_group || detailFlow.data_source}
          <div class="border-t border-gray-800 pt-4 mt-4">
            <p class="text-xs text-gray-400 mb-2 font-medium">ST 2110 / ST 2022-7 / NMOS (Advanced)</p>
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2">
                <p class="text-xs text-gray-500 mb-1">Data Source</p>
                <p class="text-sm text-gray-300">{detailFlow.data_source || "manual"}</p>
              </div>

              <div>
                <p class="text-xs text-gray-500 mb-1">Path A (Source)</p>
                <p class="text-sm text-gray-300 font-mono">{detailFlow.source_addr_a || "-"}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 mb-1">Path A (Multicast / Port)</p>
                <p class="text-sm text-gray-300 font-mono">
                  {detailFlow.multicast_addr_a || "-"}:{detailFlow.group_port_a || "-"}
                </p>
              </div>

              <div>
                <p class="text-xs text-gray-500 mb-1">Path B (Source)</p>
                <p class="text-sm text-gray-300 font-mono">{detailFlow.source_addr_b || "-"}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 mb-1">Path B (Multicast / Port)</p>
                <p class="text-sm text-gray-300 font-mono">
                  {detailFlow.multicast_addr_b || "-"}:{detailFlow.group_port_b || "-"}
                </p>
              </div>

              <div>
                <p class="text-xs text-gray-500 mb-1">Media Type</p>
                <p class="text-sm text-gray-300">{detailFlow.media_type || "-"}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500 mb-1">ST 2110 Format</p>
                <p class="text-sm text-gray-300">
                  {#if detailFlow.st2110_format}
                    {formatShortMap[detailFlow.st2110_format] || detailFlow.st2110_format}
                  {:else}
                    -
                  {/if}
                </p>
              </div>

              <div class="col-span-2 flex flex-wrap items-center gap-2 mt-1">
                <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-slate-900 text-slate-100 border border-slate-700">
                  Transport: {detailFlow.transport_protocol || "RTP/UDP"}
                </span>
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium border
                    {detailFlow.source_addr_b || detailFlow.multicast_addr_b
                      ? 'bg-emerald-950 text-emerald-200 border-emerald-700'
                      : 'bg-slate-900 text-slate-200 border-slate-700'}"
                >
                  {detailFlow.source_addr_b || detailFlow.multicast_addr_b ? "2022-7 A/B" : "Single-path"}
                </span>
                {#if detailFlow.redundancy_group}
                  <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-sky-950 text-sky-100 border border-sky-700">
                    Redundancy group
                  </span>
                {/if}
              </div>

              {#if detailFlow.nmos_node_id || detailFlow.nmos_flow_id || detailFlow.nmos_sender_id || detailFlow.nmos_device_id}
                <div class="col-span-2">
                  <p class="text-xs text-gray-500 mb-1">NMOS IDs</p>
                  <div class="text-[11px] text-gray-300 font-mono space-y-1">
                    {#if detailFlow.nmos_node_id}<div>node: {detailFlow.nmos_node_id}</div>{/if}
                    {#if detailFlow.nmos_device_id}<div>device: {detailFlow.nmos_device_id}</div>{/if}
                    {#if detailFlow.nmos_sender_id}<div>sender: {detailFlow.nmos_sender_id}</div>{/if}
                    {#if detailFlow.nmos_flow_id}<div>flow: {detailFlow.nmos_flow_id}</div>{/if}
                  </div>
                </div>
              {/if}

              {#if canEdit && onSyncFromNMOS}
                <div class="col-span-2 border-t border-gray-800 pt-3">
                  <p class="text-xs text-gray-500 mb-2">Sync from NMOS (Pull → DB)</p>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">IS-04 Base URL</label>
                      <input
                        type="text"
                        bind:value={nmosSyncIs04}
                        placeholder="http://node-ip:port"
                        class="w-full px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                      />
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">IS-05 Base URL (optional)</label>
                      <input
                        type="text"
                        bind:value={nmosSyncIs05}
                        placeholder="http://node-ip:port"
                        class="w-full px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                      />
                    </div>
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <button
                      onclick={async () => {
                        if (!detailFlow?.id || !nmosSyncIs04.trim()) return;
                        nmosSyncError = "";
                        nmosSyncLoading = true;
                        try {
                          const result = await onSyncFromNMOS(detailFlow.id, nmosSyncIs04.trim(), nmosSyncIs05.trim());
                          if (result?.flow) detailFlow = result.flow;
                        } catch (e) {
                          nmosSyncError = e.message || "Sync failed";
                        } finally {
                          nmosSyncLoading = false;
                        }
                      }}
                      disabled={nmosSyncLoading || !nmosSyncIs04.trim()}
                      class="px-4 py-2 rounded-md bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-medium"
                    >
                      {nmosSyncLoading ? "Syncing..." : "Sync from NMOS"}
                    </button>
                    <p class="text-xs text-gray-500">This will update the flow record from NMOS/SDP/IS-05.</p>
                  </div>
                  {#if nmosSyncError}
                    <p class="text-xs text-red-400 mt-1">{nmosSyncError}</p>
                  {/if}
                </div>
              {/if}

              {#if onCheckIS05Receiver}
                <div class="col-span-2 border-t border-gray-800 pt-3">
                  <p class="text-xs text-gray-500 mb-2">IS-05 Receiver State Check</p>
                  <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">IS-05 Base URL</label>
                      <input
                        type="text"
                        bind:value={is05CheckBaseUrl}
                        placeholder="http://node-ip:port"
                        class="w-full px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                      />
                    </div>
                    <div>
                      <label class="block text-xs text-gray-500 mb-1">Receiver ID</label>
                      <input
                        type="text"
                        bind:value={is05CheckReceiverId}
                        placeholder="receiver UUID"
                        class="w-full px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                      />
                    </div>
                  </div>

                  <div class="mt-2 flex items-center gap-2">
                    <button
                      onclick={async () => {
                        if (!detailFlow?.id || !is05CheckBaseUrl.trim() || !is05CheckReceiverId.trim()) return;
                        is05CheckError = "";
                        is05CheckResult = null;
                        is05CheckLoading = true;
                        try {
                          const res = await onCheckIS05Receiver(detailFlow.id, is05CheckBaseUrl.trim(), is05CheckReceiverId.trim());
                          is05CheckResult = res;
                        } catch (e) {
                          is05CheckError = e.message || "Check failed";
                        } finally {
                          is05CheckLoading = false;
                        }
                      }}
                      disabled={is05CheckLoading || !is05CheckBaseUrl.trim() || !is05CheckReceiverId.trim()}
                      class="px-4 py-2 rounded-md bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white text-sm font-medium"
                    >
                      {is05CheckLoading ? "Checking..." : "Check Receiver State"}
                    </button>
                    <p class="text-xs text-gray-500">Reads active/staged transport_params and compares to this flow.</p>
                  </div>

                  {#if is05CheckError}
                    <p class="text-xs text-red-400 mt-1">{is05CheckError}</p>
                  {/if}

                  {#if is05CheckResult}
                    <div class="mt-3 grid grid-cols-1 md:grid-cols-2 gap-2">
                      <div class="rounded-lg border border-gray-800 bg-gray-950 p-3">
                        <p class="text-xs text-gray-500 mb-1">Active</p>
                        <p class="text-[11px] text-gray-300">
                          A match: <span class={is05CheckResult.active?.matches_a ? "text-emerald-300" : "text-amber-300"}>{is05CheckResult.active?.matches_a ? "YES" : "NO"}</span>
                        </p>
                        <p class="text-[11px] text-gray-300">
                          B match: <span class={is05CheckResult.active?.matches_b ? "text-emerald-300" : "text-amber-300"}>{is05CheckResult.active?.matches_b ? "YES" : "NO"}</span>
                        </p>
                      </div>
                      <div class="rounded-lg border border-gray-800 bg-gray-950 p-3">
                        <p class="text-xs text-gray-500 mb-1">Staged</p>
                        <p class="text-[11px] text-gray-300">
                          A match: <span class={is05CheckResult.staged?.matches_a ? "text-emerald-300" : "text-amber-300"}>{is05CheckResult.staged?.matches_a ? "YES" : "NO"}</span>
                        </p>
                        <p class="text-[11px] text-gray-300">
                          B match: <span class={is05CheckResult.staged?.matches_b ? "text-emerald-300" : "text-amber-300"}>{is05CheckResult.staged?.matches_b ? "YES" : "NO"}</span>
                        </p>
                      </div>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/if}

        {#if detailFlow.sdp_url || detailFlow.sdp_cache || (canEdit && onFetchSDP)}
          <div class="border-t border-gray-800 pt-4 mt-4">
            <p class="text-xs text-gray-400 mb-2 font-medium">SDP (Session Description Protocol)</p>
            {#if detailFlow.sdp_url}
              <div class="mb-2">
                <p class="text-xs text-gray-500 mb-1">Manifest URL</p>
                <a
                  href={detailFlow.sdp_url}
                  target="_blank"
                  rel="noreferrer"
                  class="text-sm text-orange-400 hover:text-orange-300 break-all"
                >
                  {detailFlow.sdp_url}
                </a>
              </div>
            {/if}
            {#if detailFlow.sdp_cache}
              <details class="bg-gray-950 rounded-lg border border-gray-800 mb-3">
                <summary class="px-4 py-2 cursor-pointer text-sm text-gray-300 hover:text-gray-100">
                  View SDP Content
                </summary>
                <pre class="p-4 text-xs font-mono text-green-200 overflow-auto max-h-48 whitespace-pre-wrap break-words">{detailFlow.sdp_cache}</pre>
              </details>
            {/if}
            {#if canEdit && onFetchSDP}
              <div class="flex gap-2 items-end">
                <div class="flex-1">
                  <label class="block text-xs text-gray-500 mb-1">Fetch SDP from manifest URL</label>
                  <input
                    type="url"
                    bind:value={fetchSdpManifestUrl}
                    placeholder="http://node-ip:port/sdp/..."
                    class="w-full px-3 py-2 bg-gray-950 border border-gray-700 rounded-md text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                  />
                </div>
                <button
                  onclick={async () => {
                    if (!fetchSdpManifestUrl.trim() || !detailFlow?.id) return;
                    fetchSdpError = "";
                    fetchSdpLoading = true;
                    try {
                      const result = await onFetchSDP(detailFlow.id, fetchSdpManifestUrl.trim());
                      if (result?.flow) detailFlow = result.flow;
                    } catch (e) {
                      fetchSdpError = e.message || "Fetch failed";
                    } finally {
                      fetchSdpLoading = false;
                    }
                  }}
                  disabled={fetchSdpLoading || !fetchSdpManifestUrl.trim()}
                  class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:opacity-50 text-white text-sm font-medium"
                >
                  {fetchSdpLoading ? "Fetching..." : "Fetch SDP"}
                </button>
              </div>
              {#if fetchSdpError}
                <p class="text-xs text-red-400 mt-1">{fetchSdpError}</p>
              {/if}
            {/if}
          </div>
        {/if}
      </div>

      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={closeDetailModal}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Close
        </button>
        {#if canEdit}
          <button
            onclick={() => {
              const flowToEdit = detailFlow; // Save flow before closing modal
              closeDetailModal();
              handleEditFlow(flowToEdit);
            }}
            class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
          >
            Edit Flow
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Flow Check Modal -->
{#if showCheckModal && checkFlow}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in"
    onclick={closeCheckModal}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-2xl animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 class="text-xl font-bold text-gray-100">Check NMOS: {checkFlow.display_name}</h2>
        <button
          onclick={closeCheckModal}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-6 space-y-4">
        <div class="space-y-2">
          <label for="check-base-url" class="block text-sm font-medium text-gray-300">
            NMOS Node Base URL
          </label>
          <input
            id="check-base-url"
            type="text"
            bind:value={checkBaseUrl}
            placeholder="http://192.168.1.100:8080"
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
        </div>

        {#if checkResult}
          <div class="rounded-lg border {checkResult.error ? 'border-red-800 bg-red-950/50' : 'border-gray-800 bg-gray-800/50'} p-4">
            {#if checkResult.error}
              <p class="text-red-300 font-semibold mb-2">Error</p>
              <p class="text-red-200 text-sm">{checkResult.error}</p>
            {:else}
              <div class="space-y-3">
                <div>
                  <p class="text-xs text-gray-400 mb-1">Flow ID</p>
                  <p class="text-sm font-mono text-gray-200">{checkResult.flow_id}</p>
                </div>
                <div>
                  <p class="text-xs text-gray-400 mb-1">Display Name</p>
                  <p class="text-sm text-gray-200">{checkResult.display_name}</p>
                </div>
                <div>
                  <p class="text-xs text-gray-400 mb-1">Match Status</p>
                  <span
                    class="inline-flex items-center rounded-full px-3 py-1 text-xs font-medium {checkResult.is_match
                      ? 'bg-emerald-900 text-emerald-200 border border-emerald-700'
                      : 'bg-red-900 text-red-200 border border-red-700'}"
                  >
                    {checkResult.is_match ? "Match Found" : "No Match"}
                  </span>
                </div>
                <div>
                  <p class="text-xs text-gray-400 mb-1">Matched Count</p>
                  <p class="text-sm text-gray-200">{checkResult.matched_count || 0} sender(s)</p>
                </div>
                {#if checkResult.nmos_matches && checkResult.nmos_matches.length > 0}
                  <div>
                    <p class="text-xs text-gray-400 mb-2">Matching Senders</p>
                    <div class="space-y-2">
                      {#each checkResult.nmos_matches as match}
                        <div class="bg-gray-900 p-3 rounded-md border border-gray-700">
                          <p class="text-sm font-medium text-gray-200">{match.label || match.id}</p>
                          <p class="text-xs text-gray-400 font-mono">{match.id}</p>
                          {#if match.manifest_href}
                            <p class="text-xs text-gray-500 mt-1">{match.manifest_href}</p>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={closeCheckModal}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Close
        </button>
        <button
          onclick={handleCheckFlow}
          disabled={checking || !checkBaseUrl.trim()}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
        >
          {checking ? "Checking..." : "Check Flow"}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Lock/Unlock Info Modal -->
{#if showLockInfoModal && lockInfoFlow}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-fade-in"
    onclick={() => {
      showLockInfoModal = false;
      lockInfoFlow = null;
      lockInfoMessage = "";
    }}
    role="dialog"
    aria-modal="true"
  >
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-md animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <div class="flex items-center gap-3">
          {#if lockInfoFlow.locked}
            <svg class="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          {:else}
            <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
            </svg>
          {/if}
          <h2 class="text-xl font-bold text-gray-100">
            {lockInfoFlow.locked ? "Flow Locked" : "Flow Unlocked"}
          </h2>
        </div>
        <button
          onclick={() => {
            showLockInfoModal = false;
            lockInfoFlow = null;
            lockInfoMessage = "";
          }}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="p-6 space-y-4">
        <div class="space-y-2">
          <p class="text-sm text-gray-300">{lockInfoMessage}</p>
          <div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
            <div class="space-y-2">
              <div>
                <p class="text-xs text-gray-400 mb-1">Flow Name</p>
                <p class="text-sm font-medium text-gray-100">{lockInfoFlow.display_name}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 mb-1">Flow ID</p>
                <p class="text-sm text-gray-300 font-mono">{lockInfoFlow.flow_id}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 mb-1">Multicast Address</p>
                <p class="text-sm text-gray-300">{lockInfoFlow.multicast_ip}:{lockInfoFlow.port}</p>
              </div>
              <div>
                <p class="text-xs text-gray-400 mb-1">Status</p>
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {lockInfoFlow.locked
                    ? 'bg-amber-900 text-amber-200 border border-amber-700'
                    : 'bg-emerald-900 text-emerald-200 border border-emerald-700'}"
                >
                  {lockInfoFlow.locked ? "Locked" : "Unlocked"}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={() => {
            showLockInfoModal = false;
            lockInfoFlow = null;
            lockInfoMessage = "";
          }}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(-20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  .animate-fade-in {
    animation: fade-in 0.2s ease-out;
  }
  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>
