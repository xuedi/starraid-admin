// Command admin is the StarRaid game-master web console backend (see docs/admin.md).
// It shares the server's DB models (the schema contract) and serves the Svelte UI.
package main

import (
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
	// TODO: JSON admin API (world events, faction directives, fixtures, monitoring);
	// serve the built Svelte SPA from web/dist.

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
