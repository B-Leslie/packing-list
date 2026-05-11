package trips

import (
	"context"
	"database/sql"
)

type Pack struct{ db *sql.DB }

func NewPack(db *sql.DB) *Pack { return &Pack{db: db} }

func (p *Pack) Toggle(ctx context.Context, tripID, itemID string, packed bool) error {
	pInt := 0
	if packed {
		pInt = 1
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO trip_pack_state(trip_id,item_id,packed,packed_at)
		VALUES (?,?,?,CASE WHEN ?=1 THEN CURRENT_TIMESTAMP ELSE NULL END)
		ON CONFLICT(trip_id,item_id) DO UPDATE
		SET packed = excluded.packed, packed_at = excluded.packed_at`,
		tripID, itemID, pInt, pInt)
	return err
}

func (p *Pack) Progress(ctx context.Context, tripID string) (packedCount, totalCount int, err error) {
	// Note: total is computed at the renderer level (we need the merge logic).
	// Pack only exposes the packed count for callers who already know the total.
	err = p.db.QueryRowContext(ctx,
		`SELECT count(*) FROM trip_pack_state WHERE trip_id = ? AND packed = 1`, tripID).Scan(&packedCount)
	return
}
