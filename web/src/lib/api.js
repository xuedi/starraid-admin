// Tiny fetch helper for the admin JSON API (base /api). Throws an Error carrying
// the backend's {error} message + HTTP status so views can render a clean state.
const BASE = '/api'

export async function get(path) {
  let res
  try {
    res = await fetch(BASE + path, { headers: { accept: 'application/json' } })
  } catch {
    throw Object.assign(new Error('network error — is the admin backend running?'), { status: 0 })
  }
  return handle(res)
}

export async function post(path, body) {
  let res
  try {
    res = await fetch(BASE + path, {
      method: 'POST',
      headers: { 'content-type': 'application/json', accept: 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    throw Object.assign(new Error('network error — is the admin backend running?'), { status: 0 })
  }
  return handle(res)
}

async function handle(res) {
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json()
      if (body && body.error) msg = body.error
    } catch {
      /* non-JSON error body */
    }
    throw Object.assign(new Error(msg), { status: res.status })
  }
  return res.json()
}
