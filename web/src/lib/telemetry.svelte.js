// Shared live-telemetry poller. One interval feeds both the topbar up/down dot
// and the Dashboard cards, ref-counted so it stops when nothing is mounted.
// Server-down is a normal state (up:false) — never an error (see /api/telemetry).
import { get } from './api.js'

export const telemetry = $state({
  up: false,
  objects: null,
  sessions: null,
  tick_hz: null,
  uptime_s: null,
  loaded: false,
})

let timer = null
let subscribers = 0

async function poll() {
  try {
    const t = await get('/telemetry')
    Object.assign(telemetry, t, { loaded: true })
  } catch {
    Object.assign(telemetry, {
      up: false,
      objects: null,
      sessions: null,
      tick_hz: null,
      uptime_s: null,
      loaded: true,
    })
  }
}

// startTelemetry begins (or joins) the poll loop and returns a cleanup to call
// from a component $effect.
export function startTelemetry(intervalMs = 2000) {
  subscribers++
  if (!timer) {
    poll()
    timer = setInterval(poll, intervalMs)
  }
  return () => {
    subscribers--
    if (subscribers <= 0 && timer) {
      clearInterval(timer)
      timer = null
    }
  }
}
