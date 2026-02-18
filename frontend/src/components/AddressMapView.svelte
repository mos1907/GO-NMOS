<script>
  import EmptyState from "./EmptyState.svelte";

  let { addressMap = null } = $props();
</script>

<section class="mt-4 space-y-3">
  <div>
    <h3 class="text-sm font-semibold text-gray-100">Address Map</h3>
    <p class="text-[11px] text-gray-400">View multicast addresses organized by /24 subnets</p>
  </div>

  {#if addressMap}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-6">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <span class="text-sm font-medium text-gray-300">Total Subnets:</span>
          <span class="text-lg font-semibold text-gray-100">{addressMap.total_subnets || 0}</span>
        </div>
      </div>

      {#if (addressMap.items || []).length === 0}
        <div class="text-center py-12">
          <p class="text-gray-400 text-sm">No subnets found</p>
        </div>
      {:else}
        <div class="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-800">
              <tr>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Subnet</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Flow Count</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">IPs</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each (addressMap.items || []) as b}
                <tr class="hover:bg-gray-800/70 transition-colors">
                  <td class="px-4 py-3 text-gray-100 font-medium font-mono text-[11px]">{b.subnet}</td>
                  <td class="px-4 py-3">
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-blue-900 text-blue-200 border border-blue-700">
                      {b.count}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-gray-300 text-[11px] font-mono">
                    {Object.keys(b.flows || {}).slice(0, 6).join(", ")}
                    {#if Object.keys(b.flows || {}).length > 6}
                      <span class="text-gray-500">... (+{Object.keys(b.flows || {}).length - 6} more)</span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  {:else}
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-12">
      <EmptyState
        title="No address map data"
        message="Address map data is not available. This view shows multicast addresses organized by /24 subnets."
      />
    </div>
  {/if}
</section>
