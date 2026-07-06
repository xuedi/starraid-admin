<script>
  import { route } from './lib/router.svelte.js'
  import { auth } from './lib/auth.svelte.js'
  import Sidebar from './lib/Sidebar.svelte'
  import Topbar from './lib/Topbar.svelte'
  import Login from './lib/views/Login.svelte'
  import Dashboard from './lib/views/Dashboard.svelte'
  import Users from './lib/views/Users.svelte'
  import UserDetail from './lib/views/UserDetail.svelte'
  import Objects from './lib/views/Objects.svelte'
  import ObjectDetail from './lib/views/ObjectDetail.svelte'
  import MapView from './lib/views/MapView.svelte'
  import NotFound from './lib/views/NotFound.svelte'

  const views = {
    dashboard: Dashboard,
    users: Users,
    user: UserDetail,
    objects: Objects,
    object: ObjectDetail,
    map: MapView,
    notfound: NotFound,
  }
  let View = $derived(views[route.name] ?? NotFound)
</script>

{#if !auth.authed}
  <div class="h-full">
    <Login />
  </div>
{:else}
<div class="drawer lg:drawer-open h-full">
  <input id="nav-drawer" type="checkbox" class="drawer-toggle" />

  <div class="drawer-content flex flex-col h-full min-h-0">
    <Topbar />

    <!-- Read-only banner. The login is a placeholder gate, not a security
         boundary (real admin auth is a parked TBD). -->
    <div class="bg-warning/15 text-warning-content/90 text-xs px-4 py-1.5 border-b border-warning/20 flex items-center gap-2">
      <span class="badge badge-warning badge-sm">GM</span>
      Read-only monitor · placeholder login (not a security boundary) · exposes account emails — run behind a network boundary.
    </div>

    <main class="grow overflow-y-auto p-4 md:p-6">
      <!-- Remount views on param change (e.g. user 1 → user 2) via keyed block. -->
      {#key route.path}
        <View />
      {/key}
    </main>
  </div>

  <div class="drawer-side z-20">
    <label for="nav-drawer" aria-label="close sidebar" class="drawer-overlay"></label>
    <Sidebar />
  </div>
</div>
{/if}
