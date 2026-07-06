<script>
  import { get } from '../api.js'
  import { num, coord, orDash, kindBadge } from '../format.js'
  import State from '../State.svelte'
  import Icon from '../Icon.svelte'

  let objects = $state([])
  let sectors = $state([])
  let sectorFilter = $state('') // '' = all sectors
  let error = $state(null)
  let loading = $state(true)

  // Sector list for the filter (loaded once).
  $effect(() => {
    ;(async () => {
      try {
        sectors = await get('/sectors')
      } catch {
        /* filter degrades to none; the table still loads */
      }
    })()
  })

  // Objects reload whenever the filter changes.
  $effect(() => {
    const q = sectorFilter ? `?sector=${sectorFilter}` : ''
    ;(async () => {
      loading = true
      error = null
      try {
        objects = await get(`/objects${q}`)
      } catch (e) {
        error = e.message
      } finally {
        loading = false
      }
    })()
  })
</script>

<div class="flex flex-wrap items-center justify-between gap-3 mb-4">
  <h1 class="text-xl font-semibold">Objects</h1>
  <label class="flex items-center gap-2 text-sm">
    <span class="opacity-60">Sector</span>
    <select class="select select-sm select-bordered" bind:value={sectorFilter}>
      <option value="">All</option>
      {#each sectors as s}
        <option value={s.id}>{s.name}</option>
      {/each}
    </select>
  </label>
</div>

<State {loading} {error} empty={objects.length === 0} emptyText="No objects in this scope.">
  <div class="card bg-base-200 shadow">
    <div class="overflow-x-auto">
      <table class="table table-sm">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Class</th>
            <th>Kind</th>
            <th>Sector</th>
            <th>Owner</th>
            <th>Position</th>
            <th>Hull</th>
            <th>Shield</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each objects as o}
            <tr class="hover">
              <td class="tabular opacity-60">{o.id}</td>
              <td class="font-medium">{orDash(o.name)}</td>
              <td>{o.class_name}</td>
              <td><span class="badge badge-sm {kindBadge(o.kind)}">{o.kind}</span></td>
              <td>{o.sector}</td>
              <td>{orDash(o.owner)}</td>
              <td class="tabular text-xs opacity-70">{coord(o.x, o.y)}</td>
              <td class="tabular">{num(o.health)}</td>
              <td class="tabular">{num(o.shield)}</td>
              <td><span class="badge badge-ghost badge-sm">{o.status}</span></td>
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
  </div>
</State>
