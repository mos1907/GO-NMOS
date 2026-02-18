<script>
  import EmptyState from "./EmptyState.svelte";
  import { IconRefresh } from "../lib/icons.js";

  let {
    checkerResult = null,
    onRunCollisionCheck,
  } = $props();
</script>

<section class="mt-4 space-y-3">
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <div>
      <h3 class="text-sm font-semibold text-gray-100">Collision Checker</h3>
      <p class="text-[11px] text-gray-400">Detect multicast address and port collisions</p>
    </div>
    <div class="flex gap-2">
      <button
        on:click={onRunCollisionCheck}
        class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
      >
        {@html IconRefresh}
        Run Collision Check
      </button>
    </div>
  </div>

  {#if checkerResult}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-6">
      <div class="flex items-center gap-4 mb-4">
        <div class="flex items-center gap-2">
          <div
            class="w-3 h-3 rounded-full {(checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0) > 0
              ? 'bg-red-500'
              : 'bg-green-500'}"
          ></div>
          <span class="text-lg font-semibold text-gray-100">
            Total Collisions: {checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0}
          </span>
        </div>
      </div>

      {#if (checkerResult.result?.items || checkerResult.items || []).length > 0}
        <div class="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-800">
              <tr>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Multicast IP</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Port</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Count</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Flows</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each (checkerResult.result?.items || checkerResult.items || []) as item}
                <tr class="hover:bg-gray-800/70 transition-colors">
                  <td class="px-4 py-3 text-gray-100 font-medium">{item.multicast_ip}</td>
                  <td class="px-4 py-3 text-gray-300">{item.port}</td>
                  <td class="px-4 py-3">
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-red-900 text-red-200 border border-red-700">
                      {item.count}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-gray-300 text-[11px]">
                    {(item.flow_names || []).join(", ")}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else if (checkerResult.result?.total_collisions ?? checkerResult.total_collisions ?? 0) === 0}
        <div class="text-center py-8">
          <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-900/20 mb-4">
            <svg class="w-8 h-8 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <p class="text-gray-300 font-medium">No collisions detected</p>
          <p class="text-gray-400 text-sm mt-1">All flows have unique multicast addresses and ports</p>
        </div>
      {/if}
    </div>
  {:else}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-12">
      <EmptyState
        title="No check results"
        message="Click 'Run Collision Check' to analyze your flows for address and port conflicts."
        icon={IconRefresh}
      />
    </div>
  {/if}
</section>
