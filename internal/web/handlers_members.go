package web

import (
	"net/http"
	"strings"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) inviteMember(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		http.Error(w, "email required", 400)
		return
	}
	if !s.RateLimit.Allow("invite:" + auth.UserFrom(r.Context())) {
		http.Error(w, "too many invites; try again later", 429)
		return
	}
	uid, created, err := s.Users.FindOrCreate(r.Context(), email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := s.Trips.AddMember(r.Context(), tID, uid, "editor"); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// If newly created, also send a magic link so they can sign in.
	if created {
		if err := s.Magic.Issue(r.Context(), email); err != nil {
			s.Logger.Warn("invite magic link", "err", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Trips.RemoveMember(r.Context(), tID, r.PathValue("uid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
