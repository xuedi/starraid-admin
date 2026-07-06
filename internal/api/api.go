// Package api is the admin console's read-only JSON API over the shared DB, plus
// a telemetry proxy to the running server's /stats. Handlers are written against
// the Store interface so they unit-test with a fake (no DB in CI). When the DB is
// down the store is nil and the DB routes answer 503 — the process still boots
// and binds (see the plan's "crash on start" note).
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/xuedi/starraid-admin/internal/db"
)

// Store is the read surface the handlers need. *db.DB implements it; tests use a
// fake. User/Object return db.ErrNotFound when the id has no row.
type Store interface {
	Summary(ctx context.Context) (db.Summary, error)
	Users(ctx context.Context) ([]db.UserRow, error)
	User(ctx context.Context, id int64) (*db.UserDetail, error)
	Objects(ctx context.Context, sectorID *int64) ([]db.ObjectRow, error)
	Object(ctx context.Context, id int64) (*db.ObjectDetail, error)
	Sectors(ctx context.Context) ([]db.Sector, error)
	MapObjects(ctx context.Context, sectorID int64) ([]db.MapObject, error)
	CredByEmail(ctx context.Context, email string) (*db.AccountCred, error)
}

// Server holds the API dependencies. store may be nil (DB unavailable) — the DB
// routes then return 503; telemetry + healthz still work.
type Server struct {
	store    Store
	statsURL string
	httpc    *http.Client
}

// New builds an API server. statsURL is the running game server's /stats
// endpoint, proxied by /api/telemetry.
func New(store Store, statsURL string) *Server {
	return &Server{
		store:    store,
		statsURL: statsURL,
		httpc:    &http.Client{Timeout: 2 * time.Second},
	}
}

// Mount registers the API + healthz routes on the mux (Go 1.22 method+path
// patterns). Static/SPA serving is mounted separately by main at "/".
func (s *Server) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/summary", s.handleSummary)
	mux.HandleFunc("GET /api/users", s.handleUsers)
	mux.HandleFunc("GET /api/users/{id}", s.handleUser)
	mux.HandleFunc("GET /api/objects", s.handleObjects)
	mux.HandleFunc("GET /api/objects/{id}", s.handleObject)
	mux.HandleFunc("GET /api/sectors", s.handleSectors)
	mux.HandleFunc("GET /api/map", s.handleMap)
	mux.HandleFunc("GET /api/telemetry", s.handleTelemetry)
	mux.HandleFunc("POST /api/login", s.handleLogin)
}

// handleLogin is the console's placeholder front door: it bcrypt-verifies the
// posted email/password against the accounts table and returns {ok, email}. It
// is NOT a security boundary yet — the SPA gates itself on the result, but the
// read API stays open (real session-gated admin auth is a parked TBD,
// docs/admin.md). A uniform 401 hides whether the email or the password was
// wrong.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&req); err != nil {
		apiError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cred, err := s.store.CredByEmail(r.Context(), strings.TrimSpace(req.Email))
	if errors.Is(err, db.ErrNotFound) || cred == nil && err == nil {
		apiError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		slog.Error("admin login lookup failed", "err", err)
		apiError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(req.Password)) != nil {
		apiError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "email": cred.Email})
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	v, err := s.store.Summary(r.Context())
	writeResult(w, v, err)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	v, err := s.store.Users(r.Context())
	writeResult(w, v, err)
}

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.store.User(r.Context(), id)
	writeResult(w, v, err)
}

func (s *Server) handleObjects(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	sector, ok := queryID(w, r, "sector")
	if !ok {
		return
	}
	v, err := s.store.Objects(r.Context(), sector)
	writeResult(w, v, err)
}

func (s *Server) handleObject(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.store.Object(r.Context(), id)
	writeResult(w, v, err)
}

func (s *Server) handleSectors(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	v, err := s.store.Sectors(r.Context())
	writeResult(w, v, err)
}

func (s *Server) handleMap(w http.ResponseWriter, r *http.Request) {
	if !s.haveStore(w) {
		return
	}
	sector, ok := queryID(w, r, "sector")
	if !ok {
		return
	}
	if sector == nil {
		apiError(w, http.StatusBadRequest, "missing ?sector=")
		return
	}
	v, err := s.store.MapObjects(r.Context(), *sector)
	writeResult(w, v, err)
}

// haveStore reports whether the DB is available, answering 503 (JSON) if not so
// the SPA can show a "database offline" state instead of a hard failure.
func (s *Server) haveStore(w http.ResponseWriter) bool {
	if s.store == nil {
		apiError(w, http.StatusServiceUnavailable, "database unavailable")
		return false
	}
	return true
}

// pathID parses the {id} path value as a positive int64.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		apiError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

// queryID parses an optional numeric query param; ok=false means it was present
// but malformed (a 400 was already written). A missing param yields (nil, true).
func queryID(w http.ResponseWriter, r *http.Request, key string) (*int64, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		apiError(w, http.StatusBadRequest, "invalid "+key)
		return nil, false
	}
	return &id, true
}

// writeResult renders a query result: 404 for ErrNotFound, 500 for other errors
// (logged), 200 with the JSON body otherwise. A nil pointer value (single-row
// lookups) is treated as not-found for safety.
func writeResult(w http.ResponseWriter, v any, err error) {
	if errors.Is(err, db.ErrNotFound) {
		apiError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		slog.Error("admin api query failed", "err", err)
		apiError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if isNilPtr(v) {
		apiError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("admin api encode failed", "err", err)
	}
}

func apiError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
