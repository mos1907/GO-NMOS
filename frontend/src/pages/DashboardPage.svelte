<script>
  import { onMount, onDestroy } from "svelte";
  import { api, apiWithMeta } from "../lib/api.js";
  import { connectMQTT, disconnectMQTT } from "../lib/mqtt.js";
  import NMOSView from "../components/NMOSView.svelte";
  import NMOSPatchView from "../components/NMOSPatchView.svelte";
  import TopologyView from "../components/TopologyView.svelte";
  import DashboardHomeView from "../components/DashboardHomeView.svelte";
  import FlowsView from "../components/FlowsView.svelte";
  import SearchView from "../components/SearchView.svelte";
  import NewFlowView from "../components/NewFlowView.svelte";
  import UsersView from "../components/UsersView.svelte";
  import SettingsView from "../components/SettingsView.svelte";
  import CheckerView from "../components/CheckerView.svelte";
  import AutomationJobsView from "../components/AutomationJobsView.svelte";
  import PlannerView from "../components/PlannerView.svelte";
  import AddressMapView from "../components/AddressMapView.svelte";
  import LogsView from "../components/LogsView.svelte";

  export let token;
  export let user;
  export let onLogout;

  let currentView = "dashboard";
  let loading = true;
  let error = "";
  let success = "";

  let summary = { total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 };
  let flows = [];
  let flowLimit = 50;
  let flowOffset = 0;
  let flowTotal = 0;
  let flowSortBy = "updated_at";
  let flowSortOrder = "desc";
  let users = [];
  let settings = {};
  let searchTerm = "";
  let searchResults = [];
  let searchLimit = 50;
  let searchOffset = 0;
  let searchTotal = 0;
  let importing = false;
  let nmosBaseUrl = "";
  let nmosResult = null;
  let nmosIS05Base = "";
  let selectedNMOSFlow = null;
  let selectedNMOSReceiver = null;
  let nmosTakeBusy = false;
  let checkerResult = null;
  let automationJobs = [];
  let addressMap = null;
  let logsKind = "api";
  let logsLines = [];

  // Internal NMOS registry (IS-04 style) state
  let registryNodes = [];
  let registryDevices = [];
  let registrySenders = [];
  let registryReceivers = [];
  let selectedRegistryNodeId = "";
  let selectedRegistryDeviceId = "";

  // NMOS Patch-style view state (sender/receiver selection)
  let nmosNodes = [];
  let showAddNodeModal = false;
  let newNodeName = "";
  let newNodeUrl = "";
  let selectedSenderNodeId = "";
  let selectedReceiverNodeId = "";
  let senderNodeSenders = [];
  let receiverNodeReceivers = [];
  let selectedPatchSender = null;
  let selectedPatchReceiver = null;
  let nmosPatchStatus = "";
  let nmosPatchError = "";
  let senderFilterText = "";
  let receiverFilterText = "";
  let senderFormatFilter = "";
  let receiverFormatFilter = "";
  let plannerRoots = [];
  let plannerChildren = [];
  let selectedPlannerRoot = null;
  let newPlannerParent = { parent_id: null, name: "", cidr: "", description: "", color: "" };
  let newPlannerChild = { parent_id: null, name: "", cidr: "", description: "", color: "" };

  let newFlow = {
    display_name: "",
    multicast_ip: "",
    source_ip: "",
    port: 5004,
    flow_status: "active",
    availability: "available",
    transport_protocol: "RTP/UDP",
    note: ""
  };

  const isAdmin = user?.role === "admin";
  const canEdit = user?.role === "admin" || user?.role === "editor";

  // Basit UI sürüm bilgisi (frontend build versiyonu)
  const uiVersion = "go-NMOS UI v0.2.0 (router beta)";
  let showBuildModal = true;

  async function loadDashboard() {
    loading = true;
    error = "";
    try {
      [summary, flows] = await Promise.all([
        api("/flows/summary", { token }),
        loadFlows()
      ]);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function loadFlows() {
    const { data, headers } = await apiWithMeta(
      `/flows?limit=${flowLimit}&offset=${flowOffset}&sort_by=${encodeURIComponent(flowSortBy)}&sort_order=${encodeURIComponent(flowSortOrder)}`,
      { token }
    );
    flowTotal = Number(headers.get("X-Total-Count") || 0);
    return data;
  }

  async function loadUsers() {
    if (!(user?.role === "admin" || user?.role === "editor")) return;
    users = await api("/users", { token });
  }

  async function loadSettings() {
    settings = await api("/settings", { token });
  }

  async function refreshAll() {
    success = "";
    await loadDashboard();
    await loadUsers().catch(() => {});
    await loadSettings().catch(() => {});
    await loadCheckerLatest().catch(() => {});
    await loadAutomationJobs().catch(() => {});
    await loadAddressMap().catch(() => {});
    await loadPlannerRoots().catch(() => {});
  }

  async function runSearch() {
    error = "";
    try {
      if (!searchTerm.trim()) {
        searchResults = [];
        searchTotal = 0;
        return;
      }
      const { data, headers } = await apiWithMeta(
        `/flows/search?q=${encodeURIComponent(searchTerm)}&limit=${searchLimit}&offset=${searchOffset}&sort_by=${encodeURIComponent(flowSortBy)}&sort_order=${encodeURIComponent(flowSortOrder)}`,
        { token }
      );
      searchResults = data;
      searchTotal = Number(headers.get("X-Total-Count") || 0);
    } catch (e) {
      error = e.message;
    }
  }

  async function createFlow() {
    error = "";
    success = "";
    try {
      await api("/flows", { method: "POST", token, body: newFlow });
      success = "Flow created successfully.";
      newFlow = {
        display_name: "",
        multicast_ip: "",
        source_ip: "",
        port: 5004,
        flow_status: "active",
        availability: "available",
        transport_protocol: "RTP/UDP",
        note: ""
      };
      await refreshAll();
      currentView = "flows";
    } catch (e) {
      error = e.message;
    }
  }

  async function toggleFlowLock(flow) {
    error = "";
    success = "";
    try {
      const result = await api(`/flows/${flow.id}/lock`, {
        method: "POST",
        token,
        body: { locked: !flow.locked }
      });
      flow.locked = result.locked;
      success = result.locked ? "Flow locked." : "Flow unlocked.";
      await loadFlows().then((data) => {
        flows = data;
      });
      await api("/flows/summary", { token }).then((s) => {
        summary = s;
      });
    } catch (e) {
      error = e.message;
    }
  }

  async function deleteFlow(flow) {
    if (!confirm(`Delete flow '${flow.display_name}'?`)) return;
    error = "";
    success = "";
    try {
      await api(`/flows/${flow.id}`, { method: "DELETE", token });
      success = "Flow deleted.";
      await refreshAll();
    } catch (e) {
      error = e.message;
    }
  }

  async function nextFlowPage() {
    if (flowOffset + flowLimit >= flowTotal) return;
    flowOffset += flowLimit;
    flows = await loadFlows();
  }

  async function prevFlowPage() {
    flowOffset = Math.max(0, flowOffset - flowLimit);
    flows = await loadFlows();
  }

  async function applyFlowSort() {
    flowOffset = 0;
    flows = await loadFlows();
  }

  async function nextSearchPage() {
    if (searchOffset + searchLimit >= searchTotal) return;
    searchOffset += searchLimit;
    await runSearch();
  }

  async function prevSearchPage() {
    searchOffset = Math.max(0, searchOffset - searchLimit);
    await runSearch();
  }

  async function saveSetting(key) {
    error = "";
    success = "";
    try {
      await api(`/settings/${key}`, {
        method: "PATCH",
        token,
        body: { value: settings[key] ?? "" }
      });
      success = `Setting '${key}' updated.`;
    } catch (e) {
      error = e.message;
    }
  }

  async function exportFlows() {
    error = "";
    try {
      const data = await api("/flows/export", { token });
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "flows-export.json";
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      error = e.message;
    }
  }

  async function importFlowsFromFile(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    importing = true;
    error = "";
    success = "";
    try {
      const text = await file.text();
      const payload = JSON.parse(text);
      const result = await api("/flows/import", { method: "POST", token, body: payload });
      success = `Import complete. ${result.imported ?? 0} flow processed.`;
      await refreshAll();
    } catch (e) {
      error = e.message;
    } finally {
      importing = false;
      event.target.value = "";
    }
  }

  async function discoverNMOS() {
    error = "";
    success = "";
    try {
      nmosResult = await api("/nmos/discover", {
        method: "POST",
        token,
        body: { base_url: nmosBaseUrl }
      });

      // IS-05 base URL varsayılanı: <base>/x-nmos/connection/<version>
      const base = nmosResult.base_url?.replace(/\/$/, "") || nmosBaseUrl.replace(/\/$/, "");
      const ver = (nmosResult.is04_version || "").replace(/^\//, "");
      nmosIS05Base = `${base}/x-nmos/connection/${ver}`;

      // Varsayılan seçimler: ilk flow ve ilk receiver
      selectedNMOSFlow = flows[0] || null;
      selectedNMOSReceiver = (nmosResult.receivers || [])[0] || null;

      success = "NMOS discovery completed.";
    } catch (e) {
      error = e.message;
    }
  }

  function isTakeReady() {
    return !!(nmosResult && selectedNMOSFlow && selectedNMOSReceiver && nmosIS05Base && !nmosTakeBusy);
  }

  async function executeNMOSTake() {
    if (!isTakeReady()) return;
    error = "";
    success = "";
    nmosTakeBusy = true;
    try {
      const base = nmosIS05Base.replace(/\/$/, "");
      const receiverId = selectedNMOSReceiver.id;
      const connectionUrl = `${base}/single/receivers/${receiverId}/staged`;

      // Uygun NMOS sender varsa flow_id'ye göre eşle
      let senderId = "";
      if (nmosResult?.senders && selectedNMOSFlow?.flow_id) {
        const match = nmosResult.senders.find((s) => s.flow_id === selectedNMOSFlow.flow_id);
        if (match) senderId = match.id;
      }

      await api(`/flows/${selectedNMOSFlow.id}/nmos/apply`, {
        method: "POST",
        token,
        body: {
          connection_url: connectionUrl,
          sender_id: senderId || undefined
        }
      });

      success = `TAKE OK: ${selectedNMOSFlow.display_name} → ${selectedNMOSReceiver.label}`;
    } catch (e) {
      error = e.message;
    } finally {
      nmosTakeBusy = false;
    }
  }

  // NMOS Patch GUI node management (frontend-only, localStorage)
  const NODES_KEY = "go_nmos_nodes";

  function loadNodes() {
    try {
      const raw = localStorage.getItem(NODES_KEY);
      if (!raw) {
        nmosNodes = [];
        return;
      }
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) {
        nmosNodes = parsed;
      }
    } catch {
      nmosNodes = [];
    }
  }

  function saveNodes() {
    try {
      localStorage.setItem(NODES_KEY, JSON.stringify(nmosNodes));
    } catch {
      // ignore
    }
  }

  function openAddNodeModal() {
    newNodeName = "";
    newNodeUrl = "";
    showAddNodeModal = true;
  }

  function cancelAddNode() {
    showAddNodeModal = false;
  }

  function addNode() {
    const name = newNodeName.trim();
    const url = newNodeUrl.trim();
    if (!name || !url) return;
    const id = `${Date.now()}-${Math.random().toString(16).slice(2)}`;
    nmosNodes = [...nmosNodes, { id, name, base_url: url }];
    saveNodes();
    showAddNodeModal = false;
  }

  async function loadSenderNode(nodeId) {
    const node = nmosNodes.find((n) => n.id === nodeId);
    if (!node) return;
    nmosPatchError = "";
    nmosPatchStatus = "";
    try {
      const res = await api("/nmos/discover", {
        method: "POST",
        token,
        body: { base_url: node.base_url }
      });
      senderNodeSenders = res.senders || [];
      selectedPatchSender = senderNodeSenders[0] || null;
      nmosPatchStatus = `Loaded ${senderNodeSenders.length} senders from ${node.name}`;
    } catch (e) {
      nmosPatchError = e.message;
    }
  }

  async function loadReceiverNode(nodeId) {
    const node = nmosNodes.find((n) => n.id === nodeId);
    if (!node) return;
    nmosPatchError = "";
    nmosPatchStatus = "";
    try {
      const res = await api("/nmos/discover", {
        method: "POST",
        token,
        body: { base_url: node.base_url }
      });
      receiverNodeReceivers = res.receivers || [];
      selectedPatchReceiver = receiverNodeReceivers[0] || null;
      const base = res.base_url?.replace(/\/$/, "") || node.base_url.replace(/\/$/, "");
      const ver = (res.is04_version || "").replace(/^\//, "");
      nmosIS05Base = `${base}/x-nmos/connection/${ver}`;
      nmosPatchStatus = `Loaded ${receiverNodeReceivers.length} receivers from ${node.name}`;
    } catch (e) {
      nmosPatchError = e.message;
    }
  }

  function isPatchTakeReady() {
    return !!(selectedPatchSender && selectedPatchReceiver && nmosIS05Base && !nmosTakeBusy);
  }

  async function executePatchTake() {
    if (!isPatchTakeReady()) return;
    nmosPatchError = "";
    nmosPatchStatus = "";
    nmosTakeBusy = true;
    try {
      const base = nmosIS05Base.replace(/\/$/, "");
      const receiverId = selectedPatchReceiver.id;
      const connectionUrl = `${base}/single/receivers/${receiverId}/staged`;

      const internalFlow =
        flows.find((f) => f.flow_id && f.flow_id === selectedPatchSender?.flow_id) ||
        flows.find((f) => f.flow_id && f.flow_id === selectedPatchSender?.flow_id?.toString?.());
      if (!internalFlow) {
        throw new Error(
          `No internal flow found for selected sender flow_id=${selectedPatchSender?.flow_id || "?"}. Import/sync flows first (or select a sender that matches an existing flow).`
        );
      }

      // sender_id + transport_params (from internal flow) ile backend'e delege et
      await api(`/flows/${internalFlow.id}/nmos/apply`, {
        method: "POST",
        token,
        body: {
          connection_url: connectionUrl,
          sender_id: selectedPatchSender.id
        }
      });

      nmosPatchStatus = `TAKE OK: ${selectedPatchSender.label} → ${selectedPatchReceiver.label} (flow: ${internalFlow.display_name})`;
    } catch (e) {
      nmosPatchError = e.message;
    } finally {
      nmosTakeBusy = false;
    }
  }

  function filterSenders(list) {
    return (list || []).filter((s) => {
      const txt = senderFilterText.trim().toLowerCase();
      const fmt = senderFormatFilter.trim();
      const okText =
        !txt ||
        s.label?.toLowerCase().includes(txt) ||
        s.flow_id?.toLowerCase().includes(txt) ||
        s.description?.toLowerCase().includes(txt);
      const okFmt = !fmt || (s.format || "").toLowerCase().includes(fmt);
      return okText && okFmt;
    });
  }

  function filterReceivers(list) {
    return (list || []).filter((r) => {
      const txt = receiverFilterText.trim().toLowerCase();
      const fmt = receiverFormatFilter.trim();
      const okText =
        !txt ||
        r.label?.toLowerCase().includes(txt) ||
        r.description?.toLowerCase().includes(txt) ||
        r.id?.toLowerCase().includes(txt);
      const okFmt = !fmt || (r.format || "").toLowerCase().includes(fmt);
      return okText && okFmt;
    });
  }

  async function runCollisionCheck() {
    error = "";
    try {
      checkerResult = await api("/checker/collisions", { token });
      success = "Collision check completed.";
      await loadDashboard();
    } catch (e) {
      error = e.message;
    }
  }

  async function loadCheckerLatest() {
    checkerResult = await api("/checker/latest?kind=collisions", { token });
  }

  async function loadAutomationJobs() {
    if (!(user?.role === "admin" || user?.role === "editor")) return;
    automationJobs = await api("/automation/jobs", { token });
  }

  async function toggleAutomationJob(job, enabled) {
    error = "";
    try {
      await api(`/automation/jobs/${job.job_id}/${enabled ? "enable" : "disable"}`, {
        method: "POST",
        token
      });
      await loadAutomationJobs();
    } catch (e) {
      error = e.message;
    }
  }

  async function loadAddressMap() {
    addressMap = await api("/address-map", { token });
  }

  async function loadPlannerRoots() {
    plannerRoots = await api("/address/buckets/privileged", { token });
    if (!selectedPlannerRoot && plannerRoots.length > 0) {
      await selectPlannerRoot(plannerRoots[0]);
    }
  }

  async function selectPlannerRoot(root) {
    selectedPlannerRoot = root;
    plannerChildren = await api(`/address/buckets/${root.id}/children`, { token });
    newPlannerParent.parent_id = root.id;
  }

  async function createPlannerParent() {
    error = "";
    try {
      await api("/address/buckets/parent", {
        method: "POST",
        token,
        body: newPlannerParent
      });
      newPlannerParent = { parent_id: selectedPlannerRoot?.id || null, name: "", cidr: "", description: "", color: "" };
      await loadPlannerRoots();
      if (selectedPlannerRoot) await selectPlannerRoot(selectedPlannerRoot);
      success = "Planner parent bucket created.";
    } catch (e) {
      error = e.message;
    }
  }

  async function createPlannerChild(parent) {
    error = "";
    try {
      await api("/address/buckets/child", {
        method: "POST",
        token,
        body: { ...newPlannerChild, parent_id: parent.id }
      });
      newPlannerChild = { parent_id: null, name: "", cidr: "", description: "", color: "" };
      await selectPlannerRoot(selectedPlannerRoot);
      success = "Planner child bucket created.";
    } catch (e) {
      error = e.message;
    }
  }

  async function exportBuckets() {
    try {
      const data = await api("/address/buckets/export", { token });
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "planner-buckets-export.json";
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      error = e.message;
    }
  }

  async function importBucketsFromFile(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    try {
      const payload = JSON.parse(await file.text());
      await api("/address/buckets/import", { method: "POST", token, body: payload });
      await loadPlannerRoots();
      if (selectedPlannerRoot) await selectPlannerRoot(selectedPlannerRoot);
      success = "Planner buckets imported.";
    } catch (e) {
      error = e.message;
    } finally {
      event.target.value = "";
    }
  }

  async function plannerQuickEdit(item) {
    const newName = prompt("Bucket name", item.name);
    if (newName == null) return;
    const newDesc = prompt("Bucket description", item.description || "");
    if (newDesc == null) return;
    try {
      await api(`/address/buckets/${item.id}`, {
        method: "PATCH",
        token,
        body: { name: newName, description: newDesc }
      });
      if (selectedPlannerRoot) await selectPlannerRoot(selectedPlannerRoot);
      success = "Bucket updated.";
    } catch (e) {
      error = e.message;
    }
  }

  async function plannerDelete(item) {
    if (!confirm(`Delete bucket '${item.name}'?`)) return;
    try {
      await api(`/address/buckets/${item.id}`, { method: "DELETE", token });
      if (selectedPlannerRoot) await selectPlannerRoot(selectedPlannerRoot);
      success = "Bucket deleted.";
    } catch (e) {
      error = e.message;
    }
  }

  async function loadLogs() {
    error = "";
    try {
      const data = await api(`/logs?kind=${encodeURIComponent(logsKind)}&lines=300`, { token });
      logsLines = data.lines || [];
    } catch (e) {
      error = e.message;
    }
  }

  async function loadNMOSRegistry() {
    try {
      const [nodes, devices, senders, receivers] = await Promise.all([
        api("/nmos/registry/nodes", { token }),
        api("/nmos/registry/devices", { token }),
        api("/nmos/registry/senders", { token }),
        api("/nmos/registry/receivers", { token })
      ]);
      registryNodes = nodes;
      registryDevices = devices;
      registrySenders = senders;
      registryReceivers = receivers;
    } catch (e) {
      console.error("Failed to load NMOS registry", e);
    }
  }

  function handleMQTTEvent(event) {
    if (event.event === "created" || event.event === "updated") {
      // Refresh flows if we're on flows view
      if (currentView === "flows" || currentView === "dashboard") {
        loadFlows().then((data) => {
          flows = data;
        });
      }
      // Refresh summary
      api("/flows/summary", { token }).then((s) => {
        summary = s;
      });
    } else if (event.event === "deleted") {
      // Remove from list if visible
      flows = flows.filter((f) => f.flow_id !== event.flow_id);
      // Refresh summary
      api("/flows/summary", { token }).then((s) => {
        summary = s;
      });
    }
  }

  onMount(() => {
    refreshAll();
    
    // Connect to MQTT if WebSocket URL is available
    const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsHost = window.location.hostname;
    const wsPort = "9001";
    const wsUrl = `${wsProtocol}//${wsHost}:${wsPort}`;
    const topicPrefix = "go-nmos/flows/events";
    
    try {
      connectMQTT(wsUrl, topicPrefix, handleMQTTEvent);
    } catch (e) {
      console.log("MQTT connection skipped:", e.message);
    }
  });

  onDestroy(() => {
    disconnectMQTT();
  });
</script>

<main class="min-h-screen bg-gradient-to-b from-slate-950 via-slate-950 to-slate-900 text-slate-50">
  <div class="max-w-6xl mx-auto px-4 py-8 space-y-5">
  <header class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-4">
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight text-slate-50">
        go-NMOS
        <span class="ml-2 text-[11px] align-middle rounded-full border border-slate-700 bg-slate-900/70 px-2 py-0.5 font-medium text-slate-200 uppercase tracking-[0.18em]">
          Dashboard
        </span>
      </h1>
      <p class="text-xs text-slate-400">
        Signed in as <span class="font-semibold text-slate-100">{user?.username}</span>
        <span class="mx-2 h-3 w-px inline-block bg-slate-600 align-middle"></span>
        <span class="uppercase text-[11px] tracking-wide px-2 py-0.5 rounded-full bg-slate-900 border border-slate-700 text-slate-100">
          {user?.role}
        </span>
      </p>
    </div>
    <div class="flex items-center gap-2 text-xs">
      <button
        class="px-3 py-1.5 rounded-md bg-slate-900 hover:bg-slate-800 border border-slate-700 text-slate-100 font-medium shadow-sm transition"
        on:click={refreshAll}
      >
        Refresh
      </button>
      <button
        class="px-3 py-1.5 rounded-md bg-svelte hover:bg-orange-400 text-slate-950 font-semibold shadow-sm transition"
        on:click={onLogout}
      >
        Logout
      </button>
    </div>
  </header>

  <nav class="flex flex-wrap gap-2 border-b border-slate-800 pb-3 text-xs mt-3">
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'dashboard'
        ? 'border-svelte bg-slate-900 text-svelte-soft font-semibold shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "dashboard")}
    >
      Dashboard
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'flows'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "flows")}
    >
      Flows
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'search'
        ? 'border-slate-900 bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-nmos-bg text-black/70 hover:border-slate-300 hover:bg-slate-900/5'}"
      on:click={() => (currentView = "search")}
    >
      Search
    </button>
    {#if canEdit}
      <button
        class="px-3 py-1.5 rounded-md border {currentView === 'newFlow'
          ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
          : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
        on:click={() => (currentView = "newFlow")}
      >
        New Flow
      </button>
    {/if}
    {#if user?.role === "admin" || user?.role === "editor"}
      <button
        class="px-3 py-1.5 rounded-md border {currentView === 'users'
          ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
          : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
        on:click={() => (currentView = "users")}
      >
        Users
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'nmos'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "nmos")}
    >
      NMOS
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'topology'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => {
        currentView = "topology";
        loadNMOSRegistry();
      }}
    >
      Topology
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'nmosPatch'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "nmosPatch")}
    >
      NMOS Patch
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'checker'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "checker")}
    >
      Checker
    </button>
    {#if user?.role === "admin" || user?.role === "editor"}
      <button
        class="px-3 py-1.5 rounded-md border {currentView === 'automation'
          ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
          : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
        on:click={() => (currentView = "automation")}
      >
        Automation
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'planner'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "planner")}
    >
      Planner
    </button>
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'addressMap'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "addressMap")}
    >
      Address Map
    </button>
    {#if isAdmin}
      <button
        class="px-3 py-1.5 rounded-md border {currentView === 'logs'
          ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
          : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
        on:click={() => {
          currentView = "logs";
          loadLogs();
        }}
      >
        Logs
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border {currentView === 'settings'
        ? 'border-svelte bg-slate-900 text-svelte-soft shadow-sm'
        : 'border-transparent bg-slate-900/40 text-slate-300 hover:border-slate-600 hover:bg-slate-900/40'}"
      on:click={() => (currentView = "settings")}
    >
      Settings
    </button>
  </nav>

  {#if showBuildModal}
    <!-- Sürüm / build popup -->
    <div class="fixed inset-0 z-40 flex items-center justify-center bg-black/60">
      <div class="w-full max-w-sm rounded-2xl border border-slate-200 bg-white shadow-2xl shadow-slate-900/40 p-5 space-y-3">
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-1">
            <div class="inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1">
              <span class="inline-flex h-2.5 w-2.5 rounded-full bg-emerald-500 shadow-[0_0_0_3px_rgba(16,185,129,0.35)]"></span>
              <span class="text-[11px] font-semibold uppercase tracking-[0.16em] text-emerald-900">
                UI Build Status
              </span>
            </div>
            <p class="text-sm font-semibold text-black mt-1">
              You are running the latest frontend build.
            </p>
            <p class="text-xs text-black/70">
              Current version: <span class="font-semibold">{uiVersion}</span>
            </p>
          </div>
          <button
            class="shrink-0 rounded-full border border-slate-200 bg-slate-50 text-slate-500 hover:bg-slate-100 hover:text-slate-700 w-7 h-7 flex items-center justify-center text-xs"
            on:click={() => (showBuildModal = false)}
            aria-label="Close"
          >
            ✕
          </button>
        </div>
        <div class="flex justify-end">
          <button
            class="px-3 py-1.5 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-black"
            on:click={() => (showBuildModal = false)}
          >
            Close
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if success}
    <p class="text-xs font-semibold text-svelte">{success}</p>
  {/if}

  {#if loading}
    <p class="text-sm text-black">Loading...</p>
  {:else if error}
    <p class="text-sm text-red-600">{error}</p>
  {:else}
    {#if currentView === "dashboard"}
      <DashboardHomeView {summary} {flows} {flowTotal} />
    {/if}

    {#if currentView === "flows"}
      <FlowsView
        {flows}
        bind:flowLimit
        bind:flowOffset
        {flowTotal}
        bind:flowSortBy
        bind:flowSortOrder
        {canEdit}
        {isAdmin}
        onApplyFlowSort={applyFlowSort}
        onPrevFlowPage={prevFlowPage}
        onNextFlowPage={nextFlowPage}
        onToggleFlowLock={toggleFlowLock}
        onDeleteFlow={deleteFlow}
      />
    {/if}

    {#if currentView === "search"}
      <SearchView
        bind:searchTerm
        {searchResults}
        bind:searchLimit
        bind:searchOffset
        {searchTotal}
        onRunSearch={runSearch}
        onPrevSearchPage={prevSearchPage}
        onNextSearchPage={nextSearchPage}
      />
    {/if}

    {#if currentView === "newFlow" && canEdit}
      <NewFlowView {newFlow} onCreateFlow={createFlow} />
    {/if}

    {#if currentView === "users" && (user?.role === "admin" || user?.role === "editor")}
      <UsersView {users} />
    {/if}

    {#if currentView === "settings"}
      <SettingsView
        {settings}
        {isAdmin}
        {canEdit}
        {importing}
        onSaveSetting={saveSetting}
        onExportFlows={exportFlows}
        onImportFlowsFromFile={importFlowsFromFile}
      />
    {/if}

    {#if currentView === "nmos"}
      <NMOSView
        {nmosBaseUrl}
        {nmosResult}
        {nmosIS05Base}
        {flows}
        {selectedNMOSFlow}
        {selectedNMOSReceiver}
        {nmosTakeBusy}
        onBaseUrlChange={(v) => (nmosBaseUrl = v)}
        onDiscoverNMOS={discoverNMOS}
        onIS05BaseChange={(v) => (nmosIS05Base = v)}
        onSelectFlow={(f) => (selectedNMOSFlow = f)}
        onSelectReceiver={(r) => (selectedNMOSReceiver = r)}
        onExecuteTake={executeNMOSTake}
        isTakeReady={isTakeReady}
      />
    {/if}

    {#if currentView === "topology"}
      <TopologyView
        {registryNodes}
        {registryDevices}
        {registrySenders}
        {registryReceivers}
        {selectedRegistryNodeId}
        {selectedRegistryDeviceId}
        onSelectNode={(id) => {
          selectedRegistryNodeId = id;
          selectedRegistryDeviceId = "";
        }}
        onSelectDevice={(id) => (selectedRegistryDeviceId = id)}
        {isPatchTakeReady}
        {selectedPatchSender}
        {selectedPatchReceiver}
        {nmosIS05Base}
        {nmosTakeBusy}
        onSelectPatchSender={(s) => (selectedPatchSender = s)}
        onSelectPatchReceiver={(r) => (selectedPatchReceiver = r)}
        onExecutePatchTake={executePatchTake}
      />
    {/if}

    {#if currentView === "nmosPatch"}
      <NMOSPatchView
        {nmosNodes}
        {selectedSenderNodeId}
        {selectedReceiverNodeId}
        {senderNodeSenders}
        {receiverNodeReceivers}
        {selectedPatchSender}
        {selectedPatchReceiver}
        {nmosIS05Base}
        {nmosPatchStatus}
        {nmosPatchError}
        {nmosTakeBusy}
        {senderFilterText}
        {receiverFilterText}
        {senderFormatFilter}
        {receiverFormatFilter}
        {showAddNodeModal}
        {newNodeName}
        {newNodeUrl}
        onReloadNodes={loadNodes}
        onOpenAddNode={openAddNodeModal}
        onCancelAddNode={cancelAddNode}
        onConfirmAddNode={addNode}
        onChangeNewNodeName={(v) => (newNodeName = v)}
        onChangeNewNodeUrl={(v) => (newNodeUrl = v)}
        onSelectSenderNode={(id) => {
          selectedSenderNodeId = id;
          loadSenderNode(id);
        }}
        onSelectReceiverNode={(id) => {
          selectedReceiverNodeId = id;
          loadReceiverNode(id);
        }}
        onUpdateSenderFilterText={(v) => (senderFilterText = v)}
        onUpdateReceiverFilterText={(v) => (receiverFilterText = v)}
        onUpdateSenderFormatFilter={(v) => (senderFormatFilter = v)}
        onUpdateReceiverFormatFilter={(v) => (receiverFormatFilter = v)}
        onSelectPatchSender={(s) => (selectedPatchSender = s)}
        onSelectPatchReceiver={(r) => (selectedPatchReceiver = r)}
        onExecutePatchTake={executePatchTake}
        {isPatchTakeReady}
        {filterSenders}
        {filterReceivers}
      />
    {/if}

    {#if currentView === "checker"}
      <CheckerView {checkerResult} onRunCollisionCheck={runCollisionCheck} />
    {/if}

    {#if currentView === "automation" && (user?.role === "admin" || user?.role === "editor")}
      <AutomationJobsView {automationJobs} {isAdmin} onToggleAutomationJob={toggleAutomationJob} />
    {/if}

    {#if currentView === "planner"}
      <PlannerView
        {plannerRoots}
        {plannerChildren}
        {selectedPlannerRoot}
        {newPlannerParent}
        {newPlannerChild}
        {canEdit}
        {isAdmin}
        onSelectPlannerRoot={selectPlannerRoot}
        onPlannerQuickEdit={plannerQuickEdit}
        onPlannerDelete={plannerDelete}
        onCreatePlannerParent={createPlannerParent}
        onCreatePlannerChild={createPlannerChild}
        onExportBuckets={exportBuckets}
        onImportBucketsFromFile={importBucketsFromFile}
      />
    {/if}

    {#if currentView === "addressMap"}
      <AddressMapView {addressMap} />
    {/if}

    {#if currentView === "logs" && isAdmin}
      <LogsView bind:logsKind {logsLines} onLoadLogs={loadLogs} />
    {/if}
  {/if}
  </div>
</main>
