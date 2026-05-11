package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestAttachBundleThenPackToggle(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	uid, _, _ := s.Users.FindOrCreate(ctx, "u@example.com")

	// Seed catalog.
	itID, _ := s.Items.Create(ctx, catalog.Item{Name: "Toothbrush", DefaultQty: 1, CreatedBy: uid})
	bID, _ := s.Bundles.Create(ctx, catalog.Bundle{Name: "wash", CreatedBy: uid})
	s.Bundles.AddItem(ctx, bID, itID, nil)

	// Create trip via HTTP.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=T&nights=1"))
	loc := w.Header().Get("Location")
	tID := strings.TrimPrefix(loc, "/trips/")

	// Attach bundle.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/bundles", "bundle_id="+bID))
	if w.Code != 200 {
		t.Fatalf("attach: %d", w.Code)
	}

	// Pack the item.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/pack/"+itID, "packed=1"))
	if w.Code != 200 {
		t.Fatalf("pack: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "packed") {
		t.Errorf("expected packed class in row, got %s", w.Body.String())
	}
}
