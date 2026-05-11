package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFGetSetsCookie(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	w := httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Fatal("handler should be called for GET")
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), CSRFCookieName+"=") {
		t.Errorf("expected csrf cookie set, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestCSRFPostRequiresHeaderMatch(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	// No cookie + no header -> 403.
	w := httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 without csrf, got %d", w.Code)
	}

	// Cookie + matching header -> ok.
	r := httptest.NewRequest("POST", "/", nil)
	r.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "tok"})
	r.Header.Set(CSRFHeaderName, "tok")
	w = httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid csrf, got %d", w.Code)
	}

	// Cookie + form value match -> ok.
	r = httptest.NewRequest("POST", "/", strings.NewReader("csrf_token=tok"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "tok"})
	w = httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with form csrf, got %d", w.Code)
	}
}
