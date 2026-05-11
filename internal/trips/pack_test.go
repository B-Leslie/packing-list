package trips

import (
	"context"
	"testing"
)

func TestPackToggle(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i", "X", "g", false, 1)
	h.addBundle("b", "b")
	h.bundleItem("b", "i", nil)
	h.attach(trip, "b")

	p := NewPack(h.render.db)
	ctx := context.Background()

	if err := p.Toggle(ctx, trip, "i", true); err != nil {
		t.Fatalf("toggle on: %v", err)
	}
	list, _ := h.render.Render(ctx, trip)
	if !list[0].Packed {
		t.Fatal("expected packed=true after toggle on")
	}

	if err := p.Toggle(ctx, trip, "i", false); err != nil {
		t.Fatalf("toggle off: %v", err)
	}
	list, _ = h.render.Render(ctx, trip)
	if list[0].Packed {
		t.Fatal("expected packed=false after toggle off")
	}
}
