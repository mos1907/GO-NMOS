<script>
  let {
    token = "",
    isAdmin = false,
    canEdit = false,
  } = $props();

  // Scheduled playbook executions
  let playbooks = $state([]);
  let scheduledExecs = $state([]);
  let scheduledLoading = $state(false);
  let scheduledError = $state("");

  // Maintenance windows
  let maintenanceWindows = $state([]);
  let maintenanceLoading = $state(false);
  let maintenanceError = $state("");

  // New scheduled playbook execution form
  let newSchedule = $state({
    playbook_id: "",
    scheduledDateTime: "",
    paramsText: "{}",
  });

  // New maintenance window form
  let newWindow = $state({
    name: "",
    description: "",
    startDateTime: "",
    endDateTime: "",
    routingPolicyId: "",
    enabled: true,
  });

  async function loadPlaybooks() {
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/playbooks", { token });
      playbooks = Array.isArray(data) ? data : [];
    } catch (e) {
      console.error("Failed to load playbooks:", e);
      playbooks = [];
    }
  }

  async function loadScheduledExecs() {
    scheduledLoading = true;
    scheduledError = "";
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/schedule/playbooks?limit=100", { token });
      scheduledExecs = Array.isArray(data) ? data : [];
    } catch (e) {
      scheduledError = e.message || "Failed to load scheduled playbook executions";
      scheduledExecs = [];
    } finally {
      scheduledLoading = false;
    }
  }

  async function loadMaintenanceWindows() {
    maintenanceLoading = true;
    maintenanceError = "";
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/maintenance/windows", { token });
      maintenanceWindows = Array.isArray(data) ? data : [];
    } catch (e) {
      maintenanceError = e.message || "Failed to load maintenance windows";
      maintenanceWindows = [];
    } finally {
      maintenanceLoading = false;
    }
  }

  async function initSchedulingView() {
    await Promise.all([loadPlaybooks(), loadScheduledExecs(), loadMaintenanceWindows()]);
  }

  $effect(() => {
    initSchedulingView();
  });

  async function createScheduledExecution() {
    if (!newSchedule.playbook_id || !newSchedule.scheduledDateTime) {
      alert("Please select a playbook and time.");
      return;
    }
    let params = {};
    if (newSchedule.paramsText && newSchedule.paramsText.trim()) {
      try {
        params = JSON.parse(newSchedule.paramsText);
      } catch (e) {
        alert("Parameters must be valid JSON.");
        return;
      }
    }
    const dt = new Date(newSchedule.scheduledDateTime);
    if (Number.isNaN(dt.getTime())) {
      alert("Please provide a valid date/time.");
      return;
    }
    const scheduled_at = dt.toISOString();
    try {
      const { api } = await import("../lib/api.js");
      await api(`/playbooks/${newSchedule.playbook_id}/schedule`, {
        method: "POST",
        token,
        body: {
          parameters: params,
          scheduled_at,
        },
      });
      await loadScheduledExecs();
      // Reset form
      newSchedule = {
        playbook_id: "",
        scheduledDateTime: "",
        paramsText: "{}",
      };
      alert("Scheduled playbook execution created.");
    } catch (e) {
      alert("Failed to schedule playbook: " + e.message);
    }
  }

  async function cancelScheduledExecution(id) {
    if (!confirm("Cancel this scheduled playbook execution?")) return;
    try {
      const { api } = await import("../lib/api.js");
      await api(`/schedule/playbooks/${id}`, { method: "DELETE", token });
      await loadScheduledExecs();
    } catch (e) {
      alert("Failed to cancel scheduled execution: " + e.message);
    }
  }

  function formatDateTime(value) {
    if (!value) return "-";
    try {
      return new Date(value).toLocaleString();
    } catch {
      return value;
    }
  }

  async function createMaintenanceWindow() {
    if (!newWindow.name || !newWindow.startDateTime || !newWindow.endDateTime) {
      alert("Name, start time and end time are required.");
      return;
    }
    const start = new Date(newWindow.startDateTime);
    const end = new Date(newWindow.endDateTime);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
      alert("Invalid start or end time.");
      return;
    }
    if (end <= start) {
      alert("End time must be after start time.");
      return;
    }

    let routingPolicyId = undefined;
    if (newWindow.routingPolicyId && newWindow.routingPolicyId.trim()) {
      const n = Number(newWindow.routingPolicyId);
      if (!Number.isNaN(n)) {
        routingPolicyId = n;
      }
    }

    try {
      const { api } = await import("../lib/api.js");
      await api("/maintenance/windows", {
        method: "POST",
        token,
        body: {
          name: newWindow.name,
          description: newWindow.description,
          start_time: start.toISOString(),
          end_time: end.toISOString(),
          routing_policy_id: routingPolicyId,
          enabled: newWindow.enabled,
        },
      });
      await loadMaintenanceWindows();
      newWindow = {
        name: "",
        description: "",
        startDateTime: "",
        endDateTime: "",
        routingPolicyId: "",
        enabled: true,
      };
      alert("Maintenance window created.");
    } catch (e) {
      alert("Failed to create maintenance window: " + e.message);
    }
  }

  async function toggleMaintenanceWindowEnabled(window) {
    try {
      const { api } = await import("../lib/api.js");
      await api(`/maintenance/windows/${window.id}`, {
        method: "PUT",
        token,
        body: { enabled: !window.enabled },
      });
      await loadMaintenanceWindows();
    } catch (e) {
      alert("Failed to update maintenance window: " + e.message);
    }
  }

  async function deleteMaintenanceWindow(id) {
    if (!confirm("Delete this maintenance window?")) return;
    try {
      const { api } = await import("../lib/api.js");
      await api(`/maintenance/windows/${id}`, { method: "DELETE", token });
      await loadMaintenanceWindows();
    } catch (e) {
      alert("Failed to delete maintenance window: " + e.message);
    }
  }

  const now = $derived(new Date());
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-gray-100 mb-2">Scheduling & Maintenance (E.2)</h2>
    <p class="text-gray-400 text-sm">
      Time-based execution of playbooks and maintenance windows for routing policies.
    </p>
  </div>

  <!-- Scheduled playbook executions -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-semibold text-gray-100">Scheduled Playbook Executions</h3>
        <p class="text-sm text-gray-400">
          Configure time-based execution of operational playbooks.
        </p>
      </div>
      <button
        class="px-3 py-1.5 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-xs hover:bg-gray-700"
        onclick={loadScheduledExecs}
        disabled={scheduledLoading}
      >
        {scheduledLoading ? "Refreshing..." : "Refresh"}
      </button>
    </div>

    {#if canEdit}
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 bg-gray-950/40 border border-gray-800 rounded-lg p-4">
        <div class="space-y-2">
          <label class="block text-sm font-medium text-gray-300 mb-1">Playbook</label>
          <select
            bind:value={newSchedule.playbook_id}
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 text-sm"
          >
            <option value="">— Select playbook —</option>
            {#each playbooks as playbook}
              <option value={playbook.id}>{playbook.name}</option>
            {/each}
          </select>
        </div>
        <div class="space-y-2">
          <label class="block text-sm font-medium text-gray-300 mb-1">Time</label>
          <input
            type="datetime-local"
            bind:value={newSchedule.scheduledDateTime}
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
          />
        </div>
        <div class="space-y-2">
          <label class="block text-sm font-medium text-gray-300 mb-1">Parameters (JSON)</label>
          <textarea
            bind:value={newSchedule.paramsText}
            rows="2"
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-xs font-mono focus:outline-none focus:border-orange-500"
          ></textarea>
        </div>
        <div class="md:col-span-3 flex justify-end">
          <button
            class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium disabled:bg-gray-700 disabled:cursor-not-allowed"
            onclick={createScheduledExecution}
            disabled={scheduledLoading}
          >
            Schedule Playbook
          </button>
        </div>
      </div>
    {/if}

    {#if scheduledError}
      <div class="p-3 rounded-md bg-red-900/30 border border-red-700 text-red-200 text-sm">
        Error: {scheduledError}
      </div>
    {/if}

    {#if scheduledLoading && !scheduledExecs.length}
      <div class="text-center py-6 text-gray-500 text-sm">Loading scheduled executions...</div>
    {:else if !scheduledExecs.length}
      <div class="text-center py-6 text-gray-500 text-sm">
        No scheduled playbook executions.
      </div>
    {:else}
      <div class="space-y-2 max-h-80 overflow-y-auto">
        {#each scheduledExecs as exec}
          <div class="flex items-start justify-between gap-3 px-3 py-2 rounded-md border border-gray-800 bg-gray-900">
            <div class="space-y-0.5 text-xs text-gray-300">
              <div class="flex items-center gap-2">
                <span class="text-gray-100 font-semibold text-sm">{exec.playbook_id}</span>
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-medium {exec.status === 'pending'
                    ? 'bg-yellow-900/60 text-yellow-100 border border-yellow-700/60'
                    : exec.status === 'executed'
                    ? 'bg-green-900/60 text-green-100 border border-green-700/60'
                    : exec.status === 'failed'
                    ? 'bg-red-900/60 text-red-100 border border-red-700/60'
                    : 'bg-gray-800 text-gray-200 border border-gray-700'}"
                >
                  {exec.status}
                </span>
              </div>
              <div>Scheduled: <span class="text-gray-100">{formatDateTime(exec.scheduled_at)}</span></div>
              <div>
                Executed:
                <span class="text-gray-100">
                  {exec.executed_at ? formatDateTime(exec.executed_at) : "—"}
                </span>
              </div>
            </div>
            {#if canEdit && exec.status === "pending"}
              <button
                class="px-3 py-1.5 rounded-md bg-red-900 hover:bg-red-800 text-red-100 text-xs font-medium"
                onclick={() => cancelScheduledExecution(exec.id)}
              >
                Cancel
              </button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Maintenance windows -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-semibold text-gray-100">Maintenance Windows</h3>
        <p class="text-sm text-gray-400">
          Define maintenance periods and attach routing policies.
        </p>
      </div>
      <button
        class="px-3 py-1.5 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-xs hover:bg-gray-700"
        onclick={loadMaintenanceWindows}
        disabled={maintenanceLoading}
      >
        {maintenanceLoading ? "Refreshing..." : "Refresh"}
      </button>
    </div>

    {#if canEdit}
      <div class="bg-gray-950/40 border border-gray-800 rounded-lg p-4 space-y-3">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300 mb-1">Name</label>
            <input
              type="text"
              bind:value={newWindow.name}
              placeholder="e.g. Nightly Maintenance"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300 mb-1">Routing Policy ID (optional)</label>
            <input
              type="text"
              bind:value={newWindow.routingPolicyId}
              placeholder="e.g. 1"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300 mb-1">Start</label>
            <input
              type="datetime-local"
              bind:value={newWindow.startDateTime}
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
          <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300 mb-1">End</label>
            <input
              type="datetime-local"
              bind:value={newWindow.endDateTime}
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Description</label>
          <textarea
            bind:value={newWindow.description}
            rows="2"
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 text-sm focus:outline-none focus:border-orange-500"
          ></textarea>
        </div>
        <div class="flex items-center justify-between">
          <label class="flex items-center gap-2 text-sm text-gray-300">
            <input type="checkbox" bind:checked={newWindow.enabled} class="rounded" />
            Enabled
          </label>
          <button
            class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium disabled:bg-gray-700 disabled:cursor-not-allowed"
            onclick={createMaintenanceWindow}
            disabled={maintenanceLoading}
          >
            Create Window
          </button>
        </div>
      </div>
    {/if}

    {#if maintenanceError}
      <div class="p-3 rounded-md bg-red-900/30 border border-red-700 text-red-200 text-sm">
        Error: {maintenanceError}
      </div>
    {/if}

    {#if maintenanceLoading && !maintenanceWindows.length}
      <div class="text-center py-6 text-gray-500 text-sm">Loading maintenance windows...</div>
    {:else if !maintenanceWindows.length}
      <div class="text-center py-6 text-gray-500 text-sm">
        No maintenance windows defined.
      </div>
    {:else}
      <div class="space-y-3 max-h-80 overflow-y-auto">
        {#each maintenanceWindows as window}
          <div class="flex items-start justify-between gap-3 px-3 py-2 rounded-md border border-gray-800 bg-gray-900">
            <div class="space-y-1 text-xs text-gray-300">
              <div class="flex items-center gap-2">
                <span class="text-sm font-semibold text-gray-100">{window.name}</span>
                {#if window.enabled}
                  <span class="px-2 py-0.5 rounded text-[10px] font-medium bg-green-900/60 text-green-100 border border-green-700/60">
                    Enabled
                  </span>
                {:else}
                  <span class="px-2 py-0.5 rounded text-[10px] font-medium bg-gray-700 text-gray-200 border border-gray-600">
                    Disabled
                  </span>
                {/if}
              </div>
              {#if window.description}
                <div class="text-gray-400">{window.description}</div>
              {/if}
              <div>
                {formatDateTime(window.start_time)} → {formatDateTime(window.end_time)}
              </div>
              <div class="text-gray-400">
                Routing policy ID:
                <span class="text-gray-100">{window.routing_policy_id ?? "—"}</span>
              </div>
            </div>
            {#if canEdit}
              <div class="flex flex-col gap-2">
                <button
                  class="px-3 py-1.5 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-200 text-xs font-medium"
                  onclick={() => toggleMaintenanceWindowEnabled(window)}
                >
                  {window.enabled ? "Disable" : "Enable"}
                </button>
                {#if isAdmin}
                  <button
                    class="px-3 py-1.5 rounded-md bg-red-900 hover:bg-red-800 text-red-100 text-xs font-medium"
                    onclick={() => deleteMaintenanceWindow(window.id)}
                  >
                    Delete
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

