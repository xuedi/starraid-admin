// Command admin is the StarRaid game-master web console backend (see
// docs/admin.md). It serves a read-only JSON API over the same PostgreSQL the
// server owns, proxies the server's live telemetry, and serves the built Svelte
// SPA. Read-only first cut: it surfaces DB + telemetry state; authoring,
// fixtures, and admin auth are parked (see .claude/plans/admin-web-console.md).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xuedi/starraid-admin/internal/api"
	"github.com/xuedi/starraid-admin/internal/db"
	"github.com/xuedi/starraid-admin/internal/web"
)

func main() {
	addr := envOr("ADMIN_ADDR", ":8090")
	dsn := envOr("DATABASE_URL", "postgres://starraid:starraid@localhost:5432/starraid?sslmode=disable")
	webDir := envOr("ADMIN_WEB_DIR", "web/dist")
	statsURL := envOr("SERVER_STATS_URL", "http://localhost:8080/stats")

	// Open the DB fail-soft: a failure logs a warning and leaves the store nil,
	// so the API answers 503 on DB routes but the process still boots and binds
	// (telemetry + SPA keep working). This deliberately avoids the "crash on
	// start" pattern that motivated this work (see the plan).
	var store api.Store
	dbc, err := db.Open(context.Background(), dsn)
	if err != nil {
		slog.Warn("admin: database unavailable at startup — DB routes will 503", "err", err)
	} else {
		store = dbc
		defer dbc.Close()
	}

	mux := http.NewServeMux()
	api.New(store, statsURL).Mount(mux)
	mux.Handle("/", web.Handler(webDir))

	srv := &http.Server{Addr: addr, Handler: mux}

	// Clean shutdown on SIGINT/SIGTERM so a stackctl stop/restart doesn't orphan
	// the process squatting the port (the crash-context hardening note).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("starraid admin backend starting", "addr", addr, "web_dir", webDir, "stats", statsURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("admin server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("admin: shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("admin: graceful shutdown failed", "err", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
