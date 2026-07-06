// Command admin is the StarRaid game-master web console backend (see docs/admin.md).
// It shares the server's DB models (the schema contract) and serves the Svelte UI.
package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	addr := envOr("ADMIN_ADDR", ":8090")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	// Root: a placeholder console landing so the backend URL shows something. The
	// JSON admin API + Svelte SPA (world map, factions, settings/fixtures) are the
	// TODO — see docs/admin.md. "/" catches unmatched paths, so 404 the rest.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, landingHTML)
	})

	slog.Info("starraid admin backend starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("admin server failed", "err", err)
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// landingHTML is the placeholder console home served at "/". It intentionally
// carries no build step (no Svelte, no assets) so the backend URL is live the
// moment the process starts; the real GM console replaces it (see docs/admin.md).
const landingHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>StarRaid — GM Console</title>
<style>
  :root { color-scheme: light dark; }
  body { margin: 0; min-height: 100vh; display: grid; place-items: center;
         font: 16px/1.6 system-ui, sans-serif; background: #0b1020; color: #e6ecff; }
  main { max-width: 34rem; padding: 2rem; }
  h1 { margin: 0 0 .25rem; font-size: 1.5rem; letter-spacing: .02em; }
  .sub { color: #8ea0c8; margin: 0 0 1.5rem; }
  ul { list-style: none; padding: 0; margin: 0; }
  li { padding: .5rem .75rem; border: 1px solid #24314f; border-radius: .5rem;
       margin-bottom: .5rem; display: flex; justify-content: space-between; align-items: center; }
  li.todo { opacity: .55; }
  a { color: #6ea8ff; text-decoration: none; }
  .tag { font-size: .7rem; text-transform: uppercase; letter-spacing: .08em;
         color: #8ea0c8; border: 1px solid #24314f; border-radius: 1rem; padding: .1rem .55rem; }
  .live { color: #57d9a3; border-color: #1f5e46; }
</style>
</head>
<body>
<main>
  <h1>StarRaid — Game-Master Console</h1>
  <p class="sub">Admin backend is up. World &amp; settings tooling is scaffolding
  (see <code>docs/admin.md</code>); here is what exists today.</p>
  <ul>
    <li><a href="/healthz">/healthz</a> <span class="tag live">live</span></li>
    <li class="todo">World map &amp; objects <span class="tag">todo</span></li>
    <li class="todo">Factions &amp; contracts <span class="tag">todo</span></li>
    <li class="todo">Settings &amp; fixtures <span class="tag">todo</span></li>
  </ul>
</main>
</body>
</html>
`
