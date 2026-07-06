<script>
  import { get } from '../api.js'
  import { ts } from '../format.js'
  import State from '../State.svelte'
  import Icon from '../Icon.svelte'

  let users = $state([])
  let error = $state(null)
  let loading = $state(true)

  $effect(() => {
    ;(async () => {
      loading = true
      error = null
      try {
        users = await get('/users')
      } catch (e) {
        error = e.message
      } finally {
        loading = false
      }
    })()
  })
</script>

<h1 class="text-xl font-semibold mb-4">Users</h1>

<State {loading} {error} empty={users.length === 0} emptyText="No accounts.">
  <div class="card bg-base-200 shadow">
    <div class="overflow-x-auto">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Email</th>
            <th>Status</th>
            <th>Characters</th>
            <th>Last login</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each users as u}
            <tr class="hover">
              <td class="tabular opacity-60">{u.id}</td>
              <td class="font-medium">{u.email}</td>
              <td>
                <span class="badge badge-sm {u.status === 'active' ? 'badge-success' : 'badge-ghost'}">{u.status}</span>
              </td>
              <td class="tabular">{u.characters}</td>
              <td class="text-sm opacity-70">{ts(u.last_login_at)}</td>
              <td class="text-right">
                <a class="btn btn-ghost btn-xs" href={`#/users/${u.id}`} aria-label="open account">
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
