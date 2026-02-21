<script>
  import { api } from "../lib/api.js";
  import EmptyState from "./EmptyState.svelte";

  let { 
    addressMap = null, 
    token = "",
    onNavigateToPlanner = null,
  } = $props();

  let selectedSubnet = $state(null);
  let subnetAnalysis = $state(null);
  let loadingAnalysis = $state(false);
  let showAnalysisModal = $state(false);

  async function loadSubnetAnalysis(subnet) {
    if (!subnet || !token) return;
    loadingAnalysis = true;
    selectedSubnet = subnet;
    try {
      subnetAnalysis = await api(`/address-map/subnet/analysis?subnet=${encodeURIComponent(subnet)}`, { token });
      showAnalysisModal = true;
    } catch (e) {
      console.error('Failed to load subnet analysis:', e);
      subnetAnalysis = null;
    } finally {
      loadingAnalysis = false;
    }
  }

  function closeAnalysisModal() {
    showAnalysisModal = false;
    selectedSubnet = null;
    subnetAnalysis = null;
  }

  function getUsageColor(percentage) {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    if (percentage >= 50) return 'bg-blue-500';
    return 'bg-green-500';
  }
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
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Planner Bucket</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Flow Count</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Usage</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">IPs</th>
                <th class="text-left border-b border-gray-800 px-4 py-3 font-medium text-gray-200">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each (addressMap.items || []) as b}
                <tr class="hover:bg-gray-800/70 transition-colors">
                  <td class="px-4 py-3 text-gray-100 font-medium font-mono text-[11px]">{b.subnet}</td>
                  <td class="px-4 py-3">
                    {#if b.bucket_id}
                      {#if onNavigateToPlanner}
                        <button
                          onclick={() => onNavigateToPlanner(b.bucket_id)}
                          class="text-blue-400 hover:text-blue-300 hover:underline text-[11px] transition-colors"
                        >
                          {b.bucket_name || `Bucket #${b.bucket_id}`}
                        </button>
                      {:else}
                        <span class="text-gray-300 text-[11px]">{b.bucket_name || `Bucket #${b.bucket_id}`}</span>
                      {/if}
                      {#if b.bucket_cidr}
                        <span class="text-gray-500 text-[10px] ml-1">({b.bucket_cidr})</span>
                      {/if}
                    {:else}
                      <span class="text-gray-500 text-[10px]">-</span>
                    {/if}
                  </td>
                  <td class="px-4 py-3">
                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium bg-blue-900 text-blue-200 border border-blue-700">
                      {b.count}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    {#if b.usage_percentage !== undefined}
                      <div class="space-y-1">
                        <div class="flex items-center gap-2">
                          <div class="flex-1 bg-gray-700 rounded-full h-2 overflow-hidden">
                            <div
                              class={`h-full ${getUsageColor(b.usage_percentage)}`}
                              style="width: {Math.min(b.usage_percentage, 100)}%"
                            ></div>
                          </div>
                          <span class="text-[10px] text-gray-400 min-w-[45px] text-right">
                            {b.usage_percentage.toFixed(1)}%
                          </span>
                        </div>
                        <div class="text-[9px] text-gray-500">
                          {b.used_ips || 0} / {b.total_ips || 254} IPs
                        </div>
                      </div>
                    {:else}
                      <span class="text-[10px] text-gray-500">-</span>
                    {/if}
                  </td>
                  <td class="px-4 py-3 text-gray-300 text-[11px] font-mono">
                    {Object.keys(b.flows || {}).slice(0, 6).join(", ")}
                    {#if Object.keys(b.flows || {}).length > 6}
                      <span class="text-gray-500">... (+{Object.keys(b.flows || {}).length - 6} more)</span>
                    {/if}
                  </td>
                  <td class="px-4 py-3">
                    <button
                      onclick={() => loadSubnetAnalysis(b.subnet)}
                      disabled={loadingAnalysis}
                      class="px-2 py-1 text-[10px] bg-gray-700 hover:bg-gray-600 text-gray-200 rounded disabled:opacity-50"
                    >
                      Details
                    </button>
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

  <!-- Subnet Analysis Modal -->
  {#if showAnalysisModal && subnetAnalysis}
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onclick={closeAnalysisModal}
      role="dialog"
      aria-modal="true"
    >
      <div
        class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-y-auto"
        onclick={(e) => e.stopPropagation()}
      >
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
          <h2 class="text-xl font-bold text-gray-100">Subnet Analysis: {selectedSubnet}</h2>
          <button
            onclick={closeAnalysisModal}
            class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
            aria-label="Close"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-6 space-y-6">
          <!-- Summary Stats -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="bg-gray-800 rounded-lg p-4">
              <div class="text-sm text-gray-400 mb-1">Total IPs</div>
              <div class="text-2xl font-bold text-gray-100">{subnetAnalysis.total_ips || 254}</div>
            </div>
            <div class="bg-gray-800 rounded-lg p-4">
              <div class="text-sm text-gray-400 mb-1">Used IPs</div>
              <div class="text-2xl font-bold text-orange-400">{subnetAnalysis.used_ips || 0}</div>
            </div>
            <div class="bg-gray-800 rounded-lg p-4">
              <div class="text-sm text-gray-400 mb-1">Available IPs</div>
              <div class="text-2xl font-bold text-green-400">{subnetAnalysis.available_ips || 0}</div>
            </div>
            <div class="bg-gray-800 rounded-lg p-4">
              <div class="text-sm text-gray-400 mb-1">Usage</div>
              <div class="text-2xl font-bold text-gray-100">{subnetAnalysis.usage_percentage?.toFixed(1) || 0}%</div>
            </div>
          </div>

          <!-- Used IPs -->
          {#if subnetAnalysis.used_ip_list && subnetAnalysis.used_ip_list.length > 0}
            <div>
              <h3 class="text-lg font-semibold text-gray-100 mb-3">Used IP Addresses ({subnetAnalysis.used_ip_list.length})</h3>
              <div class="bg-gray-800 rounded-lg p-4 max-h-60 overflow-y-auto">
                <div class="grid grid-cols-4 md:grid-cols-6 gap-2">
                  {#each subnetAnalysis.used_ip_list as ip}
                    <div class="text-xs font-mono text-gray-300 bg-gray-700 px-2 py-1 rounded">
                      {ip}
                    </div>
                  {/each}
                </div>
              </div>
            </div>
          {/if}

          <!-- Available IPs (first 100) -->
          {#if subnetAnalysis.available_ip_list && subnetAnalysis.available_ip_list.length > 0}
            <div>
              <h3 class="text-lg font-semibold text-gray-100 mb-3">
                Available IP Addresses
                {#if subnetAnalysis.available_ips > 100}
                  <span class="text-sm text-gray-400">(showing first 100 of {subnetAnalysis.available_ips})</span>
                {:else}
                  <span class="text-sm text-gray-400">({subnetAnalysis.available_ips})</span>
                {/if}
              </h3>
              <div class="bg-gray-800 rounded-lg p-4 max-h-60 overflow-y-auto">
                <div class="grid grid-cols-4 md:grid-cols-6 gap-2">
                  {#each subnetAnalysis.available_ip_list as ip}
                    <div class="text-xs font-mono text-green-300 bg-gray-700 px-2 py-1 rounded">
                      {ip}
                    </div>
                  {/each}
                </div>
              </div>
            </div>
          {/if}

          <!-- Flows by IP -->
          {#if subnetAnalysis.flows_by_ip && Object.keys(subnetAnalysis.flows_by_ip).length > 0}
            <div>
              <h3 class="text-lg font-semibold text-gray-100 mb-3">Flows by IP Address</h3>
              <div class="bg-gray-800 rounded-lg overflow-hidden">
                <table class="min-w-full text-xs">
                  <thead class="bg-gray-700">
                    <tr>
                      <th class="text-left px-4 py-2 font-medium text-gray-200">IP Address</th>
                      <th class="text-left px-4 py-2 font-medium text-gray-200">Flow Count</th>
                      <th class="text-left px-4 py-2 font-medium text-gray-200">Flows</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-700">
                    {#each Object.entries(subnetAnalysis.flows_by_ip) as [ip, flows]}
                      <tr class="hover:bg-gray-700/50">
                        <td class="px-4 py-2 font-mono text-gray-300">{ip}</td>
                        <td class="px-4 py-2">
                          <span class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium bg-blue-900 text-blue-200">
                            {flows.length}
                          </span>
                        </td>
                        <td class="px-4 py-2 text-gray-400">
                          {flows.map(f => f.display_name || f.id).join(", ")}
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</section>
