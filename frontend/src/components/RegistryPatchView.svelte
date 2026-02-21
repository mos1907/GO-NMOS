<script>
  let {
    registryNodes = [],
    registryDevices = [],
    registrySenders = [],
    registryReceivers = [],
    /** Node whose senders are shown (source side) */
    selectedSenderNodeId = "",
    /** Node whose receivers are shown and where IS-05 is called (destination side) */
    selectedReceiverNodeId = "",
    selectedRegistryDeviceId = "",
    onSelectSenderNode,
    onSelectReceiverNode,
    onSelectDevice,
    isPatchTakeReady,
    selectedPatchSender = null,
    selectedPatchReceiver = null,
    nmosIS05Base = "",
    nmosTakeBusy = false,
    nmosPatchStatus = "",
    nmosPatchError = "",
    nmosPatchWarning = "",
    onSelectPatchSender,
    onSelectPatchReceiver,
    onExecutePatchTake,
    flows = [],
    token = "",
    /** Increment to trigger refetch of receiver active state (e.g. after TAKE). */
    refreshReceiverActiveTrigger = 0,
    // Discover at URL → ask "Register?" before adding to registry
    discoverAtUrlResult = null,
    pendingRegisterUrl = "",
    discoverAtUrlLoading = false,
    onDiscoverAtUrl = () => {},
    onRegisterDiscoveredNode = () => {},
    onCloseRegisterConfirm = () => {},
    /** Remove node from registry (editor/admin). Called with nodeId after confirm. */
    onDeleteNode,
    canEdit = false,
    // RDS: bulk node discovery and register to registry (NMOS Patch style)
    showConnectRDSModal = false,
    rdsQueryUrl = "",
    rdsDiscovering = false,
    rdsNodes = [],
    rdsSelectedIds = [],
    rdsError = "",
    onOpenRDS,
    onCloseRDS,
    onChangeRegistryQueryUrl,
    onDiscoverRegistryNodes,
    onToggleRegistryNode,
    onSelectAllRegistryNodes,
    onRegisterSelectedToRegistry,
    onReloadRegistry,
    reloadRegistrySyncing = false,
    /** Configured RDS from Settings (for display and actions) */
    registryConfigs = [],
    /** Per-RDS stats (nodes, devices, senders, receivers) from GET /registry/config/stats */
    registryStats = [],
    onReloadRegistryUrl,
    onUpdateRegistry,
    onRemoveRegistry,
  } = $props();

  let discoverUrlInput = $state("");
  let editingRds = $state(null);
  let editForm = $state({ name: "", query_url: "", role: "prod", enabled: true });
  let reloadingRdsUrl = $state("");
  let updateRdsLoading = $state(false);

  function statsForRds(queryUrl) {
    const norm = (u) => (u || "").trim().replace(/\/+$/, "");
    return (registryStats || []).find((s) => norm(s.query_url) === norm(queryUrl));
  }

  function handleRDSSelectAll() {
    onSelectAllRegistryNodes && onSelectAllRegistryNodes();
  }

  // Ensure arrays are never null (reactive)
  let safeRegistryNodes = $derived(registryNodes || []);
  let safeRegistryDevices = $derived(registryDevices || []);
  let safeRegistrySenders = $derived(registrySenders || []);
  let safeRegistryReceivers = $derived(registryReceivers || []);

  // A.4: Topology path model (extensible for SDN/IS-06: path_id, link_ids, capacity, etc.)
  /** @typedef {{ senderId: string, senderLabel: string, receiverId: string, receiverLabel: string, segmentLabel: string, flowId?: string, pathId?: string, capacity?: string }} TopologyPath */

  // A.4: Group by site / room / device type
  let groupBy = $state("none"); // "none" | "site" | "room" | "device_type"

  /** @type {TopologyPath | null} */
  const currentPath = $derived.by(() => {
    if (!selectedPatchSender || !selectedPatchReceiver) return null;
    return {
      senderId: selectedPatchSender.id,
      senderLabel: selectedPatchSender.label || selectedPatchSender.id,
      receiverId: selectedPatchReceiver.id,
      receiverLabel: selectedPatchReceiver.label || selectedPatchReceiver.id,
      segmentLabel: "IS-05",
      flowId: selectedPatchSender.flow_id || undefined
      // pathId, capacity: reserved for SDN/IS-06
    };
  });

  // ST 2110 / NMOS: Sender = output (source), Receiver = input (destination). A node can be input-only, output-only, or both (e.g. router).
  const nodeIdsWithSenders = $derived.by(() => {
    const set = new Set();
    for (const s of safeRegistrySenders) {
      const dev = safeRegistryDevices.find((d) => d.id === s.device_id);
      if (dev?.node_id) set.add(dev.node_id);
    }
    return set;
  });
  const nodeIdsWithReceivers = $derived.by(() => {
    const set = new Set();
    for (const r of safeRegistryReceivers) {
      const dev = safeRegistryDevices.find((d) => d.id === r.device_id);
      if (dev?.node_id) set.add(dev.node_id);
    }
    return set;
  });
  const senderNodes = $derived(safeRegistryNodes.filter((n) => nodeIdsWithSenders.has(n.id)));
  const receiverNodes = $derived(safeRegistryNodes.filter((n) => nodeIdsWithReceivers.has(n.id)));

  const sendersFromSelectedSenderNode = $derived.by(() => {
    if (!selectedSenderNodeId) return [];
    return safeRegistrySenders.filter((s) => {
      const dev = safeRegistryDevices.find((d) => d.id === s.device_id);
      const onNode = dev && dev.node_id === selectedSenderNodeId;
      const deviceOk = !selectedRegistryDeviceId || s.device_id === selectedRegistryDeviceId;
      return onNode && deviceOk;
    });
  });

  const receiversFromSelectedReceiverNode = $derived.by(() => {
    if (!selectedReceiverNodeId) return [];
    return safeRegistryReceivers.filter((r) => {
      const dev = safeRegistryDevices.find((d) => d.id === r.device_id);
      return dev && dev.node_id === selectedReceiverNodeId;
    });
  });

  const nodeGroups = $derived.by(() => {
    const list = safeRegistryNodes;
    if (groupBy === "none" || groupBy === "device_type") return [{ key: "", label: "All", nodes: list }];
    const map = new Map();
    for (const n of list) {
      const site = n.tags?.site?.[0] || "";
      const room = n.tags?.room?.[0] || "";
      const key = groupBy === "site" ? (site || "_ungrouped") : groupBy === "room" ? (room || "_ungrouped") : "";
      if (!map.has(key)) {
        const label = key === "_ungrouped" ? "Ungrouped" : key || "All";
        map.set(key, { key, label, nodes: [] });
      }
      map.get(key).nodes.push(n);
    }
    return [...map.values()].sort((a, b) => (a.label === "Ungrouped" ? 1 : b.label === "Ungrouped" ? -1 : a.label.localeCompare(b.label)));
  });

  const deviceGroups = $derived.by(() => {
    const list = safeRegistryDevices.filter((d) => !selectedSenderNodeId || d.node_id === selectedSenderNodeId);
    if (groupBy !== "device_type") return [{ key: "", label: "All", devices: list }];
    const map = new Map();
    for (const d of list) {
      const type = d.type || "";
      const short = type.split(":").pop() || type || "_ungrouped";
      const key = short || "_ungrouped";
      if (!map.has(key)) map.set(key, { key, label: key === "_ungrouped" ? "Ungrouped" : key, devices: [] });
      map.get(key).devices.push(d);
    }
    return [...map.values()].sort((a, b) => a.label.localeCompare(b.label));
  });

  function deviceTypeLabel(type) {
    if (!type) return "—";
    const part = type.split(":").pop();
    return part || type;
  }

  // IS-05 is called on the receiver's node (connection API lives there).
  const currentTreeNodeId = $derived.by(() => {
    if (selectedReceiverNodeId) return selectedReceiverNodeId;
    if (selectedPatchReceiver?.device_id) {
      const dev = safeRegistryDevices.find((d) => d.id === selectedPatchReceiver.device_id);
      if (dev?.node_id) return dev.node_id;
    }
    if (nmosIS05Base && safeRegistryNodes.length > 0) {
      try {
        const u = new URL(nmosIS05Base);
        const host = u.hostname || u.host?.split(":")[0] || "";
        if (host) {
          const node = safeRegistryNodes.find(
            (n) => n.hostname === host || n.hostname === u.host || (n.hostname && n.hostname.includes(host))
          );
          if (node?.id) return node.id;
        }
      } catch (_) {}
    }
    return "";
  });

  const receiverIdsOnCurrentNode = $derived.by(() => {
    if (!currentTreeNodeId) return [];
    return safeRegistryReceivers
      .filter((r) => {
        const dev = safeRegistryDevices.find((d) => d.id === r.device_id);
        return dev && dev.node_id === currentTreeNodeId;
      })
      .map((r) => r.id);
  });

  // BCC-style: IS-05 active state per receiver (Receiving: sender, master_enable)
  let receiverActiveMap = $state({});
  let receiverActiveLoading = $state(false);

  async function loadReceiverActive() {
    if (!nmosIS05Base || receiverIdsOnCurrentNode.length === 0 || !token) {
      receiverActiveMap = {};
      return;
    }
    receiverActiveLoading = true;
    try {
      const { api } = await import("../lib/api.js");
      const ids = receiverIdsOnCurrentNode.join(",");
      const list = await api(
        `/nmos/receivers-active?is05_base=${encodeURIComponent(nmosIS05Base)}&receiver_ids=${encodeURIComponent(ids)}`,
        { token }
      );
      const map = {};
      for (const item of list || []) {
        if (item.receiver_id) {
          map[item.receiver_id] = {
            sender_id: item.sender_id || "",
            master_enable: item.master_enable,
          };
        }
      }
      receiverActiveMap = map;
    } catch (e) {
      console.warn("Failed to load receiver active state:", e);
      receiverActiveMap = {};
    } finally {
      receiverActiveLoading = false;
    }
  }

  $effect(() => {
    const base = nmosIS05Base;
    const nodeId = currentTreeNodeId;
    const ids = receiverIdsOnCurrentNode;
    const trigger = refreshReceiverActiveTrigger;
    if (!base || !nodeId) {
      receiverActiveMap = {};
      return;
    }
    if (ids.length > 0 || trigger > 0) {
      loadReceiverActive();
    }
  });

  function getSenderLabel(senderId) {
    if (!senderId) return "";
    const s = safeRegistrySenders.find((x) => x.id === senderId);
    return s?.label || (senderId.length > 8 ? senderId.slice(0, 8) + "…" : senderId);
  }

  function getReceiverLabel(receiverId) {
    if (!receiverId) return "";
    const r = safeRegistryReceivers.find((x) => x.id === receiverId);
    return r?.label || (receiverId.length > 8 ? receiverId.slice(0, 8) + "…" : receiverId);
  }

  // Per sender: list of receiver ids that are actively receiving from this sender (master_enable)
  const activeDestinationsBySender = $derived.by(() => {
    const out = {};
    for (const [receiverId, active] of Object.entries(receiverActiveMap)) {
      if (active?.sender_id && active.master_enable !== false) {
        if (!out[active.sender_id]) out[active.sender_id] = [];
        out[active.sender_id].push(receiverId);
      }
    }
    return out;
  });

  async function disableReceiver(receiverId, e) {
    if (e) e.stopPropagation();
    if (!receiverId || !nmosIS05Base || !token) return;
    try {
      const { api } = await import("../lib/api.js");
      await api("/nmos/receiver-disable", {
        method: "POST",
        token,
        body: { is05_base: nmosIS05Base, receiver_id: receiverId },
      });
      await loadReceiverActive();
    } catch (err) {
      console.error("Receiver disable failed:", err);
      alert("Disable failed: " + (err.message || err));
    }
  }

  // A.5: Snapshot export & conformance
  let conformanceReport = $state(null);
  let conformanceLoading = $state(false);
  let showConformanceModal = $state(false);

  async function exportSnapshot() {
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/nmos/snapshot", { token });
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `nmos-snapshot-${new Date().toISOString().split("T")[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      console.error("Failed to export snapshot:", e);
      alert("Failed to export snapshot: " + e.message);
    }
  }

  // B.1: Receiver connection state & history
  let receiverConnections = $state([]);
  let receiverConnectionHistory = $state([]);
  let connectionStateLoading = $state(false);

  async function loadReceiverConnectionState(receiverId) {
    if (!receiverId || !token) return;
    connectionStateLoading = true;
    receiverConnections = [];
    receiverConnectionHistory = [];
    try {
      const { api } = await import("../lib/api.js");
      const [conns, hist] = await Promise.all([
        api(`/receiver/connections?receiver_id=${encodeURIComponent(receiverId)}`, { token }),
        api(`/receiver/${encodeURIComponent(receiverId)}/history?limit=10`, { token })
      ]);
      receiverConnections = Array.isArray(conns) ? conns : [];
      receiverConnectionHistory = Array.isArray(hist) ? hist : [];
    } catch (e) {
      console.warn("Failed to load receiver connection state:", e);
    } finally {
      connectionStateLoading = false;
    }
  }

  $effect(() => {
    const rec = selectedPatchReceiver;
    if (rec?.id) loadReceiverConnectionState(rec.id);
    else {
      receiverConnections = [];
      receiverConnectionHistory = [];
    }
  });

  async function checkConformance() {
    conformanceLoading = true;
    conformanceReport = null;
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/nmos/conformance", { token });
      conformanceReport = data;
      showConformanceModal = true;
    } catch (e) {
      console.error("Failed to check conformance:", e);
      alert("Failed to check conformance: " + e.message);
    } finally {
      conformanceLoading = false;
    }
  }

  // B.2: Patch plan (bulk patch)
  let patchPlanReceivers = $state([]);
  let bulkPatchMode = $state("immediate"); // "immediate" | "safe_switch"
  let bulkPatchResult = $state(null);
  let bulkPatchBusy = $state(false);

  // B.3: Policy check & warnings
  let policyViolations = $state([]);
  let showPolicyWarning = $state(false);
  let overridePolicy = $state(false);

  // B.2: Scheduled activations
  let showScheduleModal = $state(false);
  let scheduledDateTime = $state("");
  let scheduledActivations = $state([]);
  let scheduledLoading = $state(false);

  function addReceiverToPlan(rec) {
    if (!rec?.id) return;
    if (patchPlanReceivers.some((r) => r.id === rec.id)) return;
    patchPlanReceivers = [...patchPlanReceivers, rec];
  }

  function removeReceiverFromPlan(recId) {
    patchPlanReceivers = patchPlanReceivers.filter((r) => r.id !== recId);
  }

  function clearPatchPlan() {
    patchPlanReceivers = [];
    bulkPatchResult = null;
    policyViolations = [];
    overridePolicy = false;
    showPolicyWarning = false;
  }

  const internalFlowForSender = $derived.by(() => {
    const sender = selectedPatchSender;
    if (!sender?.flow_id) return null;
    const f = (flows || []).find(
      (x) => x.flow_id && (x.flow_id === sender.flow_id || x.flow_id === sender.flow_id?.toString?.())
    );
    return f;
  });

  const isBulkPatchReady = $derived(
    !!(
      internalFlowForSender?.id &&
      patchPlanReceivers.length > 0 &&
      nmosIS05Base &&
      !bulkPatchBusy
    )
  );

  async function executeBulkPatch() {
    if (!isBulkPatchReady) return;
    // B.3: Policy check before executing
    if (policyViolations.length === 0) {
      await checkPolicyForBulkPatch();
    }
    if (policyViolations.length > 0 && !overridePolicy) {
      showPolicyWarning = true;
      return;
    }
    bulkPatchBusy = true;
    bulkPatchResult = null;
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/nmos/bulk-patch", {
        method: "POST",
        token,
        body: {
          flow_id: internalFlowForSender.id,
          receiver_ids: patchPlanReceivers.map((r) => r.id),
          is05_base_url: nmosIS05Base,
          sender_id: selectedPatchSender?.id || undefined,
          mode: bulkPatchMode,
          override_policy: overridePolicy || undefined
        }
      });
      bulkPatchResult = data;
      overridePolicy = false;
      policyViolations = [];
      showPolicyWarning = false;
    } catch (e) {
      bulkPatchResult = { success: 0, failed: patchPlanReceivers.length, error: e.message };
    } finally {
      bulkPatchBusy = false;
    }
  }

  // B.2: Scheduled activations
  async function loadScheduledActivations() {
    scheduledLoading = true;
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/nmos/scheduled-activations?limit=20", { token });
      scheduledActivations = Array.isArray(data) ? data : [];
    } catch (e) {
      console.error("Failed to load scheduled activations:", e);
      scheduledActivations = [];
    } finally {
      scheduledLoading = false;
    }
  }

  function openScheduleModal() {
    if (!isBulkPatchReady) return;
    const now = new Date();
    now.setMinutes(now.getMinutes() + 5); // Default: 5 minutes from now
    scheduledDateTime = now.toISOString().slice(0, 16); // YYYY-MM-DDTHH:mm format
    showScheduleModal = true;
  }

  async function createScheduledActivation() {
    if (!isBulkPatchReady || !scheduledDateTime) return;
    try {
      const { api } = await import("../lib/api.js");
      const scheduledAt = new Date(scheduledDateTime).toISOString();
      await api("/nmos/scheduled-activations", {
        method: "POST",
        token,
        body: {
          flow_id: internalFlowForSender.id,
          receiver_ids: patchPlanReceivers.map((r) => r.id),
          is05_base_url: nmosIS05Base,
          sender_id: selectedPatchSender?.id || undefined,
          scheduled_at: scheduledAt,
          mode: bulkPatchMode
        }
      });
      showScheduleModal = false;
      scheduledDateTime = "";
      await loadScheduledActivations();
      clearPatchPlan();
    } catch (e) {
      alert("Failed to create scheduled activation: " + e.message);
    }
  }

  async function cancelScheduledActivation(id) {
    if (!confirm("Cancel this scheduled activation?")) return;
    try {
      const { api } = await import("../lib/api.js");
      await api(`/nmos/scheduled-activations/${id}`, { method: "DELETE", token });
      await loadScheduledActivations();
    } catch (e) {
      alert("Failed to cancel scheduled activation: " + e.message);
    }
  }

  // B.3: Policy check
  async function checkPolicyForConnection(senderId, receiverId, flowId, flowLabel, senderLabel) {
    if (!senderId || !receiverId) return null;
    try {
      const { api } = await import("../lib/api.js");
      const result = await api("/routing/check", {
        method: "POST",
        token,
        body: {
          sender_id: senderId,
          receiver_id: receiverId,
          flow_id: flowId || undefined,
          flow_label: flowLabel || undefined,
          sender_label: senderLabel || undefined,
        },
      });
      return result;
    } catch (e) {
      console.warn("Policy check failed:", e);
      return null;
    }
  }

  async function checkPolicyForBulkPatch() {
    if (!selectedPatchSender?.id || patchPlanReceivers.length === 0 || !internalFlowForSender) return;
    policyViolations = [];
    for (const rec of patchPlanReceivers) {
      const check = await checkPolicyForConnection(
        selectedPatchSender.id,
        rec.id,
        internalFlowForSender.id,
        internalFlowForSender.display_name,
        selectedPatchSender.label
      );
      if (check && !check.allowed && check.violations) {
        policyViolations.push(...check.violations.map((v) => ({ ...v, receiver_id: rec.id, receiver_label: rec.label })));
      }
    }
    if (policyViolations.length > 0) {
      showPolicyWarning = true;
    }
  }

  $effect(() => {
    loadScheduledActivations();
    const interval = setInterval(loadScheduledActivations, 30000); // Refresh every 30s
    return () => clearInterval(interval);
  });
</script>

<section class="mt-4 space-y-4">
  <!-- Discover node at URL: after discover, parent shows Register? modal -->
  <div class="rounded-lg bg-slate-900/60 border border-slate-700 p-3 flex flex-wrap items-end gap-3">
    <div class="flex flex-col gap-1 min-w-[280px]">
      <label for="discover-node-url" class="text-xs font-medium text-slate-400">Discover node at URL</label>
      <input
        id="discover-node-url"
        type="text"
        bind:value={discoverUrlInput}
        placeholder="http://host:port"
        class="px-3 py-2 rounded-md bg-slate-950 border border-slate-700 text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-orange-500"
      />
    </div>
    <button
      type="button"
      disabled={discoverAtUrlLoading || !discoverUrlInput.trim()}
      onclick={() => onDiscoverAtUrl(discoverUrlInput.trim())}
      class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
    >
      {discoverAtUrlLoading ? "Discovering…" : "Discover"}
    </button>
    <p class="text-[11px] text-slate-500">If IS-04/IS-05 is found, you will be asked whether to register the node.</p>
  </div>

  <!-- Register? modal (shown when parent sets discoverAtUrlResult + pendingRegisterUrl) -->
  {#if pendingRegisterUrl && (discoverAtUrlResult != null)}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" role="dialog" aria-modal="true" aria-labelledby="register-confirm-title">
      <div class="bg-slate-900 border border-slate-600 rounded-xl shadow-xl max-w-md w-full p-5 space-y-4">
        <h3 id="register-confirm-title" class="text-base font-semibold text-slate-100">Register node to registry?</h3>
        <p class="text-sm text-slate-300">
          Found IS-04/IS-05 at <code class="px-1 py-0.5 rounded bg-slate-800 text-slate-200 break-all">{pendingRegisterUrl}</code>.
          {#if discoverAtUrlResult?.counts}
            <span class="block mt-2 text-slate-400">Senders: {discoverAtUrlResult.counts.senders ?? "—"}, Receivers: {discoverAtUrlResult.counts.receivers ?? "—"}, Flows: {discoverAtUrlResult.counts.flows ?? "—"}</span>
          {/if}
        </p>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            onclick={() => onCloseRegisterConfirm()}
            class="px-3 py-2 rounded-md border border-slate-600 text-slate-300 hover:bg-slate-800 text-sm"
          >
            Skip
          </button>
          <button
            type="button"
            onclick={() => { onRegisterDiscoveredNode(pendingRegisterUrl); onCloseRegisterConfirm(); }}
            class="px-3 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium"
          >
            Register
          </button>
        </div>
      </div>
    </div>
  {/if}

  <header class="flex flex-wrap items-center justify-between gap-3">
    <div>
      <h3 class="text-sm font-semibold text-slate-50">Registry & Patch</h3>
      <p class="text-[11px] text-slate-400">
        Select Sender Node → Receiver Node → Source (Sender) and destination (Receiver) → TAKE. ST 2110: Sender = output, Receiver = input; a node may be input-only, output-only, or both (e.g. router).
      </p>
    </div>
    <div class="flex flex-wrap items-center gap-3">
      <div class="flex items-center gap-2">
        <label for="registry-patch-group-by" class="text-[11px] text-slate-400">Group by</label>
        <select
          id="registry-patch-group-by"
          class="px-2 py-1 rounded-md bg-slate-900 border border-slate-700 text-[11px] text-slate-200"
          bind:value={groupBy}
        >
          <option value="none">None</option>
          <option value="site">Site</option>
          <option value="room">Room</option>
          <option value="device_type">Device type</option>
        </select>
      </div>
      <div class="flex items-center gap-2">
        <button
          class="px-2 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-[11px] text-slate-200 border border-slate-700"
          onclick={exportSnapshot}
          title="Export registry snapshot (JSON)"
        >
          Export snapshot
        </button>
        <button
          class="px-2 py-1 rounded-md bg-indigo-900 hover:bg-indigo-800 text-[11px] text-indigo-100 border border-indigo-700 disabled:opacity-50"
          onclick={checkConformance}
          disabled={conformanceLoading}
          title="Check conformance (required fields, formats)"
        >
          {conformanceLoading ? "Checking..." : "Check conformance"}
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-md bg-slate-800 hover:bg-slate-700 text-[11px] text-slate-200 border border-slate-600"
          onclick={() => onOpenRDS?.()}
          title="Discover nodes from RDS and bulk-register to registry"
        >
          Connect RDS
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-md bg-slate-800 hover:bg-slate-700 text-[11px] text-slate-200 border border-slate-600 disabled:opacity-50 disabled:cursor-not-allowed"
          disabled={reloadRegistrySyncing}
          onclick={() => onReloadRegistry?.()}
          title="Re-fetch from all registries configured in Settings and sync devices, flows, senders, receivers"
        >
          {reloadRegistrySyncing ? "Syncing…" : "Reload registry"}
        </button>
      </div>
      <div class="text-[11px] text-slate-400 space-y-0.5 text-right">
        <p>Registry: Nodes {safeRegistryNodes.length} · Devices {safeRegistryDevices.length}</p>
        <p>Endpoints: Senders {safeRegistrySenders.length} · Receivers {safeRegistryReceivers.length}</p>
      </div>
    </div>
  </header>

  <!-- Configured RDS list: Reload / Edit / Delete per row -->
  <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 mb-4">
    <h4 class="text-xs font-semibold text-slate-200 uppercase tracking-wide mb-3">Configured RDS</h4>
    {#if (registryConfigs || []).length === 0}
      <p class="text-xs text-slate-500">No RDS configured. Use <strong>Connect RDS</strong> or <strong>Settings → NMOS Registries</strong> to add.</p>
    {:else}
      <ul class="space-y-2">
        {#each (registryConfigs || []) as r}
          {@const stats = statsForRds(r.query_url)}
          <li class="flex flex-wrap items-center gap-2 py-2 px-3 rounded-lg border border-slate-800 bg-slate-900/40">
            <div class="min-w-0 flex-1">
              <span class="font-medium text-slate-200 text-sm">{r.name || r.query_url || "—"}</span>
              <span class="text-xs text-slate-500 block truncate" title={r.query_url}>{r.query_url}</span>
              {#if stats}
                {#if stats.error}
                  <span class="text-[10px] text-amber-400 mt-0.5 block">Unreachable: {stats.error}</span>
                {:else}
                  <span class="text-[10px] text-slate-400 mt-0.5 block">
                    Nodes: {stats.nodes} · Devices: {stats.devices} · Senders: {stats.senders} · Receivers: {stats.receivers}
                  </span>
                {/if}
              {:else}
                <span class="text-[10px] text-slate-500 mt-0.5 block">—</span>
              {/if}
            </div>
            {#if !r.enabled}
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-slate-700 text-slate-400">Disabled</span>
            {/if}
            <div class="flex items-center gap-1 shrink-0">
              <button
                type="button"
                class="px-2 py-1 rounded text-[11px] bg-slate-700 hover:bg-slate-600 text-slate-200 border border-slate-600 disabled:opacity-50"
                disabled={reloadingRdsUrl === r.query_url || reloadRegistrySyncing}
                title="Reload this RDS"
                onclick={async () => {
                  reloadingRdsUrl = r.query_url;
                  try {
                    await onReloadRegistryUrl?.(r.query_url);
                  } finally {
                    reloadingRdsUrl = "";
                  }
                }}
              >
                {reloadingRdsUrl === r.query_url ? "…" : "Reload"}
              </button>
              {#if canEdit}
                <button
                  type="button"
                  class="px-2 py-1 rounded text-[11px] bg-slate-700 hover:bg-slate-600 text-slate-200 border border-slate-600"
                  title="Edit RDS"
                  onclick={() => {
                    editingRds = r;
                    editForm = { name: r.name || "", query_url: r.query_url || "", role: r.role || "prod", enabled: r.enabled !== false };
                  }}
                >
                  Edit
                </button>
                <button
                  type="button"
                  class="px-2 py-1 rounded text-[11px] text-red-300 hover:bg-red-900/40 border border-red-800"
                  title="Remove RDS and all its nodes"
                  onclick={() => {
                    if (confirm("Remove this RDS? All nodes from this registry (and their senders/receivers) will be deleted.")) {
                      onRemoveRegistry?.(r.query_url);
                    }
                  }}
                >
                  Delete
                </button>
              {/if}
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

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
    <div class="grid md:grid-cols-[2fr_2fr_2fr] gap-4 items-start">
      <!-- Sender Nodes + Devices + Senders -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 space-y-3">
        <div class="flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Sender Nodes</h4>
          <span class="text-[11px] text-slate-400">{senderNodes.length} node</span>
        </div>
        <div class="space-y-1 max-h-32 overflow-auto pr-1">
          {#each senderNodes as node}
            {@const siteTag = node.tags?.site?.[0] || ""}
            {@const roomTag = node.tags?.room?.[0] || ""}
            {@const nodeTooltip = [siteTag && `Site: ${siteTag}`, roomTag && `Room: ${roomTag}`].filter(Boolean).join(" · ") || node.hostname || node.id}
            <div class="flex items-center gap-1 w-full group/row">
              <button
                type="button"
                class="flex-1 min-w-0 text-left px-3 py-1.5 rounded-lg border text-[11px] {selectedSenderNodeId === node.id
                  ? 'border-emerald-500 bg-slate-900 text-slate-50'
                  : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-600'}"
                title={nodeTooltip}
                onclick={() => onSelectSenderNode?.(node.id)}
              >
                <span class="font-medium truncate block">{node.label || node.id}</span>
                <span class="block text-[10px] text-slate-400 truncate">{node.hostname}</span>
              </button>
              {#if canEdit && onDeleteNode}
                <button type="button" class="shrink-0 p-1.5 rounded border border-slate-700 bg-slate-800/80 text-slate-400 hover:bg-red-900/40 hover:border-red-700 hover:text-red-200" title="Remove from registry" onclick={(e) => { e.stopPropagation(); if (confirm("Remove this node from registry?")) onDeleteNode(node.id); }}>
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              {/if}
            </div>
          {/each}
        </div>

        <div class="flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Devices</h4>
          <span class="text-[11px] text-slate-400">
            {#if selectedSenderNodeId}
              {safeRegistryDevices.filter((d) => d.node_id === selectedSenderNodeId).length} of {safeRegistryDevices.length}
            {:else}
              {safeRegistryDevices.length} devices
            {/if}
          </span>
        </div>
        <div class="space-y-2 max-h-48 overflow-auto pr-1">
          {#each deviceGroups as { key, label, devices }}
            {#if groupBy === "device_type"}
              <p class="text-[10px] font-semibold text-slate-400 uppercase tracking-wide sticky top-0 bg-slate-950/95 py-0.5">{label}</p>
            {/if}
            <div class="space-y-1">
              {#each devices as dev}
                {@const siteTag = dev.tags?.site?.[0] || ""}
                {@const roomTag = dev.tags?.room?.[0] || ""}
                {@const caps = dev.meta?.capabilities}
                {@const capSummary = caps && typeof caps === "object" ? Object.keys(caps).slice(0, 3).join(", ") : ""}
                {@const devTooltip = [siteTag && `Site: ${siteTag}`, roomTag && `Room: ${roomTag}`, capSummary && `Capabilities: ${capSummary}`].filter(Boolean).join(" · ")}
                <button
                  type="button"
                  class="w-full text-left px-3 py-1.5 rounded-lg border text-[11px] {selectedRegistryDeviceId === dev.id
                    ? 'border-svelte bg-slate-900 text-slate-50'
                    : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-600'}"
                  title={devTooltip || dev.label || dev.id}
                  onclick={() => onSelectDevice?.(dev.id)}
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-medium truncate flex-1">{dev.label || dev.id}</span>
                    {#if (groupBy !== "device_type") && (siteTag || roomTag)}
                      <div class="flex gap-1 shrink-0">
                        {#if siteTag}
                          <span class="px-1.5 py-0.5 rounded text-[9px] bg-indigo-900/60 text-indigo-200 border border-indigo-700/50" title="Site: {siteTag}">
                            {siteTag}
                          </span>
                        {/if}
                        {#if roomTag}
                          <span class="px-1.5 py-0.5 rounded text-[9px] bg-purple-900/60 text-purple-200 border border-purple-700/50" title="Room: {roomTag}">
                            {roomTag}
                          </span>
                        {/if}
                      </div>
                    {/if}
                  </div>
                  <span class="block text-[10px] text-slate-400 truncate">{deviceTypeLabel(dev.type)}</span>
                </button>
              {/each}
            </div>
          {/each}
        </div>

        <div class="flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Sources (Senders)</h4>
          <span class="text-[11px] text-slate-400">{selectedSenderNodeId ? sendersFromSelectedSenderNode.length : 0} sender</span>
        </div>
        <div class="space-y-1 max-h-44 overflow-auto pr-1 text-[11px]">
          {#if !selectedSenderNodeId}
            <p class="text-slate-500 italic text-[11px]">Select a Sender Node.</p>
          {:else if sendersFromSelectedSenderNode.length === 0}
            <p class="text-slate-500 italic text-[11px]">No senders on this node.</p>
          {:else}
            {#each sendersFromSelectedSenderNode as s}
              {@const activeRecvIds = activeDestinationsBySender[s.id] || []}
              <button
                type="button"
                class="w-full text-left px-3 py-2 rounded-lg border border-slate-800 bg-slate-900/60 hover:border-svelte/70 hover:bg-slate-900 flex flex-col gap-0.5"
                onclick={() => onSelectPatchSender?.(s)}
              >
                <span class="text-[13px] font-medium text-slate-50 truncate">{s.label}</span>
                <span class="text-[11px] text-slate-400 truncate">{s.flow_id}</span>
                <span class="text-[10px] text-slate-500 truncate uppercase">{s.transport}</span>
                {#if activeRecvIds.length > 0}
                  <div class="mt-1 pt-1 border-t border-slate-800/80 text-[10px] text-emerald-400" title={activeRecvIds.map((id) => getReceiverLabel(id)).join(", ")}>
                    → Receiving at: {activeRecvIds.length === 1 ? getReceiverLabel(activeRecvIds[0]) : activeRecvIds.length === 2 ? activeRecvIds.map((id) => getReceiverLabel(id)).join(", ") : `${activeRecvIds.length} destinations`}
                  </div>
                {/if}
              </button>
            {/each}
          {/if}
        </div>
      </div>

      <!-- Path + TAKE (orta) -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/80 p-4 space-y-4">
        <div class="border-slate-800 text-[11px] text-slate-400 space-y-2">
          <p class="font-semibold text-slate-200">Path (sender → segment → receiver)</p>
          <!-- A.4 Path visualisation: diagram style, data model ready for SDN/IS-06 -->
          {#if currentPath}
            <div class="flex items-stretch gap-0 rounded-lg overflow-hidden border border-slate-700 bg-slate-900/80" role="img" aria-label="Path from {currentPath.senderLabel} to {currentPath.receiverLabel}">
              <div class="flex-1 min-w-0 py-2 px-2.5 rounded-l-md bg-emerald-950/50 border-r border-slate-700" title={currentPath.senderLabel}>
                <p class="text-[10px] uppercase text-slate-500 font-semibold">Source</p>
                <p class="font-medium text-slate-200 truncate text-xs">{currentPath.senderLabel}</p>
              </div>
              <div class="flex items-center px-1.5 bg-slate-800/80 border-r border-slate-700" aria-hidden="true">
                <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/></svg>
              </div>
              <div class="flex-1 min-w-0 py-2 px-2.5 bg-slate-800/60 border-r border-slate-700" title="Network segment (IS-05)">
                <p class="text-[10px] uppercase text-slate-500 font-semibold">Segment</p>
                <p class="text-slate-400 truncate text-xs">{currentPath.segmentLabel}</p>
              </div>
              <div class="flex items-center px-1.5 bg-slate-800/80 border-r border-slate-700" aria-hidden="true">
                <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/></svg>
              </div>
              <div class="flex-1 min-w-0 py-2 px-2.5 rounded-r-md bg-violet-950/30" title={currentPath.receiverLabel}>
                <p class="text-[10px] uppercase text-slate-500 font-semibold">Destination</p>
                <p class="font-medium text-slate-200 truncate text-xs">{currentPath.receiverLabel}</p>
              </div>
            </div>
          {:else}
            <div class="py-3 px-3 rounded-lg bg-slate-900/60 border border-slate-800 border-dashed text-slate-500 text-center">
              Select a sender and a receiver to see path.
            </div>
          {/if}
          <p class="truncate text-[10px]">IS-05 base: {nmosIS05Base || "not set"}</p>
          {#if nmosPatchStatus}
            <p class="text-[11px] text-green-400 mt-2">{nmosPatchStatus}</p>
          {/if}
          {#if nmosPatchWarning}
            <p class="text-[11px] text-amber-400 mt-2">{nmosPatchWarning}</p>
          {/if}
          {#if nmosPatchError}
            <p class="text-[11px] text-red-400 mt-2">{nmosPatchError}</p>
          {/if}
          <div class="pt-4 flex flex-col items-center gap-2">
            <button
              class="px-6 py-3 rounded-xl bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-bold transition-colors"
              disabled={!isPatchTakeReady?.()}
              onclick={onExecutePatchTake}
            >
              {nmosTakeBusy ? "TAKING..." : "TAKE PATCH"}
            </button>
            <span class="text-[10px] text-slate-500">{isPatchTakeReady?.() ? "Source and destination selected" : "Select Sender and Receiver"}</span>
          </div>
        </div>
      </div>

      <!-- Receiver Nodes + Destinations -->
      <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 space-y-3">
        <div class="flex items-center justify-between mb-1">
          <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Receiver Nodes</h4>
          <span class="text-[11px] text-slate-400">{receiverNodes.length} node</span>
        </div>
        <div class="space-y-1 max-h-32 overflow-auto pr-1">
          {#each receiverNodes as node}
            {@const siteTag = node.tags?.site?.[0] || ""}
            {@const roomTag = node.tags?.room?.[0] || ""}
            {@const nodeTooltip = [siteTag && `Site: ${siteTag}`, roomTag && `Room: ${roomTag}`].filter(Boolean).join(" · ") || node.hostname || node.id}
            <div class="flex items-center gap-1 w-full group/row">
              <button
                type="button"
                class="flex-1 min-w-0 text-left px-3 py-1.5 rounded-lg border text-[11px] {selectedReceiverNodeId === node.id
                  ? 'border-violet-500 bg-slate-900 text-slate-50'
                  : 'border-slate-800 bg-slate-900/60 text-slate-300 hover:border-slate-600'}"
                title={nodeTooltip}
                onclick={() => onSelectReceiverNode?.(node.id)}
              >
                <span class="font-medium truncate block">{node.label || node.id}</span>
                <span class="block text-[10px] text-slate-400 truncate">{node.hostname}</span>
              </button>
              {#if canEdit && onDeleteNode}
                <button type="button" class="shrink-0 p-1.5 rounded border border-slate-700 bg-slate-800/80 text-slate-400 hover:bg-red-900/40 hover:border-red-700 hover:text-red-200" title="Remove from registry" onclick={(e) => { e.stopPropagation(); if (confirm("Remove this node from registry?")) onDeleteNode(node.id); }}>
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              {/if}
            </div>
          {/each}
        </div>

        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Destinations (Receivers)</h4>
            <p class="text-[11px] text-slate-400">Select a Receiver Node, then choose destination.</p>
          </div>
          <span class="text-[11px] text-slate-400">{receiversFromSelectedReceiverNode.length} receiver</span>
        </div>
        <div class="flex items-center gap-2 mb-1">
          {#if receiverActiveLoading}
            <span class="text-[10px] text-slate-500">Connection status…</span>
          {:else if nmosIS05Base && receiverIdsOnCurrentNode.length > 0}
            <button
              type="button"
              class="px-2 py-0.5 rounded text-[10px] bg-slate-800 hover:bg-slate-700 text-slate-300"
              onclick={() => loadReceiverActive()}
            >
              Refresh status
            </button>
          {/if}
        </div>
        <div class="space-y-1 max-h-56 overflow-auto pr-1 text-[11px]">
          {#if !selectedReceiverNodeId}
            <p class="text-slate-500 italic">Select a Receiver Node.</p>
          {:else if receiversFromSelectedReceiverNode.length === 0}
            <p class="text-slate-500 italic">No receivers on this node.</p>
          {:else}
            {#each receiversFromSelectedReceiverNode as r}
              {@const active = receiverActiveMap[r.id]}
              <div class="flex flex-col gap-0.5">
                <button
                  type="button"
                  class="w-full text-left px-3 py-2 rounded-lg border border-slate-800 bg-slate-900/60 hover:border-svelte/70 hover:bg-slate-900 flex flex-col gap-0.5"
                  onclick={() => onSelectPatchReceiver?.(r)}
                >
                  <span class="text-[13px] font-medium text-slate-50 truncate">{r.label}</span>
                  <span class="text-[11px] text-slate-400 truncate">{r.description}</span>
                  <span class="text-[10px] text-slate-500 truncate uppercase">{r.format} · {r.transport}</span>
                  {#if active !== undefined}
                    <div class="mt-1 pt-1 border-t border-slate-800/80 flex flex-col gap-0.5">
                      {#if active.sender_id}
                        <div class="flex items-center justify-between gap-2 flex-wrap">
                          <span class="text-[10px] text-emerald-400" title="Receiving from {active.sender_id}">
                            Receiving: {getSenderLabel(active.sender_id)}
                          </span>
                          {#if active.master_enable === false}
                            <span class="px-1.5 py-0.5 rounded text-[9px] bg-amber-900/60 text-amber-200">Disabled</span>
                          {:else}
                            <span class="px-1.5 py-0.5 rounded text-[9px] bg-emerald-900/60 text-emerald-200">Enabled</span>
                          {/if}
                        </div>
                        {#if active.flow?.display_name}
                          <span class="text-[10px] text-slate-400" title="Flow in use">
                            Flow: {active.flow.display_name}{#if active.flow.multicast_ip && active.flow.port != null} ({active.flow.multicast_ip}:{active.flow.port}){/if}
                          </span>
                        {/if}
                      {:else}
                        <span class="text-[10px] text-slate-500">Unpatched</span>
                      {/if}
                    </div>
                  {/if}
                </button>
                {#if active?.sender_id && active?.master_enable !== false}
                  <button
                    type="button"
                    class="self-end px-2 py-0.5 rounded text-[10px] text-amber-300 hover:bg-amber-900/40 border border-amber-800/60"
                    onclick={(e) => disableReceiver(r.id, e)}
                    title="Disable this receiver (un-TAKE)"
                  >
                    Disable
                  </button>
                {/if}
              </div>
            {/each}
          {/if}
        </div>

        <!-- B.2: Patch plan (bulk patch) -->
        <div class="mt-4 pt-3 border-t border-slate-800 space-y-2">
          <h4 class="text-[11px] font-semibold text-slate-200 uppercase tracking-wide">Patch plan</h4>
          <p class="text-[10px] text-slate-500">Add receivers for bulk patch, then execute.</p>
          {#if selectedPatchReceiver}
            <button
              type="button"
              class="px-2 py-1 rounded text-[10px] bg-slate-800 hover:bg-slate-700 text-slate-200 disabled:opacity-50"
              disabled={patchPlanReceivers.some((r) => r.id === selectedPatchReceiver.id)}
              onclick={() => addReceiverToPlan(selectedPatchReceiver)}
            >
              + Add &quot;{selectedPatchReceiver.label}&quot; to plan
            </button>
          {/if}
          {#if patchPlanReceivers.length > 0}
            <div class="space-y-1 max-h-24 overflow-y-auto">
              {#each patchPlanReceivers as rec}
                <div class="flex items-center justify-between gap-2 px-2 py-1 rounded bg-slate-900/60 border border-slate-800 text-[11px]">
                  <span class="truncate flex-1">{rec.label}</span>
                  <button
                    type="button"
                    class="shrink-0 px-1.5 py-0.5 rounded text-[10px] text-red-300 hover:bg-red-900/40"
                    onclick={() => removeReceiverFromPlan(rec.id)}
                  >
                    remove
                  </button>
                </div>
              {/each}
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <label class="flex items-center gap-1.5 text-[10px] text-slate-400 cursor-pointer">
                <input type="radio" name="bulkMode" checked={bulkPatchMode === "immediate"} onchange={() => (bulkPatchMode = "immediate")} />
                Immediate
              </label>
              <label class="flex items-center gap-1.5 text-[10px] text-slate-400 cursor-pointer">
                <input type="radio" name="bulkMode" checked={bulkPatchMode === "safe_switch"} onchange={() => (bulkPatchMode = "safe_switch")} />
                Safe switch (audio first)
              </label>
            </div>
            <div class="flex gap-2 flex-wrap">
              <button
                class="px-3 py-1.5 rounded-lg bg-svelte text-slate-950 text-xs font-semibold disabled:opacity-40"
                disabled={!isBulkPatchReady}
                title={!internalFlowForSender ? "Select a sender with a matching flow first" : ""}
                onclick={executeBulkPatch}
              >
                {bulkPatchBusy ? "PATCHING..." : `EXECUTE BULK (${patchPlanReceivers.length})`}
              </button>
              <button
                class="px-3 py-1.5 rounded-lg border border-slate-700 text-slate-300 text-xs hover:bg-slate-800 disabled:opacity-40"
                disabled={!isBulkPatchReady}
                onclick={openScheduleModal}
              >
                Schedule
              </button>
              <button
                class="px-2 py-1.5 rounded-lg border border-slate-700 text-slate-300 text-xs hover:bg-slate-800"
                onclick={clearPatchPlan}
              >
                Clear
              </button>
            </div>
            <!-- B.3: Policy violations warning -->
            {#if policyViolations.length > 0 && !overridePolicy}
              <div class="p-3 rounded bg-amber-950/50 border border-amber-800 text-[11px]">
                <div class="flex items-center gap-2 mb-2">
                  <span class="text-amber-300 font-semibold">⚠ Policy Violations</span>
                </div>
                <div class="space-y-1 mb-2">
                  {#each policyViolations.slice(0, 3) as v}
                    <div class="text-amber-200">
                      <span class="font-medium">{v.policy_name}</span>: {v.reason}
                      {#if v.receiver_label}
                        <span class="text-amber-300"> (→ {v.receiver_label})</span>
                      {/if}
                    </div>
                  {/each}
                  {#if policyViolations.length > 3}
                    <div class="text-amber-300 italic">+ {policyViolations.length - 3} more violations</div>
                  {/if}
                </div>
                <div class="flex gap-2">
                  <button
                    class="px-2 py-1 rounded text-[10px] bg-amber-900/60 hover:bg-amber-900 text-amber-200"
                    onclick={() => (overridePolicy = true)}
                  >
                    Override & Continue
                  </button>
                  <button
                    class="px-2 py-1 rounded text-[10px] border border-amber-700 text-amber-300 hover:bg-amber-900/40"
                    onclick={() => {
                      policyViolations = [];
                      showPolicyWarning = false;
                    }}
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            {/if}
            {#if bulkPatchResult}
              <div class="p-2 rounded bg-slate-900/80 border border-slate-700 text-[11px]">
                <span class="text-emerald-300 font-medium">{bulkPatchResult.success} OK</span>
                {#if bulkPatchResult.failed > 0}
                  <span class="text-amber-300"> · {bulkPatchResult.failed} failed</span>
                {/if}
                {#if bulkPatchResult.error}
                  <p class="mt-1 text-red-300">{bulkPatchResult.error}</p>
                {/if}
              </div>
            {/if}
          {:else}
            <p class="text-[10px] text-slate-500 italic">No receivers in plan. Select a receiver and click &quot;Add to plan&quot;.</p>
          {/if}
        </div>

        <!-- B.1: Receiver connection state & history -->
        {#if selectedPatchReceiver}
          <div class="mt-4 pt-3 border-t border-slate-800 space-y-2">
            <h4 class="text-[11px] font-semibold text-slate-200 uppercase tracking-wide">Connection state</h4>
            {#if connectionStateLoading}
              <p class="text-[11px] text-slate-500 italic">Loading…</p>
            {:else if receiverConnections.length === 0 && receiverConnectionHistory.length === 0}
              <p class="text-[11px] text-slate-500 italic">No connection state recorded yet.</p>
            {:else}
              {#if receiverConnections.length > 0}
                <div class="space-y-1">
                  {#each receiverConnections as conn}
                    <div class="p-2 rounded bg-slate-900/60 border border-slate-800 text-[11px]">
                      <div class="flex items-center justify-between gap-2">
                        <span class="px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase {conn.state === 'active' ? 'bg-emerald-900/60 text-emerald-200' : 'bg-amber-900/60 text-amber-200'}">{conn.state}</span>
                        <span class="text-slate-400">{conn.role}</span>
                      </div>
                      <div class="mt-1 text-slate-300 truncate" title={conn.sender_id}>Sender: {conn.sender_id || "—"}</div>
                      {#if conn.changed_at}
                        <div class="mt-0.5 text-[10px] text-slate-500">{new Date(conn.changed_at).toLocaleString()}{#if conn.changed_by} · {conn.changed_by}{/if}</div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
              {#if receiverConnectionHistory.length > 0}
                <div class="mt-2">
                  <p class="text-[10px] font-medium text-slate-400 mb-1">Recent history</p>
                  <div class="space-y-0.5 max-h-24 overflow-y-auto">
                    {#each receiverConnectionHistory.slice(0, 5) as h}
                      <div class="text-[10px] text-slate-500 flex items-center gap-2">
                        <span class="px-1 py-0.5 rounded {h.action === 'connect' ? 'bg-emerald-900/40' : 'bg-slate-800'}">{h.action}</span>
                        <span class="truncate">{h.sender_id || "—"}</span>
                        <span class="text-slate-600">{new Date(h.changed_at).toLocaleTimeString()}</span>
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}
            {/if}
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- A.5 Conformance modal -->
  {#if showConformanceModal}
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
      role="dialog"
      aria-modal="true"
      aria-labelledby="conformance-title"
      tabindex="-1"
      onclick={() => (showConformanceModal = false)}
      onkeydown={(e) => e.key === "Escape" && (showConformanceModal = false)}
    >
      <div class="w-full max-w-3xl rounded-xl border border-slate-800 bg-slate-950 p-6 max-h-[80vh] overflow-y-auto" role="document" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between mb-4">
          <h3 id="conformance-title" class="text-lg font-semibold text-slate-50">NMOS Conformance Report</h3>
          <button
            class="px-3 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm"
            onclick={() => (showConformanceModal = false)}
          >
            Close
          </button>
        </div>
        {#if conformanceReport}
          <div class="space-y-4">
            <div class="flex items-center justify-between gap-4 p-3 rounded-lg bg-slate-900 border border-slate-800">
              <span class="text-sm text-slate-300">Total issues:</span>
              <span class="text-lg font-semibold {conformanceReport.total_issues === 0 ? 'text-emerald-300' : 'text-amber-300'}">
                {conformanceReport.total_issues}
              </span>
              {#if conformanceReport.issues && conformanceReport.issues.length > 0}
                {@const errors = conformanceReport.issues.filter((i) => i.severity === "error")}
                {@const warnings = conformanceReport.issues.filter((i) => i.severity === "warning")}
                <div class="flex gap-2 text-[10px]">
                  {#if errors.length > 0}
                    <span class="px-2 py-0.5 rounded bg-red-900/60 text-red-200">{errors.length} error(s)</span>
                  {/if}
                  {#if warnings.length > 0}
                    <span class="px-2 py-0.5 rounded bg-amber-900/60 text-amber-200">{warnings.length} warning(s)</span>
                  {/if}
                </div>
              {/if}
            </div>
            {#if conformanceReport.summary && Object.keys(conformanceReport.summary).length > 0}
              <div class="grid grid-cols-5 gap-2 text-xs">
                {#each Object.entries(conformanceReport.summary) as [type, count]}
                  <div class="p-2 rounded bg-slate-900 border border-slate-800 text-center">
                    <div class="font-semibold text-slate-200">{count}</div>
                    <div class="text-slate-400">{type}</div>
                  </div>
                {/each}
              </div>
            {/if}
            {#if conformanceReport.issues && conformanceReport.issues.length > 0}
              <div class="space-y-2">
                <h4 class="text-sm font-semibold text-slate-200">Issues (required fields, formats, tags)</h4>
                <div class="space-y-1 max-h-96 overflow-y-auto">
                  {#each conformanceReport.issues as issue}
                    <div class="p-2 rounded bg-slate-900/60 border border-slate-800 text-xs">
                      <div class="flex items-center justify-between gap-2">
                        <span class="font-medium text-slate-200">{issue.resource_type}</span>
                        <span class="px-2 py-0.5 rounded text-[10px] uppercase {issue.severity === 'error' ? 'bg-red-900/60 text-red-200' : 'bg-amber-900/60 text-amber-200'}">
                          {issue.severity || "error"}
                        </span>
                        <span class="px-2 py-0.5 rounded bg-slate-800 text-slate-400">{issue.issue}</span>
                      </div>
                      <div class="mt-1 text-slate-400">
                        {issue.resource_label || issue.resource_id}
                      </div>
                      <div class="mt-1 text-slate-500">
                        <span class="font-medium">{issue.field}</span>: {issue.message}
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {:else}
              <div class="p-4 rounded-lg bg-emerald-950/50 border border-emerald-800 text-center">
                <p class="text-emerald-200 font-semibold">No conformance issues found.</p>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- B.2: Scheduled activations list -->
  <div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4 space-y-3">
    <div class="flex items-center justify-between">
      <h4 class="text-xs font-semibold text-slate-100 uppercase tracking-wide">Scheduled activations</h4>
      <button
        class="px-2 py-1 rounded text-[10px] bg-slate-800 hover:bg-slate-700 text-slate-200"
        onclick={loadScheduledActivations}
      >
        Refresh
      </button>
    </div>
    {#if scheduledLoading}
      <p class="text-[11px] text-slate-500 italic">Loading…</p>
    {:else if !scheduledActivations || scheduledActivations.length === 0}
      <p class="text-[11px] text-slate-500 italic">No scheduled activations.</p>
    {:else}
      <div class="space-y-1 max-h-64 overflow-y-auto">
        {#each scheduledActivations as act}
          <div class="p-2 rounded bg-slate-900/60 border border-slate-800 text-[11px]">
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class="px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase {act.status === 'pending' ? 'bg-amber-900/60 text-amber-200' : act.status === 'executed' ? 'bg-emerald-900/60 text-emerald-200' : act.status === 'failed' ? 'bg-red-900/60 text-red-200' : 'bg-slate-700 text-slate-300'}">{act.status}</span>
              {#if act.status === 'pending'}
                <button
                  class="px-1.5 py-0.5 rounded text-[10px] text-red-300 hover:bg-red-900/40"
                  onclick={() => cancelScheduledActivation(act.id)}
                >
                  Cancel
                </button>
              {/if}
            </div>
            <div class="text-[10px] text-slate-400">
              <div>Scheduled: {new Date(act.scheduled_at).toLocaleString()}</div>
              {#if act.executed_at}
                <div>Executed: {new Date(act.executed_at).toLocaleString()}</div>
              {/if}
              <div>Receivers: {act.receiver_ids?.length || 0} · Mode: {act.mode}</div>
              {#if act.result && typeof act.result === 'object' && act.result.success !== undefined}
                <div class="mt-1">
                  <span class="text-emerald-300">{act.result.success} OK</span>
                  {#if act.result.failed > 0}
                    <span class="text-amber-300"> · {act.result.failed} failed</span>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- B.2: Schedule modal -->
  {#if showScheduleModal}
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
      role="dialog"
      aria-modal="true"
      aria-labelledby="schedule-title"
      tabindex="-1"
      onclick={() => (showScheduleModal = false)}
      onkeydown={(e) => e.key === "Escape" && (showScheduleModal = false)}
    >
      <div class="w-full max-w-md rounded-xl border border-slate-800 bg-slate-950 p-6" role="document" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between mb-4">
          <h3 id="schedule-title" class="text-lg font-semibold text-slate-50">Schedule activation</h3>
          <button
            class="px-3 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm"
            onclick={() => (showScheduleModal = false)}
          >
            Close
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label for="schedule-datetime" class="block text-sm text-slate-300 mb-1">Date & time</label>
            <input
              id="schedule-datetime"
              type="datetime-local"
              bind:value={scheduledDateTime}
              class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-slate-200"
              min={new Date().toISOString().slice(0, 16)}
            />
          </div>
          <div class="text-[11px] text-slate-400">
            <p>Flow: {internalFlowForSender?.display_name || "—"}</p>
            <p>Receivers: {patchPlanReceivers.length}</p>
            <p>Mode: {bulkPatchMode}</p>
          </div>
          <div class="flex gap-2 justify-end">
            <button
              class="px-3 py-2 rounded-lg border border-slate-700 text-slate-300 hover:bg-slate-800"
              onclick={() => (showScheduleModal = false)}
            >
              Cancel
            </button>
            <button
              class="px-3 py-2 rounded-lg bg-svelte text-slate-950 font-semibold disabled:opacity-40"
              disabled={!scheduledDateTime || new Date(scheduledDateTime) <= new Date()}
              onclick={createScheduledActivation}
            >
              Schedule
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Edit RDS modal -->
  {#if editingRds}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" role="dialog" aria-modal="true" aria-labelledby="edit-rds-title">
      <div class="bg-gray-900 border border-gray-700 rounded-xl p-6 max-w-md w-full mx-4 shadow-xl">
        <h3 id="edit-rds-title" class="text-lg font-semibold text-gray-100 mb-4">Edit RDS</h3>
        <div class="space-y-3">
          <div>
            <label for="edit-rds-name" class="block text-xs font-medium text-gray-300 mb-1">Name</label>
            <input
              id="edit-rds-name"
              type="text"
              class="w-full px-2 py-1.5 rounded-md bg-gray-800 border border-gray-600 text-sm text-gray-100"
              placeholder="e.g. Core Registry"
              bind:value={editForm.name}
            />
          </div>
          <div>
            <label for="edit-rds-url" class="block text-xs font-medium text-gray-300 mb-1">Query URL</label>
            <input
              id="edit-rds-url"
              type="url"
              class="w-full px-2 py-1.5 rounded-md bg-gray-800 border border-gray-600 text-sm text-gray-100"
              placeholder="http://host:port/x-nmos/query"
              bind:value={editForm.query_url}
            />
          </div>
          <div>
            <label for="edit-rds-role" class="block text-xs font-medium text-gray-300 mb-1">Role</label>
            <select
              id="edit-rds-role"
              class="w-full px-2 py-1.5 rounded-md bg-gray-800 border border-gray-600 text-sm text-gray-100"
              bind:value={editForm.role}
            >
              <option value="prod">prod</option>
              <option value="lab">lab</option>
              <option value="remote">remote</option>
              <option value="other">other</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <input
              id="edit-rds-enabled"
              type="checkbox"
              class="rounded border-gray-600 bg-gray-800"
              bind:checked={editForm.enabled}
            />
            <label for="edit-rds-enabled" class="text-xs text-gray-300">Enabled</label>
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-4">
          <button
            type="button"
            class="px-4 py-2 rounded-md border border-gray-600 text-gray-200 hover:bg-gray-800"
            disabled={updateRdsLoading}
            onclick={() => (editingRds = null)}
          >
            Cancel
          </button>
          <button
            type="button"
            class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white font-medium disabled:opacity-50"
            disabled={updateRdsLoading || !editForm.query_url?.trim()}
            onclick={async () => {
              if (!editingRds?.query_url) return;
              updateRdsLoading = true;
              try {
                await onUpdateRegistry?.(editingRds.query_url, { ...editForm });
                editingRds = null;
              } finally {
                updateRdsLoading = false;
              }
            }}
          >
            {updateRdsLoading ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Connect RDS Modal: discover nodes from RDS, select, bulk-register to registry -->
  {#if showConnectRDSModal}
    <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" role="dialog" aria-modal="true" aria-labelledby="rds-modal-title">
      <div class="bg-slate-900 border border-slate-600 rounded-xl p-6 w-full max-w-2xl space-y-4 shadow-xl">
        <div class="flex items-start justify-between">
          <div>
            <h3 id="rds-modal-title" class="text-lg font-semibold text-slate-50 mb-1">Connect to NMOS Registry (RDS)</h3>
            <p class="text-sm text-slate-400">
              Enter Query API URL to discover nodes; selected ones are bulk-registered to the internal registry (they appear under Sender/Receiver Nodes).
            </p>
          </div>
          <button type="button" class="px-2 py-1 rounded text-slate-400 hover:text-white hover:bg-slate-800" onclick={() => onCloseRDS && onCloseRDS()} aria-label="Close">✕</button>
        </div>
        <div class="flex flex-wrap gap-2">
          <input
            type="url"
            class="flex-1 min-w-[260px] px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-orange-500"
            placeholder="http://host:port or http://host.docker.internal:8082"
            value={rdsQueryUrl}
            oninput={(e) => onChangeRegistryQueryUrl && onChangeRegistryQueryUrl(e.target.value)}
          />
          <button
            type="button"
            class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            disabled={!(rdsQueryUrl && rdsQueryUrl.trim()) || rdsDiscovering}
            onclick={() => onDiscoverRegistryNodes && onDiscoverRegistryNodes()}
          >
            {rdsDiscovering ? "Discovering..." : "Discover Nodes"}
          </button>
        </div>
        <p class="text-xs text-slate-500">
          If the registry host runs inside Docker, use <code class="px-1 py-0.5 rounded bg-slate-800 text-slate-300">host.docker.internal</code>.
        </p>
        {#if rdsError}
          <div class="px-4 py-2 rounded-lg bg-red-900/50 border border-red-800 text-sm text-red-300">{rdsError}</div>
        {/if}
        <div class="flex items-center justify-between">
          <span class="text-sm text-slate-400">Found: <strong class="text-slate-200">{Array.isArray(rdsNodes) ? rdsNodes.length : 0}</strong> node(s) · Selected: <strong class="text-slate-200">{Array.isArray(rdsSelectedIds) ? rdsSelectedIds.length : 0}</strong></span>
          {#if Array.isArray(rdsNodes) && rdsNodes.length >= 1}
            <button type="button" class="px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm border border-slate-700" onclick={handleRDSSelectAll}>Select all</button>
          {/if}
        </div>
        <div class="max-h-[320px] overflow-auto rounded border border-slate-700">
          {#if !Array.isArray(rdsNodes) || rdsNodes.length === 0}
            <div class="p-4 text-sm text-slate-500 text-center">No nodes discovered yet. Enter URL and click Discover Nodes.</div>
          {:else}
            <div class="divide-y divide-slate-800">
              {#each rdsNodes as n}
                <label class="flex items-start gap-3 p-3 hover:bg-slate-800/60 cursor-pointer">
                  <input type="checkbox" class="mt-1" checked={Array.isArray(rdsSelectedIds) && rdsSelectedIds.includes(n.id)} onchange={() => onToggleRegistryNode && onToggleRegistryNode(n.id)} />
                  <div class="min-w-0 flex-1">
                    <div class="text-sm font-medium text-slate-100 truncate">{n.label || n.id}</div>
                    <div class="text-xs text-slate-400 truncate">{n.base_url || n.baseURL || n.BaseURL || n.href || n.Href || "—"}</div>
                  </div>
                </label>
              {/each}
            </div>
          {/if}
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm border border-slate-600" onclick={() => onCloseRDS && onCloseRDS()}>Cancel</button>
          <button
            type="button"
            class="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-semibold disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            disabled={!(Array.isArray(rdsSelectedIds) && rdsSelectedIds.length >= 1) || rdsDiscovering}
            onclick={() => onRegisterSelectedToRegistry && onRegisterSelectedToRegistry()}
          >
            {rdsDiscovering ? "Saving..." : "Register selected to registry"}
          </button>
        </div>
      </div>
    </div>
  {/if}
</section>

