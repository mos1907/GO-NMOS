<script>
  import { onMount } from 'svelte';
  import EmptyState from "./EmptyState.svelte";
  import { IconDownload, IconUpload, IconPlus } from "../lib/icons.js";
  import { api } from "../lib/api.js";

  let {
    plannerRoots = [],
    plannerChildren = [],
    selectedPlannerRoot = null,
    newPlannerParent,
    newPlannerChild,
    canEdit = false,
    isAdmin = false,
    token = "",
    onSelectPlannerRoot,
    onPlannerQuickEdit,
    onPlannerDelete,
    onCreatePlannerParent,
    onCreatePlannerChild,
    onExportBuckets,
    onImportBucketsFromFile,
  } = $props();

  // Ensure arrays are never null (reactive)
  let safePlannerRoots = $derived(plannerRoots || []);
  let safePlannerChildren = $derived(plannerChildren || []);

  // Usage statistics for buckets
  let bucketStats = $state(new Map());

  async function loadBucketStats(bucketId) {
    if (!bucketId || !token) return;
    try {
      const stats = await api(`/address/buckets/${bucketId}/usage`, { token });
      bucketStats.set(bucketId, stats);
    } catch (e) {
      console.error('Failed to load bucket stats:', e);
    }
  }

  // Load stats for all children when selected root changes
  $effect(() => {
    if (selectedPlannerRoot && safePlannerChildren.length > 0) {
      safePlannerChildren.forEach(child => {
        if (child.id) {
          loadBucketStats(child.id);
        }
      });
    }
  });

  function getUsageColor(percentage) {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    if (percentage >= 50) return 'bg-blue-500';
    return 'bg-green-500';
  }
</script>

