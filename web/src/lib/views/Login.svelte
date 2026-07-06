<script>
  import { login, FIXTURE } from '../auth.svelte.js'

  // Prefilled with the fixture (seed) credentials — a dev convenience "for now".
  let email = $state(FIXTURE.email)
  let password = $state(FIXTURE.password)
  let error = $state(null)
  let busy = $state(false)

  async function submit(e) {
    e.preventDefault()
    if (busy) return
    busy = true
    error = null
    try {
      await login(email, password)
    } catch (err) {
      error = err.message
    } finally {
      busy = false
    }
  }
</script>

<div class="min-h-full grid place-items-center p-6">
  <div class="card bg-base-200 shadow-xl w-full max-w-sm">
    <form class="card-body gap-3" onsubmit={submit}>
      <div class="text-center mb-1">
        <div class="text-xl font-semibold tracking-wide">StarRaid</div>
        <div class="text-xs opacity-60">Game-Master Console</div>
      </div>

      <label class="floating-label">
        <span>Email</span>
        <input class="input input-bordered w-full" type="email" bind:value={email} placeholder="Email" autocomplete="username" />
      </label>

      <label class="floating-label">
        <span>Password</span>
        <input class="input input-bordered w-full" type="password" bind:value={password} placeholder="Password" autocomplete="current-password" />
      </label>

      {#if error}
        <div role="alert" class="alert alert-error py-2 text-sm">
          <span>{error}</span>
        </div>
      {/if}

      <button class="btn btn-primary mt-1" type="submit" disabled={busy}>
        {#if busy}<span class="loading loading-spinner loading-sm"></span>{/if}
        Sign in
      </button>

      <p class="text-[0.7rem] leading-snug opacity-50 text-center">
        Placeholder login (not a security boundary) — prefilled with the seed account.
        Real admin auth is a TBD.
      </p>
    </form>
  </div>
</div>
