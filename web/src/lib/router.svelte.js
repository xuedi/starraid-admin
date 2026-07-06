// Hand-rolled hash router (see the plan — no SvelteKit). A reactive `route` rune
// tracks the current view + params, parsed from location.hash. The route table
// maps `#/...` patterns to a view name the App renders. ~50 lines, zero deps.

const routes = [
  { name: 'dashboard', re: /^\/?$/, keys: [] },
  { name: 'users', re: /^\/users\/?$/, keys: [] },
  { name: 'user', re: /^\/users\/(\d+)\/?$/, keys: ['id'] },
  { name: 'objects', re: /^\/objects\/?$/, keys: [] },
  { name: 'object', re: /^\/objects\/(\d+)\/?$/, keys: ['id'] },
  { name: 'map', re: /^\/map\/?$/, keys: [] },
]

function resolve(hash) {
  const path = (hash || '').replace(/^#/, '') || '/'
  for (const r of routes) {
    const m = path.match(r.re)
    if (m) {
      const params = {}
      r.keys.forEach((k, i) => (params[k] = m[i + 1]))
      return { name: r.name, params, path }
    }
  }
  return { name: 'notfound', params: {}, path }
}

export const route = $state(resolve(location.hash))

function update() {
  const r = resolve(location.hash)
  route.name = r.name
  route.params = r.params
  route.path = r.path
}

window.addEventListener('hashchange', update)

// navigate programmatically (e.g. a map blip click → object detail).
export function navigate(to) {
  if (location.hash === '#' + to) update()
  else location.hash = to
}

// The top-level nav section a route belongs to, so detail pages keep their
// parent sidebar item highlighted.
export function section(name) {
  if (name === 'user') return 'users'
  if (name === 'object') return 'objects'
  return name
}
