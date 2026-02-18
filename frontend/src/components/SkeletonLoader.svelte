<script>
  export let lines = 3;
  export let showHeader = true;
</script>

<div class="space-y-3 animate-pulse">
  {#if showHeader}
    <div class="flex items-center justify-between">
      <div class="h-4 bg-gray-800 rounded w-1/4"></div>
      <div class="h-4 bg-gray-800 rounded w-1/6"></div>
    </div>
  {/if}
  {#each Array(lines) as _, i}
    <div class="h-4 bg-gray-800 rounded" style="width: {Math.random() * 40 + 60}%"></div>
  {/each}
</div>
