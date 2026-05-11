package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInviteCreatesUserAndMembership(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=T&nights=1"))
	tID := strings.TrimPrefix(w.Header().Get("Location"), "/trips/")

	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/members", "email=bob@example.com"))
	if w.Code != 200 {
		t.Fatalf("invite: %d (%s)", w.Code, w.Body.String())
	}
	_, created, _ := s.Users.FindOrCreate(context.Background(), "bob@example.com")
	if created {
		t.Error("expected bob to already exist after invite")
	}
}
