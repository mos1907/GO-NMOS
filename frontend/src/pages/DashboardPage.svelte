<script>
  import { onMount, onDestroy } from "svelte";
  import { api, apiWithMeta } from "../lib/api.js";
  import { connectMQTT, disconnectMQTT } from "../lib/mqtt.js";
  import RegistryPatchView from "../components/RegistryPatchView.svelte";
  import TopologyView from "../components/TopologyView.svelte";
  import DashboardHomeView from "../components/DashboardHomeView.svelte";
  import FlowsView from "../components/FlowsView.svelte";
  import SearchView from "../components/SearchView.svelte";
  import UsersView from "../components/UsersView.svelte";
  import SettingsView from "../components/SettingsView.svelte";
  import CheckerView from "../components/CheckerView.svelte";
  import AutomationJobsView from "../components/AutomationJobsView.svelte";
  import PlannerView from "../components/PlannerView.svelte";
  import MigrationChecklistView from "../components/MigrationChecklistView.svelte";
  import AudioMappingView from "../components/AudioMappingView.svelte";
  import AudioChainView from "../components/AudioChainView.svelte";
  import EventsView from "../components/EventsView.svelte";
  import AddressMapView from "../components/AddressMapView.svelte";
  import LogsView from "../components/LogsView.svelte";
  import PortExplorerView from "../components/PortExplorerView.svelte";
  import MultiSiteView from "../components/MultiSiteView.svelte";
  import PlaybooksView from "../components/PlaybooksView.svelte";
  import SchedulingView from "../components/SchedulingView.svelte";
  import MetricsView from "../components/MetricsView.svelte";
  import InteropView from "../components/InteropView.svelte";
  //import PlaybooksView from "../components/PlaybooksView.svelte";
  import SkeletonLoader from "../components/SkeletonLoader.svelte";
  import EmptyState from "../components/EmptyState.svelte";
  import { addNotification } from "../lib/notifications.js";

  let {
    token,
    user,
    onLogout,
  } = $props();

  let currentView = $state("dashboard");
  let loading = $state(true);
  let error = $state("");
  let success = $state("");
  let autoOpenFlowName = $state(null);

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
  let nmosCheckerResult = $state(null);
  let automationJobs = $state([]);
  let automationSummary = $state(null);
  let systemInfo = $state(null);
  let addressMap = $state(null);
  let logsKind = $state("api");
  let logsLines = $state([]);
  let registryConfigs = $state([]);
  let registryStats = $state([]);
  let registryCompat = $state([]);
  let sitesRoomsSummary = $state(null);

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

  let sdnPingLoading = $state(false);
  let sdnPingError = $state("");
  let sdnPingResult = $state(null);

  function openMigrationBlog() {
    window.open("https://muratdemirci.com.tr/amwa-nmos/", "_blank", "noopener,noreferrer");
  }

  // Internal NMOS registry (IS-04 style) state
  let registryNodes = $state([]);
  let registryDevices = $state([]);
  let registrySenders = $state([]);
  let registryReceivers = $state([]);
  let registryFlows = $state([]);
  let registryHealth = $state(null);
  let selectedRegistryNodeId = $state("");
  let selectedRegistryDeviceId = $state("");
  /** Increment after TAKE to refresh receiver active state (BCC-style). */
  let refreshReceiverActiveTrigger = $state(0);

  // Real-time: MQTT/WebSocket primary; polling fallback. Live control without page refresh.
  // Patch state (IS-05): no backend event (read from external node), 2s polling
  $effect(() => {
    const view = currentView;
    if (view !== "topology") return;
    const tid = setInterval(() => refreshReceiverActiveTrigger++, 2000);
    return () => clearInterval(tid);
  });

  // Registry: WebSocket /ws/registry sync event triggers loadNMOSRegistry() (live). Fallback: 30s polling
  $effect(() => {
    const view = currentView;
    if (view !== "topology") return;
    const tid = setInterval(() => loadNMOSRegistry(), 30000);
    return () => clearInterval(tid);
  });

  // Flows/Dashboard: handleMQTTEvent already updates on MQTT flow events (live). Fallback: 30s silent refresh
  $effect(() => {
    const view = currentView;
    if (view !== "dashboard" && view !== "flows") return;
    const tid = setInterval(async () => {
      try {
        if (view === "flows") {
          const data = await loadFlows();
          if (Array.isArray(data)) flows = [...data];
        } else {
          const [sum, data, sys, sitesRooms] = await Promise.all([
            api("/flows/summary", { token }),
            loadFlows(),
            api("/system", { token }).catch(() => systemInfo),
            api("/nmos/registry/sites-summary", { token }).catch(() => null)
          ]);
          summary = sum;
          flows = Array.isArray(data) ? [...data] : flows;
          if (sys) systemInfo = sys;
          if (sitesRooms != null) sitesRoomsSummary = sitesRooms;
        }
      } catch (_) {}
    }, 30000);
    return () => clearInterval(tid);
  });

  // NMOS Patch-style view state (sender/receiver selection)
  let nmosNodes = $state([]);
  let selectedSenderNodeId = $state("");
  let selectedReceiverNodeId = $state("");
  let selectedPatchSender = $state(null);
  let selectedPatchReceiver = $state(null);
  let nmosPatchStatus = $state("");
  let nmosPatchError = $state("");
  let nmosPatchWarning = $state("");
  let senderFilterText = $state("");
  let receiverFilterText = $state("");
  // RDS (Registry) connect modal state
  let showConnectRDSModal = $state(false);
  let rdsQueryUrl = $state("");
  let rdsDiscovering = $state(false);
  let rdsNodes = $state([]);
  let rdsSelectedIds = $state([]);
  let rdsError = $state("");
  // Discover at URL → "Register?" confirm (Registry & Patch)
  let discoverAtUrlResult = $state(null);
  let pendingRegisterUrl = $state("");
  let discoverAtUrlLoading = $state(false);
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
    bucket_id: null,
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
  // E.3: canEdit includes operator, engineer, admin (but not viewer or automation)
  const canEdit = user?.role === "admin" || user?.role === "editor" || user?.role === "operator" || user?.role === "engineer";

  // Simple UI version label (frontend build version)
  const uiVersion = "go-NMOS UI v0.2.0 (router beta)";
  let showBuildModal = $state(true);

  async function loadDashboard() {
    loading = true;
    error = "";
    try {
      const [sum, data, sys, sitesRooms] = await Promise.all([
        api("/flows/summary", { token }),
        loadFlows(),
        api("/system", { token }).catch(() => systemInfo),
        api("/nmos/registry/sites-summary", { token }).catch(() => null)
      ]);
      summary = sum;
      // Ensure flows is always an array - use spread to trigger reactivity
      flows = Array.isArray(data) ? [...data] : [];
      systemInfo = sys || systemInfo;
      sitesRoomsSummary = sitesRooms;
    } catch (e) {
      console.error("loadDashboard error:", e);
      addNotification("error", e.message);
    } finally {
      loading = false;
    }
  }

  async function loadFlows() {
    try {
      const { data, headers } = await apiWithMeta(
        `/flows?limit=${flowLimit}&offset=${flowOffset}&sort_by=${encodeURIComponent(flowSortBy)}&sort_order=${encodeURIComponent(flowSortOrder)}`,
        { token }
      );
      flowTotal = Number(headers.get("X-Total-Count") || 0);
      return Array.isArray(data) ? data : [];
    } catch (e) {
      console.error("loadFlows error:", e);
      throw e;
    }
  }

  async function loadUsers() {
    if (!canEdit) return;
    users = await api("/users", { token });
  }

  async function loadSettings() {
    settings = await api("/settings", { token });
  }

  async function loadRegistryConfig() {
    try {
      const data = await api("/registry/config", { token });
      registryConfigs = Array.isArray(data) ? data : [];
    } catch (e) {
      console.warn("Failed to load registry config:", e.message);
      registryConfigs = [];
    }
  }

  async function loadRegistryStats() {
    try {
      const data = await api("/registry/config/stats", { token });
      registryStats = Array.isArray(data) ? data : [];
    } catch (e) {
      registryStats = [];
    }
  }

  async function loadRegistryCompat() {
    try {
      const data = await api("/registry/compat", { token });
      registryCompat = Array.isArray(data) ? data : [];
    } catch {
      registryCompat = [];
    }
  }

  async function saveRegistryConfigs(configs) {
    error = "";
    success = "";
    try {
      await api("/registry/config", {
        method: "PUT",
        token,
        body: configs,
      });
      registryConfigs = configs;
      addNotification("success", "Registry configuration saved.");
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  /** Remove one RDS from config and delete all nodes (and patch sources/destinations) from that registry. */
  async function removeRegistry(queryUrl) {
    try {
      const res = await api("/registry/config/remove", {
        method: "POST",
        token,
        body: { query_url: queryUrl },
      });
      await loadRegistryConfig();
      await loadRegistryStats();
      await loadNMOSRegistry();
      const n = res.deleted_nodes ?? 0;
      addNotification("success", res.message || `Registry removed. ${n} node(s) removed from Registry Patch.`);
    } catch (e) {
      addNotification("error", e.message || "Failed to remove registry");
      throw e;
    }
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

  async function pingSDNController() {
    sdnPingError = "";
    sdnPingResult = null;
    sdnPingLoading = true;
    try {
      const body = {};
      if (settings.sdn_controller_url) {
        body.url = settings.sdn_controller_url;
      }
      const res = await api("/sdn/ping", {
        method: "POST",
        token,
        body,
      });
      sdnPingResult = res;
    } catch (e) {
      sdnPingError = e.message;
    } finally {
      sdnPingLoading = false;
    }
  }

  async function fetchSDNTopology() {
    return api("/sdn/topology", { token });
  }

  async function fetchSDNPaths(from, to) {
    const q = new URLSearchParams();
    if (from) q.set("from", from);
    if (to) q.set("to", to);
    return api(`/sdn/paths?${q.toString()}`, { token });
  }

  async function patchFlow(flowId, updates) {
    await api(`/flows/${flowId}`, { method: "PATCH", token, body: updates });
    await refreshAll();
  }

  async function refreshAll() {
    success = "";
    await loadDashboard();
    await loadHealthDetail().catch(() => {});
    await loadUsers().catch(() => {});
    await loadSettings().catch(() => {});
    await loadRegistryConfig().catch(() => {});
    await loadRegistryStats().catch(() => {});
    await loadRegistryCompat().catch(() => {});
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
      searchResults = Array.isArray(data) ? data : [];
      searchTotal = Number(headers.get("X-Total-Count") || 0);
    } catch (e) {
      searchResults = [];
      searchTotal = 0;
      addNotification("error", e.message);
    }
  }

  function normalizeNewFlowPayload(form = newFlow) {
    const toInt = (v) => {
      const n = typeof v === "string" ? parseInt(v, 10) : v;
      return Number.isFinite(n) ? n : 0;
    };

    // Debug: log the form object to see what we're receiving
    console.debug("normalizeNewFlowPayload - form.display_name:", form?.display_name);

    const payload = {
      display_name: (form?.display_name || "").trim(),
      multicast_ip: (form?.multicast_ip || "").trim(),
      source_ip: (form?.source_ip || "").trim(),
      port: toInt(form?.port),
      flow_status: form?.flow_status || "active",
      availability: form?.availability || "available",
      transport_protocol: form?.transport_protocol || "RTP/UDP",
      note: form?.note || "",
      // optional bucket_id
      bucket_id: form?.bucket_id || null,
      // optional aliases
      alias_1: form?.alias_1 || "",
      alias_2: form?.alias_2 || "",
      alias_3: form?.alias_3 || "",
      alias_4: form?.alias_4 || "",
      alias_5: form?.alias_5 || "",
      alias_6: form?.alias_6 || "",
      alias_7: form?.alias_7 || "",
      alias_8: form?.alias_8 || "",
      // optional user fields
      user_field_1: form?.user_field_1 || "",
      user_field_2: form?.user_field_2 || "",
      user_field_3: form?.user_field_3 || "",
      user_field_4: form?.user_field_4 || "",
      user_field_5: form?.user_field_5 || "",
      user_field_6: form?.user_field_6 || "",
      user_field_7: form?.user_field_7 || "",
      user_field_8: form?.user_field_8 || "",
    };

    return payload;
  }

  // Flag to auto-open modal when switching to flows view
  let shouldAutoOpenFlowModal = $state(false);

  // Wrapper function to open modal when Create Flow is called from EmptyState in DashboardHomeView
  function handleCreateFlowFromDashboard() {
    // Reset form and switch to flows view (modal will be opened by FlowsView)
    resetNewFlowForm();
    editingFlow = null;
    shouldAutoOpenFlowModal = true;
    currentView = "flows";
    // Reset flag after a short delay
    setTimeout(() => {
      shouldAutoOpenFlowModal = false;
    }, 500);
  }

  async function createFlow(form) {
    error = "";
    success = "";
    try {
      // Use the form parameter if provided, otherwise fall back to newFlow state
      // If form is provided, use it; otherwise use current newFlow state
      const formData = form ? { ...form } : { ...newFlow };

      // Normalize payload first
      const payload = normalizeNewFlowPayload(formData);

      // Validate after normalization
      if (!payload.display_name || !payload.display_name.trim()) {
        error = "Display name is required.";
        return;
      }
      
      const result = await api("/flows", { method: "POST", token, body: payload });
      if (result?.already_exists) addNotification("warning", result?.message || "A flow with this name already exists. No duplicate created.");
      else addNotification("success", "Flow created successfully.");
      resetNewFlowForm();
      // Force reload flows list
      const reloadedFlows = await loadFlows();
      flows = Array.isArray(reloadedFlows) ? [...reloadedFlows] : [];
      summary = await api("/flows/summary", { token });
    } catch (e) {
      console.error("Error creating flow:", e);
      addNotification("error", e.message);
    }
  }

  async function updateFlow(form) {
    if (!editingFlow) return;
    error = "";
    success = "";
    try {
      const payload = normalizeNewFlowPayload(form ?? newFlow);
      await api(`/flows/${editingFlow.id}`, {
        method: "PATCH",
        token,
        body: payload,
      });
      addNotification("success", "Flow updated successfully.");
      editingFlow = null;
      resetNewFlowForm();
      await refreshAll();
    } catch (e) {
      addNotification("error", e.message);
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
      bucket_id: null,
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
      bucket_id: flow.bucket_id || null,
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
        // Empty fields = backend fills all (multicast_ip, source_ip, port, SDP, NMOS, format_summary)
        fields: [],
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
        flows = Array.isArray(data) ? [...data] : [];
      });
      await api("/flows/summary", { token }).then((s) => {
        summary = s;
      });
      // Return result for popup display
      return { locked: result.locked, flow: updatedFlow };
    } catch (e) {
      addNotification("error", e.message);
      return null;
    }
  }

  async function deleteFlow(flow) {
    if (!confirm(`Delete flow '${flow.display_name}'?`)) return;
    error = "";
    success = "";
    try {
      await api(`/flows/${flow.id}`, { method: "DELETE", token });
      addNotification("success", "Flow deleted.");
      await refreshAll();
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  async function nextFlowPage() {
    if (flowOffset + flowLimit >= flowTotal) return;
    flowOffset += flowLimit;
    const data = await loadFlows();
    flows = Array.isArray(data) ? [...data] : [];
  }

  async function prevFlowPage() {
    flowOffset = Math.max(0, flowOffset - flowLimit);
    const data = await loadFlows();
    flows = Array.isArray(data) ? [...data] : [];
  }

  async function applyFlowSort() {
    flowOffset = 0;
    const data = await loadFlows();
    flows = Array.isArray(data) ? [...data] : [];
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
      addNotification("success", `Setting '${key}' updated.`);
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("success", `User '${username}' updated successfully.`);
      await loadUsers();
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("success", `User '${username}' deleted successfully.`);
      await loadUsers();
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("error", e.message);
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
      addNotification("success", `Import complete. ${result.imported ?? 0} flow(s) processed.`);
      await refreshAll();
    } catch (e) {
      addNotification("error", e.message);
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

      // IS-05 base URL default: <base>/x-nmos/connection/<is05_version>
      const base = nmosResult.base_url?.replace(/\/$/, "") || nmosBaseUrl.replace(/\/$/, "");
      // Use IS-05 version if available, otherwise fallback to IS-04 version
      const is05Ver = (nmosResult.is05_version || nmosResult.is04_version || "v1.0").replace(/^\//, "");
      nmosIS05Base = `${base}/x-nmos/connection/${is05Ver}`;
      console.log("[NMOS Discovery] IS-05 version:", nmosResult.is05_version, "Using:", is05Ver, "Base URL:", nmosIS05Base);

      // Default selection: first flow and first receiver
      selectedNMOSFlow = flows[0] || null;
      selectedNMOSReceiver = (nmosResult.receivers || [])[0] || null;

      addNotification("success", "NMOS discovery completed.");
    } catch (e) {
      addNotification("error", e.message);
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

      // Match NMOS sender by flow_id if available
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

      addNotification("success", `TAKE OK: ${selectedNMOSFlow.display_name} → ${selectedNMOSReceiver.label}`);
    } catch (e) {
      addNotification("error", e.message);
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

  /** Normalize URL for comparison (trim, trailing slash). */
  function normalizeRegistryUrl(u) {
    if (!u || typeof u !== "string") return "";
    return u.trim().replace(/\/+$/, "");
  }

  /** Ensure this query URL is in the system registry config so Reload registry and RDS list work. */
  async function ensureRegistryInConfig(queryUrl) {
    const q = normalizeRegistryUrl(queryUrl);
    if (!q) return;
    try {
      let configs = await api("/registry/config", { token });
      if (!Array.isArray(configs)) configs = [];
      const has = configs.some((r) => normalizeRegistryUrl(r.query_url) === q);
      if (!has) {
        const label = q.replace(/^https?:\/\//, "").replace(/\/x-nmos\/query.*/i, "") || "RDS";
        configs = [...configs, { name: label, query_url: q, role: "prod", enabled: true }];
        await api("/registry/config", { method: "PUT", token, body: configs });
        registryConfigs = configs;
      }
    } catch (e) {
      console.warn("Could not add RDS to config:", e.message);
    }
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
      await ensureRegistryInConfig(q);
      await loadRegistryConfig();
      await loadRegistryStats();
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

  let reloadRegistrySyncing = $state(false);

  /** Re-fetch all nodes from system-configured registries (Settings) and sync resources; then refresh list. Same result from any client. */
  async function reloadRegistry() {
    reloadRegistrySyncing = true;
    try {
      const res = await api("/nmos/registry/sync", {
        method: "POST",
        token,
        body: {}
      });
      await loadNMOSRegistry();
      const failed = res.failed || [];
      if (res.synced != null && res.synced > 0) {
        addNotification("success", res.message || `Synced ${res.synced} node(s).`);
      } else if (res.synced === 0) {
        addNotification("info", res.message || "Configure registries in Settings to sync, or use Connect RDS to add nodes once.");
      }
      if (failed.length > 0) {
        addNotification("warning", `${failed.length} node(s) failed to sync: ${failed.slice(0, 3).join(", ")}${failed.length > 3 ? "…" : ""}`);
      }
    } catch (e) {
      addNotification("error", e.message || "Reload registry failed");
    } finally {
      reloadRegistrySyncing = false;
    }
  }

  /** Reload a single RDS by query_url; then refresh registry list and stats. */
  async function reloadRegistryByUrl(queryUrl) {
    try {
      const res = await api("/nmos/registry/sync", {
        method: "POST",
        token,
        body: { query_url: (queryUrl || "").trim() },
      });
      await loadNMOSRegistry();
      await loadRegistryStats();
      if (res.synced != null && res.synced > 0) {
        addNotification("success", res.message || `Synced ${res.synced} node(s).`);
      }
      if ((res.failed || []).length > 0) {
        addNotification("warning", `${res.failed.length} node(s) failed to sync.`);
      }
    } catch (e) {
      addNotification("error", e.message || "Reload failed");
      throw e;
    }
  }

  /** Update one RDS in config (name, query_url, role, enabled); then refresh config list. */
  async function updateRegistry(oldQueryUrl, updated) {
    const norm = (u) => (u || "").trim().replace(/\/+$/, "");
    const key = norm(oldQueryUrl);
    const configs = (registryConfigs || []).map((r) =>
      norm(r.query_url) === key ? { ...r, ...updated } : r
    );
    await saveRegistryConfigs(configs);
    await loadRegistryConfig();
    await loadRegistryStats();
  }

  /** Registry & Patch: bulk-register selected nodes from RDS into internal registry (sync). */
  async function registerSelectedNodesToRegistry() {
    const selected = new Set(rdsSelectedIds || []);
    const candidates = (rdsNodes || []).filter((n) => selected.has(n.id));
    if (candidates.length === 0) return;
    let items = candidates
      .map((n) => {
        const base_url = (n.base_url || n.baseURL || n.BaseURL || n.href || n.Href || "").trim().replace(/\/$/, "");
        const label = (n.label || n.id || base_url || "").trim();
        return { base_url, label };
      })
      .filter((x) => x.base_url);
    const beforeDedupe = items.length;
    const seenUrl = new Set();
    items = items.filter((x) => {
      if (seenUrl.has(x.base_url)) return false;
      seenUrl.add(x.base_url);
      return true;
    });
    if (beforeDedupe > items.length) {
      addNotification("info", "Duplicate source (same URL) removed: " + (beforeDedupe - items.length) + " entries.");
    }
    if (items.length === 0) {
      rdsError = "Selected nodes have no base_url.";
      return;
    }
    rdsDiscovering = true;
    rdsError = "";
    try {
      let newCount = 0;
      let alreadyCount = 0;
      for (const item of items) {
        try {
          const res = await api("/nmos/register-node", { method: "POST", token, body: { base_url: item.base_url, label: item.label } });
          if (res && res.already_registered) {
            alreadyCount++;
            addNotification("info", (res.previous_label || res.node_id || "") + " was already registered; label updated to: " + item.label);
          } else {
            newCount++;
          }
        } catch (e) {
          console.warn("Register node failed:", item.base_url, e);
        }
      }
      await loadNMOSRegistry();
      if (alreadyCount > 0 && newCount > 0) {
        addNotification("success", newCount + " new node(s) added, " + alreadyCount + " already registered (label updated).");
      } else if (alreadyCount > 0) {
        addNotification("success", "All were already registered (" + alreadyCount + " node label(s) updated).");
      } else {
        addNotification("success", newCount + " node(s) registered. They will appear in Sender/Receiver Nodes.");
      }
      showConnectRDSModal = false;
    } catch (e) {
      rdsError = e.message || "Registration failed";
    } finally {
      rdsDiscovering = false;
    }
  }

  async function discoverAtUrl(url) {
    if (!url?.trim()) return;
    discoverAtUrlLoading = true;
    discoverAtUrlResult = null;
    pendingRegisterUrl = "";
    try {
      const res = await api("/nmos/discover", {
        method: "POST",
        token,
        body: { base_url: url.trim() }
      });
      discoverAtUrlResult = res;
      pendingRegisterUrl = url.trim();
    } catch (e) {
      addNotification("error", e.message || "Discover failed");
    } finally {
      discoverAtUrlLoading = false;
    }
  }

  async function registerDiscoveredNode(url) {
    if (!url?.trim()) return;
    try {
      await api("/nmos/register-node", {
        method: "POST",
        token,
        body: { base_url: url.trim() }
      });
      discoverAtUrlResult = null;
      pendingRegisterUrl = "";
      await loadNMOSRegistry();
      addNotification("success", "Node registered to registry.");
    } catch (e) {
      addNotification("error", e.message || "Register failed");
    }
  }

  function closeRegisterConfirm() {
    discoverAtUrlResult = null;
    pendingRegisterUrl = "";
  }

  async function deleteRegistryNode(nodeId) {
    if (!nodeId || !canEdit) return;
    try {
      await api(`/nmos/registry/nodes/${encodeURIComponent(nodeId)}`, { method: "DELETE", token });
      addNotification("success", "Node removed from registry.");
      await loadNMOSRegistry();
      if (selectedRegistryNodeId === nodeId) {
        selectedRegistryNodeId = "";
        selectedRegistryDeviceId = "";
      }
      if (selectedReceiverNodeId === nodeId) selectedReceiverNodeId = "";
    } catch (e) {
      addNotification("error", e.message || "Remove failed");
    }
  }

  /** Discover IS-05 base for a node (by id). Used when user selects a node so we can show BCC/external connection status. */
  async function discoverIS05BaseForNode(nodeId) {
    if (!nodeId || !registryNodes?.length) return;
    const node = registryNodes.find((n) => n.id === nodeId);
    if (!node || !node.hostname) return;
    await discoverIS05BaseFromNodeHostname(node.hostname);
  }

  function discoverIS05BaseFromNodeHostname(hostname) {
    return (async () => {
      let baseURL = hostname;
      if (!baseURL.startsWith("http://") && !baseURL.startsWith("https://")) {
        const ports = [8080, 8081, 8082, 80, 443];
        for (const port of ports) {
          try {
            const testURL = `http://${hostname}:${port}`;
            const result = await api("/nmos/detect-is05", {
              method: "POST",
              token,
              body: { base_url: testURL }
            });
            if (result.detected && result.is05_base_url) {
              nmosIS05Base = result.is05_base_url;
              console.log("[Registry Patch] Discovered IS-05 base URL:", nmosIS05Base);
              return;
            }
          } catch (e) {
            continue;
          }
        }
        baseURL = `http://${hostname}:8080`;
      } else {
        baseURL = baseURL.replace(/\/$/, "");
      }
      try {
        const result = await api("/nmos/detect-is05", {
          method: "POST",
          token,
          body: { base_url: baseURL }
        });
        if (result.detected && result.is05_base_url) {
          nmosIS05Base = result.is05_base_url;
          console.log("[Registry Patch] Discovered IS-05 base URL:", nmosIS05Base);
        }
      } catch (e) {
        console.warn("[Registry Patch] Error discovering IS-05 base URL:", e);
      }
    })();
  }

  async function discoverIS05BaseForReceiver(receiver) {
    if (!receiver || !receiver.device_id) return;
    const device = registryDevices.find((d) => d.id === receiver.device_id);
    if (!device || !device.node_id) return;
    const node = registryNodes.find((n) => n.id === device.node_id);
    if (!node || !node.hostname) return;
    await discoverIS05BaseFromNodeHostname(node.hostname);
  }

  function isPatchTakeReady() {
    return !!(selectedPatchSender && selectedPatchReceiver && nmosIS05Base && !nmosTakeBusy);
  }

  async function executePatchTake() {
    if (!isPatchTakeReady()) return;
    nmosPatchError = "";
    nmosPatchStatus = "";
    nmosPatchWarning = "";
    nmosTakeBusy = true;
    try {
      const base = nmosIS05Base.replace(/\/$/, "");
      const receiverId = selectedPatchReceiver.id;
      const connectionUrl = `${base}/single/receivers/${receiverId}/staged`;

      let internalFlow =
        flows.find((f) => f.flow_id && f.flow_id === selectedPatchSender?.flow_id) ||
        flows.find((f) => f.flow_id && f.flow_id === selectedPatchSender?.flow_id?.toString?.()) ||
        flows.find((f) => f.nmos_flow_id && (f.nmos_flow_id === selectedPatchSender?.flow_id || f.nmos_flow_id === selectedPatchSender?.flow_id?.toString?.()));

      // If flow not found in internal flows, try to create it from registry flow (backend returns existing if flow_id already exists)
      if (!internalFlow && selectedPatchSender?.flow_id) {
        const registryFlow = registryFlows.find((f) => f.id === selectedPatchSender.flow_id);
        if (registryFlow) {
          // Create a flow from registry flow data
          try {
            const newFlow = await api("/flows", {
              method: "POST",
              token,
              body: {
                flow_id: registryFlow.id,
                display_name: registryFlow.label || `Flow ${registryFlow.id.substring(0, 8)}`,
                data_source: "nmos",
                nmos_flow_id: registryFlow.id,
                nmos_label: registryFlow.label,
                nmos_description: registryFlow.description,
                // Try to fetch SDP if manifest_href is available
                sdp_url: selectedPatchSender.manifest_href || ""
              }
            });
            internalFlow = newFlow?.flow ?? newFlow;
            if (newFlow?.already_exists) {
              nmosPatchWarning = newFlow.message || "A flow with this name already exists. Using existing flow.";
            }
            // Reload flows to include the new one (or existing)
            const flowsData = await api("/flows?limit=1000", { token });
            flows = Array.isArray(flowsData) ? flowsData : [];
            console.log("[Registry Patch] Created flow from registry:", newFlow);
          } catch (createError) {
            console.error("[Registry Patch] Failed to create flow from registry:", createError);
            throw new Error(
              `No internal flow found for sender flow_id=${selectedPatchSender.flow_id}. Failed to create flow: ${createError.message}`
            );
          }
        } else {
          throw new Error(
            `No internal flow found for selected sender flow_id=${selectedPatchSender?.flow_id || "?"}. Flow not found in registry either.`
          );
        }
      } else if (!internalFlow) {
        throw new Error(
          `No internal flow found for selected sender. Sender has no flow_id.`
        );
      }

      // sender_id + transport_params (from internal flow) ile backend'e delege et; response includes flow info
      const applyRes = await api(`/flows/${internalFlow.id}/nmos/apply`, {
        method: "POST",
        token,
        body: {
          connection_url: connectionUrl,
          sender_id: selectedPatchSender.id
        }
      });

      const flowLabel = applyRes?.flow?.display_name || internalFlow.display_name;
      const flowAddr = applyRes?.flow?.multicast_ip && applyRes?.flow?.port != null
        ? ` ${applyRes.flow.multicast_ip}:${applyRes.flow.port}`
        : "";
      nmosPatchStatus = `TAKE OK: ${selectedPatchSender.label} → ${selectedPatchReceiver.label} · Flow: ${flowLabel}${flowAddr}`;
      refreshReceiverActiveTrigger++;
    } catch (e) {
      nmosPatchError = e.message;
    } finally {
      nmosTakeBusy = false;
    }
  }

  async function runCollisionCheck() {
    error = "";
    try {
      checkerResult = await api("/checker/collisions", { token });
      addNotification("success", "Collision check completed.");
      await loadDashboard();
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  function handleCheckerFlowClick(flowName) {
    // Switch to flows view and auto-open flow detail modal
    currentView = "flows";
    autoOpenFlowName = flowName;
    // Reset after a short delay to allow re-opening if needed
    setTimeout(() => {
      autoOpenFlowName = null;
    }, 500);
  }

  async function loadCheckerLatest() {
    const [coll, nmos] = await Promise.all([
      api("/checker/latest?kind=collisions", { token }),
      api("/checker/latest?kind=nmos", { token })
    ]);
    checkerResult = coll;
    nmosCheckerResult = nmos;
  }

  async function runNmosCheck() {
    error = "";
    try {
      const result = await api("/checker/nmos?timeout=5", { token });
      nmosCheckerResult = { kind: "nmos", result, created_at: new Date().toISOString() };
      addNotification("success", "NMOS difference check completed.");
      await loadDashboard();
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  async function loadAutomationJobs() {
    if (!canEdit) return;
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
      addNotification("error", e.message);
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
      addNotification("success", "Planner parent bucket created.");
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("success", "Planner child bucket created.");
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("error", e.message);
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
      addNotification("success", "Planner buckets imported.");
    } catch (e) {
      addNotification("error", e.message);
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
      addNotification("success", "Bucket updated.");
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  async function plannerDelete(item) {
    if (!confirm(`Delete bucket '${item.name}'?`)) return;
    try {
      await api(`/address/buckets/${item.id}`, { method: "DELETE", token });
      if (selectedPlannerRoot) await selectPlannerRoot(selectedPlannerRoot);
      addNotification("success", "Bucket deleted.");
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  async function loadLogs() {
    try {
      const data = await api(`/logs?kind=${encodeURIComponent(logsKind)}&lines=300`, { token });
      logsLines = data.lines || [];
    } catch (e) {
      addNotification("error", e.message);
    }
  }

  async function loadNMOSRegistry() {
    try {
      const [nodes, devices, flows, senders, receivers] = await Promise.all([
        api("/nmos/registry/nodes", { token }),
        api("/nmos/registry/devices", { token }),
        api("/nmos/registry/flows", { token }),
        api("/nmos/registry/senders", { token }),
        api("/nmos/registry/receivers", { token })
      ]);
      registryNodes = nodes;
      registryDevices = devices;
      registryFlows = flows;
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
          flows = Array.isArray(data) ? [...data] : [];
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

    // Registry WebSocket: live update — on sync event refresh registry data (no page refresh)
    try {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const host = window.location.host;
      const ws = new WebSocket(`${proto}//${host}/ws/registry`);
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          registryEvents = [data, ...registryEvents].slice(0, 50);
          // When backend registry sync completes, refresh UI (nodes/senders/receivers)
          if (data?.kind === "sync" || data?.resource === "nmos_registry") {
            loadNMOSRegistry();
          }
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
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'topology'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => {
        currentView = "topology";
        loadNMOSRegistry();
      }}
    >
      Registry & Patch
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'systemTopology'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => {
        currentView = "systemTopology";
        loadNMOSRegistry();
      }}
    >
      Topology
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'checker'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "checker")}
    >
      Checker
    </button>
    {#if canEdit}
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'automation'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "automation")}
      >
        Automation
      </button>
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'playbooks'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "playbooks")}
      >
        Playbooks
      </button>
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'scheduling'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "scheduling")}
      >
        Schedule
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
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'migration'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "migration")}
    >
      Migration
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'audio'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "audio")}
    >
      Audio (IS-08)
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'signalChain'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "signalChain")}
    >
      Signal chain
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'events'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "events")}
    >
      Events
    </button>
    <button
      class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'multiSite'
        ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
        : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
      onclick={() => (currentView = "multiSite")}
    >
      Multi-Site
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
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'metrics'
          ? 'bg-orange-600 text-white shadow-md shadow-orange-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "metrics")}
      >
        Metrics
      </button>
      <button
        class="px-3 py-1.5 rounded-md border transition-all duration-150 {currentView === 'interop'
          ? 'bg-purple-600 text-white shadow-md shadow-purple-600/20'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700 border border-gray-700 hover:border-gray-600'}"
        onclick={() => (currentView = "interop")}
      >
        Interop Tests
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
        onCreateFlow={handleCreateFlowFromDashboard}
        {nodeCheckUrl}
        {nodeCheckLoading}
        {nodeCheckError}
        {nodeCheckResult}
        onNodeUrlChange={(v) => (nodeCheckUrl = v)}
        onRunNodeCheck={checkNodeAtUrl}
        {registryConfigs}
        {registryCompat}
        {sitesRoomsSummary}
      />
    {/if}

    {#if currentView === "flows"}
      <FlowsView
        {flows}
        {token}
        bind:flowLimit
        bind:flowOffset
        {flowTotal}
        bind:flowSortBy
        bind:flowSortOrder
        {canEdit}
        {isAdmin}
        bind:newFlow
        onApplyFlowSort={applyFlowSort}
        onPrevFlowPage={prevFlowPage}
        onNextFlowPage={nextFlowPage}
        onToggleFlowLock={toggleFlowLock}
        onDeleteFlow={deleteFlow}
        onCreateFlow={createFlow}
        onEditFlow={openEditFlowModal}
        onUpdateFlow={updateFlow}
        onPatchFlow={patchFlow}
        onCheckFlow={checkFlowNMOS}
        onFetchSDP={fetchFlowSDP}
        onSyncFromNMOS={syncFlowFromNMOS}
        onCheckIS05Receiver={checkIS05Receiver}
        {editingFlow}
        {autoOpenFlowName}
        autoOpenModal={shouldAutoOpenFlowModal}
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


    {#if currentView === "users" && isAdmin}
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
        {sdnPingLoading}
        {sdnPingError}
        {sdnPingResult}
        onPingSDN={pingSDNController}
        onFetchSDNTopology={fetchSDNTopology}
        onFetchSDNPaths={fetchSDNPaths}
        {registryConfigs}
        onSaveRegistryConfigs={saveRegistryConfigs}
        onRemoveRegistry={removeRegistry}
        {token}
      />
    {/if}

    {#if currentView === "topology"}
      <RegistryPatchView
        {registryNodes}
        {registryDevices}
        {registrySenders}
        {registryReceivers}
        selectedSenderNodeId={selectedRegistryNodeId}
        selectedReceiverNodeId={selectedReceiverNodeId}
        selectedRegistryDeviceId={selectedRegistryDeviceId}
        onSelectSenderNode={async (id) => {
          selectedRegistryNodeId = id;
        }}
        onSelectReceiverNode={async (id) => {
          selectedReceiverNodeId = id;
          await discoverIS05BaseForNode(id);
        }}
        onSelectDevice={async (id) => {
          selectedRegistryDeviceId = id;
          const dev = registryDevices.find((d) => d.id === id);
          if (dev?.node_id) selectedRegistryNodeId = dev.node_id;
        }}
        isPatchTakeReady={isPatchTakeReady}
        selectedPatchSender={selectedPatchSender}
        selectedPatchReceiver={selectedPatchReceiver}
        {nmosIS05Base}
        {nmosTakeBusy}
        onSelectPatchSender={(s) => {
          selectedPatchSender = s;
          if (selectedPatchReceiver) {
            discoverIS05BaseForReceiver(selectedPatchReceiver);
          }
        }}
        onSelectPatchReceiver={async (r) => {
          selectedPatchReceiver = r;
          await discoverIS05BaseForReceiver(r);
        }}
        onExecutePatchTake={executePatchTake}
        {flows}
        {token}
        {nmosPatchStatus}
        {nmosPatchError}
        {nmosPatchWarning}
        refreshReceiverActiveTrigger={refreshReceiverActiveTrigger}
        discoverAtUrlResult={discoverAtUrlResult}
        pendingRegisterUrl={pendingRegisterUrl}
        discoverAtUrlLoading={discoverAtUrlLoading}
        onDiscoverAtUrl={discoverAtUrl}
        onRegisterDiscoveredNode={registerDiscoveredNode}
        onCloseRegisterConfirm={closeRegisterConfirm}
        onDeleteNode={deleteRegistryNode}
        {canEdit}
        showConnectRDSModal={showConnectRDSModal}
        rdsQueryUrl={rdsQueryUrl}
        rdsDiscovering={rdsDiscovering}
        rdsNodes={rdsNodes}
        rdsSelectedIds={rdsSelectedIds}
        rdsError={rdsError}
        onOpenRDS={openRDSModal}
        onCloseRDS={closeRDSModal}
        onChangeRegistryQueryUrl={(v) => (rdsQueryUrl = v)}
        onDiscoverRegistryNodes={discoverRegistryNodes}
        onToggleRegistryNode={toggleRegistryNode}
        onSelectAllRegistryNodes={selectAllRegistryNodes}
        onRegisterSelectedToRegistry={registerSelectedNodesToRegistry}
        onReloadRegistry={reloadRegistry}
        reloadRegistrySyncing={reloadRegistrySyncing}
        registryConfigs={registryConfigs}
        registryStats={registryStats}
        onReloadRegistryUrl={reloadRegistryByUrl}
        onUpdateRegistry={updateRegistry}
        onRemoveRegistry={removeRegistry}
      />
    {/if}
    {#if currentView === "systemTopology"}
      <TopologyView
        {registryNodes}
        {registryDevices}
        {registrySenders}
        {registryReceivers}
        {token}
      />
    {/if}

    {#if currentView === "checker"}
      <CheckerView
        {checkerResult}
        nmosCheckerResult={nmosCheckerResult}
        {token}
        onRunCollisionCheck={runCollisionCheck}
        onRunNmosCheck={runNmosCheck}
        onFlowClick={handleCheckerFlowClick}
      />
    {/if}

    {#if currentView === "automation" && canEdit}
      <AutomationJobsView {automationJobs} {isAdmin} onToggleAutomationJob={toggleAutomationJob} />
    {/if}

    {#if currentView === "playbooks" && canEdit}
      <PlaybooksView {token} {isAdmin} canEdit={canEdit} userRole={user?.role || "viewer"} />
    {/if}

    {#if currentView === "scheduling" && canEdit}
      <SchedulingView {token} {isAdmin} canEdit={canEdit} />
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
        {token}
        onSelectPlannerRoot={selectPlannerRoot}
        onPlannerQuickEdit={plannerQuickEdit}
        onPlannerDelete={plannerDelete}
        onCreatePlannerParent={createPlannerParent}
        onCreatePlannerChild={createPlannerChild}
        onExportBuckets={exportBuckets}
        onImportBucketsFromFile={importBucketsFromFile}
      />
    {/if}

    {#if currentView === "migration"}
      <MigrationChecklistView onOpenBlog={openMigrationBlog} />
    {/if}
    {#if currentView === "audio"}
      <AudioMappingView {token} />
    {/if}
    {#if currentView === "signalChain"}
      <AudioChainView {token} />
    {/if}
    {#if currentView === "events"}
      <EventsView {token} />
    {/if}

    {#if currentView === "multiSite"}
      <MultiSiteView {token} {sitesRoomsSummary} />
    {/if}

    {#if currentView === "addressMap"}
      <AddressMapView 
        {addressMap} 
        {token} 
        onNavigateToPlanner={(bucketId) => {
          currentView = "planner";
          // Try to find and select the bucket's parent root
          if (bucketId && plannerRoots.length > 0) {
            // Find bucket in children
            const bucket = plannerChildren.find(b => b.id === bucketId);
            if (bucket && bucket.parent_id) {
              const parent = plannerRoots.find(r => r.id === bucket.parent_id);
              if (parent) {
                selectPlannerRoot(parent);
              }
            } else if (plannerRoots.length > 0) {
              selectPlannerRoot(plannerRoots[0]);
            }
          }
        }}
      />
    {/if}

    {#if currentView === "portExplorer" && isAdmin}
      <PortExplorerView />
    {/if}

    {#if currentView === "logs" && isAdmin}
      <LogsView bind:logsKind {logsLines} {token} onLoadLogs={loadLogs} />
    {/if}

    {#if currentView === "metrics"}
      <MetricsView {token} />
    {/if}

    {#if currentView === "interop"}
      <InteropView {token} />
    {/if}
  {/if}
  </div>
</main>
