package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestItemsCreateGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewItems(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, Item{Name: "Toothbrush", Category: "toiletries", PerNight: false, DefaultQty: 1, CreatedBy: "u_test"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Toothbrush" || got.Category != "toiletries" || got.DefaultQty != 1 {
		t.Errorf("unexpected: %+v", got)
	}
}

func TestItemsListExcludesSoftDeleted(t *testing.T) {
	db := newTestDB(t)
	repo := NewItems(db)
	ctx := context.Background()
	a, _ := repo.Create(ctx, Item{Name: "A", DefaultQty: 1, CreatedBy: "u_test"})
	b, _ := repo.Create(ctx, Item{Name: "B", DefaultQty: 1, CreatedBy: "u_test"})
	if err := repo.SoftDelete(ctx, a, "u_test"); err != nil {
		t.Fatalf("soft-delete: %v", err)
	}
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != b {
		t.Errorf("expected only B, got %+v", list)
	}
}

func TestItemsGetNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewItems(db)
	_, err := repo.Get(context.Background(), "nope")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestItemsUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewItems(db)
	ctx := context.Background()
	id, _ := repo.Create(ctx, Item{Name: "Old", DefaultQty: 1, CreatedBy: "u_test"})
	if err := repo.Update(ctx, id, Item{Name: "New", Category: "x", PerNight: true, DefaultQty: 3}); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := repo.Get(ctx, id)
	if got.Name != "New" || !got.PerNight || got.DefaultQty != 3 {
		t.Errorf("update did not apply: %+v", got)
	}
}

func TestItemsRestore(t *testing.T) {
	db := newTestDB(t)
	repo := NewItems(db)
	ctx := context.Background()
	id, _ := repo.Create(ctx, Item{Name: "X", DefaultQty: 1, CreatedBy: "u_test"})
	repo.SoftDelete(ctx, id, "u_test")
	if err := repo.Restore(ctx, id); err != nil {
		t.Fatalf("restore: %v", err)
	}
	list, _ := repo.List(ctx)
	if len(list) != 1 {
		t.Errorf("expected 1 active item after restore, got %d", len(list))
	}
}
