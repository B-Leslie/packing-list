package trips

import (
	"context"
	"database/sql"
	"sort"
	"testing"
)

// helpers ---------------------------------------------------------------------

func floatp(v float64) *float64 { return &v }

type fixture struct {
	t    *testing.T
	db   interface{ /* not used */ }
	trip string
	srcs *Sources
}

// renderFixture seeds: user u_a, trip "T" with given nights, plus convenience
// builders for items, bundles, and bundle composition.
func renderFixture(t *testing.T, nights int) (trip string, h *renderHelper) {
	db := newTestDB(t)
	trepo := NewTrips(db)
	id, err := trepo.Create(context.Background(), "T", nights, "u_a")
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}
	return id, &renderHelper{t: t, db: db, render: NewRenderer(db)}
}

type renderHelper struct {
	t      *testing.T
	db     anyDB
	render *Renderer
}

// anyDB is a type alias used for the db field on renderHelper.
// The plan used an anonymous interface; we use *sql.DB directly because
// Go's *sql.DB.Exec returns the concrete sql.Result interface, not an
// inline structural type. execDB calls h.render.db.Exec, not h.db.Exec,
// so this field is effectively just a holder. Deviation noted.
type anyDB = interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// (We use direct SQL here for terseness in tests rather than going through
// the catalog package — render is the unit under test, not catalog.)

// addItem inserts an item directly.
func (h *renderHelper) addItem(id, name, cat string, perNight bool, defQty float64) {
	h.t.Helper()
	per := 0
	if perNight {
		per = 1
	}
	if _, err := execDB(h, `INSERT INTO items(id,name,category,per_night,default_qty,created_by) VALUES (?,?,?,?,?,'u_a')`,
		id, name, cat, per, defQty); err != nil {
		h.t.Fatalf("addItem: %v", err)
	}
}

