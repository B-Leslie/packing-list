package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/auth"
)

func authedRequest(t *testing.T, s *Server, method, target string, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	// Create user + session.
	uid, _, _ := s.Users.FindOrCreate(context.Background(), "u@example.com")
	cookieVal, _ := s.Sessions.Issue(context.Background(), uid)
	r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: cookieVal})
	r.AddCookie(&http.Cookie{Name: auth.CSRFCookieName, Value: "tok"})
	r.Header.Set(auth.CSRFHeaderName, "tok")
	return r
}

func TestItemsCreateAndDelete(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/items", "name=Toothbrush&category=toiletries&default_qty=1"))
	if w.Code != 200 {
		t.Fatalf("create: status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("expected row html, got %s", w.Body.String())
	}
	// Fetch list, confirm visible.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/items", ""))
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("items page missing item: %s", w.Body.String())
	}
}
