<script>
  import { notifications, dismissNotification } from "../lib/notifications.js";
</script>

<div
  class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 max-w-md w-full pointer-events-none"
  aria-live="polite"
>
  <div class="flex flex-col gap-2 pointer-events-auto">
    {#each $notifications as n (n.id)}
      {@const isSuccess = n.type === "success"}
      {@const isError = n.type === "error"}
      {@const isWarning = n.type === "warning"}
      <div
        class="flex items-start gap-3 rounded-lg border shadow-lg p-4 {isSuccess
          ? 'bg-green-950/95 border-green-700 text-green-100'
          : isError
            ? 'bg-red-950/95 border-red-700 text-red-100'
            : 'bg-amber-950/95 border-amber-700 text-amber-100'}"
        role="alert"
      >
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium">{n.message}</p>
        </div>
        <button
          type="button"
          class="flex-shrink-0 p-1 rounded hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-white/50"
          onclick={() => dismissNotification(n.id)}
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path
              fill-rule="evenodd"
              d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
              clip-rule="evenodd"
            />
          </svg>
        </button>
      </div>
    {/each}
  </div>
</div>
