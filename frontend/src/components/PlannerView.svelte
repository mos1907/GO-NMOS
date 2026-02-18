<script>
  export let plannerRoots = [];
  export let plannerChildren = [];
  export let selectedPlannerRoot = null;
  export let newPlannerParent;
  export let newPlannerChild;
  export let canEdit = false;
  export let isAdmin = false;

  export let onSelectPlannerRoot;
  export let onPlannerQuickEdit;
  export let onPlannerDelete;
  export let onCreatePlannerParent;
  export let onCreatePlannerChild;
  export let onExportBuckets;
  export let onImportBucketsFromFile;
</script>

<h3>Planner Buckets</h3>
<div style="display:grid;grid-template-columns:1fr 2fr;gap:12px;">
  <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
    <h4>Drives</h4>
    {#each plannerRoots as root}
      <div style="display:flex;justify-content:space-between;margin:4px 0;">
        <button on:click={() => onSelectPlannerRoot?.(root)}>{root.name}</button>
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
                <button on:click={() => onPlannerQuickEdit?.(item)}>Edit</button>
                {#if isAdmin}
                  <button on:click={() => onPlannerDelete?.(item)} style="margin-left:6px;">Delete</button>
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
      <button on:click={onCreatePlannerParent} style="margin-top:8px;">Create Parent</button>
    </div>
    <div style="border:1px solid #ddd;border-radius:8px;padding:10px;">
      <h4>Create View (child)</h4>
      <input bind:value={newPlannerChild.name} placeholder="Name" />
      <input bind:value={newPlannerChild.cidr} placeholder="CIDR or range label" />
      <input bind:value={newPlannerChild.description} placeholder="Description" />
      <input bind:value={newPlannerChild.color} placeholder="Color (optional)" />
      <button
        on:click={() => onCreatePlannerChild?.(selectedPlannerRoot)}
        style="margin-top:8px;"
        disabled={!selectedPlannerRoot}
      >
        Create Child
      </button>
    </div>
  </div>
{/if}

<div style="margin-top:12px;display:flex;gap:8px;">
  <button on:click={onExportBuckets}>Export Planner</button>
  {#if canEdit}
    <label style="display:inline-flex;align-items:center;gap:8px;">
      <span>Import Planner:</span>
      <input type="file" accept="application/json" on:change={onImportBucketsFromFile} />
    </label>
  {/if}
</div>

