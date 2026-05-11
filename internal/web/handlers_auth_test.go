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

func TestPostLoginCreatesUserAndShowsSentPage(t *testing.T) {
	s := newTestServer(t)
	form := strings.NewReader("email=alice@example.com&csrf_token=t")
	r := httptest.NewRequest("POST", "/login", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: auth.CSRFCookieName, Value: "t"})
	r.Header.Set(auth.CSRFHeaderName, "t")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	body := w.Body.String()
	if !strings.Contains(body, "Check your inbox") {
		t.Errorf("expected login_sent page, got: %s", body)
	}
	_, created, err := s.Users.FindOrCreate(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if created {
		t.Error("expected user already created by login")
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
