package auth

import (
	"context"
	"testing"
)

func TestFindOrCreateUserIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	u := NewUsers(db)
	ctx := context.Background()

	id1, created1, err := u.FindOrCreate(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !created1 {
		t.Error("expected created=true on first call")
	}

	id2, created2, err := u.FindOrCreate(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if created2 {
		t.Error("expected created=false on second call")
	}
	if id1 != id2 {
		t.Errorf("expected same id, got %s vs %s", id1, id2)
	}
}

func TestFindOrCreateCaseInsensitive(t *testing.T) {
	db := newTestDB(t)
	u := NewUsers(db)
	ctx := context.Background()
	id1, _, _ := u.FindOrCreate(ctx, "Alice@Example.com")
	id2, _, _ := u.FindOrCreate(ctx, "alice@example.com")
	if id1 != id2 {
		t.Errorf("expected case-insensitive match, got %s vs %s", id1, id2)
	}
}
