package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestBundleCreateAndCycleRejected(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	uid, _, _ := s.Users.FindOrCreate(ctx, "u@example.com")

	// Create A, B; nest A->B; try B->A (cycle).
	aID, _ := s.Bundles.Create(ctx, catalog.Bundle{Name: "A", CreatedBy: uid})
	bID, _ := s.Bundles.Create(ctx, catalog.Bundle{Name: "B", CreatedBy: uid})

	// Nest A -> B via HTTP.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/bundles/"+aID+"/children", "child_id="+bID))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "B") {
		t.Fatalf("nest A->B: %d %s", w.Code, w.Body.String())
	}

	// Now B -> A should be 409 (cycle).
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/bundles/"+bID+"/children", "child_id="+aID))
	if w.Code != 409 {
		t.Fatalf("expected 409 on cycle, got %d (%s)", w.Code, w.Body.String())
	}
}
