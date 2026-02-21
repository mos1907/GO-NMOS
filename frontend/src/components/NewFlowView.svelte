<script>
  import { api } from "../lib/api.js";

  let {
    newFlow,
    onCreateFlow,
    onUpdateFlow = null,
    onClose,
    isOpen = false,
    editingFlow = null, // If set, we're editing an existing flow
    token = "",
  } = $props();

  let isSubmitting = $state(false);
  let collisionCheck = $state(null);
  let checkingCollision = $state(false);
  let buckets = $state([]);
  let loadingBuckets = $state(false);

  async function handleSubmit() {
    if (isSubmitting) return;
    // Warn if collision detected
    if (collisionCheck?.has_collision) {
      if (!confirm(`Warning: This IP:Port combination conflicts with ${collisionCheck.conflict_count} flow(s). Do you want to continue anyway?`)) {
        return;
      }
    }
    isSubmitting = true;
    try {
      if (editingFlow) {
        await onUpdateFlow?.(newFlow);
      } else {
        await onCreateFlow?.(newFlow);
      }
      onClose?.();
    } finally {
      isSubmitting = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === "Escape") {
      onClose?.();
    }
  }

  let isEditMode = $derived(!!editingFlow);
  let modalTitle = $derived(isEditMode ? "Edit Flow" : "Create New Flow");
  let submitButtonText = $derived(
    isSubmitting
      ? (isEditMode ? "Updating..." : "Creating...")
      : (isEditMode ? "Update Flow" : "Create Flow")
  );

  // Real-time collision checking
  let collisionCheckTimeout = null;
  async function checkCollision() {
    if (!newFlow.multicast_ip || !newFlow.port) {
      collisionCheck = null;
      return;
    }

    // Clear previous timeout
    if (collisionCheckTimeout) {
      clearTimeout(collisionCheckTimeout);
    }

    // Debounce: wait 500ms after user stops typing
    collisionCheckTimeout = setTimeout(async () => {
      checkingCollision = true;
      try {
        const excludeFlowID = editingFlow?.id || null;
        const url = `/checker/check?multicast_ip=${encodeURIComponent(newFlow.multicast_ip)}&port=${newFlow.port}${excludeFlowID ? `&exclude_flow_id=${excludeFlowID}` : ''}`;
        collisionCheck = await api(url, { token });
      } catch (e) {
        collisionCheck = null;
      } finally {
        checkingCollision = false;
      }
    }, 500);
  }

  // Watch for changes in multicast_ip and port
  $effect(() => {
    if (newFlow.multicast_ip && newFlow.port) {
      checkCollision();
    } else {
      collisionCheck = null;
    }
  });

  // Load buckets only when modal transitions from closed -> open
  let wasOpen = $state(false);

  $effect(() => {
    if (isOpen && !wasOpen && token) {
      loadBuckets();
    }
    wasOpen = isOpen;
  });

  async function loadBuckets() {
    if (loadingBuckets) return;
    loadingBuckets = true;
    try {
      buckets = await api("/address/buckets/all", { token });
    } catch (e) {
      console.error("Failed to load buckets:", e);
      buckets = [];
    } finally {
      loadingBuckets = false;
    }
  }

  function getBucketDisplayName(bucket) {
    if (bucket.parent_id) {
      const parent = buckets.find(b => b.id === bucket.parent_id);
      return parent ? `${parent.name} / ${bucket.name}` : bucket.name;
    }
    return bucket.name;
  }
</script>

