package web

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExportRoundTrip(t *testing.T) {
	s := newTestServer(t)
	// Seed via API.
	s.Handler().ServeHTTP(httptest.NewRecorder(), authedRequest(t, s, "POST", "/items", "name=Toothbrush&default_qty=1"))
	// Export.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/export", ""))
	if w.Code != 200 {
		t.Fatalf("export: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("export missing toothbrush: %s", w.Body.String())
	}
	dump := bytes.TrimSpace(w.Body.Bytes())
	if !bytes.HasPrefix(dump, []byte("{")) {
		t.Errorf("expected json, got %s", w.Body.String())
	}
}
