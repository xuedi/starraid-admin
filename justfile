# StarRaid admin — Go backend + Svelte frontend. Run `just` to list recipes.
#
# The backend serves a read-only JSON API over the same PostgreSQL the server
# owns, proxies the server's live telemetry, and serves the built Svelte SPA.
# Start the DB from the `server` repo (`just db-up`) or point DATABASE_URL at
# your own.
#
# Environment (all optional; defaults suit the local stack):
#   ADMIN_ADDR        backend listen address              (default :8090)
#   DATABASE_URL      Postgres DSN (read-only use)        (server's local DSN)
#   ADMIN_WEB_DIR     built SPA dir served from disk      (default web/dist)
#   SERVER_STATS_URL  game server /stats to proxy         (http://localhost:8080/stats)

# List available recipes
default:
    @just --list

# Install backend (Go) + frontend (npm) dependencies
install:
    go mod download
    cd web && npm install

# Build the backend and the frontend bundle (web/dist)
build:
    go build ./...
    cd web && npm run build

# Run the backend (serves the API + built SPA at :8090). Build first so web/dist
# exists — `just build && just run` is the out-of-the-box flow (stackctl uses it).
run:
    go run ./cmd/admin

# Run the frontend dev server (Vite, hot reload) — proxies /api to the backend
# at :8090, so run `just run` in another shell for live data.
run-web:
    cd web && npm run dev

# Seed the starting world into the DB (idempotent). Assumes the server already
# ran `just db-up && just migrate`. Override the target via DATABASE_URL.
seed:
    go run ./cmd/seed

# Format Go code
fmt:
    go fmt ./...

# Vet
vet:
    go vet ./...

# Run Go tests
test:
    go test ./...
