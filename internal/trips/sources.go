package trips

import (
	"context"
	"database/sql"
)

type Sources struct{ db *sql.DB }

func NewSources(db *sql.DB) *Sources { return &Sources{db: db} }

func (s *Sources) AttachBundle(ctx context.Context, tripID, bundleID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_bundles(trip_id,bundle_id) VALUES(?,?) ON CONFLICT DO NOTHING`,
		tripID, bundleID)
	return err
}

func (s *Sources) DetachBundle(ctx context.Context, tripID, bundleID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_bundles WHERE trip_id = ? AND bundle_id = ?`, tripID, bundleID)
	return err
}

func (s *Sources) AddExtra(ctx context.Context, tripID, itemID string, qty *int) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_extras(trip_id,item_id,qty) VALUES(?,?,?)
		 ON CONFLICT(trip_id,item_id) DO UPDATE SET qty = excluded.qty`,
		tripID, itemID, q)
	return err
}

func (s *Sources) RemoveExtra(ctx context.Context, tripID, itemID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_extras WHERE trip_id = ? AND item_id = ?`, tripID, itemID)
	return err
}

func (s *Sources) SetOverride(ctx context.Context, tripID, itemID string, removed bool, qty *int) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	r := 0
	if removed {
		r = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_overrides(trip_id,item_id,removed,qty_override) VALUES(?,?,?,?)
		 ON CONFLICT(trip_id,item_id) DO UPDATE SET removed = excluded.removed, qty_override = excluded.qty_override`,
		tripID, itemID, r, q)
	return err
}

func (s *Sources) ClearOverride(ctx context.Context, tripID, itemID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_overrides WHERE trip_id = ? AND item_id = ?`, tripID, itemID)
	return err
}

func (s *Sources) AttachedBundleIDs(ctx context.Context, tripID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT bundle_id FROM trip_bundles WHERE trip_id = ?`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