{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-fade-in"
    onclick={onClose}
    onkeydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-labelledby="new-flow-title"
  >
    <!-- Modal -->
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl shadow-gray-900/40 w-full max-w-2xl max-h-[90vh] overflow-y-auto animate-slide-in"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 id="new-flow-title" class="text-xl font-bold text-gray-100">{modalTitle}</h2>
        <button
          onclick={onClose}
          class="p-1.5 rounded-md text-gray-400 hover:text-gray-200 hover:bg-gray-800 transition-colors"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <label for="display_name" class="block text-sm font-medium text-gray-300">
              Display Name *
            </label>
            <input
              id="display_name"
              type="text"
              bind:value={newFlow.display_name}
              placeholder="My Flow"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>

          <div class="space-y-2">
            <label for="multicast_ip" class="block text-sm font-medium text-gray-300">
              Multicast IP *
            </label>
            <div class="relative">
              <input
                id="multicast_ip"
                type="text"
                bind:value={newFlow.multicast_ip}
                placeholder="239.0.0.1"
                class="w-full px-4 py-2 bg-gray-900 border rounded-md text-gray-100 placeholder-gray-500 focus:outline-none transition-colors {collisionCheck?.has_collision ? 'border-red-500 focus:border-red-500' : 'border-gray-700 focus:border-orange-500'}"
              />
              {#if checkingCollision}
                <div class="absolute right-3 top-2.5">
                  <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400"></div>
                </div>
              {/if}
            </div>
            {#if collisionCheck?.has_collision}
              <div class="mt-1 p-3 bg-red-900/20 border border-red-700 rounded text-xs text-red-300">
                <p class="font-medium">⚠️ Collision detected!</p>
                <p class="mt-1">This IP:Port combination conflicts with {collisionCheck.conflict_count} flow(s):</p>
                <ul class="mt-1 list-disc list-inside mb-2">
                  {#each (collisionCheck.conflicting_flows || []) as flowName}
                    <li>{flowName}</li>
                  {/each}
                </ul>
                {#if collisionCheck.alternatives && collisionCheck.alternatives.length > 0}
                  <div class="mt-3 pt-3 border-t border-red-700">
                    <p class="font-medium mb-2">💡 Suggested alternatives:</p>
                    <div class="space-y-1.5">
                      {#each collisionCheck.alternatives as alt}
                        <button
                          onclick={() => {
                            newFlow.multicast_ip = alt.multicast_ip;
                            newFlow.port = alt.port;
                          }}
                          class="w-full text-left px-2 py-1.5 bg-gray-800 hover:bg-gray-700 border border-gray-600 rounded text-[11px] transition-colors flex items-center justify-between group"
                        >
                          <span class="text-gray-200">
                            <span class="font-mono">{alt.multicast_ip}</span>
                            <span class="text-gray-400 mx-1">:</span>
                            <span class="font-mono">{alt.port}</span>
                            {#if alt.reason === 'same_subnet_available'}
                              <span class="ml-2 text-gray-400">(same subnet)</span>
                            {:else if alt.reason === 'different_port'}
                              <span class="ml-2 text-gray-400">(different port)</span>
                            {:else if alt.reason === 'different_subnet'}
                              <span class="ml-2 text-gray-400">(different subnet)</span>
                            {/if}
                          </span>
                          <span class="text-blue-400 opacity-0 group-hover:opacity-100 transition-opacity text-[10px]">Apply →</span>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {:else if collisionCheck && !collisionCheck.has_collision && newFlow.multicast_ip && newFlow.port}
              <div class="mt-1 text-xs text-green-400">✓ No collision detected</div>
            {/if}
          </div>

          <div class="space-y-2">
            <label for="source_ip" class="block text-sm font-medium text-gray-300">
              Source IP *
            </label>
            <input
              id="source_ip"
              type="text"
              bind:value={newFlow.source_ip}
              placeholder="192.168.1.100"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>

          <div class="space-y-2">
            <label for="port" class="block text-sm font-medium text-gray-300">
              Port *
            </label>
            <input
              id="port"
              type="number"
              bind:value={newFlow.port}
              placeholder="5004"
              class="w-full px-4 py-2 bg-gray-900 border rounded-md text-gray-100 placeholder-gray-500 focus:outline-none transition-colors {collisionCheck?.has_collision ? 'border-red-500 focus:border-red-500' : 'border-gray-700 focus:border-orange-500'}"
            />
          </div>

          <div class="space-y-2">
            <label for="flow_status" class="block text-sm font-medium text-gray-300">
              Flow Status
            </label>
            <select
              id="flow_status"
              bind:value={newFlow.flow_status}
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
            >
              <option value="active">active</option>
              <option value="unused">unused</option>
              <option value="maintenance">maintenance</option>
            </select>
          </div>

          <div class="space-y-2">
            <label for="availability" class="block text-sm font-medium text-gray-300">
              Availability
            </label>
            <select
              id="availability"
              bind:value={newFlow.availability}
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
            >
              <option value="available">available</option>
              <option value="lost">lost</option>
              <option value="maintenance">maintenance</option>
            </select>
          </div>

          <div class="space-y-2">
            <label for="transport_protocol" class="block text-sm font-medium text-gray-300">
              Transport Protocol
            </label>
            <input
              id="transport_protocol"
              type="text"
              bind:value={newFlow.transport_protocol}
              placeholder="RTP"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
          </div>

          <div class="space-y-2">
            <label for="bucket_id" class="block text-sm font-medium text-gray-300">
              Planner Bucket (Optional)
            </label>
            <select
              id="bucket_id"
              bind:value={newFlow.bucket_id}
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 focus:outline-none focus:border-orange-500 transition-colors"
            >
              <option value={null}>None</option>
              {#each buckets as bucket}
                <option value={bucket.id}>{getBucketDisplayName(bucket)} {bucket.cidr ? `(${bucket.cidr})` : ''}</option>
              {/each}
            </select>
            <p class="text-xs text-gray-400 mt-1">Assign this flow to a planner bucket for organization</p>
          </div>

          <div class="space-y-2 md:col-span-2">
            <label for="note" class="block text-sm font-medium text-gray-300">
              Note
            </label>
            <textarea
              id="note"
              bind:value={newFlow.note}
              placeholder="Additional notes..."
              rows="3"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors resize-none"
            ></textarea>
          </div>

          <!-- Alias Fields -->
          <div class="md:col-span-2 border-t border-gray-800 pt-4">
            <h3 class="text-sm font-semibold text-gray-200 mb-3">Alias Fields (Optional)</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              {#each Array(8) as _, i}
                {@const aliasNum = i + 1}
                {@const aliasKey = `alias_${aliasNum}`}
                <div class="space-y-1">
                  <label for={aliasKey} class="block text-xs font-medium text-gray-400">
                    Alias {aliasNum}
                  </label>
                  <input
                    id={aliasKey}
                    type="text"
                    bind:value={newFlow[aliasKey]}
                    placeholder={`Alias ${aliasNum}...`}
                    class="w-full px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
                  />
                </div>
              {/each}
            </div>
          </div>

          <!-- User Fields -->
          <div class="md:col-span-2 border-t border-gray-800 pt-4">
            <h3 class="text-sm font-semibold text-gray-200 mb-3">User-Defined Fields (Optional)</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <!-- Audio-focused hooks for IS-08 -->
              <div class="space-y-1">
                <label for="user_field_1" class="block text-xs font-medium text-gray-400">
                  Audio Layout (IS-08 hook)
                </label>
                <input
                  id="user_field_1"
                  type="text"
                  bind:value={newFlow.user_field_1}
                  placeholder="e.g. 5.1, stereo, mono"
                  class="w-full px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
                />
              </div>
              <div class="space-y-1">
                <label for="user_field_2" class="block text-xs font-medium text-gray-400">
                  Audio Program Name
                </label>
                <input
                  id="user_field_2"
                  type="text"
                  bind:value={newFlow.user_field_2}
                  placeholder="e.g. Main Program, International, Clean Feed"
                  class="w-full px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
                />
              </div>

              <!-- Remaining generic user fields -->
              {#each Array(6) as _, i}
                {@const userNum = i + 3}
                {@const userKey = `user_field_${userNum}`}
                <div class="space-y-1">
                  <label for={userKey} class="block text-xs font-medium text-gray-400">
                    User Field {userNum}
                  </label>
                  <input
                    id={userKey}
                    type="text"
                    bind:value={newFlow[userKey]}
                    placeholder={`User field ${userNum}...`}
                    class="w-full px-3 py-1.5 bg-gray-900 border border-gray-700 rounded-md text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
                  />
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          onclick={onClose}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={handleSubmit}
          disabled={isSubmitting}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium transition-colors"
        >
          {submitButtonText}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(-20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  .animate-fade-in {
    animation: fade-in 0.2s ease-out;
  }
  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>

