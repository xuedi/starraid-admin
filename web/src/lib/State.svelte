<script>
  // Shared loading / error / empty wrapper. Renders the children snippet only in
  // the loaded, non-empty, non-error case so every view handles all four states
  // consistently.
  let {
    loading = false,
    error = null,
    empty = false,
    emptyText = 'Nothing here yet.',
    children,
  } = $props()
</script>

{#if loading}
  <div class="flex items-center gap-3 p-6 opacity-70">
    <span class="loading loading-spinner loading-sm"></span> Loading…
  </div>
{:else if error}
  <div role="alert" class="alert alert-error my-2">
    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
    </svg>
    <span>{error}</span>
  </div>
{:else if empty}
  <div class="p-6 opacity-60">{emptyText}</div>
{:else}
  {@render children?.()}
{/if}
