<script>
  import { onMount } from "svelte";
  import { api } from "../lib/api.js";

  let token = $state("");
  let host = $state("");
  let portsInput = $state("");
  let portRangeInput = $state("");
  let concurrency = $state(10);
  let timeout = $state(3);
  let scanning = $state(false);
  let results = $state([]);
  let error = $state("");
  let isLocal = $state(true);
  let showLocalWarning = $state(false);

  // Load saved settings from localStorage
  onMount(() => {
    const savedHost = localStorage.getItem("portExplorer_host");
    const savedPorts = localStorage.getItem("portExplorer_ports");
    const savedRange = localStorage.getItem("portExplorer_range");
    const savedConcurrency = localStorage.getItem("portExplorer_concurrency");
    const savedTimeout = localStorage.getItem("portExplorer_timeout");

    if (savedHost) host = savedHost;
    if (savedPorts) portsInput = savedPorts;
    if (savedRange) portRangeInput = savedRange;
    if (savedConcurrency) concurrency = parseInt(savedConcurrency, 10);
    if (savedTimeout) timeout = parseInt(savedTimeout, 10);

    token = localStorage.getItem("token") || "";
  });

  function checkLocalHost() {
    const hostLower = host.toLowerCase().trim();
    const localPatterns = [
      "localhost",
      "127.0.0.1",
      "::1",
      "0.0.0.0",
      "192.168.",
      "10.",
      "172.16.",
      "172.17.",
      "172.18.",
      "172.19.",
      "172.20.",
      "172.21.",
      "172.22.",
      "172.23.",
      "172.24.",
      "172.25.",
      "172.26.",
      "172.27.",
      "172.28.",
      "172.29.",
      "172.30.",
      "172.31.",
    ];
    isLocal = localPatterns.some((pattern) => hostLower.startsWith(pattern));
    if (!isLocal && host) {
      showLocalWarning = true;
    } else {
      showLocalWarning = false;
    }
  }

  function parsePorts(input) {
    if (!input.trim()) return [];
    return input
      .split(",")
      .map((p) => parseInt(p.trim(), 10))
      .filter((p) => !isNaN(p) && p > 0 && p <= 65535);
  }

  async function startScan() {
    if (!host.trim()) {
      error = "Host is required";
      return;
    }

    const ports = parsePorts(portsInput);
    if (ports.length === 0 && !portRangeInput.trim()) {
      error = "Ports or port range is required";
      return;
    }

    if (concurrency < 1 || concurrency > 50) {
      error = "Concurrency must be between 1 and 50";
      return;
    }

    if (timeout < 1 || timeout > 30) {
      error = "Timeout must be between 1 and 30 seconds";
      return;
    }

    // Save settings to localStorage
    localStorage.setItem("portExplorer_host", host);
    localStorage.setItem("portExplorer_ports", portsInput);
    localStorage.setItem("portExplorer_range", portRangeInput);
    localStorage.setItem("portExplorer_concurrency", concurrency.toString());
    localStorage.setItem("portExplorer_timeout", timeout.toString());

    scanning = true;
    error = "";
    results = [];

    try {
      const response = await api("/nmos/explore-ports", {
        method: "POST",
        token,
        body: {
          host: host.trim(),
          ports: ports,
          port_range: portRangeInput.trim() || undefined,
          concurrency: concurrency,
          timeout: timeout,
        },
      });

      isLocal = response.is_local || false;
      results = response.results || [];

      // Auto-add discovered ports to port list
      const discoveredPorts = results
        .filter((r) => r.is_nmos)
        .map((r) => r.port);
      if (discoveredPorts.length > 0) {
        const existingPorts = parsePorts(portsInput);
        const newPorts = [...new Set([...existingPorts, ...discoveredPorts])].sort((a, b) => a - b);
        portsInput = newPorts.join(", ");
        localStorage.setItem("portExplorer_ports", portsInput);
      }
    } catch (e) {
      error = e.message || "Port scan failed";
    } finally {
      scanning = false;
    }
  }

  function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
      // Show brief success feedback
      const btn = event.target;
      const original = btn.textContent;
      btn.textContent = "Copied!";
      btn.classList.add("bg-green-600");
      setTimeout(() => {
        btn.textContent = original;
        btn.classList.remove("bg-green-600");
      }, 1000);
    });
  }

  function getProbabilityColor(prob) {
    if (prob >= 90) return "text-green-400";
    if (prob >= 70) return "text-yellow-400";
    if (prob >= 50) return "text-orange-400";
    return "text-gray-400";
  }

  const foundNMOS = results.filter((r) => r.is_nmos).length;
  const sortedResults = [...results].sort((a, b) => {
    if (a.is_nmos !== b.is_nmos) return b.is_nmos ? 1 : -1;
    return b.probability - a.probability;
  });
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-gray-100 mb-2">Port Explorer</h2>
    <p class="text-gray-400 text-sm">
      Scan ports on a host to discover NMOS IS-04/IS-05 endpoints. Use only on systems you own or are authorized to test.
    </p>
  </div>

  {#if showLocalWarning}
    <div class="bg-amber-950/50 border border-amber-700 rounded-lg p-4">
      <div class="flex items-start gap-3">
        <svg class="w-5 h-5 text-amber-400 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
        </svg>
        <div class="flex-1">
          <h3 class="text-amber-300 font-semibold mb-1">Non-Local Network Target</h3>
          <p class="text-amber-200/80 text-sm">
            You are scanning a non-local network address. Make sure you have authorization to scan this host.
          </p>
          <button
            on:click={() => (showLocalWarning = false)}
            class="mt-2 text-amber-300 hover:text-amber-200 text-sm underline"
          >
            I understand, continue
          </button>
        </div>
      </div>
    </div>
  {/if}

  <div class="bg-gray-900 border border-gray-800 rounded-lg p-6 space-y-4">
    <div class="grid md:grid-cols-2 gap-4">
      <div>
        <label for="host" class="block text-sm font-medium text-gray-300 mb-2">Host (IP or hostname)</label>
        <input
          id="host"
          type="text"
          bind:value={host}
          on:input={checkLocalHost}
          placeholder="192.168.1.100"
          class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
        />
      </div>

      <div>
        <label for="portRange" class="block text-sm font-medium text-gray-300 mb-2">Port Range (e.g., 8080-8090)</label>
        <input
          id="portRange"
          type="text"
          bind:value={portRangeInput}
          placeholder="8080-8090"
          class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
        />
      </div>
    </div>

    <div>
      <label for="ports" class="block text-sm font-medium text-gray-300 mb-2">Port List (comma-separated)</label>
      <input
        id="ports"
        type="text"
        bind:value={portsInput}
        placeholder="8080, 8081, 8082"
        class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500"
      />
    </div>

    <div class="grid md:grid-cols-2 gap-4">
      <div>
        <label for="concurrency" class="block text-sm font-medium text-gray-300 mb-2">
          Concurrency: {concurrency}
        </label>
        <input
          id="concurrency"
          type="range"
          min="1"
          max="50"
          bind:value={concurrency}
          class="w-full"
        />
        <p class="text-xs text-gray-500 mt-1">Max concurrent port scans (1-50)</p>
      </div>

      <div>
        <label for="timeout" class="block text-sm font-medium text-gray-300 mb-2">
          Timeout: {timeout}s
        </label>
        <input
          id="timeout"
          type="range"
          min="1"
          max="30"
          bind:value={timeout}
          class="w-full"
        />
        <p class="text-xs text-gray-500 mt-1">Timeout per port (1-30 seconds)</p>
      </div>
    </div>

    {#if error}
      <div class="bg-red-950/50 border border-red-700 rounded-lg p-3 text-red-300 text-sm">{error}</div>
    {/if}

    <button
      on:click={startScan}
      disabled={scanning}
      class="w-full px-4 py-2 bg-orange-600 hover:bg-orange-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white font-medium rounded-md transition-colors"
    >
      {#if scanning}
        Scanning ports...
      {:else}
        Start Port Scan
      {/if}
    </button>
  </div>

  {#if results.length > 0}
    <div class="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
      <div class="bg-gray-800 px-6 py-4 border-b border-gray-700">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-100">
            Scan Results ({foundNMOS} NMOS found)
          </h3>
          <button
            on:click={() => copyToClipboard(JSON.stringify(results, null, 2))}
            class="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-md transition-colors"
          >
            Copy All
          </button>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-800 border-b border-gray-700">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Port</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Type</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Version</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">URL</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-800">
            {#each sortedResults as result}
              <tr class="hover:bg-gray-800/50">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-200">{result.port}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                  {#if result.is_nmos}
                    <span class="px-2 py-1 text-xs font-semibold rounded bg-green-900/50 text-green-300 border border-green-700">
                      NMOS
                    </span>
                  {:else}
                    <span class="px-2 py-1 text-xs font-semibold rounded bg-gray-800 text-gray-400 border border-gray-700">
                      No
                    </span>
                  {/if}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                  {#if result.is_is04 && result.is_is05}
                    IS-04 + IS-05
                  {:else if result.is_is04}
                    IS-04
                  {:else if result.is_is05}
                    IS-05
                  {:else}
                    -
                  {/if}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                  {result.is04_version || result.is05_version || "-"}
                </td>
                <td class="px-6 py-4 text-sm text-gray-400">
                  {#if result.base_url}
                    <code class="text-xs">{result.base_url}</code>
                  {:else}
                    -
                  {/if}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  {#if result.base_url}
                    <button
                      on:click={() => copyToClipboard(result.base_url)}
                      class="px-2 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded text-xs transition-colors"
                    >
                      Copy URL
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<style>
  input[type="range"] {
    -webkit-appearance: none;
    appearance: none;
    height: 6px;
    background: #374151;
    border-radius: 3px;
    outline: none;
  }

  input[type="range"]::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    background: #ea580c;
    border-radius: 50%;
    cursor: pointer;
  }

  input[type="range"]::-moz-range-thumb {
    width: 18px;
    height: 18px;
    background: #ea580c;
    border-radius: 50%;
    cursor: pointer;
    border: none;
  }
</style>
