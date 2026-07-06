<script>
  import { get } from '../api.js'
  import { navigate } from '../router.svelte.js'
  import { orDash, coord } from '../format.js'
  import State from '../State.svelte'

  let sectors = $state([])
  let sectorId = $state('')
  let objects = $state([])
  let error = $state(null)
  let loading = $state(true)

  // Canvas + view transform. screen = (world - cam) * scale, y flipped so +y is up.
  let canvas
  let wrap
  let W = $state(600)
  let H = $state(400)
  let cam = $state({ x: 0, y: 0 })
  let scale = $state(0.02)
  let needFit = false

  // Interaction bookkeeping.
  let dragging = $state(false)
  let moved = 0
  let last = { x: 0, y: 0 }
  let hovered = $state(null) // { obj, sx, sy }

  const COLORS = { ship: '#38bdf8', structure: '#fbbf24', object: '#94a3b8' }
  const OWNED = '#34d399'

  function toScreen(wx, wy) {
    return { sx: (wx - cam.x) * scale + W / 2, sy: H / 2 - (wy - cam.y) * scale }
  }
  function toWorld(sx, sy) {
    return { wx: cam.x + (sx - W / 2) / scale, wy: cam.y - (sy - H / 2) / scale }
  }

  function fit() {
    if (!objects.length) {
      cam = { x: 0, y: 0 }
      scale = 0.02
      return
    }
    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity
    for (const o of objects) {
      minX = Math.min(minX, o.x); maxX = Math.max(maxX, o.x)
      minY = Math.min(minY, o.y); maxY = Math.max(maxY, o.y)
    }
    cam = { x: (minX + maxX) / 2, y: (minY + maxY) / 2 }
    const spanX = Math.max(maxX - minX, 1)
    const spanY = Math.max(maxY - minY, 1)
    scale = Math.min(W / spanX, H / spanY) * 0.8
    if (!isFinite(scale) || scale <= 0) scale = 0.02
  }

  function niceStep(raw) {
    const pow = Math.pow(10, Math.floor(Math.log10(raw)))
    const n = raw / pow
    return (n >= 5 ? 5 : n >= 2 ? 2 : 1) * pow
  }

  function draw() {
    if (!canvas) return
    const dpr = window.devicePixelRatio || 1
    canvas.width = W * dpr
    canvas.height = H * dpr
    const ctx = canvas.getContext('2d')
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, W, H)

    // Grid — world-aligned lines at a "nice" step (~8 across).
    const worldSpan = W / scale
    const step = niceStep(worldSpan / 8)
    ctx.strokeStyle = 'rgba(148,163,184,0.12)'
    ctx.lineWidth = 1
    const tl = toWorld(0, 0)
    const br = toWorld(W, H)
    const startX = Math.floor(tl.wx / step) * step
    for (let x = startX; x <= br.wx; x += step) {
      const { sx } = toScreen(x, 0)
      ctx.beginPath(); ctx.moveTo(sx, 0); ctx.lineTo(sx, H); ctx.stroke()
    }
    const startY = Math.floor(br.wy / step) * step
    for (let y = startY; y <= tl.wy; y += step) {
      const { sy } = toScreen(0, y)
      ctx.beginPath(); ctx.moveTo(0, sy); ctx.lineTo(W, sy); ctx.stroke()
    }

    // Origin crosshair.
    const o = toScreen(0, 0)
    ctx.strokeStyle = 'rgba(148,163,184,0.4)'
    ctx.beginPath(); ctx.moveTo(o.sx - 8, o.sy); ctx.lineTo(o.sx + 8, o.sy)
    ctx.moveTo(o.sx, o.sy - 8); ctx.lineTo(o.sx, o.sy + 8); ctx.stroke()

    // Blips.
    ctx.font = '11px system-ui, sans-serif'
    for (const obj of objects) {
      const { sx, sy } = toScreen(obj.x, obj.y)
      const color = COLORS[obj.kind] ?? COLORS.object
      if (obj.owned) {
        ctx.strokeStyle = OWNED
        ctx.lineWidth = 2
        ctx.beginPath(); ctx.arc(sx, sy, 9, 0, Math.PI * 2); ctx.stroke()
      }
      ctx.fillStyle = color
      ctx.beginPath(); ctx.arc(sx, sy, 5, 0, Math.PI * 2); ctx.fill()
      if (objects.length <= 60) {
        ctx.fillStyle = 'rgba(226,232,240,0.75)'
        ctx.fillText(orDash(obj.name), sx + 9, sy + 4)
      }
    }
  }

  // Hit test: nearest blip within 10px of the cursor.
  function hit(sx, sy) {
    let best = null, bestD = 12 * 12
    for (const obj of objects) {
      const p = toScreen(obj.x, obj.y)
      const d = (p.sx - sx) ** 2 + (p.sy - sy) ** 2
      if (d < bestD) { bestD = d; best = { obj, sx: p.sx, sy: p.sy } }
    }
    return best
  }

  function onPointerDown(e) {
    dragging = true; moved = 0
    last = { x: e.clientX, y: e.clientY }
    canvas.setPointerCapture(e.pointerId)
  }
  function onPointerMove(e) {
    const rect = canvas.getBoundingClientRect()
    const sx = e.clientX - rect.left, sy = e.clientY - rect.top
    if (dragging) {
      const dx = e.clientX - last.x, dy = e.clientY - last.y
      moved += Math.abs(dx) + Math.abs(dy)
      cam = { x: cam.x - dx / scale, y: cam.y + dy / scale }
      last = { x: e.clientX, y: e.clientY }
      hovered = null
    } else {
      hovered = hit(sx, sy)
    }
  }
  function onPointerUp(e) {
    dragging = false
    const rect = canvas.getBoundingClientRect()
    if (moved < 5) {
      const h = hit(e.clientX - rect.left, e.clientY - rect.top)
      if (h) navigate(`/objects/${h.obj.id}`)
    }
  }
  function onWheel(e) {
    e.preventDefault()
    const rect = canvas.getBoundingClientRect()
    const sx = e.clientX - rect.left, sy = e.clientY - rect.top
    const before = toWorld(sx, sy)
    const factor = e.deltaY < 0 ? 1.15 : 1 / 1.15
    scale = Math.max(1e-4, Math.min(50, scale * factor))
    // Keep the cursor's world point stationary.
    const after = toWorld(sx, sy)
    cam = { x: cam.x + (before.wx - after.wx), y: cam.y + (before.wy - after.wy) }
  }

  // Load the sector list once, default to the first sector.
  $effect(() => {
    ;(async () => {
      try {
        sectors = await get('/sectors')
        if (sectors.length && !sectorId) sectorId = String(sectors[0].id)
      } catch (e) {
        error = e.message
        loading = false
      }
    })()
  })

  // Load blips whenever the selected sector changes.
  $effect(() => {
    if (!sectorId) return
    ;(async () => {
      loading = true
      error = null
      try {
        objects = await get(`/map?sector=${sectorId}`)
        needFit = true
      } catch (e) {
        error = e.message
      } finally {
        loading = false
      }
    })()
  })

  // Track container size.
  $effect(() => {
    if (!wrap) return
    const ro = new ResizeObserver((entries) => {
      const r = entries[0].contentRect
      W = Math.max(200, Math.floor(r.width))
      H = Math.max(200, Math.floor(r.height))
    })
    ro.observe(wrap)
    return () => ro.disconnect()
  })

  // Auto-fit once per data load, then redraw on any view/data/size change.
  $effect(() => {
    // touch reactive deps so this runs on their change
    void objects; void W; void H; void cam; void scale
    if (needFit && objects.length && W > 0) {
      fit()
      needFit = false
    }
    draw()
  })
