<script>
  export let checkerResult = null;
  export let onRunCollisionCheck;
</script>

<h3>Collision Checker</h3>
<button on:click={onRunCollisionCheck}>Run collision check now</button>
{#if checkerResult}
  <p style="margin-top:10px;">
    Total collisions: {checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0}
  </p>
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

