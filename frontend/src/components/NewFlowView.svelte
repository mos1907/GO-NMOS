<script>
  export let newFlow;
  export let onCreateFlow;
</script>

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
  <button on:click={onCreateFlow}>Create flow</button>
</div>

