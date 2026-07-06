<script>
  import { get } from '../api.js'
  import { telemetry, startTelemetry } from '../telemetry.svelte.js'
  import { num, dur } from '../format.js'
  import State from '../State.svelte'

  let summary = $state(null)
  let error = $state(null)
  let loading = $state(true)

  async function load() {
    loading = true
    error = null
    try {
      summary = await get('/summary')
    } catch (e) {
      error = e.message
    } finally {
      loading = false
    }
  }

  $effect(() => {
    startTelemetry(2000)
    load()
    // Refresh DB counts on the same cadence as telemetry.
    const t = setInterval(load, 5000)
    return () => clearInterval(t)
  })

  function tval(v, fmt = num) {
    return telemetry.up && v !== null && v !== undefined ? fmt(v) : '—'
  }
</script>

<h1 class="text-xl font-semibold mb-4">Dashboard</h1>

<!-- Live server telemetry (proxied /stats). Server-down shows dashes. -->
<section class="mb-6">
  <div class="flex items-center gap-2 mb-2">
    <h2 class="text-sm uppercase tracking-wide opacity-60">Live server</h2>
    <span class="badge badge-sm {telemetry.up ? 'badge-success' : 'badge-error'}">
      {telemetry.up ? 'online' : 'offline'}
    </span>
  </div>
  <div class="stats stats-vertical sm:stats-horizontal shadow bg-base-200 w-full">
    <div class="stat">
      <div class="stat-title">Objects</div>
      <div class="stat-value text-2xl tabular">{tval(telemetry.objects)}</div>
      <div class="stat-desc">in the running sim</div>
    </div>
    <div class="stat">
      <div class="stat-title">Sessions</div>
      <div class="stat-value text-2xl tabular">{tval(telemetry.sessions)}</div>
      <div class="stat-desc">connected clients + bots</div>
    </div>
    <div class="stat">
      <div class="stat-title">Tick</div>
      <div class="stat-value text-2xl tabular">{telemetry.up && telemetry.tick_hz != null ? telemetry.tick_hz + ' Hz' : '—'}</div>
      <div class="stat-desc">simulation rate</div>
    </div>
    <div class="stat">
      <div class="stat-title">Uptime</div>
      <div class="stat-value text-2xl tabular">{tval(telemetry.uptime_s, dur)}</div>
      <div class="stat-desc">since server start</div>
    </div>
  </div>
</section>

<!-- Persistent DB counts. -->
<section>
  <h2 class="text-sm uppercase tracking-wide opacity-60 mb-2">Database</h2>
  <State {loading} {error}>
    {#if summary}
      <div class="stats stats-vertical sm:stats-horizontal shadow bg-base-200 w-full mb-4">
        <div class="stat">
          <div class="stat-title">Accounts</div>
          <div class="stat-value text-2xl tabular">{num(summary.accounts)}</div>
        </div>
        <div class="stat">
          <div class="stat-title">Characters</div>
          <div class="stat-value text-2xl tabular">{num(summary.characters)}</div>
        </div>
        <div class="stat">
          <div class="stat-title">Objects</div>
          <div class="stat-value text-2xl tabular">{num(summary.objects)}</div>
        </div>
        <div class="stat">
          <div class="stat-title">Sectors</div>
          <div class="stat-value text-2xl tabular">{num(summary.sectors)}</div>
        </div>
      </div>

      <div class="card bg-base-200 shadow">
        <div class="card-body p-4">
          <h3 class="text-sm font-semibold opacity-80 mb-2">Objects by class</h3>
          {#if summary.by_class && summary.by_class.length}
            <div class="overflow-x-auto">
              <table class="table table-sm">
                <thead>
                  <tr><th>Class</th><th>Key</th><th class="text-right">Count</th></tr>
                </thead>
                <tbody>
                  {#each summary.by_class as c}
                    <tr>
                      <td>{c.name}</td>
                      <td class="opacity-60 font-mono text-xs">{c.key}</td>
                      <td class="text-right tabular">{num(c.count)}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {:else}
            <p class="opacity-60 text-sm">No objects seeded.</p>
          {/if}
        </div>
      </div>
    {/if}
  </State>
</section>
