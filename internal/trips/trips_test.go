package trips

import (
	"context"
	"errors"
	"testing"
)

func TestTripCreateAddsOwnerAsMember(t *testing.T) {
	db := newTestDB(t)
	repo := NewTrips(db)
	ctx := context.Background()
	id, err := repo.Create(ctx, "Weekend Devon", 2, "u_a")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	members, err := repo.Members(ctx, id)
	if err != nil {
		t.Fatalf("members: %v", err)
	}
	if len(members) != 1 || members[0].UserID != "u_a" || members[0].Role != "owner" {
		t.Errorf("expected owner u_a, got %+v", members)
	}
}

func TestTripVisibleTo(t *testing.T) {
	db := newTestDB(t)
	repo := NewTrips(db)
	ctx := context.Background()
	id, _ := repo.Create(ctx, "T", 1, "u_a")
	got, _ := repo.ListVisibleTo(ctx, "u_a")
	if len(got) != 1 || got[0].ID != id {
		t.Errorf("a should see trip, got %+v", got)
	}
	got, _ = repo.ListVisibleTo(ctx, "u_b")
	if len(got) != 0 {
		t.Errorf("b should not see trip yet, got %+v", got)
	}
	if err := repo.AddMember(ctx, id, "u_b", "editor"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	got, _ = repo.ListVisibleTo(ctx, "u_b")
	if len(got) != 1 {
		t.Errorf("b should see trip after invite, got %+v", got)
	}
}

func TestTripRoleEnforced(t *testing.T) {
	db := newTestDB(t)
	repo := NewTrips(db)
	ctx := context.Background()
	id, _ := repo.Create(ctx, "T", 1, "u_a")
	r, err := repo.RoleOf(ctx, id, "u_a")
	if err != nil {
		t.Fatalf("role: %v", err)
	}
	if r != "owner" {
		t.Errorf("a is owner, got %q", r)
	}
	_, err = repo.RoleOf(ctx, id, "u_b")
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("expected ErrNotMember, got %v", err)
	}
}

func TestTripUpdateAndSoftDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewTrips(db)
	ctx := context.Background()
	id, _ := repo.Create(ctx, "T", 1, "u_a")
	if err := repo.Update(ctx, id, "T2", 5, "notes"); err != nil {
		t.Fatalf("update: %v", err)
	}
	tr, _ := repo.Get(ctx, id)
	if tr.Name != "T2" || tr.Nights != 5 {
		t.Errorf("update did not apply: %+v", tr)
	}
	if err := repo.SoftDelete(ctx, id, "u_a"); err != nil {
		t.Fatalf("soft-delete: %v", err)
	}
	got, _ := repo.ListVisibleTo(ctx, "u_a")
	if len(got) != 0 {
		t.Errorf("expected no trips after soft-delete, got %d", len(got))
	}
}
