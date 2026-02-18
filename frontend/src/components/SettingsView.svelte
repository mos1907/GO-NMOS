<script>
  let {
    settings = {},
    isAdmin = false,
    canEdit = false,
    importing = false,
    onSaveSetting,
    onExportFlows,
    onImportFlowsFromFile,
  } = $props();

  let savingKey = $state("");
  let saveSuccess = $state("");

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
              on:click={() => handleSave("api_base_url")}
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
              on:click={() => handleSave("anonymous_access")}
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
              on:click={() => handleSave("flow_lock_role")}
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
              on:click={() => handleSave("hard_delete_enabled")}
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

  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold text-gray-100 mb-4">Backup & Restore</h3>
    <div class="flex flex-wrap gap-3 items-center">
      <button
        on:click={onExportFlows}
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
          <span>Import JSON</span>
          <input
            type="file"
            accept="application/json"
            on:change={onImportFlowsFromFile}
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
      Export flows to JSON file for backup or import flows from a previously exported JSON file.
    </p>
  </div>
</div>

