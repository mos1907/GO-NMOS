<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconPlus } from "../lib/icons.js";

  let {
    automationJobs = [],
    isAdmin = false,
    onToggleAutomationJob,
  } = $props();
</script>

<section class="mt-4 space-y-3">
  <div>
    <h3 class="text-sm font-semibold text-gray-100">Automation Jobs</h3>
    <p class="text-[11px] text-gray-400">Manage scheduled automation tasks</p>
  </div>

  <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm overflow-hidden">
    <table class="min-w-full text-xs">
      <thead class="bg-gray-800">
        <tr>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Job ID</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Type</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Schedule</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Next run</th>
          <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Status</th>
          {#if isAdmin}
            <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Action</th>
          {/if}
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-800">
        {#if automationJobs.length === 0}
          <tr>
            <td colspan={isAdmin ? 6 : 5} class="px-6 py-12">
              <EmptyState
                title="No automation jobs"
                message="No scheduled automation jobs configured."
                icon={IconPlus}
              />
            </td>
          </tr>
        {:else}
          {#each automationJobs as job}
            <tr class="hover:bg-gray-800/70 transition-colors">
              <td class="px-4 py-3 text-gray-100 font-medium">{job.job_id}</td>
              <td class="px-4 py-3 text-gray-300">{job.job_type}</td>
              <td class="px-4 py-3 text-gray-300">
                <span class="text-[11px]">{job.schedule_type}: {job.schedule_value}</span>
              </td>
              <td class="px-4 py-3 text-gray-400 text-[11px]">
                {#if job.next_run_at}
                  {new Date(job.next_run_at).toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' })}
                {:else}
                  —
                {/if}
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium {job.enabled
                    ? 'bg-green-900 text-green-200 border border-green-700'
                    : 'bg-gray-800 text-gray-200 border border-gray-700'}"
                >
                  {job.enabled ? "Enabled" : "Disabled"}
                </span>
              </td>
              {#if isAdmin}
                <td class="px-4 py-3">
                  {#if job.enabled}
                    <button
                      on:click={() => onToggleAutomationJob?.(job, false)}
                      class="px-3 py-1.5 rounded-md border border-gray-700 bg-gray-800 text-[11px] text-gray-200 hover:bg-gray-700 transition-colors"
                    >
                      Disable
                    </button>
                  {:else}
                    <button
                      on:click={() => onToggleAutomationJob?.(job, true)}
                      class="px-3 py-1.5 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-[11px] font-medium transition-colors"
                    >
                      Enable
                    </button>
                  {/if}
                </td>
              {:else}
                <td class="px-4 py-3 text-gray-500 text-sm">-</td>
              {/if}
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>
