// Placeholder console gate (see the admin plan / docs/admin.md — real admin auth
// is a parked TBD). The backend bcrypt-verifies the seed credentials at
// POST /api/login; on success the SPA remembers it in sessionStorage and reveals
// the console. This is a "for now" front door, NOT a security boundary — the read
// API stays open. Prefilled with the fixture account for dev convenience.
import { post } from './api.js'

const KEY = 'admin_authed'
const EMAIL = 'admin_email'

export const auth = $state({
  authed: sessionStorage.getItem(KEY) === '1',
  email: sessionStorage.getItem(EMAIL) || '',
})

// Fixture credentials (from the admin seed) — prefilled into the login form.
export const FIXTURE = { email: 'test@example.org', password: '1234' }

export async function login(email, password) {
  const res = await post('/login', { email, password })
  auth.authed = true
  auth.email = res.email || email
  sessionStorage.setItem(KEY, '1')
  sessionStorage.setItem(EMAIL, auth.email)
}

export function logout() {
  auth.authed = false
  auth.email = ''
  sessionStorage.removeItem(KEY)
  sessionStorage.removeItem(EMAIL)
}
