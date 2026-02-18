<script>
  let {
    newFlow,
    onCreateFlow,
    onClose,
    isOpen = false,
  } = $props();

  function handleCreate() {
    onCreateFlow?.();
    onClose?.();
  }

  function handleKeydown(e) {
    if (e.key === "Escape") {
      onClose?.();
    }
  }
</script>

{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-fade-in"
    on:click={onClose}
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-labelledby="new-flow-title"
  >
    <!-- Modal -->
    <div
      class="bg-gray-900 border border-gray-800 rounded-xl shadow-2xl shadow-gray-900/40 w-full max-w-2xl max-h-[90vh] overflow-y-auto animate-slide-in"
      on:click|stopPropagation
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-800">
        <h2 id="new-flow-title" class="text-xl font-bold text-gray-100">Create New Flow</h2>
        <button
          on:click={onClose}
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
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-800">
        <button
          on:click={onClose}
          class="px-4 py-2 rounded-md border border-gray-700 bg-gray-800 text-gray-200 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Cancel
        </button>
        <button
          on:click={handleCreate}
          class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-colors"
        >
          Create Flow
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

