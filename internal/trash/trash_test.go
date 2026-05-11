package trash

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/trips"
)

func TestListGathersAllDeleted(t *testing.T) {
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	if _, err := d.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`); err != nil {
		t.Fatal(err)
	}

	items := catalog.NewItems(d)
	bundles := catalog.NewBundles(d)
	trps := trips.NewTrips(d)
	ctx := context.Background()

	iID, _ := items.Create(ctx, catalog.Item{Name: "I", DefaultQty: 1, CreatedBy: "u_a"})
	bID, _ := bundles.Create(ctx, catalog.Bundle{Name: "B", CreatedBy: "u_a"})
	tID, _ := trps.Create(ctx, "T", 1, "u_a")
	items.SoftDelete(ctx, iID, "u_a")
	bundles.SoftDelete(ctx, bID, "u_a")
	trps.SoftDelete(ctx, tID, "u_a")

	v := NewView(d, items, bundles, trps)
	got, err := v.For(ctx, "u_a")
	if err != nil {
		t.Fatalf("for: %v", err)
	}
	if len(got.Items) != 1 || len(got.Bundles) != 1 || len(got.Trips) != 1 {
		t.Errorf("expected 1 of each, got %+v", got)
	}
}
