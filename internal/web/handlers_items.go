package web

import (
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
)

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.Items.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "items", map[string]any{
		"Title": "Items",
		"User":  auth.UserFrom(r.Context()),
		"Items": items,
	})
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	qty, _ := strconv.ParseFloat(r.FormValue("default_qty"), 64)
	if qty <= 0 {
		qty = 1
	}
	it := catalog.Item{
		Name:       r.FormValue("name"),
		Category:   r.FormValue("category"),
		PerNight:   r.FormValue("per_night") == "1",
		DefaultQty: qty,
		CreatedBy:  auth.UserFrom(r.Context()),
	}
	id, err := s.Items.Create(r.Context(), it)
	if err != nil {
		http.Error(w, "could not create", 400)
		return
	}
	created, _ := s.Items.Get(r.Context(), id)
	s.Renderer.Partial(w, "item_row", created)
}

func (s *Server) editItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	it, err := s.Items.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	s.Renderer.Partial(w, "item_row_edit", it)
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	r.ParseForm()
	qty, _ := strconv.ParseFloat(r.FormValue("default_qty"), 64)
	if qty <= 0 {
		qty = 1
	}
	if err := s.Items.Update(r.Context(), id, catalog.Item{
		Name:       r.FormValue("name"),
		Category:   r.FormValue("category"),
		PerNight:   r.FormValue("per_night") == "1",
		DefaultQty: qty,
	}); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	updated, _ := s.Items.Get(r.Context(), id)
	s.Renderer.Partial(w, "item_row", updated)
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Items.SoftDelete(r.Context(), id, auth.UserFrom(r.Context())); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK) // empty body → HTMX removes the row
}
