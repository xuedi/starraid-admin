package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/xuedi/starraid-admin/internal/db"
)

// fakeStore is a hand-rolled Store for handler tests — no DB in CI.
type fakeStore struct {
	users     []db.UserRow
	user      *db.UserDetail
	objects   []db.ObjectRow
	object    *db.ObjectDetail
	mapObjs   []db.MapObject
	lastMapID int64
	cred      *db.AccountCred
}

func (f *fakeStore) Summary(context.Context) (db.Summary, error) {
	return db.Summary{Accounts: 1, Characters: 2, Objects: 8, Sectors: 1,
		ByClass: []db.ClassCount{{Key: "skiff", Name: "Skiff", Count: 1}}}, nil
}
func (f *fakeStore) Users(context.Context) ([]db.UserRow, error) { return f.users, nil }
func (f *fakeStore) User(_ context.Context, id int64) (*db.UserDetail, error) {
	if f.user == nil || f.user.ID != id {
		return nil, db.ErrNotFound
	}
	return f.user, nil
}
func (f *fakeStore) Objects(_ context.Context, _ *int64) ([]db.ObjectRow, error) {
	return f.objects, nil
}
func (f *fakeStore) Object(_ context.Context, id int64) (*db.ObjectDetail, error) {
	if f.object == nil || f.object.ID != id {
		return nil, db.ErrNotFound
	}
	return f.object, nil
}
func (f *fakeStore) Sectors(context.Context) ([]db.Sector, error) {
	return []db.Sector{{ID: 1, Name: "Starting Area", Objects: 8}}, nil
}
func (f *fakeStore) MapObjects(_ context.Context, id int64) ([]db.MapObject, error) {
	f.lastMapID = id
	return f.mapObjs, nil
}
func (f *fakeStore) CredByEmail(_ context.Context, email string) (*db.AccountCred, error) {
	if f.cred == nil || f.cred.Email != email {
		return nil, db.ErrNotFound
	}
	return f.cred, nil
}

func newTestServer(store Store, statsURL string) http.Handler {
	mux := http.NewServeMux()
	New(store, statsURL).Mount(mux)
	return mux
}

func do(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	return rec
}

// credStore builds a fake whose one account has a real bcrypt hash of the
// password, so login tests exercise the actual compare path.
func credStore(t *testing.T, email, password string) *fakeStore {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return &fakeStore{cred: &db.AccountCred{ID: 1, Email: email, PasswordHash: string(hash)}}
}

func TestSummaryShape(t *testing.T) {
	h := newTestServer(&fakeStore{}, "")
	rec := do(t, h, "/api/summary")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var s db.Summary
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Accounts != 1 || s.Objects != 8 || len(s.ByClass) != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}

func TestUserFound(t *testing.T) {
	h := newTestServer(&fakeStore{user: &db.UserDetail{ID: 7, Email: "a@b.c"}}, "")
	rec := do(t, h, "/api/users/7")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var u db.UserDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if u.ID != 7 || u.Email != "a@b.c" {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestUserNotFound(t *testing.T) {
	h := newTestServer(&fakeStore{}, "")
	if rec := do(t, h, "/api/users/999"); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestUserBadID(t *testing.T) {
	h := newTestServer(&fakeStore{}, "")
	if rec := do(t, h, "/api/users/abc"); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestObjectNotFound(t *testing.T) {
	h := newTestServer(&fakeStore{}, "")
	if rec := do(t, h, "/api/objects/42"); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestMapRequiresSector(t *testing.T) {
	h := newTestServer(&fakeStore{}, "")
	if rec := do(t, h, "/api/map"); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (missing sector)", rec.Code)
	}
}

func TestMapPassesSector(t *testing.T) {
	f := &fakeStore{mapObjs: []db.MapObject{{ID: 1, Kind: "ship"}}}
	h := newTestServer(f, "")
	rec := do(t, h, "/api/map?sector=3")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if f.lastMapID != 3 {
		t.Fatalf("sector id = %d, want 3", f.lastMapID)
	}
}

func TestDBDownReturns503(t *testing.T) {
	h := newTestServer(nil, "") // nil store == DB unavailable
	for _, p := range []string{"/api/summary", "/api/users", "/api/objects", "/api/sectors"} {
		if rec := do(t, h, p); rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status = %d, want 503", p, rec.Code)
		}
	}
}

func TestLoginOK(t *testing.T) {
	h := newTestServer(credStore(t, "test@example.org", "1234"), "")
	rec := postJSON(t, h, "/api/login", `{"email":"test@example.org","password":"1234"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Ok    bool   `json:"ok"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Ok || body.Email != "test@example.org" {
		t.Fatalf("unexpected login body: %+v", body)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h := newTestServer(credStore(t, "test@example.org", "1234"), "")
	if rec := postJSON(t, h, "/api/login", `{"email":"test@example.org","password":"nope"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	h := newTestServer(credStore(t, "test@example.org", "1234"), "")
	if rec := postJSON(t, h, "/api/login", `{"email":"ghost@example.org","password":"1234"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLoginDBDown(t *testing.T) {
	h := newTestServer(nil, "")
	if rec := postJSON(t, h, "/api/login", `{"email":"test@example.org","password":"1234"}`); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestTelemetryUp(t *testing.T) {
	stats := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"objects":5,"sessions":2,"tick_hz":10,"uptime_s":42}`))
	}))
	defer stats.Close()

	h := newTestServer(&fakeStore{}, stats.URL)
	rec := do(t, h, "/api/telemetry")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var tm telemetry
	if err := json.Unmarshal(rec.Body.Bytes(), &tm); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !tm.Up || tm.Objects == nil || *tm.Objects != 5 || tm.Sessions == nil || *tm.Sessions != 2 {
		t.Fatalf("unexpected telemetry: %+v", tm)
	}
}

func TestTelemetryDown(t *testing.T) {
	// An unroutable URL → the proxy fails and reports up:false with null fields.
	h := newTestServer(&fakeStore{}, "http://127.0.0.1:1/stats")
	rec := do(t, h, "/api/telemetry")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (down is a normal state)", rec.Code)
	}
	var tm telemetry
	if err := json.Unmarshal(rec.Body.Bytes(), &tm); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tm.Up || tm.Objects != nil {
		t.Fatalf("expected up:false with null fields, got %+v", tm)
	}
}
