package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestTrashRestoresSoftDeletedItem(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	uid, _, _ := s.Users.FindOrCreate(ctx, "u@example.com")
	id, _ := s.Items.Create(ctx, catalog.Item{Name: "X", DefaultQty: 1, CreatedBy: uid})
	s.Items.SoftDelete(ctx, id, uid)

	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/trash", ""))
	if !strings.Contains(w.Body.String(), "X") {
		t.Fatalf("expected X in trash, got %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trash/item/"+id+"/restore", ""))
	if w.Code != 200 {
		t.Fatalf("restore: %d", w.Code)
	}
	list, _ := s.Items.List(ctx)
	if len(list) != 1 {
		t.Errorf("expected restored item back in list, got %d items", len(list))
	}
}
