package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	users := auth.NewUsers(d)
	now := func() time.Time { return time.Now() }
	mailer := &auth.LogMailer{Logger: logger}
	cfg := config.Config{BaseURL: "http://test"}
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	s := &Server{
		Cfg:       cfg,
		Logger:    logger,
		Renderer:  r,
		Users:     users,
		Sessions:  auth.NewSessions(d, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), now),
		Magic:     auth.NewMagic(d, mailer, cfg.BaseURL, now),
		RateLimit: auth.NewRateLimiter(10, time.Minute, now),
		Items:     catalog.NewItems(d),
		Bundles:   catalog.NewBundles(d),
		Trips:     trips.NewTrips(d),
		Sources:   trips.NewSources(d),
		Pack:      trips.NewPack(d),
		Renderer2: trips.NewRenderer(d),
		IsDev:     true,
		Now:       now,
	}
	s.Trash = trash.NewView(d, s.Items, s.Bundles, s.Trips)
	return s
}

func postLoginRequest(email string) *http.Request {
	form := strings.NewReader("email=" + email + "&csrf_token=t")
	r := httptest.NewRequest("POST", "/login", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: auth.CSRFCookieName, Value: "t"})
	r.Header.Set(auth.CSRFHeaderName, "t")
	return r
}

func TestPostLoginExistingUserShowsSentPage(t *testing.T) {
	s := newTestServer(t)
	if _, _, err := s.Users.FindOrCreate(context.Background(), "alice@example.com"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, postLoginRequest("alice@example.com"))
	body := w.Body.String()
	if !strings.Contains(body, "Check your inbox") {
		t.Errorf("expected login_sent page, got: %s", body)
	}
}

// Anti-enumeration: response for an unknown email must be
// indistinguishable from the existing-user response.
func TestPostLoginUnknownEmailStillShowsSentPage(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, postLoginRequest("stranger@example.com"))
	body := w.Body.String()
	if !strings.Contains(body, "Check your inbox") {
		t.Errorf("expected login_sent page for unknown email, got: %s", body)
	}
	if _, found, err := s.Users.Find(context.Background(), "stranger@example.com"); err != nil || found {
		t.Errorf("unknown email must not be auto-created: found=%v err=%v", found, err)
	}
}

func TestGetLoginRendersForm(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/login", nil))
	if !strings.Contains(w.Body.String(), "Send magic link") {
		t.Errorf("login page: %s", w.Body.String())
	}
}
