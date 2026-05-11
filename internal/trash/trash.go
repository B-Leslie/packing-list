// Package trash aggregates soft-deleted rows for the /trash view.
package trash

import (
	"context"
	"database/sql"

	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/trips"
)

type Bin struct {
	Items   []catalog.Item
	Bundles []catalog.Bundle
	Trips   []trips.Trip
}

type View struct {
	db      *sql.DB
	items   *catalog.Items
	bundles *catalog.Bundles
	trips   *trips.Trips
}

func NewView(db *sql.DB, i *catalog.Items, b *catalog.Bundles, t *trips.Trips) *View {
	return &View{db, i, b, t}
}

func (v *View) For(ctx context.Context, userID string) (Bin, error) {
	var b Bin
	var err error
	if b.Items, err = v.items.ListDeleted(ctx); err != nil {
		return b, err
	}
	if b.Bundles, err = v.bundles.ListDeleted(ctx); err != nil {
		return b, err
	}
	if b.Trips, err = v.trips.ListDeleted(ctx, userID); err != nil {
		return b, err
	}
	return b, nil
}
