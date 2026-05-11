package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestBundlesCreateAddItem(t *testing.T) {
	db := newTestDB(t)
	items := NewItems(db)
	bundles := NewBundles(db)
	ctx := context.Background()

	tb, _ := items.Create(ctx, Item{Name: "Toothbrush", DefaultQty: 1, CreatedBy: "u_test"})
	bid, err := bundles.Create(ctx, Bundle{Name: "washbag-basic", CreatedBy: "u_test"})
	if err != nil {
		t.Fatalf("create bundle: %v", err)
	}
	if err := bundles.AddItem(ctx, bid, tb, nil); err != nil {
		t.Fatalf("add item: %v", err)
	}
	got, err := bundles.Items(ctx, bid)
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	if len(got) != 1 || got[0].ItemID != tb {
		t.Errorf("expected one item ref, got %+v", got)
	}
}

func TestBundlesNestChildPreventsCycle(t *testing.T) {
	db := newTestDB(t)
	bundles := NewBundles(db)
	ctx := context.Background()

	a, _ := bundles.Create(ctx, Bundle{Name: "A", CreatedBy: "u_test"})
	b, _ := bundles.Create(ctx, Bundle{Name: "B", CreatedBy: "u_test"})
	c, _ := bundles.Create(ctx, Bundle{Name: "C", CreatedBy: "u_test"})

	if err := bundles.AddChild(ctx, a, b); err != nil {
		t.Fatalf("a->b: %v", err)
	}
	if err := bundles.AddChild(ctx, b, c); err != nil {
		t.Fatalf("b->c: %v", err)
	}
	// c -> a would create cycle a -> b -> c -> a.
	err := bundles.AddChild(ctx, c, a)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	// self-loop guarded by CHECK.
	err = bundles.AddChild(ctx, a, a)
	if err == nil {
		t.Fatal("expected error on self-nest, got nil")
	}
}

func TestBundlesListChildren(t *testing.T) {
	db := newTestDB(t)
	bundles := NewBundles(db)
	ctx := context.Background()
	a, _ := bundles.Create(ctx, Bundle{Name: "A", CreatedBy: "u_test"})
	b, _ := bundles.Create(ctx, Bundle{Name: "B", CreatedBy: "u_test"})
	bundles.AddChild(ctx, a, b)
	got, _ := bundles.Children(ctx, a)
	if len(got) != 1 || got[0] != b {
		t.Errorf("expected [b], got %v", got)
	}
}

func TestBundlesRemoveItemAndChild(t *testing.T) {
	db := newTestDB(t)
	items := NewItems(db)
	bundles := NewBundles(db)
	ctx := context.Background()
	tb, _ := items.Create(ctx, Item{Name: "Toothbrush", DefaultQty: 1, CreatedBy: "u_test"})
	bid, _ := bundles.Create(ctx, Bundle{Name: "wash", CreatedBy: "u_test"})
	bundles.AddItem(ctx, bid, tb, nil)
	if err := bundles.RemoveItem(ctx, bid, tb); err != nil {
		t.Fatalf("remove item: %v", err)
	}
	got, _ := bundles.Items(ctx, bid)
	if len(got) != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}
