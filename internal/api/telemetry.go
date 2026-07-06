package api

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// serverStats mirrors the running server's /stats JSON (package stats.Snapshot).
type serverStats struct {
	Objects  int     `json:"objects"`
	Sessions int64   `json:"sessions"`
	TickHz   float64 `json:"tick_hz"`
	UptimeS  float64 `json:"uptime_s"`
}

// telemetry is what /api/telemetry returns: the server stats wrapped in an
// "up" flag. When the server is unreachable, up=false and the fields are null —
// the Dashboard degrades to "—" instead of erroring.
type telemetry struct {
	Up       bool     `json:"up"`
	Objects  *int     `json:"objects"`
	Sessions *int64   `json:"sessions"`
	TickHz   *float64 `json:"tick_hz"`
	UptimeS  *float64 `json:"uptime_s"`
}

// handleTelemetry proxies the game server's /stats so the browser never has to
// reach it cross-origin. It always answers 200 with the wrapper — a down server
// is a normal state (up:false), not an API error.
func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	out := telemetry{Up: false}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, s.statsURL, nil)
	if err == nil {
		resp, err := s.httpc.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var st serverStats
			if resp.StatusCode == http.StatusOK && json.NewDecoder(resp.Body).Decode(&st) == nil {
				out = telemetry{Up: true, Objects: &st.Objects, Sessions: &st.Sessions, TickHz: &st.TickHz, UptimeS: &st.UptimeS}
			}
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// isNilPtr reports whether v is a typed nil pointer, so single-row lookups that
// hand back a nil result are rendered as 404 rather than a JSON "null" body.
func isNilPtr(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}
