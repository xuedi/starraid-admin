# admin

The StarRaid game-master web console — **Go backend + Svelte frontend**. See
[../docs/admin.md](../docs/admin.md).

The GM tool for steering the living world: world events, faction directives, resources/markets,
**fixtures** (e.g. "pirate base with N ships"), and live monitoring. The backend is Go so it
can import the server's DB models/migrations directly (the schema contract — no drift); it
writes authored content to PostgreSQL with a change flag the server reconciles.

```sh
# backend
go run ./cmd/admin            # or: just run-admin   (serves :8090)

# frontend (dev)
cd web && npm install && npm run dev   # or: just run-admin-web
```

Layout: `cmd/admin` (backend entry), `internal/httpapi` (admin JSON API),
`web/` (Svelte SPA, built to `web/dist` and served by the backend).
