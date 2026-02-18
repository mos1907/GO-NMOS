<script>
  export let settings = {};
  export let isAdmin = false;
  export let canEdit = false;
  export let importing = false;

  export let onSaveSetting;
  export let onExportFlows;
  export let onImportFlowsFromFile;
</script>

<h3>Settings & Backup</h3>
<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:10px;">
  <label>api_base_url
    <input bind:value={settings.api_base_url} />
    {#if isAdmin}<button on:click={() => onSaveSetting?.("api_base_url")}>Save</button>{/if}
  </label>
  <label>anonymous_access
    <input bind:value={settings.anonymous_access} />
    {#if isAdmin}<button on:click={() => onSaveSetting?.("anonymous_access")}>Save</button>{/if}
  </label>
  <label>flow_lock_role
    <input bind:value={settings.flow_lock_role} />
    {#if isAdmin}<button on:click={() => onSaveSetting?.("flow_lock_role")}>Save</button>{/if}
  </label>
  <label>hard_delete_enabled
    <input bind:value={settings.hard_delete_enabled} />
    {#if isAdmin}<button on:click={() => onSaveSetting?.("hard_delete_enabled")}>Save</button>{/if}
  </label>
</div>
<div style="margin-top:16px;display:flex;gap:8px;align-items:center;">
  <button on:click={onExportFlows}>Export flows JSON</button>
  {#if canEdit}
    <label style="display:inline-flex;align-items:center;gap:8px;">
      <span>Import JSON:</span>
      <input type="file" accept="application/json" on:change={onImportFlowsFromFile} disabled={importing} />
    </label>
  {/if}
</div>

