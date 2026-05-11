package trips

import (
	"context"
	"database/sql"
	"sort"
)

// Row is one rendered item destined for the UI.
type Row struct {
	ItemID   string
	Name     string
	Category string
	Qty      int
	PerNight bool
	Sources  []string // unique source bundle names plus possibly "extras"
	Packed   bool
}

type Renderer struct{ db *sql.DB }

func NewRenderer(db *sql.DB) *Renderer { return &Renderer{db: db} }

// Render returns the final packing list for tripID. Pure function over DB rows.
func (r *Renderer) Render(ctx context.Context, tripID string) ([]Row, error) {
	// 1. Trip metadata.
	var nights int
	if err := r.db.QueryRowContext(ctx, `SELECT nights FROM trips WHERE id = ?`, tripID).Scan(&nights); err != nil {
		return nil, err
	}

	// 2. Resolve attached bundles, expanded recursively.
	bundleIDs, bundleNames, err := r.expandAttachedBundles(ctx, tripID)
	if err != nil {
		return nil, err
	}

	type srcEntry struct {
		qty      int
		perNight bool
		source   string
	}
	srcs := map[string][]srcEntry{} // item_id -> entries
	itemMeta := map[string]struct {
		Name, Category string
		PerNight       bool
	}{}

	// 3. For each (non-deleted) bundle, gather its non-deleted items.
	if len(bundleIDs) > 0 {
		query := `
		  SELECT bi.bundle_id, bi.item_id, COALESCE(bi.qty, i.default_qty),
		         i.name, i.category, i.per_night
		  FROM bundle_items bi
		  JOIN items i ON i.id = bi.item_id
		  JOIN bundles b ON b.id = bi.bundle_id
		  WHERE bi.bundle_id IN (` + placeholders(len(bundleIDs)) + `)
		    AND b.deleted_at IS NULL
		    AND i.deleted_at IS NULL`
		args := make([]any, 0, len(bundleIDs))
		for _, id := range bundleIDs {
			args = append(args, id)
		}
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var bID, iID, name, cat string
			var qty int
			var per int
			if err := rows.Scan(&bID, &iID, &qty, &name, &cat, &per); err != nil {
				rows.Close()
				return nil, err
			}
			srcs[iID] = append(srcs[iID], srcEntry{qty: qty, perNight: per != 0, source: bundleNames[bID]})
			itemMeta[iID] = struct {
				Name, Category string
				PerNight       bool
			}{name, cat, per != 0}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// 4. Extras.
	{
		rows, err := r.db.QueryContext(ctx, `
		  SELECT te.item_id, COALESCE(te.qty, i.default_qty), i.name, i.category, i.per_night
		  FROM trip_extras te JOIN items i ON i.id = te.item_id
		  WHERE te.trip_id = ? AND i.deleted_at IS NULL`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID, name, cat string
			var qty, per int
			if err := rows.Scan(&iID, &qty, &name, &cat, &per); err != nil {
				rows.Close()
				return nil, err
			}
			srcs[iID] = append(srcs[iID], srcEntry{qty: qty, perNight: per != 0, source: "extras"})
			itemMeta[iID] = struct {
				Name, Category string
				PerNight       bool
			}{name, cat, per != 0}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// 5. Overrides.
	overrides := map[string]struct {
		removed     bool
		qtyOverride sql.NullInt64
	}{}
	{
		rows, err := r.db.QueryContext(ctx,
			`SELECT item_id, removed, qty_override FROM trip_overrides WHERE trip_id = ?`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID string
			var rem int
			var qo sql.NullInt64
			if err := rows.Scan(&iID, &rem, &qo); err != nil {
				rows.Close()
				return nil, err
			}
			overrides[iID] = struct {
				removed     bool
				qtyOverride sql.NullInt64
			}{rem != 0, qo}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// 6. Pack state.
	packed := map[string]bool{}
	{
		rows, err := r.db.QueryContext(ctx,
			`SELECT item_id, packed FROM trip_pack_state WHERE trip_id = ?`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID string
			var p int
			if err := rows.Scan(&iID, &p); err != nil {
				rows.Close()
				return nil, err
			}
			packed[iID] = p != 0
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// 7. Merge.
	out := make([]Row, 0, len(srcs))
	for iID, entries := range srcs {
		if ov, ok := overrides[iID]; ok && ov.removed {
			continue
		}
		meta := itemMeta[iID]
		row := Row{
			ItemID:   iID,
			Name:     meta.Name,
			Category: meta.Category,
			PerNight: meta.PerNight,
			Packed:   packed[iID],
		}

		// Sources de-duplicated, in stable order.
		seen := map[string]bool{}
		for _, e := range entries {
			if !seen[e.source] {
				seen[e.source] = true
				row.Sources = append(row.Sources, e.source)
			}
		}
		sort.Strings(row.Sources)

		// Quantity: max for fixed items, sum * nights for per-night items.
		if meta.PerNight {
			sum := 0
			for _, e := range entries {
				sum += e.qty
			}
			row.Qty = sum * nights
		} else {
			max := 0
			for _, e := range entries {
				if e.qty > max {
					max = e.qty
				}
			}
			row.Qty = max
		}

		// Apply qty override last.
		if ov, ok := overrides[iID]; ok && ov.qtyOverride.Valid {
			row.Qty = int(ov.qtyOverride.Int64)
		}

		out = append(out, row)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// expandAttachedBundles returns the set of bundle IDs that contribute to the
// trip (attached bundles + all nested descendants, excluding soft-deleted).
// Also returns a map of id -> name (for sourcing labels).
func (r *Renderer) expandAttachedBundles(ctx context.Context, tripID string) ([]string, map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.name
		FROM trip_bundles tb
		JOIN bundles b ON b.id = tb.bundle_id
		WHERE tb.trip_id = ? AND b.deleted_at IS NULL`, tripID)
	if err != nil {
		return nil, nil, err
	}
	var queue []string
	names := map[string]string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, nil, err
		}
		queue = append(queue, id)
		names[id] = name
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	seen := map[string]bool{}
	for _, id := range queue {
		seen[id] = true
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		rs, err := r.db.QueryContext(ctx, `
			SELECT b.id, b.name
			FROM bundle_children bc
			JOIN bundles b ON b.id = bc.child_id
			WHERE bc.parent_id = ? AND b.deleted_at IS NULL`, cur)
		if err != nil {
			return nil, nil, err
		}
		for rs.Next() {
			var id, name string
			if err := rs.Scan(&id, &name); err != nil {
				rs.Close()
				return nil, nil, err
			}
			if !seen[id] {
				seen[id] = true
				queue = append(queue, id)
				names[id] = name
			}
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return nil, nil, err
		}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids, names, nil
}

func placeholders(n int) string {
	if n == 0 {
		return ""
	}
	out := "?"
	for i := 1; i < n; i++ {
		out += ",?"
	}
	return out
}