func (h *renderHelper) softDeleteItem(id string) {
	h.t.Helper()
	if _, err := execDB(h, `UPDATE items SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		h.t.Fatalf("softDeleteItem: %v", err)
	}
}

func (h *renderHelper) addBundle(id, name string) {
	h.t.Helper()
	if _, err := execDB(h, `INSERT INTO bundles(id,name,created_by) VALUES (?,?,'u_a')`, id, name); err != nil {
		h.t.Fatalf("addBundle: %v", err)
	}
}

func (h *renderHelper) softDeleteBundle(id string) {
	h.t.Helper()
	if _, err := execDB(h, `UPDATE bundles SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?`, id); err != nil {
		h.t.Fatalf("softDeleteBundle: %v", err)
	}
}

func (h *renderHelper) bundleItem(bundle, item string, qty *float64) {
	h.t.Helper()
	var q any
	if qty != nil {
		q = *qty
	}
	if _, err := execDB(h, `INSERT INTO bundle_items(bundle_id,item_id,qty) VALUES (?,?,?)`, bundle, item, q); err != nil {
		h.t.Fatalf("bundleItem: %v", err)
	}
}

func (h *renderHelper) nest(parent, child string) {
	h.t.Helper()
	if _, err := execDB(h, `INSERT INTO bundle_children(parent_id,child_id) VALUES (?,?)`, parent, child); err != nil {
		h.t.Fatalf("nest: %v", err)
	}
}

func (h *renderHelper) attach(trip, bundle string) {
	h.t.Helper()
	if _, err := execDB(h, `INSERT INTO trip_bundles(trip_id,bundle_id) VALUES (?,?)`, trip, bundle); err != nil {
		h.t.Fatalf("attach: %v", err)
	}
}

// indirect access to the underlying *sql.DB — Renderer keeps it private.
func execDB(h *renderHelper, q string, args ...any) (anyResult, error) {
	return h.render.db.Exec(q, args...)
}

type anyResult = interface {
	RowsAffected() (int64, error)
	LastInsertId() (int64, error)
}

// tests -----------------------------------------------------------------------

func TestRenderSimpleFixedItem(t *testing.T) {
	trip, h := renderFixture(t, 3)
	h.addItem("i_tb", "Toothbrush", "toiletries", false, 1)
	h.addBundle("b_wash", "washbag-basic")
	h.bundleItem("b_wash", "i_tb", nil)
	h.attach(trip, "b_wash")

	list, err := h.render.Render(context.Background(), trip)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if len(list) != 1 || list[0].ItemID != "i_tb" || list[0].Qty != 1 {
		t.Fatalf("unexpected list: %+v", list)
	}
	if list[0].Sources[0] != "washbag-basic" {
		t.Errorf("source: got %v", list[0].Sources)
	}
}

func TestRenderPerNightScalesByNights(t *testing.T) {
	trip, h := renderFixture(t, 4)
	h.addItem("i_u", "Underwear", "clothing", true, 1)
	h.addBundle("b_c", "clothes")
	h.bundleItem("b_c", "i_u", nil)
	h.attach(trip, "b_c")

	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].Qty != 4 {
		t.Fatalf("expected qty=4, got %+v", list)
	}
}

func TestRenderFixedItemTakesMaxAcrossBundles(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_tw", "Towel", "general", false, 1)
	h.addBundle("b_a", "a")
	h.addBundle("b_b", "b")
	h.bundleItem("b_a", "i_tw", floatp(1))
	h.bundleItem("b_b", "i_tw", floatp(2)) // bigger qty wins for fixed items
	h.attach(trip, "b_a")
	h.attach(trip, "b_b")

	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].Qty != 2 {
		t.Fatalf("expected qty=2 (max), got %+v", list)
	}
}

func TestRenderPerNightSumsAcrossBundles(t *testing.T) {
	trip, h := renderFixture(t, 2)
	h.addItem("i_s", "Socks", "clothing", true, 1)
	h.addBundle("b_a", "a")
	h.addBundle("b_b", "b")
	h.bundleItem("b_a", "i_s", floatp(1)) // 1 per night
	h.bundleItem("b_b", "i_s", floatp(1)) // 1 per night
	h.attach(trip, "b_a")
	h.attach(trip, "b_b")

	list, _ := h.render.Render(context.Background(), trip)
	// (1 + 1) per_night * 2 nights = 4
	if len(list) != 1 || list[0].Qty != 4 {
		t.Fatalf("expected qty=4, got %+v", list)
	}
}

func TestRenderOverrideRemovesItem(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_tb", "Toothbrush", "t", false, 1)
	h.addBundle("b_w", "w")
	h.bundleItem("b_w", "i_tb", nil)
	h.attach(trip, "b_w")

	srcs := NewSources(h.render.db)
	if err := srcs.SetOverride(context.Background(), trip, "i_tb", true, nil); err != nil {
		t.Fatalf("override: %v", err)
	}
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %+v", list)
	}
}

func TestRenderOverrideForcesQty(t *testing.T) {
	trip, h := renderFixture(t, 5)
	h.addItem("i_u", "Underwear", "c", true, 1)
	h.addBundle("b", "b")
	h.bundleItem("b", "i_u", nil)
	h.attach(trip, "b")

	srcs := NewSources(h.render.db)
	srcs.SetOverride(context.Background(), trip, "i_u", false, floatp(2)) // force qty=2 regardless of nights

	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].Qty != 2 {
		t.Fatalf("expected qty=2 from override, got %+v", list)
	}
}

func TestRenderIncludesExtras(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_book", "Book", "leisure", false, 1)
	srcs := NewSources(h.render.db)
	srcs.AddExtra(context.Background(), trip, "i_book", nil)

	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].ItemID != "i_book" {
		t.Fatalf("unexpected: %+v", list)
	}
	if list[0].Sources[0] != "extras" {
		t.Errorf("expected source 'extras', got %v", list[0].Sources)
	}
}

func TestRenderSkipsDeletedItems(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_x", "Old", "g", false, 1)
	h.addBundle("b", "b")
	h.bundleItem("b", "i_x", nil)
	h.attach(trip, "b")
	h.softDeleteItem("i_x")
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 0 {
		t.Fatalf("expected empty (item soft-deleted), got %+v", list)
	}
}

func TestRenderSkipsDeletedBundles(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i", "I", "g", false, 1)
	h.addBundle("b", "b")
	h.bundleItem("b", "i", nil)
	h.attach(trip, "b")
	h.softDeleteBundle("b")
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 0 {
		t.Fatalf("expected empty (bundle soft-deleted), got %+v", list)
	}
}

func TestRenderExpandsNestedBundles(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_tb", "Toothbrush", "t", false, 1)
	h.addBundle("b_wash", "washbag")
	h.bundleItem("b_wash", "i_tb", nil)
	h.addBundle("b_weekend", "weekend")
	h.nest("b_weekend", "b_wash")
	h.attach(trip, "b_weekend") // only weekend attached; toothbrush should appear via nesting

	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].ItemID != "i_tb" {
		t.Fatalf("expected toothbrush via nest, got %+v", list)
	}
}

func TestRenderDiamondInheritance(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_tb", "Toothbrush", "t", false, 1)
	h.addBundle("b_w", "wash")
	h.bundleItem("b_w", "i_tb", nil)
	h.addBundle("b_a", "a")
	h.addBundle("b_b", "b")
	h.nest("b_a", "b_w")
	h.nest("b_b", "b_w")
	h.attach(trip, "b_a")
	h.attach(trip, "b_b")
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || list[0].Qty != 1 {
		t.Fatalf("expected one toothbrush (diamond), got %+v", list)
	}
}

func TestRenderSkipsDeletedNestedChild(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i", "X", "g", false, 1)
	h.addBundle("b_inner", "inner")
	h.bundleItem("b_inner", "i", nil)
	h.addBundle("b_outer", "outer")
	h.nest("b_outer", "b_inner")
	h.attach(trip, "b_outer")
	h.softDeleteBundle("b_inner")
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 0 {
		t.Fatalf("expected empty (inner soft-deleted), got %+v", list)
	}
}

func TestRenderGroupedByCategorySortedByName(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i_b", "Boots", "clothing", false, 1)
	h.addItem("i_t", "T-shirt", "clothing", false, 1)
	h.addItem("i_p", "Passport", "documents", false, 1)
	h.addBundle("b", "b")
	for _, it := range []string{"i_b", "i_t", "i_p"} {
		h.bundleItem("b", it, nil)
	}
	h.attach(trip, "b")

	list, _ := h.render.Render(context.Background(), trip)
	cats := []string{}
	for _, row := range list {
		cats = append(cats, row.Category)
	}
	want := []string{"clothing", "clothing", "documents"}
	if !sort.IsSorted(sort.StringSlice(cats)) || len(list) != 3 {
		t.Errorf("rows not category-sorted, got %v want %v", cats, want)
	}
}

func TestRenderIncludesPackState(t *testing.T) {
	trip, h := renderFixture(t, 1)
	h.addItem("i", "X", "g", false, 1)
	h.addBundle("b", "b")
	h.bundleItem("b", "i", nil)
	h.attach(trip, "b")
	// pre-mark packed
	if _, err := execDB(h, `INSERT INTO trip_pack_state(trip_id,item_id,packed,packed_at) VALUES (?,?,1,CURRENT_TIMESTAMP)`, trip, "i"); err != nil {
		t.Fatalf("pack state insert: %v", err)
	}
	list, _ := h.render.Render(context.Background(), trip)
	if len(list) != 1 || !list[0].Packed {
		t.Fatalf("expected packed=true, got %+v", list)
	}
}
