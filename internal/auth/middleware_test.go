package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireUserRedirectsWhenAnonymous(t *testing.T) {
	db := newTestDB(t)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = UserFrom(r.Context()) // would normally use this
		w.WriteHeader(http.StatusOK)
	})
	mw := RequireUser(s)(h)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestRequireUserAllowsWithValidCookie(t *testing.T) {
	db := newTestDB(t)
	db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })
	cookieVal, _ := s.Issue(context.Background(), "u_a")

	hit := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		if got := UserFrom(r.Context()); got != "u_a" {
			t.Errorf("ctx user: %q", got)
		}
	})
	mw := RequireUser(s)(h)
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: CookieName, Value: cookieVal})
	mw.ServeHTTP(httptest.NewRecorder(), r)
	if !hit {
		t.Error("expected handler to be called")
	}
}

func TestRequireUserReturns401ForHTMX(t *testing.T) {
	db := newTestDB(t)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	RequireUser(s)(h).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for HTMX, got %d", w.Code)
	}
	if w.Header().Get("HX-Redirect") != "/login" {
		t.Errorf("expected HX-Redirect, got %q", w.Header().Get("HX-Redirect"))
	}
}
