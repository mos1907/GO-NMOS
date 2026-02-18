<script>
  import { onMount, onDestroy } from "svelte";
  import { api, apiWithMeta } from "../lib/api.js";
  import { connectMQTT, disconnectMQTT } from "../lib/mqtt.js";

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
      // IS-05 base ayarı
      const base = res.base_url?.replace(/\/$/, "") || node.base_url.replace(/\/$/, "");
      const ver = (res.is04_version || "").replace(/^\//, "");
      nmosIS05Base = `${base}/x-nmos/connection/${ver}`;
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

      // Basit: sender_id + master_enable ile backend'e delege et
      await api(`/flows/${flows[0]?.id || 1}/nmos/apply`, {
        method: "POST",
        token,
        body: {
          connection_url: connectionUrl,
          sender_id: selectedPatchSender.id
        }
      });

      nmosPatchStatus = `TAKE OK: ${selectedPatchSender.label} → ${selectedPatchReceiver.label}`;
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
      <section class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 mb-4">
        <div class="rounded-xl border border-slate-200 bg-white px-3 py-3 shadow-sm">
          <p class="text-[11px] text-black/60 font-medium uppercase tracking-wide">Total</p>
          <p class="mt-1 text-2xl font-semibold text-black">{summary.total}</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-emerald-50 px-3 py-3 shadow-sm">
          <p class="text-[11px] text-emerald-800 font-medium uppercase tracking-wide">Active</p>
          <p class="mt-1 text-2xl font-semibold text-emerald-900">{summary.active}</p>
        </div>
        <div class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-3 shadow-sm">
          <p class="text-[11px] text-amber-800 font-medium uppercase tracking-wide">Locked</p>
          <p class="mt-1 text-2xl font-semibold text-amber-900">{summary.locked}</p>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white px-3 py-3 shadow-sm">
          <p class="text-[11px] text-black/60 font-medium uppercase tracking-wide">Unused</p>
          <p class="mt-1 text-2xl font-semibold text-black">{summary.unused}</p>
        </div>
        <div class="rounded-xl border border-sky-200 bg-sky-50 px-3 py-3 shadow-sm">
          <p class="text-[11px] text-sky-800 font-medium uppercase tracking-wide">Maintenance</p>
          <p class="mt-1 text-2xl font-semibold text-sky-900">{summary.maintenance}</p>
        </div>
      </section>
      <section class="rounded-xl border border-slate-200 bg-white shadow-sm">
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100">
          <div>
            <h3 class="text-sm font-semibold text-black">Latest Flows</h3>
            <p class="text-[11px] text-black/60 mt-0.5">Son eklenen/ güncellenen akışların özeti</p>
          </div>
          <span class="text-[11px] text-black/70 bg-nmos-bg px-2 py-0.5 rounded-full border border-slate-200">
            Showing {Math.min(flows.length, 12)} of {flowTotal}
          </span>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full text-xs">
            <thead class="bg-slate-50">
              <tr>
                <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Display Name</th>
                <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Flow ID</th>
                <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Multicast</th>
                <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Port</th>
                <th class="text-left px-4 py-2 border-b border-slate-200 font-medium text-black/80">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              {#each flows.slice(0, 12) as flow}
                <tr class="hover:bg-slate-50/80">
                  <td class="px-4 py-2 text-black truncate text-[13px] font-medium">{flow.display_name}</td>
                  <td class="px-4 py-2 text-black/70 truncate">{flow.flow_id}</td>
                  <td class="px-4 py-2 text-black">{flow.multicast_ip}</td>
                  <td class="px-4 py-2 text-black">{flow.port}</td>
                  <td class="px-4 py-2">
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium {flow.flow_status === 'active'
                      ? 'bg-emerald-50 text-emerald-800 border border-emerald-200'
                      : flow.flow_status === 'maintenance'
                      ? 'bg-amber-50 text-amber-800 border border-amber-200'
                      : 'bg-slate-50 text-slate-700 border border-slate-200'}">
                      {flow.flow_status}
                    </span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>
    {/if}

    {#if currentView === "flows"}
      <section class="mt-4 space-y-3">
        <div class="flex items-center justify-between gap-3 flex-wrap">
          <div>
            <h3 class="text-sm font-semibold text-black">Flows</h3>
            <p class="text-[11px] text-black/60">Akış listesini filtreleyin ve yönetin</p>
          </div>
          <div class="flex flex-wrap items-center gap-2 text-xs">
            <label class="flex items-center gap-1 text-black/70">
              <span>Sort by</span>
              <select
                bind:value={flowSortBy}
                on:change={applyFlowSort}
                class="px-2 py-1 rounded-md border border-slate-300 bg-white text-xs"
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
            <label class="flex items-center gap-1 text-black/70">
              <span>Order</span>
              <select
                bind:value={flowSortOrder}
                on:change={applyFlowSort}
                class="px-2 py-1 rounded-md border border-slate-300 bg-white text-xs"
              >
                <option value="desc">desc</option>
                <option value="asc">asc</option>
              </select>
            </label>
            <div class="flex items-center gap-1">
              <button
                class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
                on:click={prevFlowPage}
                disabled={flowOffset === 0}
              >
                Prev
              </button>
              <button
                class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
                on:click={nextFlowPage}
                disabled={flowOffset + flowLimit >= flowTotal}
              >
                Next
              </button>
            </div>
            <span class="text-[11px] text-black/60">
              Showing {flowOffset + 1}-{Math.min(flowOffset + flowLimit, flowTotal)} / {flowTotal}
            </span>
          </div>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-slate-50">
              <tr>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Display Name</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Flow ID</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Multicast</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Source</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Port</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Status</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Availability</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Locked</th>
                {#if canEdit}
                  <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Action</th>
                {/if}
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              {#each flows as flow}
                <tr class="hover:bg-slate-50/80">
                  <td class="px-3 py-2 text-[13px] font-medium text-black">{flow.display_name}</td>
                  <td class="px-3 py-2 text-black/70">{flow.flow_id}</td>
                  <td class="px-3 py-2 text-black">{flow.multicast_ip}</td>
                  <td class="px-3 py-2 text-black">{flow.source_ip}</td>
                  <td class="px-3 py-2 text-black">{flow.port}</td>
                  <td class="px-3 py-2 text-black/80">{flow.flow_status}</td>
                  <td class="px-3 py-2 text-black/80">{flow.availability}</td>
                  <td class="px-3 py-2 text-lg">{flow.locked ? "🔒" : "🔓"}</td>
                  {#if canEdit}
                    <td class="px-3 py-2">
                      <div class="flex flex-wrap gap-1.5">
                        <button
                          class="px-2.5 py-1 rounded-md border border-slate-300 bg-white text-[11px] hover:bg-slate-50"
                          on:click={() => toggleFlowLock(flow)}
                        >
                          {flow.locked ? "Unlock" : "Lock"}
                        </button>
                        {#if isAdmin}
                          <button
                            class="px-2.5 py-1 rounded-md border border-red-300 bg-red-50 text-[11px] text-red-700 hover:bg-red-100"
                            on:click={() => deleteFlow(flow)}
                          >
                            Delete
                          </button>
                        {/if}
                      </div>
                    </td>
                  {/if}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>
    {/if}

    {#if currentView === "search"}
      <section class="mt-4 space-y-3">
        <div class="flex items-center justify-between gap-3 flex-wrap">
          <div>
            <h3 class="text-sm font-semibold text-black">Quick Search</h3>
            <p class="text-[11px] text-black/60">Akışları isim, IP veya flow ID ile hızlıca bulun</p>
          </div>
          <span class="text-[11px] text-black/60">
            {searchTotal > 0
              ? `${searchOffset + 1}-${Math.min(searchOffset + searchLimit, searchTotal)} / ${searchTotal}`
              : "0 result"}
          </span>
        </div>
        <div class="flex flex-wrap gap-2 items-center">
          <input
            bind:value={searchTerm}
            placeholder="Search by name/ip/flow id/note..."
            class="px-3 py-2 rounded-md border border-slate-300 bg-white text-sm min-w-[260px] flex-1"
          />
          <button
            class="px-3 py-2 rounded-md bg-slate-900 text-white text-xs font-semibold hover:bg-black"
            on:click={runSearch}
          >
            Search
          </button>
          <button
            class="px-2.5 py-1.5 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
            on:click={prevSearchPage}
            disabled={searchOffset === 0}
          >
            Prev
          </button>
          <button
            class="px-2.5 py-1.5 rounded-md border border-slate-300 bg-white text-xs hover:bg-slate-50 disabled:opacity-40"
            on:click={nextSearchPage}
            disabled={searchOffset + searchLimit >= searchTotal}
          >
            Next
          </button>
        </div>
        <div class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-slate-50">
              <tr>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Display Name</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Flow ID</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Multicast</th>
                <th class="text-left border-b border-slate-200 px-3 py-2 font-medium text-black/80">Port</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              {#each searchResults as flow}
                <tr class="hover:bg-slate-50/80">
                  <td class="px-3 py-2 text-[13px] font-medium text-black">{flow.display_name}</td>
                  <td class="px-3 py-2 text-black/70">{flow.flow_id}</td>
                  <td class="px-3 py-2 text-black">{flow.multicast_ip}</td>
                  <td class="px-3 py-2 text-black">{flow.port}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>
    {/if}

    {#if currentView === "newFlow" && canEdit}
      <h3>Create New Flow</h3>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:10px;">
        <input bind:value={newFlow.display_name} placeholder="Display name" />
        <input bind:value={newFlow.multicast_ip} placeholder="Multicast IP" />
        <input bind:value={newFlow.source_ip} placeholder="Source IP" />
        <input type="number" bind:value={newFlow.port} placeholder="Port" />
        <select bind:value={newFlow.flow_status}>
          <option value="active">active</option>
          <option value="unused">unused</option>
          <option value="maintenance">maintenance</option>
        </select>
        <select bind:value={newFlow.availability}>
          <option value="available">available</option>
          <option value="lost">lost</option>
          <option value="maintenance">maintenance</option>
        </select>
        <input bind:value={newFlow.transport_protocol} placeholder="Transport protocol" />
        <input bind:value={newFlow.note} placeholder="Note" />
      </div>
      <div style="margin-top:12px;">
        <button on:click={createFlow}>Create flow</button>
      </div>
    {/if}

    {#if currentView === "users" && (user?.role === "admin" || user?.role === "editor")}
      <h3>Users</h3>
      <table style="width:100%;border-collapse:collapse;">
        <thead>
          <tr>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Username</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Role</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Created</th>
          </tr>
        </thead>
        <tbody>
          {#each users as u}
            <tr>
              <td style="border-bottom:1px solid #eee;padding:8px;">{u.username}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{u.role}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{u.created_at}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}

    {#if currentView === "settings"}
      <h3>Settings & Backup</h3>
      <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:10px;">
        <label>api_base_url
          <input bind:value={settings.api_base_url} />
          {#if isAdmin}<button on:click={() => saveSetting("api_base_url")}>Save</button>{/if}
        </label>
        <label>anonymous_access
          <input bind:value={settings.anonymous_access} />
          {#if isAdmin}<button on:click={() => saveSetting("anonymous_access")}>Save</button>{/if}
        </label>
        <label>flow_lock_role
          <input bind:value={settings.flow_lock_role} />
          {#if isAdmin}<button on:click={() => saveSetting("flow_lock_role")}>Save</button>{/if}
        </label>
        <label>hard_delete_enabled
          <input bind:value={settings.hard_delete_enabled} />
          {#if isAdmin}<button on:click={() => saveSetting("hard_delete_enabled")}>Save</button>{/if}
        </label>
      </div>
      <div style="margin-top:16px;display:flex;gap:8px;align-items:center;">
        <button on:click={exportFlows}>Export flows JSON</button>
        {#if canEdit}
          <label style="display:inline-flex;align-items:center;gap:8px;">
            <span>Import JSON:</span>
            <input type="file" accept="application/json" on:change={importFlowsFromFile} disabled={importing} />
          </label>
        {/if}
      </div>
    {/if}

    {#if currentView === "nmos"}
      <div class="space-y-4">
        <div class="flex flex-wrap gap-3 items-end">
          <div class="flex flex-col gap-1">
            <label class="text-sm font-medium text-slate-300">NMOS Node Base URL</label>
            <input
              bind:value={nmosBaseUrl}
              placeholder="http://192.168.x.x:port"
              class="px-3 py-2 rounded-md bg-slate-900 border border-slate-700 text-sm min-w-[320px]"
            />
          </div>
          <button
            class="px-4 py-2 rounded-md bg-svelte hover:bg-orange-500 text-sm font-semibold text-white"
            on:click={discoverNMOS}
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
              <p class="text-sm">Senders: {nmosResult.counts?.senders} | Receivers: {nmosResult.counts?.receivers} | Flows: {nmosResult.counts?.flows}</p>
            </div>
            <div class="rounded-xl bg-slate-900/60 border border-slate-800 p-4 space-y-2">
              <label class="text-xs text-slate-400">IS-05 Base URL</label>
              <input
                bind:value={nmosIS05Base}
                class="w-full px-3 py-2 rounded-md bg-slate-950 border border-slate-700 text-xs"
              />
              <p class="text-[11px] text-slate-500">Genellikle: base_url + /x-nmos/connection/&lt;version&gt;</p>
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
                    on:click={() => (selectedNMOSFlow = f)}
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
                on:click={executeNMOSTake}
                disabled={!isTakeReady()}
              >
                <svg class="w-9 h-9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                  <polyline points="9 18 15 12 9 6" />
                </svg>
                <span>{nmosTakeBusy ? "TAKING..." : "TAKE"}</span>
              </button>
              <div class="flex items-center gap-2 text-sm text-black px-4 py-2 rounded-full bg-svelte/20 border border-svelte/40">
                <span class="inline-flex h-3 w-3 rounded-full {isTakeReady() ? 'bg-svelte' : 'bg-black/30'}"></span>
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
                    on:click={() => (selectedNMOSReceiver = r)}
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
    {/if}

    {#if currentView === "nmosPatch"}
      <div class="space-y-4">
        <!-- Üst bar: title + node/actions (nmos-patch-gui tarzı) -->
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <h3 class="text-black font-semibold text-lg">NMOS Patch</h3>
          </div>
          <div class="flex items-center gap-2">
            <button class="px-3 py-1.5 rounded-md text-sm bg-nmos-bg hover:bg-svelte/20 border border-svelte/40 text-black font-medium" on:click={loadNodes}>
              Reload Nodes
            </button>
            <button class="px-3 py-1.5 rounded-md text-sm bg-svelte hover:bg-orange-500 text-black font-semibold" on:click={openAddNodeModal}>
              Add Node
            </button>
          </div>
        </div>

        <!-- Durum satırı (status + CORS link benzeri) -->
        <div class="flex flex-wrap items-center gap-2 text-sm">
          <div class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-svelte/20 border border-svelte/40">
            <span class="inline-flex h-2.5 w-2.5 rounded-full {isPatchTakeReady() ? 'bg-svelte' : 'bg-black/30'}"></span>
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
          <!-- Source panel (sol) -->
          <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col min-h-[500px]">
            <div class="flex items-center justify-between mb-3 gap-2">
              <div>
                <h4 class="text-base font-semibold text-black">Sources</h4>
                <p class="text-[12px] text-black/70">Source selection</p>
              </div>
              <select
                bind:value={selectedSenderNodeId}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
                on:change={(e) => loadSenderNode(e.target.value)}
              >
                <option value="">Select node…</option>
                {#each nmosNodes as node}
                  <option value={node.id}>{node.name}</option>
                {/each}
              </select>
            </div>
            <!-- Filtreler -->
            <div class="flex flex-wrap gap-2 mb-3 text-sm">
              <input
                bind:value={senderFilterText}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 flex-1 min-w-[150px] text-black"
                placeholder="Search sources..."
              />
              <select
                bind:value={senderFormatFilter}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-black"
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
                {#each filterSenders(senderNodeSenders) as s}
                  <button
                    type="button"
                    class="w-full text-left px-3 py-2 hover:bg-svelte/20 flex justify-between gap-2 {selectedPatchSender && selectedPatchSender.id === s.id
                      ? 'bg-svelte/30 border-l-4 border-svelte'
                      : ''}"
                    on:click={() => (selectedPatchSender = s)}
                  >
                    <span class="truncate text-black font-medium">{s.label}</span>
                    <span class="text-[12px] text-black/60 truncate">{s.flow_id}</span>
                  </button>
                {/each}
              {/if}
            </div>
          </div>

          <!-- TAKE butonu (orta) -->
          <div class="flex flex-col items-center justify-center gap-4">
            <div class="rounded-2xl border border-slate-300 bg-white shadow-lg shadow-orange-500/20 p-3">
              <button
                class="w-40 h-40 rounded-xl bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-svelte-soft font-bold text-xl flex flex-col items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed hover:scale-105 active:scale-100 transition transform"
                on:click={executePatchTake}
                disabled={!isPatchTakeReady()}
              >
                <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-svelte to-orange-400 flex items-center justify-center shadow-inner">
                  <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <polyline points="9 18 15 12 9 6" />
                  </svg>
                </div>
                <span class="tracking-wide">{nmosTakeBusy ? "TAKING..." : "TAKE"}</span>
                <span class="text-[11px] font-medium text-slate-300">
                  {#if !isPatchTakeReady()}
                    Select endpoints to enable
                  {:else}
                    Ready to patch
                  {/if}
                </span>
              </button>
            </div>
            <div class="flex items-center gap-2 text-sm text-black px-4 py-2 rounded-full bg-svelte/20 border border-svelte/40">
              <span class="inline-flex h-3 w-3 rounded-full {isPatchTakeReady() ? 'bg-svelte' : 'bg-black/30'}"></span>
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

          <!-- Destination panel (sağ) -->
          <div class="rounded-xl bg-nmos-bg border border-svelte/40 p-4 flex flex-col min-h-[500px]">
            <div class="flex items-center justify-between mb-3 gap-2">
              <div>
                <h4 class="text-base font-semibold text-black">Destinations</h4>
                <p class="text-[12px] text-black/70">Destination selection</p>
              </div>
              <select
                bind:value={selectedReceiverNodeId}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
                on:change={(e) => loadReceiverNode(e.target.value)}
              >
                <option value="">Select node…</option>
                {#each nmosNodes as node}
                  <option value={node.id}>{node.name}</option>
                {/each}
              </select>
            </div>
            <!-- Filtreler -->
            <div class="flex flex-wrap gap-2 mb-3 text-sm">
              <input
                bind:value={receiverFilterText}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 flex-1 min-w-[150px] text-black"
                placeholder="Search destinations..."
              />
              <select
                bind:value={receiverFormatFilter}
                class="px-3 py-1.5 rounded-md bg-nmos-bg border border-svelte/40 text-black"
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
                {#each filterReceivers(receiverNodeReceivers) as r}
                  <button
                    type="button"
                    class="w-full text-left px-3 py-2 hover:bg-svelte/20 flex justify-between gap-2 {selectedPatchReceiver && selectedPatchReceiver.id === r.id
                      ? 'bg-svelte/30 border-l-4 border-svelte'
                      : ''}"
                    on:click={() => (selectedPatchReceiver = r)}
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
                    bind:value={newNodeName}
                    class="px-3 py-2 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
                    placeholder="e.g. Camera Router"
                  />
                </div>
                <div class="flex flex-col gap-1">
                  <label class="text-sm text-black/80 font-medium">IS-04 URL</label>
                  <input
                    bind:value={newNodeUrl}
                    class="px-3 py-2 rounded-md bg-nmos-bg border border-svelte/40 text-sm text-black"
                    placeholder="http://192.168.x.x:port"
                  />
                  <p class="text-xs text-black/60">
                    NMOS Node IS-04 URL. IS-05 endpoint otomatik hesaplanır.
                  </p>
                </div>
              </div>
              <div class="flex justify-end gap-2 pt-2">
                <button class="px-3 py-1.5 rounded-md text-sm bg-nmos-bg border border-svelte/40 text-black font-medium hover:bg-svelte/10" on:click={cancelAddNode}>
                  Cancel
                </button>
                <button class="px-3 py-1.5 rounded-md text-sm bg-svelte hover:bg-orange-500 text-black font-semibold" on:click={addNode}>
                  Add Node
                </button>
              </div>
            </div>
          </div>
        {/if}
      </div>
    {/if}

    {#if currentView === "checker"}
      <h3>Collision Checker</h3>
      <button on:click={runCollisionCheck}>Run collision check now</button>
      {#if checkerResult}
        <p style="margin-top:10px;">Total collisions: {checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0}</p>
        {#if checkerResult.result?.items || checkerResult.items}
          <table style="width:100%;border-collapse:collapse;">
            <thead>
              <tr>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Multicast IP</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Port</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Count</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Flows</th>
              </tr>
            </thead>
            <tbody>
              {#each (checkerResult.result?.items || checkerResult.items || []) as item}
                <tr>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.multicast_ip}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.port}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.count}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{(item.flow_names || []).join(", ")}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      {/if}
    {/if}

    {#if currentView === "automation" && (user?.role === "admin" || user?.role === "editor")}
      <h3>Automation Jobs</h3>
      <table style="width:100%;border-collapse:collapse;">
        <thead>
          <tr>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Job ID</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Type</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Schedule</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Enabled</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Action</th>
          </tr>
        </thead>
        <tbody>
          {#each automationJobs as job}
            <tr>
              <td style="border-bottom:1px solid #eee;padding:8px;">{job.job_id}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{job.job_type}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{job.schedule_type}: {job.schedule_value}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{job.enabled ? "ON" : "OFF"}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">
                {#if isAdmin}
                  {#if job.enabled}
                    <button on:click={() => toggleAutomationJob(job, false)}>Disable</button>
                  {:else}
                    <button on:click={() => toggleAutomationJob(job, true)}>Enable</button>
                  {/if}
                {:else}
                  <span>-</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}

    {#if currentView === "planner"}
      <h3>Planner Buckets</h3>
      <div style="display:grid;grid-template-columns:1fr 2fr;gap:12px;">
        <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
          <h4>Drives</h4>
          {#each plannerRoots as root}
            <div style="display:flex;justify-content:space-between;margin:4px 0;">
              <button on:click={() => selectPlannerRoot(root)}>{root.name}</button>
              <small>{root.cidr}</small>
            </div>
          {/each}
        </div>
        <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
          <h4>Folders / Views</h4>
          {#if selectedPlannerRoot}
            <p>Selected drive: <strong>{selectedPlannerRoot.name}</strong></p>
          {/if}
          <table style="width:100%;border-collapse:collapse;">
            <thead>
              <tr>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Name</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Type</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">CIDR</th>
                <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Description</th>
                {#if canEdit}
                  <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Action</th>
                {/if}
              </tr>
            </thead>
            <tbody>
              {#each plannerChildren as item}
                <tr>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.name}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.bucket_type}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.cidr}</td>
                  <td style="border-bottom:1px solid #eee;padding:8px;">{item.description}</td>
                  {#if canEdit}
                    <td style="border-bottom:1px solid #eee;padding:8px;">
                      <button on:click={() => plannerQuickEdit(item)}>Edit</button>
                      {#if isAdmin}
                        <button on:click={() => plannerDelete(item)} style="margin-left:6px;">Delete</button>
                      {/if}
                    </td>
                  {/if}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
      {#if canEdit}
        <div style="margin-top:12px;display:grid;grid-template-columns:1fr 1fr;gap:12px;">
          <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
            <h4>Create Folder (parent)</h4>
            <input bind:value={newPlannerParent.name} placeholder="Name" />
            <input bind:value={newPlannerParent.cidr} placeholder="CIDR (e.g. 239.1.0.0/16)" />
            <input bind:value={newPlannerParent.description} placeholder="Description" />
            <input bind:value={newPlannerParent.color} placeholder="Color (optional)" />
            <button on:click={createPlannerParent} style="margin-top:8px;">Create Parent</button>
          </div>
          <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
            <h4>Create View (child)</h4>
            <input bind:value={newPlannerChild.name} placeholder="Name" />
            <input bind:value={newPlannerChild.cidr} placeholder="CIDR or range label" />
            <input bind:value={newPlannerChild.description} placeholder="Description" />
            <input bind:value={newPlannerChild.color} placeholder="Color (optional)" />
            <button on:click={() => createPlannerChild(selectedPlannerRoot)} style="margin-top:8px;" disabled={!selectedPlannerRoot}>Create Child</button>
          </div>
        </div>
      {/if}
      <div style="margin-top:12px;display:flex;gap:8px;">
        <button on:click={exportBuckets}>Export Planner</button>
        {#if canEdit}
          <label style="display:inline-flex;align-items:center;gap:8px;">
            <span>Import Planner:</span>
            <input type="file" accept="application/json" on:change={importBucketsFromFile} />
          </label>
        {/if}
      </div>
    {/if}

    {#if currentView === "addressMap"}
      <h3>Address Map (/24 buckets)</h3>
      <p>Total subnets: {addressMap?.total_subnets || 0}</p>
      <table style="width:100%;border-collapse:collapse;">
        <thead>
          <tr>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Subnet</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Flow Count</th>
            <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">IPs</th>
          </tr>
        </thead>
        <tbody>
          {#each (addressMap?.items || []) as b}
            <tr>
              <td style="border-bottom:1px solid #eee;padding:8px;">{b.subnet}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{b.count}</td>
              <td style="border-bottom:1px solid #eee;padding:8px;">{Object.keys(b.flows || {}).slice(0, 6).join(", ")}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}

    {#if currentView === "logs" && isAdmin}
      <h3>Logs</h3>
      <div style="display:flex;gap:8px;align-items:center;margin-bottom:8px;">
        <select bind:value={logsKind}>
          <option value="api">api</option>
          <option value="audit">audit</option>
        </select>
        <button on:click={loadLogs}>Refresh Logs</button>
        <a href={`${location.protocol}//${location.hostname}:9090/api/logs/download?kind=${logsKind}`} target="_blank" rel="noreferrer">
          <button>Download</button>
        </a>
      </div>
      <pre style="background:#111;color:#d7ffd7;padding:12px;border-radius:8px;max-height:420px;overflow:auto;">{logsLines.join("\n")}</pre>
    {/if}
  {/if}
</main>
