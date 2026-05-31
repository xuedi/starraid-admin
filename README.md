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
every recipe (`build`, `run`, `run-web`, `seed`, `fmt`, `vet`).

## Seeding the starting world

`cmd/seed` authors the minimal Phase-3 world into PostgreSQL — one **"Starting Area"** sector,
the `test@example.org` account (password `1234`, bcrypt-hashed) + a `Test Pilot` character, and
**4 ships**: one owned by the player at `(0,0)` and three NPC ships at spread positions. It
composes instances from the `spaceship` catalog type the **server** migrated; it never invents
a type. The seed is **idempotent** — re-running yields the same world (no duplicates).

Run order (the server owns the schema, so migrate first):

```sh
cd ../server && just db-up && just migrate   # start Postgres + apply migrations
cd ../admin   && just seed                    # author the starting world
```

Override the target DB with `DATABASE_URL` (defaults to the server's local DSN).

Layout: `cmd/admin` (backend entry), `cmd/seed` (world seed CLI), `internal/httpapi`
(admin JSON API — planned), `web/` (Svelte SPA, built to `web/dist` and served by the backend).