</script>

<div class="flex flex-wrap items-center justify-between gap-3 mb-4">
  <h1 class="text-xl font-semibold">Sector Map</h1>
  <label class="flex items-center gap-2 text-sm">
    <span class="opacity-60">Sector</span>
    <select class="select select-sm select-bordered" bind:value={sectorId}>
      {#each sectors as s}
        <option value={String(s.id)}>{s.name} ({s.objects})</option>
      {/each}
    </select>
  </label>
</div>

<State {loading} {error} empty={!loading && objects.length === 0} emptyText="No objects in this sector.">
  <div class="card bg-base-200 shadow overflow-hidden">
    <div class="relative" bind:this={wrap} style="height: min(70vh, 640px);">
      <canvas
        bind:this={canvas}
        style="width:100%;height:100%;display:block;cursor:{dragging ? 'grabbing' : 'grab'}"
        onpointerdown={onPointerDown}
        onpointermove={onPointerMove}
        onpointerup={onPointerUp}
        onwheel={onWheel}
      ></canvas>

      {#if hovered}
        <div
          class="pointer-events-none absolute z-10 rounded-md bg-base-100/95 border border-base-300 px-2 py-1 text-xs shadow"
          style="left:{hovered.sx + 12}px; top:{hovered.sy + 12}px;"
        >
          <div class="font-medium">{orDash(hovered.obj.name)}</div>
          <div class="opacity-70">{hovered.obj.class} · {hovered.obj.kind}</div>
          <div class="opacity-50 tabular">{coord(hovered.obj.x, hovered.obj.y)}</div>
        </div>
      {/if}

      <!-- Legend -->
      <div class="absolute left-3 bottom-3 flex flex-col gap-1 text-xs bg-base-100/70 rounded-md px-2 py-1.5">
        <div class="flex items-center gap-2"><span class="h-2.5 w-2.5 rounded-full" style="background:#38bdf8"></span> ship</div>
        <div class="flex items-center gap-2"><span class="h-2.5 w-2.5 rounded-full" style="background:#fbbf24"></span> structure</div>
        <div class="flex items-center gap-2"><span class="h-2.5 w-2.5 rounded-full" style="background:#94a3b8"></span> object</div>
        <div class="flex items-center gap-2"><span class="h-2.5 w-2.5 rounded-full ring-2" style="background:transparent;box-shadow:0 0 0 2px #34d399"></span> owned</div>
      </div>

      <div class="absolute right-3 top-3 text-xs opacity-50 bg-base-100/60 rounded px-2 py-1">
        drag to pan · wheel to zoom · click a blip
      </div>
    </div>
  </div>
</State>
