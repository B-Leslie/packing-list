package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
)

type bundleItemView struct {
	BundleID string
	ItemID   string
	Name     string
	QtyShown float64 // 0 means "default"
}

type bundleChildView struct {
	ParentID string
	ChildID  string
	Name     string
}

func (s *Server) listBundles(w http.ResponseWriter, r *http.Request) {
	bs, err := s.Bundles.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "bundles", map[string]any{
		"Title":   "Bundles",
		"User":    auth.UserFrom(r.Context()),
		"Bundles": bs,
	})
}

func (s *Server) getBundleContents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bItems, _ := s.Bundles.Items(r.Context(), id)
	childIDs, _ := s.Bundles.Children(r.Context(), id)

	items, _ := s.Items.List(r.Context())
	itemNames := map[string]string{}
	for _, it := range items {
		itemNames[it.ID] = it.Name
	}

	rows := make([]bundleItemView, 0, len(bItems))
	for _, bi := range bItems {
		row := bundleItemView{BundleID: id, ItemID: bi.ItemID, Name: itemNames[bi.ItemID]}
		if bi.Qty.Valid {
			row.QtyShown = bi.Qty.Float64
		}
		rows = append(rows, row)
	}

	allBundles, _ := s.Bundles.List(r.Context())
	bundleNames := map[string]string{}
	for _, b := range allBundles {
		bundleNames[b.ID] = b.Name
	}

	childRows := make([]bundleChildView, 0, len(childIDs))
	for _, c := range childIDs {
		childRows = append(childRows, bundleChildView{ParentID: id, ChildID: c, Name: bundleNames[c]})
	}

	s.Renderer.Partial(w, "bundle_contents", map[string]any{
		"ID":       id,
		"Items":    rows,
		"Children": childRows,
	})
}

func (s *Server) createBundle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := s.Bundles.Create(r.Context(), catalog.Bundle{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		CreatedBy:   auth.UserFrom(r.Context()),
	})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	b, _ := s.Bundles.Get(r.Context(), id)
	s.Renderer.Partial(w, "bundle_row", b)
}

func (s *Server) editBundle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	b, err := s.Bundles.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items, _ := s.Items.List(r.Context())
	allBundles, _ := s.Bundles.List(r.Context())
	// Build current item rows with display names.
	itemNames := map[string]string{}
	for _, it := range items {
		itemNames[it.ID] = it.Name
	}
	bItems, _ := s.Bundles.Items(r.Context(), id)
	rows := make([]bundleItemView, 0, len(bItems))
	for _, bi := range bItems {
		row := bundleItemView{BundleID: id, ItemID: bi.ItemID, Name: itemNames[bi.ItemID]}
		if bi.Qty.Valid {
			row.QtyShown = bi.Qty.Float64
		}
		rows = append(rows, row)
	}
	// Children + nestable filter.
	childIDs, _ := s.Bundles.Children(r.Context(), id)
	bundleNames := map[string]string{}
	for _, b := range allBundles {
		bundleNames[b.ID] = b.Name
	}
	childRows := make([]bundleChildView, 0, len(childIDs))
	for _, c := range childIDs {
		childRows = append(childRows, bundleChildView{ParentID: id, ChildID: c, Name: bundleNames[c]})
	}
	// nestable = all bundles except self + transitive descendants (which would
	// create cycles). For listing, exclude self only — the AddChild call
	// performs the deeper cycle check.
	nestable := make([]catalog.Bundle, 0, len(allBundles))
	for _, b := range allBundles {
		if b.ID != id {
			nestable = append(nestable, b)
		}
	}
	s.Renderer.Render(w, "bundle_edit", map[string]any{
		"Title":           b.Name,
		"User":            auth.UserFrom(r.Context()),
		"Bundle":          b,
		"AllItems":        items,
		"Items":           rows,
		"Children":        childRows,
		"NestableBundles": nestable,
	})
}

func (s *Server) updateBundle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	r.ParseForm()
	if err := s.Bundles.Update(r.Context(), id, catalog.Bundle{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteBundle(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.SoftDelete(r.Context(), r.PathValue("id"), auth.UserFrom(r.Context())); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) bundleAddItem(w http.ResponseWriter, r *http.Request) {
	bID := r.PathValue("id")
	r.ParseForm()
	iID := r.FormValue("item_id")
	var qtyPtr *float64
	if v := r.FormValue("qty"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if err := s.Bundles.AddItem(r.Context(), bID, iID, qtyPtr); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	it, _ := s.Items.Get(r.Context(), iID)
	view := bundleItemView{BundleID: bID, ItemID: iID, Name: it.Name}
	if qtyPtr != nil {
		view.QtyShown = *qtyPtr
	}
	s.Renderer.Partial(w, "bundle_item_row", view)
}

func (s *Server) bundleRemoveItem(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.RemoveItem(r.Context(), r.PathValue("id"), r.PathValue("iid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) bundleAddChild(w http.ResponseWriter, r *http.Request) {
	parent := r.PathValue("id")
	r.ParseForm()
	child := r.FormValue("child_id")
	if err := s.Bundles.AddChild(r.Context(), parent, child); err != nil {
		if errors.Is(err, catalog.ErrConflict) {
			http.Error(w, "would create a cycle", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), 400)
		return
	}
	cb, _ := s.Bundles.Get(r.Context(), child)
	s.Renderer.Partial(w, "bundle_child_row", bundleChildView{ParentID: parent, ChildID: child, Name: cb.Name})
}

func (s *Server) bundleRemoveChild(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.RemoveChild(r.Context(), r.PathValue("id"), r.PathValue("cid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
