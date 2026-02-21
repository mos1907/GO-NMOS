<script>
  import { onMount } from 'svelte';
  import { api } from '../lib/api.js';

  let { token } = $props();

  let targets = $state([]);
  let selectedTarget = $state(null);
  let testResults = $state(null);
  let loading = $state(false);
  let error = $state(null);

  let customTarget = $state({
    name: '',
    type: 'node',
    base_url: '',
    vendor: '',
    model: '',
    description: ''
  });

  onMount(async () => {
    await loadTargets();
  });

  async function loadTargets() {
    try {
      const data = await api('/interop/targets', { token });
      targets = data.targets || [];
    } catch (err) {
      console.error('Failed to load targets:', err);
      error = 'Failed to load reference targets';
    }
  }

  async function runTest(target) {
    loading = true;
    error = null;
    testResults = null;

    try {
      const data = await api('/interop/test', {
        method: 'POST',
        body: { target },
        token
      });
      testResults = data;
      selectedTarget = target;
    } catch (err) {
      console.error('Test failed:', err);
      error = err.message || 'Test failed';
    } finally {
      loading = false;
    }
  }

  async function runCustomTest() {
    if (!customTarget.base_url) {
      error = 'Base URL is required';
      return;
    }
    await runTest(customTarget);
  }

  function getStatusBadgeClass(status) {
    switch (status) {
      case 'pass': return 'bg-green-100 text-green-800';
      case 'fail': return 'bg-red-100 text-red-800';
      case 'warning': return 'bg-yellow-100 text-yellow-800';
      case 'skip': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  }

  function formatDuration(ms) {
    if (ms < 1000) {
      return `${ms}ms`;
    }
    return `${(ms / 1000).toFixed(2)}s`;
  }
</script>

<div class="p-6">
  <h1 class="text-2xl font-bold mb-6">Interoperability Test Harness</h1>

  {#if error}
    <div class="mb-4 p-4 bg-red-50 border border-red-200 rounded">
      <p class="text-red-800">{error}</p>
    </div>
  {/if}

  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Test Targets -->
    <div class="bg-white rounded-lg shadow p-6">
      <h2 class="text-xl font-semibold mb-4">Test Targets</h2>

      <!-- Reference Targets -->
      <div class="mb-6">
        <h3 class="text-sm font-medium text-gray-700 mb-3">Reference Targets</h3>
        <div class="space-y-2">
          {#each targets as target}
            <div class="border rounded p-3 hover:bg-gray-50">
              <div class="flex justify-between items-start">
                <div class="flex-1">
                  <div class="font-medium">{target.name}</div>
                  <div class="text-sm text-gray-600">{target.type}</div>
                  {#if target.vendor}
                    <div class="text-xs text-gray-500">{target.vendor}</div>
                  {/if}
                  {#if target.description}
                    <div class="text-xs text-gray-500 mt-1">{target.description}</div>
                  {/if}
                </div>
                <button
                  onclick={() => runTest(target)}
                  disabled={loading}
                  class="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50"
                >
                  Test
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Custom Target -->
      <div class="border-t pt-6">
        <h3 class="text-sm font-medium text-gray-700 mb-3">Custom Target</h3>
        <div class="space-y-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              bind:value={customTarget.name}
              placeholder="Device/Registry Name"
              class="w-full px-3 py-2 border rounded"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select bind:value={customTarget.type} class="w-full px-3 py-2 border rounded">
              <option value="node">Node</option>
              <option value="registry">Registry</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Base URL *</label>
            <input
              type="text"
              bind:value={customTarget.base_url}
              placeholder="http://host:port"
              class="w-full px-3 py-2 border rounded"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Vendor</label>
            <input
              type="text"
              bind:value={customTarget.vendor}
              placeholder="Vendor Name"
              class="w-full px-3 py-2 border rounded"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Model</label>
            <input
              type="text"
              bind:value={customTarget.model}
              placeholder="Model Name"
              class="w-full px-3 py-2 border rounded"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              bind:value={customTarget.description}
              placeholder="Optional description"
              class="w-full px-3 py-2 border rounded"
              rows="2"
            ></textarea>
          </div>
          <button
            onclick={runCustomTest}
            disabled={loading || !customTarget.base_url}
            class="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            Run Test
          </button>
        </div>
      </div>
    </div>

    <!-- Test Results -->
    <div class="bg-white rounded-lg shadow p-6">
      <h2 class="text-xl font-semibold mb-4">Test Results</h2>

      {#if loading}
        <div class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-2 text-gray-600">Running tests...</p>
        </div>
      {:else if testResults}
        <!-- Summary -->
        <div class="mb-6 p-4 bg-gray-50 rounded">
          <h3 class="font-semibold mb-2">Test Summary</h3>
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-gray-600">Target:</span>
              <span class="ml-2 font-medium">{testResults.target.name || testResults.target.base_url}</span>
            </div>
            <div>
              <span class="text-gray-600">Type:</span>
              <span class="ml-2 font-medium">{testResults.target.type}</span>
            </div>
            <div>
              <span class="text-gray-600">Total Tests:</span>
              <span class="ml-2 font-medium">{testResults.summary.total}</span>
            </div>
            <div>
              <span class="text-gray-600">Passed:</span>
              <span class="ml-2 font-medium text-green-600">{testResults.summary.passed}</span>
            </div>
            <div>
              <span class="text-gray-600">Failed:</span>
              <span class="ml-2 font-medium text-red-600">{testResults.summary.failed}</span>
            </div>
            <div>
              <span class="text-gray-600">Warnings:</span>
              <span class="ml-2 font-medium text-yellow-600">{testResults.summary.warnings}</span>
            </div>
            <div>
              <span class="text-gray-600">Skipped:</span>
              <span class="ml-2 font-medium text-gray-600">{testResults.summary.skipped}</span>
            </div>
          </div>
        </div>

        <!-- Test Results List -->
        <div class="space-y-3 max-h-[600px] overflow-y-auto">
          {#each testResults.results as result}
            <div class="border rounded p-4">
              <div class="flex justify-between items-start mb-2">
                <div class="flex-1">
                  <div class="font-medium">{result.test_name}</div>
                  <div class="text-sm text-gray-600 mt-1">{result.message}</div>
                </div>
                <span class={`px-2 py-1 rounded text-xs font-medium ${getStatusBadgeClass(result.status)}`}>
                  {result.status}
                </span>
              </div>
              <div class="text-xs text-gray-500 mt-2">
                Duration: {formatDuration(result.duration / 1000000)}
                {#if result.details && Object.keys(result.details).length > 0}
                  <button
                    onclick={() => {
                      const detailsEl = document.getElementById(`details-${result.test_name}`);
                      detailsEl.classList.toggle('hidden');
                    }}
                    class="ml-4 text-blue-600 hover:underline"
                  >
                    Toggle Details
                  </button>
                {/if}
              </div>
              {#if result.details && Object.keys(result.details).length > 0}
                <div id="details-{result.test_name}" class="hidden mt-2 p-2 bg-gray-50 rounded text-xs">
                  <pre class="whitespace-pre-wrap">{JSON.stringify(result.details, null, 2)}</pre>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {:else}
        <div class="text-center py-8 text-gray-500">
          <p>No test results yet.</p>
          <p class="text-sm mt-2">Select a target and run tests to see results here.</p>
        </div>
      {/if}
    </div>
  </div>
</div>
