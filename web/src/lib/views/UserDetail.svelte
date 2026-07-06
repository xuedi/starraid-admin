<script>
  import { get } from '../api.js'
  import { route } from '../router.svelte.js'
  import { ts, orDash } from '../format.js'
  import State from '../State.svelte'
  import Icon from '../Icon.svelte'

  let user = $state(null)
  let error = $state(null)
  let loading = $state(true)

  $effect(() => {
    const id = route.params.id
    ;(async () => {
      loading = true
      error = null
      try {
        user = await get(`/users/${id}`)
      } catch (e) {
        error = e.message
      } finally {
        loading = false
      }
    })()
  })
</script>

<div class="flex items-center gap-2 mb-4">
  <a class="btn btn-ghost btn-sm" href="#/users"><Icon name="back" size={16} /> Users</a>
</div>

<State {loading} {error}>
  {#if user}
    <div class="card bg-base-200 shadow mb-6">
      <div class="card-body">
        <h1 class="card-title text-lg">{user.email}</h1>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm">
          <div><div class="opacity-60">Account ID</div><div class="tabular">{user.id}</div></div>
          <div><div class="opacity-60">Status</div><div><span class="badge badge-sm {user.status === 'active' ? 'badge-success' : 'badge-ghost'}">{user.status}</span></div></div>
          <div><div class="opacity-60">Created</div><div>{ts(user.created_at)}</div></div>
          <div><div class="opacity-60">Last login</div><div>{ts(user.last_login_at)}</div></div>
        </div>
      </div>
    </div>

    <h2 class="text-sm uppercase tracking-wide opacity-60 mb-2">
      Characters ({user.characters?.length ?? 0})
    </h2>

    {#if user.characters && user.characters.length}
      <div class="flex flex-col gap-4">
        {#each user.characters as c}
          <div class="card bg-base-200 shadow">
            <div class="card-body p-4">
              <div class="flex flex-wrap items-baseline gap-x-4 gap-y-1">
                <span class="font-semibold">{c.name}</span>
                <span class="text-sm opacity-70">prestige <span class="tabular">{c.prestige}</span></span>
                <span class="text-sm opacity-70">faction {orDash(c.faction)}</span>
                <span class="text-xs opacity-50 ml-auto">created {ts(c.created_at)}</span>
              </div>

              <div class="mt-2">
                <div class="text-xs uppercase tracking-wide opacity-50 mb-1">
                  Owned objects ({c.objects?.length ?? 0})
                </div>
                {#if c.objects && c.objects.length}
                  <div class="overflow-x-auto">
                    <table class="table table-sm">
                      <thead><tr><th>ID</th><th>Name</th><th>Class</th><th>Sector</th><th></th></tr></thead>
                      <tbody>
                        {#each c.objects as o}
                          <tr class="hover">
                            <td class="tabular opacity-60">{o.id}</td>
                            <td>{orDash(o.name)}</td>
                            <td>{o.class}</td>
                            <td>{o.sector}</td>
                            <td class="text-right">
                              <a class="btn btn-ghost btn-xs" href={`#/objects/${o.id}`} aria-label="open object">
                                <Icon name="detail" size={16} />
                              </a>
                            </td>
                          </tr>
                        {/each}
                      </tbody>
                    </table>
                  </div>
                {:else}
                  <p class="text-sm opacity-50">No owned objects.</p>
                {/if}
              </div>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <p class="opacity-60">This account has no characters.</p>
    {/if}
  {/if}
</State>
