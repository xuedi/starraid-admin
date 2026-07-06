// Package web serves the built Svelte SPA (web/dist) from disk with a
// single-page-app fallback: any path that isn't a real file falls back to
// index.html so client-side hash routing works and deep links resolve.
// Embedding the bundle into the binary is a parked hardening (see the plan).
package web

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Handler serves static files out of dir, falling back to index.html. If the
// bundle hasn't been built yet (no index.html), it serves a helpful placeholder
// telling the operator to run `just build` — so the backend URL is never blank.
func Handler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := filepath.Clean(r.URL.Path)
		full := filepath.Join(dir, clean)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		idx := filepath.Join(dir, "index.html")
		if _, err := os.Stat(idx); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, notBuiltHTML)
			return
		}
		http.ServeFile(w, r, idx)
	})
}

const notBuiltHTML = `<!doctype html>
<meta charset="utf-8">
<title>StarRaid — GM Console</title>
<style>body{margin:0;min-height:100vh;display:grid;place-items:center;
font:16px/1.6 system-ui,sans-serif;background:#0b1020;color:#e6ecff}
main{max-width:34rem;padding:2rem}code{background:#1a2440;padding:.1rem .4rem;border-radius:.3rem}</style>
<main>
<h1>StarRaid — GM Console</h1>
<p>The admin backend is up, but the frontend bundle isn't built yet.</p>
<p>Build it with <code>just build</code> (runs <code>npm run build</code> into <code>web/dist</code>),
then reload. During development run <code>just run-web</code> for hot reload at
<code>:5173</code> (it proxies <code>/api</code> here).</p>
<p>API is live: <a href="/api/summary" style="color:#6ea8ff">/api/summary</a> ·
<a href="/api/telemetry" style="color:#6ea8ff">/api/telemetry</a></p>
</main>
`
