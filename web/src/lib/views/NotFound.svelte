<script>
  import { route } from '../router.svelte.js'
</script>

<div class="hero py-16">
  <div class="hero-content text-center">
    <div>
      <h1 class="text-3xl font-bold">Not found</h1>
      <p class="opacity-70 mt-2">
        No view for <code class="font-mono">{route.path}</code>.
      </p>
      <a class="btn btn-primary btn-sm mt-4" href="#/">Back to dashboard</a>
    </div>
  </div>
</div>
