# admin

The StarRaid game-master web console — **Go backend + Svelte frontend**. See
[../docs/admin.md](../docs/admin.md).

The GM tool for steering the living world: world events, faction directives, resources/markets,
**fixtures** (e.g. "pirate base with N ships"), and live monitoring. The backend is Go so it
can import the server's DB models/migrations directly (the schema contract — no drift); it
writes authored content to PostgreSQL with a change flag the server reconciles.

## Prerequisites

- Go 1.26+
- Node.js + npm (frontend)
- PostgreSQL — the same DB the server owns. Start it from the [`server`](../server) repo
  (`just db-up`) or point `DATABASE_URL` at your own.

## Getting started

```sh
just install     # go mod download + npm install (web/)
just run         # backend, serves :8090   (env: ADMIN_ADDR)
just run-web     # frontend dev server (Vite, hot reload) in another shell
```

`just build` compiles the backend and bundles the frontend to `web/dist`. Run `just` to list
every recipe (`build`, `run`, `run-web`, `fmt`, `vet`).

Layout: `cmd/admin` (backend entry), `internal/httpapi` (admin JSON API — planned),
`web/` (Svelte SPA, built to `web/dist` and served by the backend).
