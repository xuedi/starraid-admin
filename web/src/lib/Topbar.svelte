<script>
  import { route } from './router.svelte.js'
  import { telemetry, startTelemetry } from './telemetry.svelte.js'
  import { auth, logout } from './auth.svelte.js'

  const titles = {
    dashboard: 'Dashboard',
    users: 'Users',
    user: 'Account',
    objects: 'Objects',
    object: 'Object',
    map: 'Sector Map',
    notfound: 'Not found',
  }
  let title = $derived(titles[route.name] ?? 'StarRaid')

  $effect(() => startTelemetry(2000))
</script>

<div class="navbar bg-base-100/80 backdrop-blur border-b border-base-300 min-h-14 px-3 sticky top-0 z-10">
  <div class="flex-none lg:hidden">
    <label for="nav-drawer" aria-label="open sidebar" class="btn btn-square btn-ghost btn-sm">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </label>
  </div>

  <div class="flex-1 px-2 font-semibold">{title}</div>

  <div class="flex-none flex items-center gap-2">
    <div class="badge badge-ghost gap-2 py-3">
      <span
        class="inline-block h-2 w-2 rounded-full {telemetry.up ? 'bg-success' : 'bg-error'}"
        class:animate-pulse={telemetry.up}
      ></span>
      <span class="text-xs">server {telemetry.up ? 'up' : 'down'}</span>
    </div>

    <div class="dropdown dropdown-end">
      <div tabindex="0" role="button" class="btn btn-ghost btn-sm gap-2">
        <span class="hidden sm:inline text-xs opacity-80">{auth.email}</span>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM4 20a8 8 0 0 1 16 0" />
        </svg>
      </div>
      <ul class="dropdown-content menu bg-base-100 rounded-box z-30 mt-2 w-44 p-2 shadow border border-base-300">
        <li class="menu-title text-xs truncate">{auth.email}</li>
        <li><button onclick={logout}>Sign out</button></li>
      </ul>
    </div>
  </div>
</div>
