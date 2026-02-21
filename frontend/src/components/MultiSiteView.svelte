<script>
  let {
    token = "",
    sitesRoomsSummary = null,
  } = $props();

  let selectedSite = $state("");
  let viewMode = $state("single"); // "single" | "cross-site" | "global"
  let crossSiteRoutings = $state([]);
  let crossSiteLoading = $state(false);
  let crossSiteError = $state("");

  // Filtered data for single site view
  let filteredNodes = $state([]);
  let filteredDevices = $state([]);
  let filteredFlows = $state([]);
  let filteredLoading = $state(false);

  // Available sites from summary
  let availableSites = $derived(
    sitesRoomsSummary?.sites?.map((s) => s.site) || []
  );

  async function loadCrossSiteRoutings() {
    crossSiteLoading = true;
    crossSiteError = "";
    try {
      const { api } = await import("../lib/api.js");
      const result = await api("/flows/cross-site", { token });
      crossSiteRoutings = result.routings || [];
    } catch (e) {
      crossSiteError = e.message || "Failed to load cross-site routings";
      crossSiteRoutings = [];
    } finally {
      crossSiteLoading = false;
    }
  }

  async function loadFilteredData() {
    if (!selectedSite) {
      filteredNodes = [];
      filteredDevices = [];
      filteredFlows = [];
      return;
    }
    filteredLoading = true;
    try {
      const { api } = await import("../lib/api.js");
      const [nodes, devices, flows] = await Promise.all([
        api(`/nmos/registry/nodes?site=${encodeURIComponent(selectedSite)}`, { token }),
        api(`/nmos/registry/devices?site=${encodeURIComponent(selectedSite)}`, { token }),
        api(`/nmos/registry/flows?site=${encodeURIComponent(selectedSite)}`, { token }),
      ]);
      filteredNodes = Array.isArray(nodes) ? nodes : [];
      filteredDevices = Array.isArray(devices) ? devices : [];
      filteredFlows = Array.isArray(flows) ? flows : [];
    } catch (e) {
      console.error("Failed to load filtered data:", e);
      filteredNodes = [];
      filteredDevices = [];
      filteredFlows = [];
    } finally {
      filteredLoading = false;
    }
  }

  $effect(() => {
    if (viewMode === "cross-site") {
      loadCrossSiteRoutings();
    } else if (viewMode === "single" && selectedSite) {
      loadFilteredData();
    }
  });

  $effect(() => {
    if (viewMode === "single" && selectedSite) {
      loadFilteredData();
    }
  });
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-gray-100 mb-2">Multi-Site Views (D.3)</h2>
    <p class="text-gray-400 text-sm">View and filter resources by site, or analyze cross-site routing</p>
  </div>

  <!-- View Mode Selector -->
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
    <div class="flex items-center gap-3">
      <span class="text-sm font-medium text-gray-300">View Mode:</span>
      <button
        onclick={() => {
          viewMode = "single";
          selectedSite = "";
        }}
        class="px-4 py-2 rounded-md text-sm font-medium transition-colors {viewMode === 'single'
          ? 'bg-orange-600 text-white'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700'}"
      >
        Single Site
      </button>
      <button
        onclick={() => {
          viewMode = "cross-site";
          loadCrossSiteRoutings();
        }}
        class="px-4 py-2 rounded-md text-sm font-medium transition-colors {viewMode === 'cross-site'
          ? 'bg-orange-600 text-white'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700'}"
      >
        Cross-Site Routing
      </button>
      <button
        onclick={() => {
          viewMode = "global";
          selectedSite = "";
        }}
        class="px-4 py-2 rounded-md text-sm font-medium transition-colors {viewMode === 'global'
          ? 'bg-orange-600 text-white'
          : 'bg-gray-800 text-gray-300 hover:bg-gray-700'}"
      >
        Global Overview
      </button>
    </div>
  </div>

  {#if viewMode === "single"}
    <!-- Single Site View -->
    <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
      <h3 class="text-lg font-semibold text-gray-100 mb-4">Single Site Filter</h3>
      <div class="space-y-4">
        <div>
          <label for="site-select" class="block text-sm font-medium text-gray-300 mb-2">
            Select Site
          </label>
          <select
            id="site-select"
            bind:value={selectedSite}
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500"
          >
            <option value="">— Select a site —</option>
            {#each availableSites as site}
              <option value={site}>{site}</option>
            {/each}
          </select>
        </div>

        {#if selectedSite}
          {#if filteredLoading}
            <div class="text-center py-8 text-gray-400">Loading...</div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div class="bg-gray-800 rounded-lg p-4">
                <p class="text-sm text-gray-400 mb-1">Nodes</p>
                <p class="text-2xl font-bold text-gray-100">{filteredNodes.length}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-4">
                <p class="text-sm text-gray-400 mb-1">Devices</p>
                <p class="text-2xl font-bold text-gray-100">{filteredDevices.length}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-4">
                <p class="text-sm text-gray-400 mb-1">Flows</p>
                <p class="text-2xl font-bold text-gray-100">{filteredFlows.length}</p>
              </div>
            </div>

            {#if filteredNodes.length > 0 || filteredDevices.length > 0 || filteredFlows.length > 0}
              <div class="mt-4 space-y-4">
                {#if filteredNodes.length > 0}
                  <div>
                    <h4 class="text-sm font-semibold text-gray-200 mb-2">Nodes ({filteredNodes.length})</h4>
                    <div class="space-y-1">
                      {#each filteredNodes.slice(0, 10) as node}
                        <div class="px-3 py-2 bg-gray-800 rounded text-sm text-gray-300">
                          {node.label || node.id}
                        </div>
                      {/each}
                      {#if filteredNodes.length > 10}
                        <p class="text-xs text-gray-500">... and {filteredNodes.length - 10} more</p>
                      {/if}
                    </div>
                  </div>
                {/if}

                {#if filteredDevices.length > 0}
                  <div>
                    <h4 class="text-sm font-semibold text-gray-200 mb-2">Devices ({filteredDevices.length})</h4>
                    <div class="space-y-1">
                      {#each filteredDevices.slice(0, 10) as device}
                        <div class="px-3 py-2 bg-gray-800 rounded text-sm text-gray-300">
                          {device.label || device.id}
                        </div>
                      {/each}
                      {#if filteredDevices.length > 10}
                        <p class="text-xs text-gray-500">... and {filteredDevices.length - 10} more</p>
                      {/if}
                    </div>
                  </div>
                {/if}

                {#if filteredFlows.length > 0}
                  <div>
                    <h4 class="text-sm font-semibold text-gray-200 mb-2">Flows ({filteredFlows.length})</h4>
                    <div class="space-y-1">
                      {#each filteredFlows.slice(0, 10) as flow}
                        <div class="px-3 py-2 bg-gray-800 rounded text-sm text-gray-300">
                          {flow.label || flow.id}
                        </div>
                      {/each}
                      {#if filteredFlows.length > 10}
                        <p class="text-xs text-gray-500">... and {filteredFlows.length - 10} more</p>
                      {/if}
                    </div>
                  </div>
                {/if}
              </div>
            {:else}
              <div class="text-center py-8 text-gray-500">
                No resources found for site "{selectedSite}"
              </div>
            {/if}
          {/if}
        {:else}
          <div class="text-center py-8 text-gray-500">
            Select a site to view its resources
          </div>
        {/if}
      </div>
    </div>
  {/if}

  {#if viewMode === "cross-site"}
    <!-- Cross-Site Routing View -->
    <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-100">Cross-Site Routing</h3>
        <button
          onclick={loadCrossSiteRoutings}
          disabled={crossSiteLoading}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
        >
          {crossSiteLoading ? "Loading..." : "Refresh"}
        </button>
      </div>

      {#if crossSiteError}
        <div class="mb-4 p-3 rounded-md bg-red-900/30 border border-red-700 text-red-200 text-sm">
          Error: {crossSiteError}
        </div>
      {/if}

      {#if crossSiteLoading}
        <div class="text-center py-8 text-gray-400">Loading cross-site routings...</div>
      {:else if crossSiteRoutings.length === 0}
        <div class="text-center py-8 text-gray-500">
          No cross-site routings found. All active connections are within the same site.
        </div>
      {:else}
        <div class="space-y-3">
          <div class="text-sm text-gray-400 mb-4">
            Found <span class="font-semibold text-orange-400">{crossSiteRoutings.length}</span> cross-site routing(s)
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-700">
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Flow</th>
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Source Site</th>
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Source Device</th>
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Target Site</th>
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Target Device</th>
                  <th class="px-3 py-2 text-left text-gray-300 font-medium">Sender → Receiver</th>
                </tr>
              </thead>
              <tbody>
                {#each crossSiteRoutings as routing}
                  <tr class="border-b border-gray-800 hover:bg-gray-800/50">
                    <td class="px-3 py-2 text-gray-200">{routing.flow_label || routing.flow_id}</td>
                    <td class="px-3 py-2">
                      <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-indigo-900/60 text-indigo-200 border border-indigo-700/50">
                        {routing.source_site}
                      </span>
                    </td>
                    <td class="px-3 py-2 text-gray-300">{routing.source_device}</td>
                    <td class="px-3 py-2">
                      <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-purple-900/60 text-purple-200 border border-purple-700/50">
                        {routing.target_site}
                      </span>
                    </td>
                    <td class="px-3 py-2 text-gray-300">{routing.target_device}</td>
                    <td class="px-3 py-2 text-gray-400 text-xs">
                      {routing.sender_label || routing.sender_id} → {routing.receiver_label || routing.receiver_id}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {/if}
    </div>
  {/if}

  {#if viewMode === "global"}
    <!-- Global Overview -->
    <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
      <h3 class="text-lg font-semibold text-gray-100 mb-4">Global Overview</h3>
      <p class="text-gray-400 text-sm mb-4">
        View all sites and their resource counts. This is the default view showing all resources across all sites.
      </p>
      {#if sitesRoomsSummary && sitesRoomsSummary.sites && sitesRoomsSummary.sites.length > 0}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {#each sitesRoomsSummary.sites as siteData}
            <div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
              <div class="flex items-center justify-between mb-3">
                <h4 class="text-sm font-semibold text-gray-200">{siteData.site}</h4>
                <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-indigo-900/60 text-indigo-200 border border-indigo-700/50">
                  Site
                </span>
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span class="text-gray-400">Nodes:</span>
                  <span class="ml-1 font-semibold text-gray-200">{siteData.nodes || 0}</span>
                </div>
                <div>
                  <span class="text-gray-400">Devices:</span>
                  <span class="ml-1 font-semibold text-gray-200">{siteData.devices || 0}</span>
                </div>
                <div>
                  <span class="text-gray-400">Flows:</span>
                  <span class="ml-1 font-semibold text-gray-200">{siteData.flows || 0}</span>
                </div>
                <div>
                  <span class="text-gray-400">Senders:</span>
                  <span class="ml-1 font-semibold text-gray-200">{siteData.senders || 0}</span>
                </div>
                <div class="col-span-2">
                  <span class="text-gray-400">Receivers:</span>
                  <span class="ml-1 font-semibold text-gray-200">{siteData.receivers || 0}</span>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {:else}
        <div class="text-center py-8 text-gray-500">
          No sites found. Sites are determined by the "site" tag on nodes, devices, and flows.
        </div>
      {/if}
    </div>
  {/if}
</div>
