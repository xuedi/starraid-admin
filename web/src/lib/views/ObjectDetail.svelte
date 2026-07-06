<script>
  import { get } from '../api.js'
  import { route } from '../router.svelte.js'
  import { num, coord, orDash, kindBadge } from '../format.js'
  import State from '../State.svelte'
  import Icon from '../Icon.svelte'

  let obj = $state(null)
  let error = $state(null)
  let loading = $state(true)

  $effect(() => {
    const id = route.params.id
    ;(async () => {
      loading = true
      error = null
      try {
        obj = await get(`/objects/${id}`)
      } catch (e) {
        error = e.message
      } finally {
        loading = false
      }
    })()
  })

  // module_types.params is forwarded as raw JSON; render the set fields as chips.
  function paramChips(params) {
    if (!params || typeof params !== 'object') return []
    return Object.entries(params)
      .filter(([, v]) => v !== 0 && v !== '' && v !== null && v !== undefined)
      .map(([k, v]) => `${k.replaceAll('_', ' ')}: ${typeof v === 'number' ? num(v) : v}`)
  }

  let slots = $derived(Array.isArray(obj?.class?.slots) ? obj.class.slots : [])
</script>

<div class="flex items-center gap-2 mb-4">
  <a class="btn btn-ghost btn-sm" href="#/objects"><Icon name="back" size={16} /> Objects</a>
</div>

<State {loading} {error}>
  {#if obj}
    <!-- Identity + live condition -->
    <div class="card bg-base-200 shadow mb-6">
      <div class="card-body">
        <div class="flex flex-wrap items-center gap-3">
          <h1 class="card-title text-lg">{orDash(obj.name)}</h1>
          <span class="badge {kindBadge(obj.class.kind)}">{obj.class.kind}</span>
          <span class="badge badge-ghost">{obj.status}</span>
          <span class="text-xs opacity-50 tabular ml-auto">#{obj.id}</span>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-sm mt-2">
          <div><div class="opacity-60">Class</div><div>{obj.class.name} <span class="opacity-50 font-mono text-xs">({obj.class.key})</span></div></div>
          <div><div class="opacity-60">Size</div><div>{obj.class.size_class}</div></div>
          <div><div class="opacity-60">Sector</div><div>{obj.sector}</div></div>
          <div><div class="opacity-60">Position</div><div class="tabular">{coord(obj.x, obj.y)}</div></div>
          <div><div class="opacity-60">Base mass</div><div class="tabular">{num(obj.class.base_mass)}</div></div>
          <div><div class="opacity-60">Cargo volume</div><div class="tabular">{num(obj.class.base_cargo_volume)}</div></div>
          <div>
            <div class="opacity-60">Owner</div>
            <div>
              {#if obj.owner_account_id}
                <a class="link link-hover" href={`#/users/${obj.owner_account_id}`}>{orDash(obj.owner)}</a>
              {:else}
                <span class="opacity-70">NPC / unowned</span>
              {/if}
            </div>
          </div>
          <div><div class="opacity-60">Updated</div><div class="text-xs opacity-70">{obj.updated_at?.slice(0, 19).replace('T', ' ')}</div></div>
        </div>

        <div class="flex flex-wrap gap-6 mt-3">
          <div>
            <div class="opacity-60 text-xs">Hull</div>
            <div class="text-lg tabular font-semibold">{num(obj.health)}</div>
          </div>
          <div>
            <div class="opacity-60 text-xs">Shield</div>
            <div class="text-lg tabular font-semibold">{num(obj.shield)}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Fitting -->
    <section class="mb-6">
      <h2 class="text-sm uppercase tracking-wide opacity-60 mb-2">
        Fitting ({obj.modules.length})
        {#if slots.length}
          <span class="opacity-50 normal-case tracking-normal font-normal ml-1">
            · slots: {slots.map((s) => `${s.count}×${s.kind}/${s.size}`).join(', ')}
          </span>
        {/if}
      </h2>
      <div class="card bg-base-200 shadow">
        <div class="overflow-x-auto">
          {#if obj.modules.length}
            <table class="table table-sm">
              <thead><tr><th>Module</th><th>Slot</th><th>Quality</th><th>Status</th><th>Parameters</th></tr></thead>
              <tbody>
                {#each obj.modules as m}
                  <tr>
                    <td class="font-medium">{m.name} <span class="opacity-50 font-mono text-xs">({m.key})</span></td>
                    <td class="text-xs opacity-70">{m.slot_kind} #{m.slot_index}</td>
                    <td class="tabular">{m.quality}</td>
                    <td><span class="badge badge-ghost badge-sm">{m.status}</span></td>
                    <td>
                      <div class="flex flex-wrap gap-1">
                        {#each paramChips(m.params) as chip}
                          <span class="badge badge-outline badge-sm">{chip}</span>
                        {/each}
                      </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {:else}
            <p class="p-4 text-sm opacity-60">No modules fitted.</p>
          {/if}
        </div>
      </div>
    </section>

    <!-- Cargo -->
    <section>
      <h2 class="text-sm uppercase tracking-wide opacity-60 mb-2">Cargo ({obj.cargo.length})</h2>
      <div class="card bg-base-200 shadow">
        <div class="overflow-x-auto">
          {#if obj.cargo.length}
            <table class="table table-sm">
              <thead><tr><th>Item</th><th>Category</th><th class="text-right">Quantity</th></tr></thead>
              <tbody>
                {#each obj.cargo as c}
                  <tr>
                    <td class="font-medium">{c.name} <span class="opacity-50 font-mono text-xs">({c.key})</span></td>
                    <td><span class="badge badge-ghost badge-sm">{c.category}</span></td>
                    <td class="text-right tabular">{num(c.quantity)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {:else}
            <p class="p-4 text-sm opacity-60">Empty hold.</p>
          {/if}
        </div>
      </div>
    </section>
  {/if}
</State>
