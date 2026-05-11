package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTripCreateAndDetail(t *testing.T) {
	s := newTestServer(t)
	// Create the trip.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=Weekend+Devon&nights=2"))
	if w.Code != 303 {
		t.Fatalf("create: status %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "/trips/") {
		t.Fatalf("redirect location: %q", loc)
	}
	// Detail page renders.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", loc, ""))
	if w.Code != 200 {
		t.Fatalf("detail: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Weekend Devon") {
		t.Errorf("expected trip name in body, got %s", w.Body.String())
	}
}
