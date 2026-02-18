<script>
  export let title = "No data available";
  export let message = "There's nothing to display here yet.";
  export let actionLabel = "";
  export let onAction = null;
  export let icon = null;
</script>

<div class="flex flex-col items-center justify-center py-12 px-4 text-center animate-fade-in">
  {#if icon}
    <div class="mb-4 text-gray-500 opacity-60">
      {@html icon}
    </div>
  {:else}
    <svg class="w-16 h-16 text-gray-600 mb-4 opacity-60" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
    </svg>
  {/if}
  <h3 class="text-lg font-semibold text-gray-200 mb-2">{title}</h3>
  <p class="text-sm text-gray-400 mb-4 max-w-md">{message}</p>
  {#if actionLabel && onAction}
    <button
      on:click={onAction}
      class="px-4 py-2 rounded-md bg-orange-600 hover:bg-orange-500 text-white text-sm font-medium transition-all duration-150 hover:scale-105 active:scale-95"
    >
      {actionLabel}
    </button>
  {/if}
</div>

<style>
  @keyframes fade-in {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .animate-fade-in {
    animation: fade-in 0.3s ease-out;
  }
</style>
