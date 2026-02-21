<script>
  let {
    token = "",
    isAdmin = false,
    canEdit = false,
    userRole = "viewer", // E.3: Current user's role
  } = $props();

  let playbooks = $state([]);
  let loading = $state(false);
  let error = $state("");
  let selectedPlaybook = $state(null);
  let showExecuteModal = $state(false);
  let executeParams = $state({});
  let executing = $state(false);
  let executions = $state([]);
  let showExecutions = $state(false);
  let showCreateModal = $state(false);
  let newPlaybook = $state({
    id: "",
    name: "",
    description: "",
    steps: "[]",
    parameters: "{}",
    allowed_roles: ["engineer", "admin"], // E.3: Default allowed roles
    enabled: true,
  });

  const stepsPlaceholder = '[{"action": "disconnect_receiver", "receiver_id": "{{receiver_id}}"}]';
  const parametersPlaceholder = '{"receiver_id": {"type": "string", "description": "Receiver ID", "required": true}}';

  async function loadPlaybooks() {
    loading = true;
    error = "";
    try {
      const { api } = await import("../lib/api.js");
      const data = await api("/playbooks", { token });
      playbooks = Array.isArray(data) ? data : [];
    } catch (e) {
      error = e.message || "Failed to load playbooks";
      playbooks = [];
    } finally {
      loading = false;
    }
  }

  async function loadExecutions(playbookId) {
    try {
      const { api } = await import("../lib/api.js");
      const data = await api(`/playbooks/${playbookId}/executions?limit=20`, { token });
      executions = Array.isArray(data) ? data : [];
    } catch (e) {
      console.error("Failed to load executions:", e);
      executions = [];
    }
  }

  async function executePlaybook(playbook) {
    selectedPlaybook = playbook;
    executeParams = {};
    showExecuteModal = true;
    // Parse parameters to initialize form
    try {
      const params = JSON.parse(playbook.parameters || "{}");
      for (const [key, def] of Object.entries(params)) {
        executeParams[key] = "";
      }
    } catch (e) {
      console.error("Failed to parse parameters:", e);
    }
  }

  async function doExecute() {
    executing = true;
    try {
      const { api } = await import("../lib/api.js");
      const result = await api(`/playbooks/${selectedPlaybook.id}/execute`, {
        method: "POST",
        token,
        body: { parameters: executeParams },
      });
      alert(`Playbook executed successfully!\nExecution ID: ${result.execution_id}`);
      showExecuteModal = false;
      if (showExecutions) {
        await loadExecutions(selectedPlaybook.id);
      }
    } catch (e) {
      alert("Failed to execute playbook: " + e.message);
    } finally {
      executing = false;
    }
  }

  async function savePlaybook() {
    try {
      const { api } = await import("../lib/api.js");
      await api(`/playbooks/${newPlaybook.id}`, {
        method: "PUT",
        token,
        body: {
          name: newPlaybook.name,
          description: newPlaybook.description,
          steps: JSON.parse(newPlaybook.steps || "[]"),
          parameters: JSON.parse(newPlaybook.parameters || "{}"),
          allowed_roles: newPlaybook.allowed_roles || ["engineer", "admin"],
          enabled: newPlaybook.enabled,
        },
      });
      alert("Playbook saved successfully!");
      showCreateModal = false;
      await loadPlaybooks();
    } catch (e) {
      alert("Failed to save playbook: " + e.message);
    }
  }

  async function deletePlaybook(playbook) {
    if (!confirm(`Delete playbook "${playbook.name}"?`)) return;
    try {
      const { api } = await import("../lib/api.js");
      await api(`/playbooks/${playbook.id}`, { method: "DELETE", token });
      await loadPlaybooks();
    } catch (e) {
      alert("Failed to delete playbook: " + e.message);
    }
  }

  $effect(() => {
    loadPlaybooks();
  });
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-2xl font-bold text-gray-100 mb-2">Operational Playbooks (E.1)</h2>
      <p class="text-gray-400 text-sm">Reusable workflows for TV campus operations</p>
    </div>
    {#if isAdmin}
      <button
        onclick={() => {
          newPlaybook = { id: "", name: "", description: "", steps: "[]", parameters: "{}", allowed_roles: ["engineer", "admin"], enabled: true };
          showCreateModal = true;
        }}
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
      >
        + New Playbook
      </button>
    {/if}
  </div>

  {#if error}
    <div class="p-4 rounded-md bg-red-900/30 border border-red-700 text-red-200 text-sm">
      Error: {error}
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-8 text-gray-400">Loading playbooks...</div>
  {:else if playbooks.length === 0}
    <div class="text-center py-8 text-gray-500">No playbooks found. Create one to get started.</div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each playbooks as playbook}
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-start justify-between mb-2">
            <div class="flex-1">
              <h3 class="text-lg font-semibold text-gray-100">{playbook.name}</h3>
              <p class="text-sm text-gray-400 mt-1">{playbook.description || "No description"}</p>
            </div>
            {#if playbook.enabled}
              <span class="px-2 py-1 rounded text-xs font-medium bg-green-900/60 text-green-200 border border-green-700/50">
                Enabled
              </span>
            {:else}
              <span class="px-2 py-1 rounded text-xs font-medium bg-gray-700 text-gray-300 border border-gray-600">
                Disabled
              </span>
            {/if}
          </div>
          <!-- E.3: Show allowed roles -->
          {#if playbook.allowed_roles && playbook.allowed_roles.length > 0}
            <div class="mt-2 flex flex-wrap gap-1">
              <span class="text-xs text-gray-500">Allowed roles:</span>
              {#each playbook.allowed_roles as role}
                <span class="px-1.5 py-0.5 rounded text-[10px] font-medium bg-indigo-900/60 text-indigo-200 border border-indigo-700/50">
                  {role}
                </span>
              {/each}
            </div>
          {/if}
          <div class="mt-4 flex gap-2">
            {#if canEdit}
              <button
                onclick={() => executePlaybook(playbook)}
                disabled={
                  !(
                    playbook.enabled &&
                    canEdit &&
                    (playbook.allowed_roles?.includes(userRole) || userRole === "admin" || isAdmin)
                  )
                }
                class="flex-1 px-3 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
                title={
                  !(
                    playbook.enabled &&
                    canEdit &&
                    (playbook.allowed_roles?.includes(userRole) || userRole === "admin" || isAdmin)
                  )
                    ? "Your role is not allowed to execute this playbook"
                    : ""
                }
              >
                Execute
              </button>
              <button
                onclick={async () => {
                  selectedPlaybook = playbook;
                  showExecutions = true;
                  await loadExecutions(playbook.id);
                }}
                class="px-3 py-2 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                History
              </button>
            {/if}
            {#if isAdmin}
              <button
                onclick={() => deletePlaybook(playbook)}
                class="px-3 py-2 rounded-md bg-red-900 hover:bg-red-800 text-red-200 text-sm font-medium transition-colors"
              >
                Delete
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Execute Modal -->
{#if showExecuteModal && selectedPlaybook}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
    <div class="w-full max-w-md rounded-lg border border-gray-800 bg-gray-900 p-6 space-y-4">
      <h3 class="text-lg font-semibold text-gray-100">Execute: {selectedPlaybook.name}</h3>
      <p class="text-sm text-gray-400">{selectedPlaybook.description}</p>
      <div class="space-y-3">
        {#each Object.entries(executeParams) as [key, value]}
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">{key}</label>
            <input
              type="text"
              bind:value={executeParams[key]}
              placeholder="Enter {key}"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
            />
          </div>
        {/each}
      </div>
      <div class="flex gap-3">
        <button
          onclick={doExecute}
          disabled={executing}
          class="flex-1 px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
        >
          {executing ? "Executing..." : "Execute"}
        </button>
        <button
          onclick={() => (showExecuteModal = false)}
          class="px-4 py-2 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Executions History Modal -->
{#if showExecutions && selectedPlaybook}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
    <div class="w-full max-w-2xl rounded-lg border border-gray-800 bg-gray-900 p-6 space-y-4 max-h-[80vh] overflow-y-auto">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-100">Execution History: {selectedPlaybook.name}</h3>
        <button
          onclick={() => {
            showExecutions = false;
            executions = [];
          }}
          class="px-3 py-1 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm"
        >
          Close
        </button>
      </div>
      {#if executions.length === 0}
        <div class="text-center py-8 text-gray-500">No executions yet</div>
      {:else}
        <div class="space-y-3">
          {#each executions as exec}
            <div class="p-4 rounded-md border {exec.status === 'success' ? 'bg-green-900/20 border-green-700' : exec.status === 'error' ? 'bg-red-900/20 border-red-700' : 'bg-gray-800 border-gray-700'}">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-gray-200">
                  {new Date(exec.started_at).toLocaleString()}
                </span>
                <span class="px-2 py-1 rounded text-xs font-medium {exec.status === 'success' ? 'bg-green-700 text-green-100' : exec.status === 'error' ? 'bg-red-700 text-red-100' : 'bg-gray-700 text-gray-300'}">
                  {exec.status}
                </span>
              </div>
              {#if exec.result}
                {@const result = typeof exec.result === 'string' ? JSON.parse(exec.result) : exec.result}
                {#if result.steps}
                  <div class="mt-2 space-y-1 text-xs text-gray-400">
                    {#each result.steps as step}
                      <div>Step {step.step}: {step.action} {step.success ? '✓' : step.error ? '✗' : ''}</div>
                    {/each}
                  </div>
                {/if}
                {#if result.error}
                  <div class="mt-2 text-xs text-red-400">{result.error}</div>
                {/if}
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}

<!-- Create/Edit Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
    <div class="w-full max-w-2xl rounded-lg border border-gray-800 bg-gray-900 p-6 space-y-4 max-h-[80vh] overflow-y-auto">
      <h3 class="text-lg font-semibold text-gray-100">New Playbook</h3>
      <div class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">ID</label>
          <input
            type="text"
            bind:value={newPlaybook.id}
            placeholder="e.g. failover_backup_encoder"
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Name</label>
          <input
            type="text"
            bind:value={newPlaybook.name}
            placeholder="e.g. Failover to Backup Encoder"
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Description</label>
          <textarea
            bind:value={newPlaybook.description}
            placeholder="Describe what this playbook does"
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
            rows="2"
          ></textarea>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Steps (JSON array)</label>
          <textarea
            bind:value={newPlaybook.steps}
            placeholder={stepsPlaceholder}
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 font-mono text-xs"
            rows="6"
          ></textarea>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Parameters (JSON object)</label>
          <textarea
            bind:value={newPlaybook.parameters}
            placeholder={parametersPlaceholder}
            class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 font-mono text-xs"
            rows="4"
          ></textarea>
        </div>
        <!-- E.3: Allowed roles selection -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Allowed Roles</label>
          <div class="flex flex-wrap gap-2">
            {#each ["viewer", "operator", "engineer", "admin", "automation"] as role}
              <label class="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-gray-700 bg-gray-900 hover:bg-gray-800 cursor-pointer">
                <input
                  type="checkbox"
                  checked={newPlaybook.allowed_roles?.includes(role) || false}
                  onchange={(e) => {
                    if (!newPlaybook.allowed_roles) newPlaybook.allowed_roles = [];
                    if (e.currentTarget.checked) {
                      if (!newPlaybook.allowed_roles.includes(role)) {
                        newPlaybook.allowed_roles = [...newPlaybook.allowed_roles, role];
                      }
                    } else {
                      newPlaybook.allowed_roles = newPlaybook.allowed_roles.filter((r) => r !== role);
                    }
                  }}
                  class="rounded"
                />
                <span class="text-xs text-gray-300">{role}</span>
              </label>
            {/each}
          </div>
          <p class="text-xs text-gray-500 mt-1">Select which roles can execute this playbook</p>
        </div>
        <div>
          <label class="flex items-center gap-2">
            <input type="checkbox" bind:checked={newPlaybook.enabled} class="rounded" />
            <span class="text-sm text-gray-300">Enabled</span>
          </label>
        </div>
      </div>
      <div class="flex gap-3">
        <button
          onclick={savePlaybook}
          class="flex-1 px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
        >
          Save
        </button>
        <button
          onclick={() => (showCreateModal = false)}
          class="px-4 py-2 rounded-md bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}
