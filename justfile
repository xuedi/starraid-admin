# StarRaid admin — Go backend + Svelte frontend. Run `just` to list recipes.
#
# The backend writes authored content to the same PostgreSQL the server owns. Start
# that DB from the `server` repo (`just db-up`) or point DATABASE_URL at your own.

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

# Run the backend (serves :8090 by default; env: ADMIN_ADDR)
run:
    go run ./cmd/admin

# Run the frontend dev server (Vite, hot reload)
run-web:
    cd web && npm run dev

# Format Go code
fmt:
    go fmt ./...

# Vet
vet:
    go vet ./...
