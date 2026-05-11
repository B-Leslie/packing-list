package web

import (
	"net/http"
	"strconv"
)

func (s *Server) attachBundle(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	bID := r.FormValue("bundle_id")
	if err := s.Sources.AttachBundle(r.Context(), tID, bID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) detachBundle(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Sources.DetachBundle(r.Context(), tID, r.PathValue("bid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) addExtra(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	iID := r.FormValue("item_id")
	var qtyPtr *int
	if v := r.FormValue("qty"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if err := s.Sources.AddExtra(r.Context(), tID, iID, qtyPtr); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) overrideItem(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	iID := r.PathValue("iid")
	removed := r.FormValue("removed") == "1"
	var qtyPtr *int
	if v := r.FormValue("qty"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if !removed && qtyPtr == nil {
		// "clear override" form.
		if err := s.Sources.ClearOverride(r.Context(), tID, iID); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	} else {
		if err := s.Sources.SetOverride(r.Context(), tID, iID, removed, qtyPtr); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
	// Empty body — caller refreshes trip detail.
	w.WriteHeader(http.StatusOK)
}

func (s *Server) togglePack(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	iID := r.PathValue("iid")
	r.ParseForm()
	// Checkbox semantics: presence of "packed" key = true.
	packed := r.FormValue("packed") == "1"
	// If the cookie pre-existed (HTMX checkbox toggling), we flip based on current value.
	// Simpler: trust the form value.
	if err := s.Pack.Toggle(r.Context(), tID, iID, packed); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Re-render the row + progress.
	rows, err := s.Renderer2.Render(r.Context(), tID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var row *struct {
		TripID string
		Row    any
	}
	for i := range rows {
		if rows[i].ItemID == iID {
			row = &struct {
				TripID string
				Row    any
			}{TripID: tID, Row: rows[i]}
			break
		}
	}
	if row == nil {
		// Item no longer in list (unlikely). Send empty 200.
		w.WriteHeader(http.StatusOK)
		return
	}
	// Also push progress refresh via HX-Trigger so progress bar updates.
	packedCount := 0
	for _, x := range rows {
		if x.Packed {
			packedCount++
		}
	}
	w.Header().Set("HX-Trigger-After-Swap",
		`{"refreshProgress":{"packed":`+strconv.Itoa(packedCount)+`,"total":`+strconv.Itoa(len(rows))+`}}`)
	s.Renderer.Partial(w, "pack_row", row)
}
