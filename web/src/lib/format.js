// Consistent formatting for the console: thousands-separated numbers, (x, y)
// coordinates, relative-ish timestamps, and safe fallbacks for null fields.

export function num(n) {
  if (n === null || n === undefined) return '—'
  return Number(n).toLocaleString('en-US')
}

export function coord(x, y) {
  return `(${num(x)}, ${num(y)})`
}

// A short, locale-stable timestamp. Nulls render as an em dash.
export function ts(s) {
  if (!s) return '—'
  const d = new Date(s)
  if (isNaN(d)) return '—'
  return d.toLocaleString('en-GB', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Uptime seconds → compact "1d 2h 3m" / "4m 5s".
export function dur(s) {
  if (s === null || s === undefined) return '—'
  s = Math.floor(Number(s))
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (d) return `${d}d ${h}h ${m}m`
  if (h) return `${h}h ${m}m`
  if (m) return `${m}m ${sec}s`
  return `${sec}s`
}

// Fallback for a possibly-null string (object/character names).
export function orDash(v) {
  return v === null || v === undefined || v === '' ? '—' : v
}

// A DaisyUI badge colour class per object kind.
export function kindBadge(kind) {
  switch (kind) {
    case 'ship':
      return 'badge-info'
    case 'structure':
      return 'badge-warning'
    case 'object':
      return 'badge-neutral'
    default:
      return 'badge-ghost'
  }
}
