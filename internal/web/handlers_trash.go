package web

import (
	"net/http"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) getTrash(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	bin, err := s.Trash.For(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "trash", map[string]any{
		"Title": "Trash",
		"User":  uid,
		"Bin":   bin,
	})
}

func (s *Server) restore(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id := r.PathValue("id")
	var err error
	switch kind {
	case "item":
		err = s.Items.Restore(r.Context(), id)
	case "bundle":
		err = s.Bundles.Restore(r.Context(), id)
	case "trip":
		// only owner can restore.
		role, rerr := s.Trips.RoleOf(r.Context(), id, auth.UserFrom(r.Context()))
		if rerr != nil || role != "owner" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		err = s.Trips.Restore(r.Context(), id)
	default:
		http.Error(w, "bad kind", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) purge(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id := r.PathValue("id")
	var err error
	switch kind {
	case "item":
		err = s.Items.Purge(r.Context(), id)
	case "bundle":
		err = s.Bundles.Purge(r.Context(), id)
	case "trip":
		role, rerr := s.Trips.RoleOf(r.Context(), id, auth.UserFrom(r.Context()))
		if rerr != nil || role != "owner" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		err = s.Trips.Purge(r.Context(), id)
	default:
		http.Error(w, "bad kind", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
