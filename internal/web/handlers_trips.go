package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/trips"
)

func (s *Server) getTripsIndex(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	list, err := s.Trips.ListVisibleTo(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "trips_index", map[string]any{
		"Title": "Trips",
		"User":  uid,
		"Trips": list,
	})
}

func (s *Server) getTripNew(w http.ResponseWriter, r *http.Request) {
	s.Renderer.Render(w, "trip_new", map[string]any{
		"Title": "New trip",
		"User":  auth.UserFrom(r.Context()),
	})
}

func (s *Server) postTripCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	nights, _ := strconv.Atoi(r.FormValue("nights"))
	id, err := s.Trips.Create(r.Context(), name, nights, auth.UserFrom(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/trips/"+id, http.StatusSeeOther)
}

// requireMember ensures the caller is a member of the trip. Returns the role.
func (s *Server) requireMember(r *http.Request) (string, string, error) {
	tID := r.PathValue("id")
	uid := auth.UserFrom(r.Context())
	role, err := s.Trips.RoleOf(r.Context(), tID, uid)
	if err != nil {
		return tID, uid, err
	}
	return tID, role, nil
}

func (s *Server) getTripDetail(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	tr, err := s.Trips.Get(r.Context(), tID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rows, err := s.Renderer2.Render(r.Context(), tID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	packed := 0
	for _, row := range rows {
		if row.Packed {
			packed++
		}
	}
	attachedIDs, _ := s.Sources.AttachedBundleIDs(r.Context(), tID)
	attached := make([]struct{ ID, Name string }, 0, len(attachedIDs))
	for _, id := range attachedIDs {
		b, err := s.Bundles.Get(r.Context(), id)
		if err == nil && !b.DeletedAt.Valid {
			attached = append(attached, struct{ ID, Name string }{b.ID, b.Name})
		}
	}
	allItems, _ := s.Items.List(r.Context())
	allBundles, _ := s.Bundles.List(r.Context())
	members, _ := s.Trips.Members(r.Context(), tID)

	s.Renderer.Render(w, "trip_detail", map[string]any{
		"Title":            tr.Name,
		"User":             auth.UserFrom(r.Context()),
		"Trip":             tr,
		"Rows":             rows,
		"Progress":         struct{ Packed, Total int }{packed, len(rows)},
		"AttachedBundles":  attached,
		"AvailableBundles": allBundles,
		"AllItems":         allItems,
		"Members":          members,
	})
	// silence unused
	_ = errors.New
	_ = trips.ErrForbidden
}

func (s *Server) patchTrip(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || (role != "owner" && role != "editor") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	tr, err := s.Trips.Get(r.Context(), tID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		name = tr.Name
	}
	nights := tr.Nights
	if v := r.FormValue("nights"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			nights = n
		}
	}
	notes := r.FormValue("notes")
	if err := s.Trips.Update(r.Context(), tID, name, nights, notes); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteTrip(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Trips.SoftDelete(r.Context(), tID, auth.UserFrom(r.Context())); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/")
	}
	w.WriteHeader(http.StatusOK)
}