<section class="mt-4 space-y-4">
  <div>
    <h3 class="text-sm font-semibold text-gray-100">Planner Buckets</h3>
    <p class="text-[11px] text-gray-400">Organize multicast addresses into drives, folders, and views</p>
  </div>

  <div class="grid md:grid-cols-[1fr_2fr] gap-4">
    <!-- Drives Panel -->
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
      <h4 class="text-sm font-semibold text-gray-100 mb-3">Drives</h4>
      <div class="space-y-2">
        {#if safePlannerRoots.length === 0}
          <div class="text-center py-8 text-gray-500 text-sm">
            No drives available
          </div>
        {:else}
          {#each safePlannerRoots as root}
            <button
              onclick={() => onSelectPlannerRoot?.(root)}
              class="w-full text-left px-3 py-2 rounded-md border transition-all duration-150 {selectedPlannerRoot && selectedPlannerRoot.id === root.id
                ? 'bg-orange-600/20 border-orange-600 text-white'
                : 'bg-gray-800 border-gray-700 text-gray-300 hover:bg-gray-700 hover:border-gray-600'}"
            >
              <div class="flex items-center justify-between">
                <span class="font-medium text-sm">{root.name}</span>
                <span class="text-xs text-gray-400">{root.cidr}</span>
              </div>
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Folders / Views Panel -->
    <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
      <div class="flex items-center justify-between mb-4">
        <h4 class="text-sm font-semibold text-gray-100">Folders / Views</h4>
        {#if selectedPlannerRoot}
          <span class="text-xs text-gray-400">
            Drive: <span class="font-medium text-gray-300">{selectedPlannerRoot.name}</span>
          </span>
        {/if}
      </div>

      {#if !selectedPlannerRoot}
        <div class="text-center py-12">
          <p class="text-gray-400 text-sm">Select a drive to view its folders and views</p>
        </div>
      {:else if safePlannerChildren.length === 0}
        <div class="text-center py-12">
          <p class="text-gray-400 text-sm">No folders or views in this drive</p>
        </div>
      {:else}
        <div class="rounded-xl border border-gray-800 bg-gray-900 overflow-hidden">
          <table class="min-w-full text-xs">
            <thead class="bg-gray-800">
              <tr>
                <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Name</th>
                <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Type</th>
                <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">CIDR</th>
                <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Usage</th>
                <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Description</th>
                {#if canEdit}
                  <th class="text-left border-b border-gray-800 px-3 py-2 font-medium text-gray-200">Action</th>
                {/if}
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-800">
              {#each safePlannerChildren as item}
                {@const stats = bucketStats.get(item.id)}
                <tr class="hover:bg-gray-800/70 transition-colors">
                  <td class="px-3 py-2 text-gray-100 font-medium">{item.name}</td>
                  <td class="px-3 py-2">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium {item.bucket_type === 'parent'
                        ? 'bg-blue-900 text-blue-200 border border-blue-700'
                        : 'bg-purple-900 text-purple-200 border border-purple-700'}"
                    >
                      {item.bucket_type}
                    </span>
                  </td>
                  <td class="px-3 py-2 text-gray-300 text-[11px]">{item.cidr}</td>
                  <td class="px-3 py-2">
                    {#if stats}
                      <div class="space-y-1">
                        <div class="flex items-center gap-2">
                          <div class="flex-1 bg-gray-700 rounded-full h-2 overflow-hidden">
                            <div
                              class={`h-full ${getUsageColor(stats.usage_percentage)}`}
                              style="width: {Math.min(stats.usage_percentage, 100)}%"
                            ></div>
                          </div>
                          <span class="text-[10px] text-gray-400 min-w-[45px] text-right">
                            {stats.usage_percentage.toFixed(1)}%
                          </span>
                        </div>
                        <div class="text-[9px] text-gray-500">
                          {stats.used_ips || 0} / {stats.total_ips || 0} IPs
                          {#if stats.used_flow_count > 0}
                            <span class="ml-1">({stats.used_flow_count} flows)</span>
                          {/if}
                        </div>
                      </div>
                    {:else}
                      <span class="text-[10px] text-gray-500">-</span>
                    {/if}
                  </td>
                  <td class="px-3 py-2 text-gray-400 text-[11px]">{item.description || "-"}</td>
                  {#if canEdit}
                    <td class="px-3 py-2">
                      <div class="flex gap-1.5">
                        <button
                          onclick={() => onPlannerQuickEdit?.(item)}
                          class="px-2.5 py-1 rounded-md border border-gray-700 bg-gray-800 text-[11px] text-gray-200 hover:bg-gray-700 transition-colors"
                        >
                          Edit
                        </button>
                        {#if isAdmin}
                          <button
                            onclick={() => onPlannerDelete?.(item)}
                            class="px-2.5 py-1 rounded-md border border-red-800 bg-red-900/60 text-[11px] text-red-200 hover:bg-red-900 transition-colors"
                          >
                            Delete
                          </button>
                        {/if}
                      </div>
                    </td>
                  {/if}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </div>

  {#if canEdit}
    <div class="grid md:grid-cols-2 gap-4">
      <!-- Create Folder -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-sm font-semibold text-gray-100 mb-4">Create Folder (Parent)</h4>
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Name</label>
            <input
              bind:value={newPlannerParent.name}
              placeholder="Folder name"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">CIDR</label>
            <input
              bind:value={newPlannerParent.cidr}
              placeholder="239.1.0.0/16"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Description</label>
            <input
              bind:value={newPlannerParent.description}
              placeholder="Description"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Color (optional)</label>
            <input
              bind:value={newPlannerParent.color}
              placeholder="#FF5733"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <button
            onclick={onCreatePlannerParent}
            class="w-full px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
          >
            Create Parent
          </button>
        </div>
      </div>

      <!-- Create View -->
      <div class="rounded-xl border border-gray-800 bg-gray-900 shadow-sm p-4">
        <h4 class="text-sm font-semibold text-gray-100 mb-4">Create View (Child)</h4>
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Name</label>
            <input
              bind:value={newPlannerChild.name}
              placeholder="View name"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">CIDR or Range</label>
            <input
              bind:value={newPlannerChild.cidr}
              placeholder="CIDR or range label"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Description</label>
            <input
              bind:value={newPlannerChild.description}
              placeholder="Description"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-300 mb-1">Color (optional)</label>
            <input
              bind:value={newPlannerChild.color}
              placeholder="#33FF57"
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>
          <button
            onclick={() => onCreatePlannerChild?.(selectedPlannerRoot)}
            disabled={!selectedPlannerRoot}
            class="w-full px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
          >
            Create Child
          </button>
        </div>
      </div>
    </div>
  {/if}

  <div class="flex flex-wrap gap-3 items-center">
    <button
      onclick={onExportBuckets}
      class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors flex items-center gap-2"
    >
      {@html IconDownload}
      Export Planner
    </button>
    {#if canEdit}
      <label class="inline-flex items-center gap-2 px-4 py-2 rounded-md border border-gray-700 bg-gray-800 hover:bg-gray-700 text-gray-200 text-sm font-medium cursor-pointer transition-colors">
        {@html IconUpload}
        <span>Import Planner</span>
        <input
          type="file"
          accept="application/json"
          onchange={onImportBucketsFromFile}
          class="hidden"
        />
      </label>
    {/if}
  </div>
</section>
