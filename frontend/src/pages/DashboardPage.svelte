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
  import UsersView from "../components/UsersView.svelte";
  import SettingsView from "../components/SettingsView.svelte";
  import CheckerView from "../components/CheckerView.svelte";
  import AutomationJobsView from "../components/AutomationJobsView.svelte";
  import PlannerView from "../components/PlannerView.svelte";
  import AddressMapView from "../components/AddressMapView.svelte";
  import LogsView from "../components/LogsView.svelte";
  import PortExplorerView from "../components/PortExplorerView.svelte";
  import SkeletonLoader from "../components/SkeletonLoader.svelte";
  import EmptyState from "../components/EmptyState.svelte";

  let {
    token,
    user,
    onLogout,
  } = $props();

  let currentView = $state("dashboard");
  let loading = $state(true);
  let error = $state("");
  let success = $state("");

  let summary = $state({ total: 0, active: 0, locked: 0, unused: 0, maintenance: 0 });
  let flows = $state([]);
  let flowLimit = $state(50);
  let flowOffset = $state(0);
  let flowTotal = $state(0);
  let flowSortBy = $state("updated_at");
  let flowSortOrder = $state("desc");
  let users = $state([]);
  let settings = $state({});
  let searchTerm = $state("");
  let searchResults = $state([]);
  let searchLimit = $state(50);
  let searchOffset = $state(0);
  let searchTotal = $state(0);
  let importing = $state(false);
  let nmosBaseUrl = $state("");
  let nmosResult = $state(null);
  let nmosIS05Base = $state("");
  let selectedNMOSFlow = $state(null);
  let selectedNMOSReceiver = $state(null);
  let nmosTakeBusy = $state(false);
  let checkerResult = $state(null);
  let automationJobs = $state([]);
  let automationSummary = $state(null);
  let systemInfo = $state(null);
  let addressMap = $state(null);
  let logsKind = $state("api");
  let logsLines = $state([]);

  // Diagnostics / Health panel state
  let healthDetail = $state(null);
  let healthLoading = $state(false);
  let healthError = $state("");
  let lastHealthLoadedAt = $state("");

  // Diagnostics: Check Node at URL
  let nodeCheckUrl = $state("");
  let nodeCheckLoading = $state(false);
  let nodeCheckError = $state("");
  let nodeCheckResult = $state(null);

  // Internal NMOS registry (IS-04 style) state
  let registryNodes = $state([]);
  let registryDevices = $state([]);
  let registrySenders = $state([]);
  let registryReceivers = $state([]);
  let registryHealth = $state(null);
  let selectedRegistryNodeId = $state("");
  let selectedRegistryDeviceId = $state("");

  // NMOS Patch-style view state (sender/receiver selection)
  let nmosNodes = $state([]);
  let showAddNodeModal = $state(false);
  let newNodeName = $state("");
  let newNodeUrl = $state("");
  let selectedSenderNodeId = $state("");
  let selectedReceiverNodeId = $state("");
  let senderNodeSenders = $state([]);
  let receiverNodeReceivers = $state([]);
  let selectedPatchSender = $state(null);
  let selectedPatchReceiver = $state(null);
  let nmosPatchStatus = $state("");
  let nmosPatchError = $state("");
  let senderFilterText = $state("");
  let receiverFilterText = $state("");
  let senderFormatFilter = $state("");
  let receiverFormatFilter = $state("");
  // RDS (Registry) connect modal state
  let showConnectRDSModal = $state(false);
  let rdsQueryUrl = $state("");
  let rdsDiscovering = $state(false);
  let rdsNodes = $state([]);
  let rdsSelectedIds = $state([]);
  let rdsError = $state("");
  let plannerRoots = $state([]);
  let plannerChildren = $state([]);
  let selectedPlannerRoot = $state(null);
  let newPlannerParent = $state({ parent_id: null, name: "", cidr: "", description: "", color: "" });
  let newPlannerChild = $state({ parent_id: null, name: "", cidr: "", description: "", color: "" });

  let newFlow = $state({
    display_name: "",
    multicast_ip: "",
    source_ip: "",
    port: 5004,
    flow_status: "active",
    availability: "available",
    transport_protocol: "RTP/UDP",
    note: "",
    alias_1: "",
    alias_2: "",
    alias_3: "",
    alias_4: "",
    alias_5: "",
    alias_6: "",
    alias_7: "",
    alias_8: "",
    user_field_1: "",
    user_field_2: "",
    user_field_3: "",
    user_field_4: "",
    user_field_5: "",
    user_field_6: "",
    user_field_7: "",
    user_field_8: "",
  });

  let editingFlow = $state(null);

  const isAdmin = user?.role === "admin";
  const canEdit = user?.role === "admin" || user?.role === "editor";

  // Basit UI sürüm bilgisi (frontend build versiyonu)
  const uiVersion = "go-NMOS UI v0.2.0 (router beta)";
  let showBuildModal = $state(true);

  async function loadDashboard() {
    loading = true;
    error = "";
    try {
      const [sum, data, sys] = await Promise.all([
        api("/flows/summary", { token }),
        loadFlows(),
        api("/system", { token }).catch(() => systemInfo)
      ]);
      summary = sum;
      flows = data;
      systemInfo = sys || systemInfo;
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

  async function loadHealthDetail() {
    healthError = "";
    healthLoading = true;
    try {
      const res = await api("/health/detail", { token });
      healthDetail = res;
      lastHealthLoadedAt = new Date().toISOString();
    } catch (e) {
      healthError = e.message;
    } finally {
      healthLoading = false;
    }
  }

  async function checkNodeAtUrl() {
    const url = nodeCheckUrl.trim();
    if (!url) {
      nodeCheckError = "URL is required";
      nodeCheckResult = null;
      return;
    }
    nodeCheckError = "";
    nodeCheckLoading = true;
    nodeCheckResult = null;
    try {
      const res = await api("/health/check-node", {
        method: "POST",
        token,
        body: { url, timeout_sec: 5 },
      });
      nodeCheckResult = res;
    } catch (e) {
      nodeCheckError = e.message;
    } finally {
      nodeCheckLoading = false;
    }
  }

  async function refreshAll() {
    success = "";
    await loadDashboard();
    await loadHealthDetail().catch(() => {});
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
      resetNewFlowForm();
      await refreshAll();
    } catch (e) {
      error = e.message;
    }
  }

  async function updateFlow() {
    if (!editingFlow) return;
    error = "";
    success = "";
    try {
      await api(`/flows/${editingFlow.id}`, {
        method: "PATCH",
        token,
        body: newFlow,
      });
      success = "Flow updated successfully.";
      editingFlow = null;
      resetNewFlowForm();
      await refreshAll();
    } catch (e) {
      error = e.message;
    }
  }

  function resetNewFlowForm() {
    newFlow = {
      display_name: "",
      multicast_ip: "",
      source_ip: "",
      port: 5004,
      flow_status: "active",
      availability: "available",
      transport_protocol: "RTP/UDP",
      note: "",
      sdp_url: "",
      sdp_cache: "",
      alias_1: "",
      alias_2: "",
      alias_3: "",
      alias_4: "",
      alias_5: "",
      alias_6: "",
      alias_7: "",
      alias_8: "",
      user_field_1: "",
      user_field_2: "",
      user_field_3: "",
      user_field_4: "",
      user_field_5: "",
      user_field_6: "",
      user_field_7: "",
      user_field_8: "",
    };
  }

  function openEditFlowModal(flow) {
    if (!flow) return;
    editingFlow = flow;
    // Copy flow data to newFlow for editing
    newFlow = {
      display_name: flow.display_name || "",
      multicast_ip: flow.multicast_ip || "",
      source_ip: flow.source_ip || "",
      port: flow.port || 5004,
      flow_status: flow.flow_status || "active",
      availability: flow.availability || "available",
      transport_protocol: flow.transport_protocol || "",
      note: flow.note || "",
      sdp_url: flow.sdp_url || "",
      sdp_cache: flow.sdp_cache || "",
      alias_1: flow.alias_1 || "",
      alias_2: flow.alias_2 || "",
      alias_3: flow.alias_3 || "",
      alias_4: flow.alias_4 || "",
      alias_5: flow.alias_5 || "",
      alias_6: flow.alias_6 || "",
      alias_7: flow.alias_7 || "",
      alias_8: flow.alias_8 || "",
      user_field_1: flow.user_field_1 || "",
      user_field_2: flow.user_field_2 || "",
      user_field_3: flow.user_field_3 || "",
      user_field_4: flow.user_field_4 || "",
      user_field_5: flow.user_field_5 || "",
      user_field_6: flow.user_field_6 || "",
      user_field_7: flow.user_field_7 || "",
      user_field_8: flow.user_field_8 || "",
    };
  }

  async function checkFlowNMOS(flowId, baseUrl) {
    error = "";
    try {
      const result = await api(`/flows/${flowId}/nmos/check?base_url=${encodeURIComponent(baseUrl)}`, {
        token,
      });
      return result;
    } catch (e) {
      throw new Error(e.message || "Failed to check flow");
    }
  }

  async function fetchFlowSDP(flowId, manifestUrl) {
    const result = await api(`/flows/${flowId}/fetch-sdp`, {
      method: "POST",
      token,
      body: { manifest_url: manifestUrl },
    });
    return result;
  }

  async function syncFlowFromNMOS(flowId, is04BaseUrl, is05BaseUrl) {
    const result = await api(`/flows/${flowId}/nmos/sync`, {
      method: "POST",
      token,
      body: {
        is04_base_url: is04BaseUrl || "",
        is05_base_url: is05BaseUrl || "",
        timeout: 6,
        fields: [
          "data_source",
          "nmos_node_id",
          "nmos_flow_id",
          "nmos_sender_id",
          "nmos_device_id",
          "nmos_node_label",
          "nmos_node_description",
          "nmos_is04_base_url",
          "nmos_is05_base_url",
          "nmos_is04_version",
          "sdp_url",
          "sdp_cache",
          "media_type",
          "redundancy_group",
          "transport_protocol",
          "st2110_format",
          "source_addr_a",
          "source_port_a",
          "multicast_addr_a",
          "group_port_a",
          "source_addr_b",
          "source_port_b",
          "multicast_addr_b",
          "group_port_b"
        ],
      },
    });
    await refreshAll();
    return result;
  }

  async function checkIS05Receiver(flowId, is05BaseUrl, receiverId) {
    const result = await api(`/flows/${flowId}/is05/receiver-check`, {
      method: "POST",
      token,
      body: {
        is05_base_url: is05BaseUrl || "",
        receiver_id: receiverId || "",
        timeout_sec: 6
      }
    });
    return result;
  }

  async function toggleFlowLock(flow) {
    error = "";
    success = ""; // Clear success message - popup will show instead
    try {
      const result = await api(`/flows/${flow.id}/lock`, {
        method: "POST",
        token,
        body: { locked: !flow.locked }
      });
      // Update flow object
      const updatedFlow = { ...flow, locked: result.locked };
      // Don't set success message - popup will show instead
      await loadFlows().then((data) => {
        flows = data;
      });
      await api("/flows/summary", { token }).then((s) => {
        summary = s;
      });
      // Return result for popup display
      return { locked: result.locked, flow: updatedFlow };
    } catch (e) {
      error = e.message;
      return null;
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

  async function updateUser(username, formData) {
    error = "";
    success = "";
    try {
      const body = {};
      if (formData.password) {
        body.password = formData.password;
      }
      if (formData.role) {
        body.role = formData.role;
      }
      await api(`/users/${username}`, {
        method: "PATCH",
        token,
        body,
      });
      success = `User '${username}' updated successfully.`;
      await loadUsers();
    } catch (e) {
      error = e.message;
    }
  }

  async function deleteUser(username) {
    error = "";
    success = "";
    try {
      await api(`/users/${username}`, {
        method: "DELETE",
        token,
      });
      success = `User '${username}' deleted successfully.`;
      await loadUsers();
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
  const RDS_QUERY_KEY = "go_nmos_rds_query_url";

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

  function openRDSModal() {
    try {
      rdsQueryUrl = localStorage.getItem(RDS_QUERY_KEY) || rdsQueryUrl;
    } catch {
      // ignore
    }
    rdsError = "";
    rdsNodes = [];
    rdsSelectedIds = [];
    showConnectRDSModal = true;
  }

  function closeRDSModal() {
    showConnectRDSModal = false;
    rdsDiscovering = false;
  }

  async function discoverRegistryNodes() {
    rdsError = "";
    rdsNodes = [];
    rdsSelectedIds = [];
    rdsDiscovering = true;
    const q = rdsQueryUrl.trim();
    try {
      try {
        localStorage.setItem(RDS_QUERY_KEY, q);
      } catch {
        // ignore
      }
      const res = await api("/nmos/registry/discover-nodes", {
        method: "POST",
        token,
        body: { query_url: q }
      });
      rdsNodes = res.nodes || [];
      rdsSelectedIds = (rdsNodes || []).map((n) => n.id).filter(Boolean);
    } catch (e) {
      rdsError = e.message;
    } finally {
      rdsDiscovering = false;
    }
  }

  function toggleRegistryNode(id) {
    if (!id) return;
    const set = new Set(rdsSelectedIds || []);
    if (set.has(id)) set.delete(id);
    else set.add(id);
    rdsSelectedIds = Array.from(set);
  }

  function selectAllRegistryNodes() {
    rdsSelectedIds = (rdsNodes || []).map((n) => n.id).filter(Boolean);
  }

  function addSelectedRegistryNodes() {
    const selected = new Set(rdsSelectedIds || []);
    const candidates = (rdsNodes || []).filter((n) => selected.has(n.id));
    if (candidates.length === 0) return;

    const norm = (s) => (s || "").trim().replace(/\/$/, "");
    const existing = new Set((nmosNodes || []).map((n) => norm(n.base_url)));

    const toAdd = [];
    for (const n of candidates) {
      const base = norm(n.base_url || "");
      if (!base) continue;
      if (existing.has(base)) continue;
      existing.add(base);
      toAdd.push({
        id: `rds-${n.id || Date.now()}-${Math.random().toString(16).slice(2)}`,
        name: (n.label || n.id || base).trim(),
        base_url: base
      });
    }

    if (toAdd.length === 0) {
      rdsError = "No new nodes to add (duplicates or missing base_url).";
      return;
    }

    nmosNodes = [...nmosNodes, ...toAdd];
    saveNodes();
    showConnectRDSModal = false;
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
    automationSummary = await api("/automation/summary", { token }).catch(() => automationSummary);
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
      registryHealth = await api("/nmos/registry/health", { token }).catch(() => registryHealth);
    } catch (e) {
      console.error("Failed to load NMOS registry", e);
    }
  }

  let realtimeEvents = $state([]);
  let registryEvents = $state([]);

  function handleMQTTEvent(event) {
    if (event && event.timestamp) {
      // Prepend newest event, keep max 50
      realtimeEvents = [event, ...realtimeEvents].slice(0, 50);
    }
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
    
    // Connect to MQTT using backend realtime config (with graceful fallback)
    (async () => {
      try {
        const cfg = await api("/realtime/config", { token });
        if (cfg?.mqtt_enabled && cfg.ws_url && cfg.topic_prefix) {
          connectMQTT(cfg.ws_url, cfg.topic_prefix, handleMQTTEvent);
          return;
        }
      } catch (e) {
        console.log("Realtime config not available, falling back to default MQTT settings:", e.message);
      }

      try {
        const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const wsHost = window.location.hostname;
        const wsPort = "9001";
        const wsUrl = `${wsProtocol}//${wsHost}:${wsPort}`;
        const topicPrefix = "go-nmos/flows/events";
        connectMQTT(wsUrl, topicPrefix, handleMQTTEvent);
      } catch (e) {
        console.log("MQTT connection skipped:", e.message);
      }
    })();

    // Connect to registry WebSocket feed (no auth, informational only)
    try {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const host = window.location.host;
      const ws = new WebSocket(`${proto}//${host}/ws/registry`);
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          registryEvents = [data, ...registryEvents].slice(0, 50);
        } catch (e) {
          console.error("Registry WS parse error:", e);
        }
      };
      ws.onerror = (e) => {
        console.log("Registry WS error:", e);
      };
      ws.onclose = () => {
        console.log("Registry WS closed");
      };
    } catch (e) {
      console.log("Registry WS connection skipped:", e.message);
    }
  });

  onDestroy(() => {
    disconnectMQTT();
  });
</script>

<main class="min-h-screen bg-[#0a0d14] text-gray-100">
    <div class="max-w-7xl mx-auto px-6 py-6 space-y-6">
  <header class="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-gray-800">
    <div class="space-y-2">
      <h1 class="text-2xl font-bold tracking-tight text-white">
        go-NMOS
        <span class="ml-2 text-xs align-middle px-2 py-0.5 rounded bg-gray-800 text-gray-300 font-medium uppercase tracking-wider">
          Dashboard
        </span>
      </h1>
      <p class="text-sm text-gray-400">
        Signed in as <span class="font-medium text-gray-200">{user?.username}</span>
        <span class="mx-2 text-gray-600">•</span>
        <span class="px-2 py-0.5 rounded bg-gray-800 text-gray-300 text-xs font-medium uppercase">
          {user?.role}
        </span>
      </p>
    </div>
    <div class="flex items-center gap-2">
      <button
        class="px-4 py-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium transition-colors border border-gray-700"
        onclick={refreshAll}
      >
        Refresh
      </button>
      <button
        class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold transition-colors"
        onclick={onLogout}
      >
        Logout
      </button>
    </div>
  </header>

  <nav class="flex flex-wrap gap-2 pb-4 border-b border-gray-800">
    <button
      class="px-4 py-2 rounded-lg text-sm font-medium transition-all duration-150 {currentView === 'dashboard'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "dashboard")}
    >
      Dashboard
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'flows'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "flows")}
    >
      Flows
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'search'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "search")}
    >
      Search
    </button>
    {#if user?.role === "admin" || user?.role === "editor"}
      <button
        class="px-3 py-1.5 rounded-md border {currentView === 'users'
          ? 'bg-orange-600 text-white'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700'}"
        onclick={() => (currentView = "users")}
      >
        Users
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'nmos'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "nmos")}
    >
      NMOS
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'topology'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => {
        currentView = "topology";
        loadNMOSRegistry();
      }}
    >
      Topology
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'nmosPatch'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "nmosPatch")}
    >
      NMOS Patch
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'checker'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "checker")}
    >
      Checker
    </button>
    {#if user?.role === "admin" || user?.role === "editor"}
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'automation'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "automation")}
      >
        Automation
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'planner'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "planner")}
    >
      Planner
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'addressMap'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "addressMap")}
    >
      Address Map
    </button>
    {#if isAdmin}
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'portExplorer'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "portExplorer")}
      >
        Port Explorer
      </button>
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'logs'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => {
          currentView = "logs";
          loadLogs();
        }}
      >
        Logs
      </button>
    {/if}
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'settings'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "settings")}
    >
      Settings
    </button>
  </nav>

  {#if showBuildModal}
    <div class="fixed inset-0 z-40 flex items-center justify-center bg-black/70">
      <div class="w-full max-w-sm rounded-lg border border-gray-800 bg-gray-900 p-6 space-y-4">
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-2">
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-green-900/50 border border-green-800">
              <span class="w-2 h-2 rounded-full bg-green-500"></span>
              <span class="text-xs font-semibold uppercase tracking-wider text-green-300">
                UI Build Status
              </span>
            </div>
            <p class="text-sm font-semibold text-white">
              You are running the latest frontend build.
            </p>
            <p class="text-xs text-gray-400">
              Current version: <span class="font-semibold text-gray-200">{uiVersion}</span>
            </p>
          </div>
          <button
            class="shrink-0 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-white w-7 h-7 flex items-center justify-center text-sm border border-gray-700"
            onclick={() => (showBuildModal = false)}
            aria-label="Close"
          >
            ✕
          </button>
        </div>
        <div class="flex justify-end">
          <button
            class="px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-500 text-white text-sm font-semibold transition-colors"
            onclick={() => (showBuildModal = false)}
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
    <div class="space-y-6">
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3 mb-4">
        {#each Array(5) as _}
          <div class="rounded-xl border border-gray-800 bg-gray-900 px-3 py-3 animate-pulse">
            <div class="h-3 bg-gray-800 rounded w-1/2 mb-2"></div>
            <div class="h-8 bg-gray-800 rounded w-3/4"></div>
          </div>
        {/each}
      </div>
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-6">
        <SkeletonLoader lines={8} showHeader={true} />
      </div>
    </div>
  {:else if error}
    <div class="bg-red-950/50 border border-red-700 rounded-lg p-4">
      <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-red-400 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
        <div class="flex-1">
          <h3 class="text-red-300 font-semibold mb-1">Error</h3>
          <p class="text-red-200/80 text-sm">{error}</p>
        </div>
      </div>
    </div>
  {:else}
    {#if currentView === "dashboard"}
      <DashboardHomeView
        {summary}
        {flows}
        {flowTotal}
        {systemInfo}
        {realtimeEvents}
        {registryEvents}
        {registryHealth}
        {automationSummary}
        {healthDetail}
        {healthLoading}
        {healthError}
        {lastHealthLoadedAt}
        onRunHealthDetail={loadHealthDetail}
        {nodeCheckUrl}
        {nodeCheckLoading}
        {nodeCheckError}
        {nodeCheckResult}
        onNodeUrlChange={(v) => (nodeCheckUrl = v)}
        onRunNodeCheck={checkNodeAtUrl}
      />
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
        {newFlow}
        onApplyFlowSort={applyFlowSort}
        onPrevFlowPage={prevFlowPage}
        onNextFlowPage={nextFlowPage}
        onToggleFlowLock={toggleFlowLock}
        onDeleteFlow={deleteFlow}
        onCreateFlow={createFlow}
        onEditFlow={openEditFlowModal}
        onUpdateFlow={updateFlow}
        onCheckFlow={checkFlowNMOS}
        onFetchSDP={fetchFlowSDP}
        onSyncFromNMOS={syncFlowFromNMOS}
        onCheckIS05Receiver={checkIS05Receiver}
        {editingFlow}
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


    {#if currentView === "users" && (user?.role === "admin" || user?.role === "editor")}
      <UsersView
        {users}
        {isAdmin}
        onUpdateUser={updateUser}
        onDeleteUser={deleteUser}
      />
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
        {showConnectRDSModal}
        registryQueryUrl={rdsQueryUrl}
        registryDiscovering={rdsDiscovering}
        registryNodes={rdsNodes}
        registrySelectedIds={rdsSelectedIds}
        registryError={rdsError}
        onReloadNodes={loadNodes}
        onOpenAddNode={openAddNodeModal}
        onCancelAddNode={cancelAddNode}
        onConfirmAddNode={addNode}
        onChangeNewNodeName={(v) => (newNodeName = v)}
        onChangeNewNodeUrl={(v) => (newNodeUrl = v)}
        onOpenRDS={openRDSModal}
        onCloseRDS={closeRDSModal}
        onChangeRegistryQueryUrl={(v) => (rdsQueryUrl = v)}
        onDiscoverRegistryNodes={discoverRegistryNodes}
        onToggleRegistryNode={toggleRegistryNode}
        onSelectAllRegistryNodes={selectAllRegistryNodes}
        onAddSelectedRegistryNodes={addSelectedRegistryNodes}
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

    {#if currentView === "portExplorer" && isAdmin}
      <PortExplorerView />
    {/if}

    {#if currentView === "logs" && isAdmin}
      <LogsView bind:logsKind {logsLines} {token} onLoadLogs={loadLogs} />
    {/if}
  {/if}
  </div>
</main>
