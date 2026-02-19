<script>
  let {
    newFlow,
    onCreateFlow,
    onUpdateFlow = null,
    onClose,
    isOpen = false,
    editingFlow = null, // If set, we're editing an existing flow
  } = $props();

  let isSubmitting = $state(false);

  async function handleSubmit() {
    if (isSubmitting) return;
    isSubmitting = true;
    try {
      if (editingFlow) {
        await onUpdateFlow?.();
      } else {
        await onCreateFlow?.();
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
            <input
              id="multicast_ip"
              type="text"
              bind:value={newFlow.multicast_ip}
              placeholder="239.0.0.1"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
            />
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
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-md text-gray-100 placeholder-gray-500 focus:outline-none focus:border-orange-500 transition-colors"
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

