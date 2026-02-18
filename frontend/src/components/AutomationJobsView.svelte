<script>
  export let automationJobs = [];
  export let isAdmin = false;
  export let onToggleAutomationJob;
</script>

<h3>Automation Jobs</h3>
<table style="width:100%;border-collapse:collapse;">
  <thead>
    <tr>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Job ID</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Type</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Schedule</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Enabled</th>
      <th style="text-align:left;border-bottom:1px solid #ddd;padding:8px;">Action</th>
    </tr>
  </thead>
  <tbody>
    {#each automationJobs as job}
      <tr>
        <td style="border-bottom:1px solid #eee;padding:8px;">{job.job_id}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{job.job_type}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{job.schedule_type}: {job.schedule_value}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">{job.enabled ? "ON" : "OFF"}</td>
        <td style="border-bottom:1px solid #eee;padding:8px;">
          {#if isAdmin}
            {#if job.enabled}
              <button on:click={() => onToggleAutomationJob?.(job, false)}>Disable</button>
            {:else}
              <button on:click={() => onToggleAutomationJob?.(job, true)}>Enable</button>
            {/if}
          {:else}
            <span>-</span>
          {/if}
        </td>
      </tr>
    {/each}
  </tbody>
</table>

