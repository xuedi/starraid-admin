# admin

The StarRaid game-master web console — **Go backend + Svelte frontend**. See
[../docs/admin.md](../docs/admin.md).

The eventual GM tool for steering the living world (world events, faction directives, markets,
**fixtures**, monitoring). This first cut is a **read-only monitor**: it *surfaces* DB state and
live server telemetry; authoring/editing and the change-flag write path are a later phase.

A browser console with a left sidebar — **Dashboard · Users · Objects · Map** — over the same
PostgreSQL the server owns:

- **Dashboard** — DB counts + live server telemetry (objects/sessions/tick/uptime), degrading to
  dashes when the server is down.
- **Users** — accounts table → account detail (characters → their owned objects).
- **Objects** — objects table (sector-filtered) → object detail (class/attributes, position,
  owner, hull/shield, **fitting** + **cargo**).
- **Map** — a `<canvas>` plotting a sector's objects by `(x, y)`; drag-pan, wheel-zoom, click a
  blip → its detail page.

Styled with **Tailwind v4 + DaisyUI v5** (dark theme).

> ⚠️ **Placeholder login — not a security boundary.** The console shows a login screen
> (prefilled with the seed account, `test@example.org` / `1234`) that bcrypt-verifies against the
> `accounts` table, but it only gates the SPA client-side — the read API stays open, so this is a
> "for now" front door, not real auth. It also exposes account emails; run it locally / behind a
> network boundary. Real session-gated admin auth is a parked TBD
> ([docs/admin.md](../docs/admin.md)); the UI shows a banner to that effect.

## Prerequisites

- Go 1.26+
- Node.js + npm (frontend)
- PostgreSQL — the same DB the server owns. Start it from the [`server`](../server) repo
  (`just db-up`) or point `DATABASE_URL` at your own.

## Getting started

```sh
just install         # go mod download + npm install (web/)
just build && just run   # build the SPA + backend, then serve the console at :8090
```

For frontend hot reload during development, run the Vite dev server alongside the backend:

```sh
just run       # backend on :8090 (in one shell)
just run-web   # Vite on :5173, proxies /api → :8090 (in another)
```

`just build` compiles the backend and bundles the frontend to `web/dist`, which the backend
serves from disk. Run `just` to list every recipe (`build`, `run`, `run-web`, `seed`, `fmt`,
`vet`, `test`).

### Environment

| Var | Default | Purpose |
|-----|---------|---------|
| `ADMIN_ADDR` | `:8090` | backend listen address |
| `DATABASE_URL` | server's local DSN | Postgres DSN (read-only use) |
| `ADMIN_WEB_DIR` | `web/dist` | built SPA directory served from disk |
| `SERVER_STATS_URL` | `http://localhost:8080/stats` | game server `/stats` proxied by `/api/telemetry` |

The backend opens the DB **fail-soft**: if it's unavailable at startup, the process still boots
and binds, the DB API routes answer `503`, and telemetry + the SPA keep working.

## API (read-only)

`GET /api/summary`, `/api/users`, `/api/users/{id}`, `/api/objects[?sector=]`,
`/api/objects/{id}`, `/api/sectors`, `/api/map?sector=`, `/api/telemetry`, plus
`POST /api/login` (bcrypt-verifies the placeholder login) and `/healthz`. All JSON; the query
layer re-declares its own read queries against the shared schema (no cross-repo import of the
server).

## Seeding the starting world

`cmd/seed` authors the starting world into PostgreSQL — one **"Starting Area"** sector, the
`test@example.org` account (password `1234`, bcrypt-hashed) + a `Test Pilot` character, and a
varied scene (player skiff + NPC ships + a station + asteroids). It composes instances from the
catalog the **server** synced; it never invents a type. Idempotent — re-running yields the same
world.

Run order (the server owns the schema, so migrate first):

```sh
cd ../server && just db-up && just migrate   # start Postgres + apply migrations + sync catalog
cd ../admin   && just seed                    # author the starting world
```

Override the target DB with `DATABASE_URL` (defaults to the server's local DSN).

## Layout

`cmd/admin` (backend entry), `cmd/seed` (world seed CLI), `internal/db` (pgx pool + read
queries), `internal/api` (JSON API + telemetry proxy), `internal/web` (static/SPA serving),
`web/` (Svelte SPA, built to `web/dist`).
