<script>
  let {
    settings = {},
    isAdmin = false,
    canEdit = false,
    importing = false,
    onSaveSetting,
    onExportFlows,
    onImportFlowsFromFile,
    sdnPingLoading = false,
    sdnPingError = "",
    sdnPingResult = null,
    onPingSDN,
    onFetchSDNTopology = null,
    onFetchSDNPaths = null,
    registryConfigs = [],
    onSaveRegistryConfigs,
    onRemoveRegistry,
    token = "",
  } = $props();

  let savingKey = $state("");
  let saveSuccess = $state("");
  let savingRegistries = $state(false);
  let registrySaveError = $state("");

  // B.4: SDN topology & paths (demo)
  let sdnTopologyLoading = $state(false);
  let sdnTopologyError = $state("");
  let sdnTopology = $state(null);
  let sdnPathsFrom = $state("");
  let sdnPathsTo = $state("");
  let sdnPathsLoading = $state(false);
  let sdnPathsError = $state("");
  let sdnPaths = $state([]);

  // D.2: System Parameters Validation
  let validationLoading = $state(false);
  let validationError = $state("");
  let validationResult = $state(null);

  let editableRegistries = $state((registryConfigs || []).map((r) => ({ ...r })));
  $effect(() => {
    editableRegistries = (registryConfigs || []).map((r) => ({ ...r }));
  });

  let registryToRemove = $state(null);
  let removeRegistryLoading = $state(false);

  // System backup & restore
  let backupLoading = $state(false);
  let restoreLoading = $state(false);
  let restoreOptions = $state({
    restore_flows: true,
    restore_settings: true,
    restore_policies: true,
    restore_registry_config: true,
  });
  let showRestoreModal = $state(false);

  // B.3: Routing policies
  let routingPolicies = $state([]);
  let routingPoliciesLoading = $state(false);
  let showPolicyModal = $state(false);
  let editingPolicy = $state(null);
  let policyForm = $state({
    name: "",
    policy_type: "forbidden_pair",
    enabled: true,
    source_pattern: "",
    destination_pattern: "",
    require_path_a: false,
    require_path_b: false,
    constraint_field: "",
    constraint_value: "",
    constraint_operator: "equals",
    description: "",
    priority: 100,
  });

  async function loadRoutingPolicies() {
    routingPoliciesLoading = true;
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/routing/policies", { token });
      routingPolicies = Array.isArray(data) ? data : [];
    } catch (e) {
      console.error("Failed to load routing policies:", e);
      routingPolicies = [];
    } finally {
      routingPoliciesLoading = false;
    }
  }

  function openPolicyModal(policy = null) {
    editingPolicy = policy;
    if (policy) {
      policyForm = {
        name: policy.name || "",
        policy_type: policy.policy_type || "forbidden_pair",
        enabled: policy.enabled ?? true,
        source_pattern: policy.source_pattern || "",
        destination_pattern: policy.destination_pattern || "",
        require_path_a: policy.require_path_a || false,
        require_path_b: policy.require_path_b || false,
        constraint_field: policy.constraint_field || "",
        constraint_value: policy.constraint_value || "",
        constraint_operator: policy.constraint_operator || "equals",
        description: policy.description || "",
        priority: policy.priority || 100,
      };
    } else {
      policyForm = {
        name: "",
        policy_type: "forbidden_pair",
        enabled: true,
        source_pattern: "",
        destination_pattern: "",
        require_path_a: false,
        require_path_b: false,
        constraint_field: "",
        constraint_value: "",
        constraint_operator: "equals",
        description: "",
        priority: 100,
      };
    }
    showPolicyModal = true;
  }

  async function savePolicy() {
    if (!policyForm.name.trim()) {
      alert("Policy name is required");
      return;
    }
    try {
      const { api } = await import("../lib/api.js");
      if (editingPolicy) {
        await api(`/routing/policies/${editingPolicy.id}`, {
          method: "PUT",
          token,
          body: policyForm,
        });
      } else {
        await api("/routing/policies", {
          method: "POST",
          token,
          body: policyForm,
        });
      }
      showPolicyModal = false;
      editingPolicy = null;
      await loadRoutingPolicies();
    } catch (e) {
      alert("Failed to save policy: " + e.message);
    }
  }

  async function deletePolicy(id) {
    if (!confirm("Delete this routing policy?")) return;
    try {
      const { api } = await import("../lib/api.js");
      await api(`/routing/policies/${id}`, { method: "DELETE", token });
      await loadRoutingPolicies();
    } catch (e) {
      alert("Failed to delete policy: " + e.message);
    }
  }

  async function togglePolicyEnabled(id, enabled) {
    try {
      const { api } = await import("../lib/api.js");
      await api(`/routing/policies/${id}`, {
        method: "PUT",
        token,
        body: { enabled: !enabled },
      });
      await loadRoutingPolicies();
    } catch (e) {
      alert("Failed to update policy: " + e.message);
    }
  }

  async function exportSystemBackup() {
    backupLoading = true;
    try {
      const { api } = await import("../lib/api.js");
      const backup = await api("/system/backup", { token });
      const blob = new Blob([JSON.stringify(backup, null, 2)], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `system-backup-${new Date().toISOString().split("T")[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      alert("Failed to export backup: " + e.message);
    } finally {
      backupLoading = false;
    }
  }

  async function importSystemBackup(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    restoreLoading = true;
    try {
      const text = await file.text();
      const backup = JSON.parse(text);
      if (!backup.version || !backup.timestamp) {
        alert("Invalid backup file format");
        return;
      }
      if (!confirm(`Restore backup from ${backup.timestamp}? This will overwrite current data.`)) {
        return;
      }
      const { api } = await import("../lib/api.js");
      const result = await api("/system/restore", {
        method: "POST",
        token,
        body: {
          backup,
          ...restoreOptions,
        },
      });
      alert(`Restore complete. Restored: ${JSON.stringify(result.restored)}`);
      // Reload page to reflect changes
      window.location.reload();
    } catch (e) {
      alert("Failed to restore backup: " + e.message);
    } finally {
      restoreLoading = false;
      event.target.value = "";
    }
  }

  $effect(() => {
    if (canEdit) loadRoutingPolicies();
  });

  async function handleSave(key) {
    savingKey = key;
    saveSuccess = "";
    try {
      await onSaveSetting?.(key);
      saveSuccess = key;
      setTimeout(() => (saveSuccess = ""), 2000);
    } finally {
      savingKey = "";
    }
  }
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-gray-100 mb-2">Settings & Backup</h2>
    <p class="text-gray-400 text-sm">Configure system settings and manage flow backups</p>
  </div>

  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">System Settings</h3>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <label for="api_base_url" class="block text-sm font-medium text-gray-300">
          API Base URL
        </label>
        <div class="flex gap-2">
          <input
            id="api_base_url"
            type="text"
            bind:value={settings.api_base_url}
            placeholder="http://api.example.com"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("api_base_url")}
              disabled={savingKey === "api_base_url"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "api_base_url"}
                Saving...
              {:else if saveSuccess === "api_base_url"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Base URL for external API integrations</p>
      </div>

      <div class="space-y-2">
        <label for="anonymous_access" class="block text-sm font-medium text-gray-300">
          Anonymous Access
        </label>
        <div class="flex gap-2">
          <input
            id="anonymous_access"
            type="text"
            bind:value={settings.anonymous_access}
            placeholder="false"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("anonymous_access")}
              disabled={savingKey === "anonymous_access"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "anonymous_access"}
                Saving...
              {:else if saveSuccess === "anonymous_access"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Allow unauthenticated access (true/false)</p>
      </div>

      <div class="space-y-2">
        <label for="flow_lock_role" class="block text-sm font-medium text-gray-300">
          Flow Lock Role
        </label>
        <div class="flex gap-2">
          <select
            id="flow_lock_role"
            bind:value={settings.flow_lock_role}
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
          >
            <option value="viewer">viewer</option>
            <option value="editor">editor</option>
            <option value="admin">admin</option>
          </select>
          {#if isAdmin}
            <button
              onclick={() => handleSave("flow_lock_role")}
              disabled={savingKey === "flow_lock_role"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "flow_lock_role"}
                Saving...
              {:else if saveSuccess === "flow_lock_role"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Minimum role required to lock flows</p>
      </div>

      <div class="space-y-2">
        <label for="hard_delete_enabled" class="block text-sm font-medium text-gray-300">
          Hard Delete Enabled
        </label>
        <div class="flex gap-2">
          <select
            id="hard_delete_enabled"
            bind:value={settings.hard_delete_enabled}
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
          >
            <option value="false">false</option>
            <option value="true">true</option>
          </select>
          {#if isAdmin}
            <button
              onclick={() => handleSave("hard_delete_enabled")}
              disabled={savingKey === "hard_delete_enabled"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "hard_delete_enabled"}
                Saving...
              {:else if saveSuccess === "hard_delete_enabled"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Enable permanent deletion of flows (admin only)</p>
      </div>
    </div>
  </div>

  <!-- D.2: System Parameters (IS-09-inspired) -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">System Parameters (IS-09)</h3>
    <p class="text-sm text-gray-400 mb-4">
      Configure PTP domain, GMID, and expected NMOS API versions. These are used for validation and diagnostics.
    </p>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <label for="system_ptp_domain" class="block text-sm font-medium text-gray-300">
          PTP Domain
        </label>
        <div class="flex gap-2">
          <input
            id="system_ptp_domain"
            type="text"
            bind:value={settings.system_ptp_domain}
            placeholder="e.g. 0"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("system_ptp_domain")}
              disabled={savingKey === "system_ptp_domain"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "system_ptp_domain"}
                Saving...
              {:else if saveSuccess === "system_ptp_domain"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">PTP domain number (0-127)</p>
      </div>

      <div class="space-y-2">
        <label for="system_ptp_gmid" class="block text-sm font-medium text-gray-300">
          PTP GMID (Grandmaster ID)
        </label>
        <div class="flex gap-2">
          <input
            id="system_ptp_gmid"
            type="text"
            bind:value={settings.system_ptp_gmid}
            placeholder="e.g. 00:1B:21:FF:FF:00:00:00"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("system_ptp_gmid")}
              disabled={savingKey === "system_ptp_gmid"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "system_ptp_gmid"}
                Saving...
              {:else if saveSuccess === "system_ptp_gmid"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">PTP Grandmaster ID (MAC address format)</p>
      </div>

      <div class="space-y-2">
        <label for="system_expected_is04" class="block text-sm font-medium text-gray-300">
          Expected IS-04 Version
        </label>
        <div class="flex gap-2">
          <input
            id="system_expected_is04"
            type="text"
            bind:value={settings.system_expected_is04}
            placeholder="e.g. v1.3"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("system_expected_is04")}
              disabled={savingKey === "system_expected_is04"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "system_expected_is04"}
                Saving...
              {:else if saveSuccess === "system_expected_is04"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Expected IS-04 Query/Node API version (e.g. v1.3)</p>
      </div>

      <div class="space-y-2">
        <label for="system_expected_is05" class="block text-sm font-medium text-gray-300">
          Expected IS-05 Version
        </label>
        <div class="flex gap-2">
          <input
            id="system_expected_is05"
            type="text"
            bind:value={settings.system_expected_is05}
            placeholder="e.g. v1.1"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("system_expected_is05")}
              disabled={savingKey === "system_expected_is05"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "system_expected_is05"}
                Saving...
              {:else if saveSuccess === "system_expected_is05"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">Expected IS-05 Connection API version (e.g. v1.1)</p>
      </div>
    </div>
  </div>

  <!-- D.2: System Parameters Validation -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">System Parameters Validation</h3>
    <p class="text-sm text-gray-400 mb-4">
      Validate registry and node information against configured system parameters. Check for version mismatches and PTP domain inconsistencies.
    </p>
    <div class="flex items-center justify-between mb-4">
      <div>
        <p class="text-sm text-gray-300">Run validation check</p>
        <p class="text-xs text-gray-500">Compare expected values with actual registry/node values</p>
      </div>
      <button
        onclick={async () => {
          validationLoading = true;
          validationError = "";
          try {
            const { api } = await import("../lib/api.js");
            const result = await api("/system/validation", { token });
            validationResult = result;
          } catch (e) {
            validationError = e.message || "Failed to run validation";
            validationResult = null;
          } finally {
            validationLoading = false;
          }
        }}
        disabled={validationLoading}
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
      >
        {validationLoading ? "Running..." : "Run Validation"}
      </button>
    </div>

    {#if validationError}
      <div class="mb-4 p-3 rounded-md bg-red-900/30 border border-red-700 text-red-200 text-sm">
        Error: {validationError}
      </div>
    {/if}

    {#if validationResult}
      <div class="space-y-4">
        <div class="flex items-center justify-between p-3 rounded-md bg-gray-800 border border-gray-700">
          <div>
            <p class="text-sm font-medium text-gray-200">Validation Summary</p>
            <p class="text-xs text-gray-400 mt-1">
              Found <span class="font-semibold text-orange-400">{validationResult.count || 0}</span> issue(s)
            </p>
          </div>
          {#if validationResult.expected}
            <div class="text-xs text-gray-400 space-y-1">
              <div>Expected IS-04: <span class="text-gray-200">{validationResult.expected.is04_version || "-"}</span></div>
              <div>Expected IS-05: <span class="text-gray-200">{validationResult.expected.is05_version || "-"}</span></div>
              <div>Expected PTP Domain: <span class="text-gray-200">{validationResult.expected.ptp_domain || "-"}</span></div>
            </div>
          {/if}
        </div>

        {#if validationResult.issues && validationResult.issues.length > 0}
          <div class="space-y-2">
            <p class="text-sm font-medium text-gray-300">Issues Found:</p>
            <div class="space-y-2">
              {#each validationResult.issues as issue}
                <div class="p-3 rounded-md border {issue.severity === 'error' ? 'bg-red-900/20 border-red-700' : 'bg-yellow-900/20 border-yellow-700'}">
                  <div class="flex items-start justify-between">
                    <div class="flex-1">
                      <div class="flex items-center gap-2 mb-1">
                        <span class="px-2 py-0.5 rounded text-[10px] font-medium uppercase {issue.severity === 'error' ? 'bg-red-700 text-red-100' : 'bg-yellow-700 text-yellow-100'}">
                          {issue.severity}
                        </span>
                        <span class="px-2 py-0.5 rounded text-[10px] font-medium bg-gray-700 text-gray-300">
                          {issue.type}
                        </span>
                        <span class="text-xs font-medium text-gray-300">{issue.field}</span>
                      </div>
                      <p class="text-sm text-gray-200 mb-1">{issue.message}</p>
                      <div class="text-xs text-gray-400 space-y-0.5">
                        <div>
                          <span class="font-medium">Resource:</span> {issue.resource_label || issue.resource_id}
                        </div>
                        <div>
                          <span class="font-medium">Expected:</span> <span class="text-gray-300">{issue.expected}</span> | 
                          <span class="font-medium">Actual:</span> <span class="text-gray-300">{issue.actual}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {:else}
          <div class="p-4 rounded-md bg-green-900/20 border border-green-700 text-green-200 text-sm text-center">
            ✓ No validation issues found. All registry and node values match expected parameters.
          </div>
        {/if}
      </div>
    {:else if !validationLoading && !validationError}
      <div class="text-sm text-gray-500 text-center py-4">
        Click "Run Validation" to check for mismatches
      </div>
    {/if}
  </div>

  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">Network Controller (IS-06)</h3>
    <div class="space-y-3">
      <div class="space-y-2">
        <label for="sdn_controller_url" class="block text-sm font-medium text-gray-300">
          SDN Controller URL
        </label>
        <div class="flex gap-2">
          <input
            id="sdn_controller_url"
            type="text"
            bind:value={settings.sdn_controller_url}
            placeholder="http://sdn-controller:port/health"
            class="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
          />
          {#if isAdmin}
            <button
              onclick={() => handleSave("sdn_controller_url")}
              disabled={savingKey === "sdn_controller_url"}
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            >
              {#if savingKey === "sdn_controller_url"}
                Saving...
              {:else if saveSuccess === "sdn_controller_url"}
                ✓ Saved
              {:else}
                Save
              {/if}
            </button>
          {/if}
        </div>
        <p class="text-xs text-gray-500">
          Base URL of your SDN / Network Controller. This is used by the ping check below.
        </p>
      </div>

      <div class="flex items-center justify-between pt-2 border-t border-gray-800 mt-2">
        <div class="space-y-1">
          <p class="text-sm font-medium text-gray-200">Ping Controller</p>
          <p class="text-xs text-gray-500">
            Verify that the SDN controller is reachable from GO-NMOS.
          </p>
        </div>
        <button
          class="px-3 py-1.5 rounded-md border border-sky-700 bg-sky-900 text-sky-100 hover:bg-sky-800 disabled:opacity-60 disabled:cursor-not-allowed text-sm font-medium"
          disabled={sdnPingLoading || !onPingSDN}
          onclick={() => onPingSDN?.()}
        >
          {sdnPingLoading ? "Pinging..." : "Ping"}
        </button>
      </div>

      {#if sdnPingError}
        <p class="text-xs text-red-400 mt-1">Error: {sdnPingError}</p>
      {:else if sdnPingResult}
        <div class="mt-2 flex items-center justify-between text-xs">
          <span class="text-gray-300">
            {sdnPingResult.url} →
            {sdnPingResult.httpCode} ({sdnPingResult.status})
          </span>
          <span class="text-gray-400 ml-2">
            {sdnPingResult.latency}
          </span>
        </div>
      {/if}

      <!-- B.4: SDN Topology & Paths (stub demo) -->
      {#if onFetchSDNTopology || onFetchSDNPaths}
        <div class="border-t border-gray-800 pt-4 mt-4 space-y-3">
          <p class="text-sm font-medium text-gray-200">SDN Topology & Paths (B.4)</p>
          <div class="flex flex-wrap gap-2 items-center">
            {#if onFetchSDNTopology}
              <button
                class="px-3 py-1.5 rounded-md border border-gray-600 bg-gray-800 text-gray-200 hover:bg-gray-700 text-sm"
                disabled={sdnTopologyLoading}
                onclick={async () => {
                  sdnTopologyLoading = true;
                  sdnTopologyError = "";
                  sdnTopology = null;
                  try {
                    sdnTopology = await onFetchSDNTopology();
                  } catch (e) {
                    sdnTopologyError = e.message;
                  } finally {
                    sdnTopologyLoading = false;
                  }
                }}
              >
                {sdnTopologyLoading ? "Loading..." : "Fetch topology"}
              </button>
            {/if}
            {#if onFetchSDNPaths}
              <input
                type="text"
                bind:value={sdnPathsFrom}
                placeholder="From (node id)"
                class="w-28 px-2 py-1.5 bg-gray-950 border border-gray-700 rounded text-sm text-gray-200"
              />
              <input
                type="text"
                bind:value={sdnPathsTo}
                placeholder="To (node id)"
                class="w-28 px-2 py-1.5 bg-gray-950 border border-gray-700 rounded text-sm text-gray-200"
              />
              <button
                class="px-3 py-1.5 rounded-md border border-gray-600 bg-gray-800 text-gray-200 hover:bg-gray-700 text-sm"
                disabled={sdnPathsLoading}
                onclick={async () => {
                  sdnPathsLoading = true;
                  sdnPathsError = "";
                  sdnPaths = [];
                  try {
                    const res = await onFetchSDNPaths(sdnPathsFrom, sdnPathsTo);
                    sdnPaths = res?.paths ?? [];
                  } catch (e) {
                    sdnPathsError = e.message;
                  } finally {
                    sdnPathsLoading = false;
                  }
                }}
              >
                {sdnPathsLoading ? "Loading..." : "Query paths"}
              </button>
            {/if}
          </div>
          {#if sdnTopologyError}
            <p class="text-xs text-red-400">Topology: {sdnTopologyError}</p>
          {/if}
          {#if sdnTopology && (sdnTopology.nodes?.length || sdnTopology.links?.length)}
            <div class="text-xs text-gray-400">
              Topology: {sdnTopology.nodes?.length ?? 0} nodes, {sdnTopology.links?.length ?? 0} links
              {#if sdnTopology.nodes?.length}
                <span class="ml-1">— {sdnTopology.nodes.map((n) => n.label || n.id).join(", ")}</span>
              {/if}
            </div>
          {/if}
          {#if sdnPathsError}
            <p class="text-xs text-red-400">Paths: {sdnPathsError}</p>
          {/if}
          {#if sdnPaths.length}
            <ul class="text-xs text-gray-300 space-y-1">
              {#each sdnPaths as p}
                <li>{p.name ?? p.id} (from: {p.from}, to: {p.to})</li>
              {/each}
            </ul>
          {/if}
        </div>
      {/if}
    </div>
  </div>

  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-2">NMOS Registries (IS-04 Query APIs)</h3>
    <p class="text-gray-400 text-sm mb-4">
      Configure external IS-04 Query endpoints (core, lab, remote sites). These are used by the NMOS and Registry & Patch views.
    </p>

    <div class="space-y-3">
      <div class="flex justify-between items-center">
        <div class="text-xs text-gray-400">
          Total: <span class="text-gray-100 font-semibold">{editableRegistries.length}</span>
          {#if editableRegistries.length > 0}
            · Enabled:
            <span class="text-emerald-300 font-semibold">
              {editableRegistries.filter((r) => r.enabled).length}
            </span>
          {/if}
        </div>
        {#if canEdit}
          <button
            class="px-3 py-1.5 rounded-md bg-gray-800 hover:bg-gray-700 text-xs text-gray-100 border border-gray-700"
            onclick={() => {
              editableRegistries = [
                ...editableRegistries,
                { name: "", query_url: "", role: "prod", enabled: true },
              ];
            }}
          >
            + Add registry
          </button>
        {/if}
      </div>

      {#if registrySaveError}
        <p class="text-xs text-red-400">{registrySaveError}</p>
      {/if}

      {#if editableRegistries.length === 0}
        <p class="text-xs text-gray-500">
          No registries configured. Add at least one IS-04 Query endpoint (e.g. core-registry, lab-registry).
        </p>
      {:else}
        <div class="space-y-2">
          {#each editableRegistries as reg, i}
            <div class="border border-gray-800 rounded-lg p-3 bg-gray-950/40 space-y-2">
              <div class="flex items-center justify-between gap-2">
                <div class="flex-1 space-y-1">
                  <label class="block text-xs font-medium text-gray-300">
                    Name
                  </label>
                  <input
                    type="text"
                    class="w-full px-2 py-1.5 rounded-md bg-gray-900 border border-gray-700 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                    placeholder="Core Registry / Lab Registry"
                    value={reg.name}
                    oninput={(e) => (editableRegistries[i].name = e.target.value)}
                  />
                </div>
                <div class="w-28 space-y-1">
                  <label class="block text-xs font-medium text-gray-300">
                    Role
                  </label>
                  <select
                    class="w-full px-2 py-1.5 rounded-md bg-gray-900 border border-gray-700 text-xs text-gray-100 focus:outline-none focus:border-orange-500"
                    value={reg.role}
                    onchange={(e) => (editableRegistries[i].role = e.target.value)}
                  >
                    <option value="prod">prod</option>
                    <option value="lab">lab</option>
                    <option value="remote">remote</option>
                    <option value="other">other</option>
                  </select>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <div class="flex-1 space-y-1">
                  <label class="block text-xs font-medium text-gray-300">
                    Query URL
                  </label>
                  <input
                    type="text"
                    class="w-full px-2 py-1.5 rounded-md bg-gray-900 border border-gray-700 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
                    placeholder="http://host:port/x-nmos/query"
                    value={reg.query_url}
                    oninput={(e) => (editableRegistries[i].query_url = e.target.value)}
                  />
                </div>
                <div class="flex items-center gap-2 pt-5">
                  <label class="flex items-center gap-1 text-xs text-gray-200">
                      <input
                        type="checkbox"
                        class="rounded border-gray-600 bg-gray-900"
                        checked={reg.enabled}
                        onchange={(e) => (editableRegistries[i].enabled = e.target.checked)}
                      />
                    Enabled
                  </label>
                  {#if canEdit}
                    <button
                      class="px-2 py-1 rounded-md text-[11px] text-red-300 hover:text-red-200 hover:bg-red-900/40 border border-transparent hover:border-red-700"
                      onclick={() => {
                        registryToRemove = { name: reg.name || reg.query_url || "RDS", query_url: reg.query_url };
                      }}
                    >
                      Remove
                    </button>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      {#if canEdit}
        <div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-800 mt-2">
          {#if saveSuccess === "registry_configs"}
            <span class="text-xs text-emerald-300">Saved.</span>
          {/if}
          <button
            class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
            disabled={savingRegistries || !onSaveRegistryConfigs}
            onclick={async () => {
              registrySaveError = "";
              savingRegistries = true;
              try {
                await onSaveRegistryConfigs?.(editableRegistries);
                saveSuccess = "registry_configs";
                setTimeout(() => {
                  if (saveSuccess === "registry_configs") saveSuccess = "";
                }, 2000);
              } catch (e) {
                registrySaveError = e?.message || "Failed to save registry configuration";
              } finally {
                savingRegistries = false;
              }
            }}
          >
            {savingRegistries ? "Saving..." : "Save registries"}
          </button>
        </div>
      {/if}
    </div>
  </div>

  <!-- Remove RDS confirmation modal -->
  {#if registryToRemove}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" role="dialog" aria-modal="true" aria-labelledby="remove-registry-title">
      <div class="bg-gray-900 border border-gray-700 rounded-xl p-6 max-w-md w-full mx-4 shadow-xl">
        <h3 id="remove-registry-title" class="text-lg font-semibold text-gray-100 mb-2">Remove RDS</h3>
        <p class="text-sm text-gray-300 mb-4">
          <strong>{registryToRemove.name}</strong> will be removed from configuration.
          All nodes discovered from this registry (and their senders/receivers in Registry Patch) will also be deleted.
        </p>
        <p class="text-xs text-amber-200/90 mb-4">
          This cannot be undone. Reload registry later to re-add nodes.
        </p>
        <div class="flex justify-end gap-3">
          <button
            type="button"
            class="px-4 py-2 rounded-md border border-gray-600 text-gray-200 hover:bg-gray-800"
            disabled={removeRegistryLoading}
            onclick={() => (registryToRemove = null)}
          >
            Cancel
          </button>
          <button
            type="button"
            class="px-4 py-2 rounded-md bg-red-600 hover:bg-red-500 disabled:opacity-50 text-white font-medium"
            disabled={removeRegistryLoading}
            onclick={async () => {
              removeRegistryLoading = true;
              try {
                await onRemoveRegistry?.(registryToRemove.query_url);
                registryToRemove = null;
              } finally {
                removeRegistryLoading = false;
              }
            }}
          >
            {removeRegistryLoading ? "Removing…" : "Remove"}
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- B.3: Routing Policies -->
  {#if canEdit}
    <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-lg font-semibold text-gray-100">Routing Policies</h3>
          <p class="text-sm text-gray-400 mt-1">Define allowed/forbidden source-destination pairs and constraints</p>
        </div>
        <button
          onclick={() => openPolicyModal()}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
        >
          + New Policy
        </button>
      </div>

      {#if routingPoliciesLoading}
        <p class="text-sm text-gray-400 italic">Loading policies...</p>
      {:else if !routingPolicies || routingPolicies.length === 0}
        <p class="text-sm text-gray-400 italic">No routing policies defined.</p>
      {:else}
        <div class="space-y-2">
          {#each routingPolicies as policy}
            <div class="p-4 rounded-lg border border-gray-800 bg-gray-800/50 flex items-start justify-between gap-4">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <span class="px-2 py-0.5 rounded text-xs font-semibold uppercase {policy.enabled ? 'bg-emerald-900/60 text-emerald-200' : 'bg-gray-700 text-gray-400'}">{policy.enabled ? "Enabled" : "Disabled"}</span>
                  <span class="px-2 py-0.5 rounded text-xs bg-blue-900/60 text-blue-200">{policy.policy_type}</span>
                  <span class="text-sm font-semibold text-gray-100">{policy.name}</span>
                  <span class="text-xs text-gray-500">Priority: {policy.priority}</span>
                </div>
                {#if policy.source_pattern || policy.destination_pattern}
                  <p class="text-xs text-gray-400">
                    Source: <span class="text-gray-300">{policy.source_pattern || "*"}</span> → Destination: <span class="text-gray-300">{policy.destination_pattern || "*"}</span>
                  </p>
                {/if}
                {#if policy.constraint_field}
                  <p class="text-xs text-gray-400">
                    Constraint: <span class="text-gray-300">{policy.constraint_field}</span> {policy.constraint_operator} <span class="text-gray-300">{policy.constraint_value}</span>
                  </p>
                {/if}
                {#if policy.description}
                  <p class="text-xs text-gray-500 mt-1">{policy.description}</p>
                {/if}
              </div>
              <div class="flex gap-2 shrink-0">
                <button
                  onclick={() => togglePolicyEnabled(policy.id, policy.enabled)}
                  class="px-2 py-1 rounded text-xs border border-gray-700 hover:bg-gray-700 text-gray-300"
                >
                  {policy.enabled ? "Disable" : "Enable"}
                </button>
                <button
                  onclick={() => openPolicyModal(policy)}
                  class="px-2 py-1 rounded text-xs border border-gray-700 hover:bg-gray-700 text-gray-300"
                >
                  Edit
                </button>
                <button
                  onclick={() => deletePolicy(policy.id)}
                  class="px-2 py-1 rounded text-xs border border-red-700 hover:bg-red-900/40 text-red-300"
                >
                  Delete
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">Backup & Restore</h3>
    
    <!-- System Backup -->
    <div class="mb-6 pb-6 border-b border-gray-800">
      <h4 class="text-sm font-semibold text-gray-200 mb-3">System Backup</h4>
      <div class="flex flex-wrap gap-3 items-center">
        <button
          onclick={exportSystemBackup}
          disabled={backupLoading}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors flex items-center gap-2"
        >
          {#if backupLoading}
            <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Exporting...
          {:else}
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Export Full Backup
          {/if}
        </button>
        {#if isAdmin}
          <label class="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-gray-700 bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium cursor-pointer transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            <span>{restoreLoading ? "Restoring..." : "Restore Backup"}</span>
            <input
              type="file"
              accept="application/json"
              onchange={importSystemBackup}
              disabled={restoreLoading}
              class="hidden"
            />
          </label>
        {/if}
      </div>
      <p class="text-xs text-gray-500 mt-3">
        Export complete system backup (settings, policies, registry config, flows) or restore from backup file.
      </p>
    </div>

    <!-- Flows Only -->
    <div>
      <h4 class="text-sm font-semibold text-gray-200 mb-3">Flows Only</h4>
      <div class="flex flex-wrap gap-3 items-center">
        <button
          onclick={onExportFlows}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Export Flows JSON
        </button>
        {#if canEdit}
          <label class="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-gray-700 bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium cursor-pointer transition-colors">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            <span>Import Flows JSON</span>
            <input
              type="file"
              accept="application/json"
              onchange={onImportFlowsFromFile}
              disabled={importing}
              class="hidden"
            />
          </label>
          {#if importing}
            <span class="text-sm text-gray-400 flex items-center gap-2">
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Importing...
            </span>
          {/if}
        {/if}
      </div>
      <p class="text-xs text-gray-500 mt-3">
        Export only flows to JSON file for backup or import flows from a previously exported JSON file.
      </p>
    </div>
  </div>

  <!-- B.3: Policy Modal -->
  {#if showPolicyModal}
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70"
      role="dialog"
      aria-modal="true"
      onclick={() => (showPolicyModal = false)}
      onkeydown={(e) => e.key === "Escape" && (showPolicyModal = false)}
    >
      <div class="w-full max-w-2xl rounded-xl border border-gray-800 bg-gray-900 p-6 max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-100">{editingPolicy ? "Edit" : "New"} Routing Policy</h3>
          <button
            class="px-3 py-1 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm"
            onclick={() => (showPolicyModal = false)}
          >
            Close
          </button>
        </div>

        <div class="space-y-4">
          <div>
            <label for="policy-name" class="block text-sm font-medium text-gray-300 mb-1">Name *</label>
            <input
              id="policy-name"
              type="text"
              bind:value={policyForm.name}
              placeholder="e.g. No test feeds on TX outputs"
              class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
            />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="policy-type" class="block text-sm font-medium text-gray-300 mb-1">Policy Type</label>
              <select id="policy-type" bind:value={policyForm.policy_type} class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100">
                <option value="forbidden_pair">Forbidden Pair</option>
                <option value="allowed_pair">Allowed Pair</option>
                <option value="path_requirement">Path Requirement</option>
                <option value="constraint">Constraint</option>
              </select>
            </div>
            <div>
              <label for="policy-priority" class="block text-sm font-medium text-gray-300 mb-1">Priority (lower = higher)</label>
              <input
                id="policy-priority"
                type="number"
                bind:value={policyForm.priority}
                class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
              />
            </div>
          </div>

          {#if policyForm.policy_type === "forbidden_pair" || policyForm.policy_type === "allowed_pair"}
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label for="source-pattern" class="block text-sm font-medium text-gray-300 mb-1">Source Pattern</label>
                <input
                  id="source-pattern"
                  type="text"
                  bind:value={policyForm.source_pattern}
                  placeholder="sender:* or flow:test-*"
                  class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
                />
                <p class="text-xs text-gray-500 mt-1">Use * for wildcard, e.g. sender:*</p>
              </div>
              <div>
                <label for="destination-pattern" class="block text-sm font-medium text-gray-300 mb-1">Destination Pattern</label>
                <input
                  id="destination-pattern"
                  type="text"
                  bind:value={policyForm.destination_pattern}
                  placeholder="receiver:* or device:TX-*"
                  class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
                />
              </div>
            </div>
          {/if}

          {#if policyForm.policy_type === "path_requirement"}
            <div class="flex gap-4">
              <label class="flex items-center gap-2 text-sm text-gray-300">
                <input type="checkbox" bind:checked={policyForm.require_path_a} />
                Require Path-A
              </label>
              <label class="flex items-center gap-2 text-sm text-gray-300">
                <input type="checkbox" bind:checked={policyForm.require_path_b} />
                Require Path-B
              </label>
            </div>
          {/if}

          {#if policyForm.policy_type === "constraint"}
            <div class="grid grid-cols-3 gap-4">
              <div>
                <label for="constraint-field" class="block text-sm font-medium text-gray-300 mb-1">Field</label>
                <select id="constraint-field" bind:value={policyForm.constraint_field} class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100">
                  <option value="">Select...</option>
                  <option value="flow_label">Flow Label</option>
                  <option value="sender_label">Sender Label</option>
                </select>
              </div>
              <div>
                <label for="constraint-operator" class="block text-sm font-medium text-gray-300 mb-1">Operator</label>
                <select id="constraint-operator" bind:value={policyForm.constraint_operator} class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100">
                  <option value="equals">Equals</option>
                  <option value="contains">Contains</option>
                  <option value="starts_with">Starts With</option>
                  <option value="ends_with">Ends With</option>
                </select>
              </div>
              <div>
                <label for="constraint-value" class="block text-sm font-medium text-gray-300 mb-1">Value</label>
                <input
                  id="constraint-value"
                  type="text"
                  bind:value={policyForm.constraint_value}
                  placeholder="e.g. test"
                  class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
                />
              </div>
            </div>
          {/if}

          <div>
            <label for="policy-description" class="block text-sm font-medium text-gray-300 mb-1">Description</label>
            <textarea
              id="policy-description"
              bind:value={policyForm.description}
              rows="2"
              class="w-full px-3 py-2 rounded-md bg-gray-800 border border-gray-700 text-gray-100"
            ></textarea>
          </div>

          <div class="flex items-center gap-2">
            <label class="flex items-center gap-2 text-sm text-gray-300">
              <input type="checkbox" bind:checked={policyForm.enabled} />
              Enabled
            </label>
          </div>

          <div class="flex gap-2 justify-end pt-2">
            <button
              class="px-4 py-2 rounded-md border border-gray-700 text-gray-300 hover:bg-gray-800"
              onclick={() => (showPolicyModal = false)}
            >
              Cancel
            </button>
            <button
              class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white font-semibold"
              onclick={savePolicy}
            >
              {editingPolicy ? "Update" : "Create"} Policy
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

