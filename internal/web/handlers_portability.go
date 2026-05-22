package web

import (
	"encoding/json"
	"net/http"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/trips"
)

type exportDoc struct {
	Items   []catalog.Item   `json:"items"`
	Bundles []bundleExport   `json:"bundles"`
	Trips   []tripExport     `json:"trips"`
}

type bundleExport struct {
	catalog.Bundle
	Items    []catalog.BundleItem `json:"items"`
	Children []string             `json:"children"`
}

type tripExport struct {
	ID        string
	Name      string
	Nights    int
	Notes     string
	Bundles   []string
	Extras    []trips.Extra
	Overrides []trips.Override
}

func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	items, _ := s.Items.List(r.Context())
	bundles, _ := s.Bundles.List(r.Context())

	be := make([]bundleExport, 0, len(bundles))
	for _, b := range bundles {
		bi, _ := s.Bundles.Items(r.Context(), b.ID)
		ch, _ := s.Bundles.Children(r.Context(), b.ID)
		be = append(be, bundleExport{Bundle: b, Items: bi, Children: ch})
	}

	tripList, _ := s.Trips.ListVisibleTo(r.Context(), uid)
	tdocs := make([]tripExport, 0, len(tripList))
	for _, t := range tripList {
		bIDs, _ := s.Sources.AttachedBundleIDs(r.Context(), t.ID)
		extras, _ := s.Sources.QueryExtras(r.Context(), t.ID)
		overrides, _ := s.Sources.QueryOverrides(r.Context(), t.ID)
		tdocs = append(tdocs, tripExport{
			ID:        t.ID,
			Name:      t.Name,
			Nights:    t.Nights,
			Notes:     t.Notes,
			Bundles:   bIDs,
			Extras:    extras,
			Overrides: overrides,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="packing-list-export.json"`)
	json.NewEncoder(w).Encode(exportDoc{Items: items, Bundles: be, Trips: tdocs})
}

func (s *Server) importJSON(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5 MiB cap
	var doc exportDoc
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, "bad json: "+err.Error(), 400)
		return
	}
	// Upsert items + bundles. Skip if ID already exists (caller can purge first).
	for _, it := range doc.Items {
		// Reuse Create if not present.
		if _, err := s.Items.Get(r.Context(), it.ID); err != nil {
			it.CreatedBy = uid
			s.Items.Create(r.Context(), it)
		}
	}
	for _, b := range doc.Bundles {
		if _, err := s.Bundles.Get(r.Context(), b.ID); err != nil {
			s.Bundles.Create(r.Context(), b.Bundle)
		}
		for _, bi := range b.Items {
			var qp *float64
			if bi.Qty.Valid {
				v := bi.Qty.Float64
				qp = &v
			}
			s.Bundles.AddItem(r.Context(), b.ID, bi.ItemID, qp)
		}
		for _, child := range b.Children {
			s.Bundles.AddChild(r.Context(), b.ID, child)
		}
	}
	for _, t := range doc.Trips {
		// Trips: create as the caller (cannot import someone else's owner).
		tID, _ := s.Trips.Create(r.Context(), t.Name, t.Nights, uid)
		for _, bID := range t.Bundles {
			s.Sources.AttachBundle(r.Context(), tID, bID)
		}
		for _, e := range t.Extras {
			s.Sources.AddExtra(r.Context(), tID, e.ItemID, e.Qty)
		}
		for _, ov := range t.Overrides {
			s.Sources.SetOverride(r.Context(), tID, ov.ItemID, ov.Removed, ov.Qty)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
