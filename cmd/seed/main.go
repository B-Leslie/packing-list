// Command seed populates a fresh database with starter items + bundles.
// Idempotent: if any items exist it bails out.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/ids"
)

type item struct {
	key        string
	name       string
	category   string
	perNight   bool
	defaultQty int
}

type bundle struct {
	key, name, description string
	items                  []bundleItem
	children               []string
}

type bundleItem struct {
	itemKey string
	qty     *int // nil = default
}

func main() {
	force := flag.Bool("force", false, "seed even if items table is non-empty")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		// Fall back to defaults for local dev: BASE_URL not required for seed.
		cfg = config.Config{DataDir: orEnv("DATA_DIR", "./data")}
	}

	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		logger.Error("mkdir", "err", err)
		os.Exit(1)
	}
	d, err := pdb.Open(filepath.Join(cfg.DataDir, "data.db"))
	if err != nil {
		logger.Error("open", "err", err)
		os.Exit(1)
	}
	defer d.Close()

	if !*force {
		var n int
		if err := d.QueryRow(`SELECT count(*) FROM items`).Scan(&n); err != nil {
			logger.Error("count items", "err", err)
			os.Exit(1)
		}
		if n > 0 {
			logger.Info("items exist; skipping (pass -force to seed anyway)")
			return
		}
	}

	// System user for created_by FK.
	if _, err := d.Exec(`INSERT INTO users(id,email) VALUES (?,?)`,
		"u_system", "system@local"); err != nil {
		// User might already exist on -force runs; that's OK.
		logger.Info("system user", "err", err)
	}

	ctx := context.Background()
	itemKeyToID := map[string]string{}

	items := starterItems()
	itemsRepo := catalog.NewItems(d)
	for _, it := range items {
		id, err := itemsRepo.Create(ctx, catalog.Item{
			Name:       it.name,
			Category:   it.category,
			PerNight:   it.perNight,
			DefaultQty: it.defaultQty,
			CreatedBy:  "u_system",
		})
		if err != nil {
			logger.Error("create item", "name", it.name, "err", err)
			os.Exit(1)
		}
		itemKeyToID[it.key] = id
	}

	bundleKeyToID := map[string]string{}
	bundles := starterBundles()
	bundlesRepo := catalog.NewBundles(d)
	// First pass: create the bundles themselves.
	for _, b := range bundles {
		id, err := bundlesRepo.Create(ctx, catalog.Bundle{
			Name:        b.name,
			Description: b.description,
			CreatedBy:   "u_system",
		})
		if err != nil {
			logger.Error("create bundle", "name", b.name, "err", err)
			os.Exit(1)
		}
		bundleKeyToID[b.key] = id
	}
	// Second pass: items + children (children depend on bundle existence).
	for _, b := range bundles {
		bID := bundleKeyToID[b.key]
		for _, bi := range b.items {
			iID := itemKeyToID[bi.itemKey]
			if iID == "" {
				logger.Error("missing seed item ref", "key", bi.itemKey)
				os.Exit(1)
			}
			if err := bundlesRepo.AddItem(ctx, bID, iID, bi.qty); err != nil {
				logger.Error("add bundle item", "err", err)
				os.Exit(1)
			}
		}
		for _, childKey := range b.children {
			cID := bundleKeyToID[childKey]
			if cID == "" {
				logger.Error("missing seed bundle ref", "key", childKey)
				os.Exit(1)
			}
			if err := bundlesRepo.AddChild(ctx, bID, cID); err != nil {
				logger.Error("nest bundle", "err", err)
				os.Exit(1)
			}
		}
	}
	logger.Info("seed complete",
		"items", len(items), "bundles", len(bundles), "at", time.Now().Format(time.RFC3339))
	_ = ids.New // keep import in case of edits
	_ = sql.ErrNoRows
	fmt.Println("seed: ok")
}

func orEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func starterItems() []item {
	return []item{
		// Toiletries
		{"toothbrush", "Toothbrush", "toiletries", false, 1},
		{"toothpaste", "Toothpaste", "toiletries", false, 1},
		{"floss", "Dental floss", "toiletries", false, 1},
		{"shampoo", "Shampoo", "toiletries", false, 1},
		{"conditioner", "Conditioner", "toiletries", false, 1},
		{"shower-gel", "Shower gel", "toiletries", false, 1},
		{"deodorant", "Deodorant", "toiletries", false, 1},
		{"razor", "Razor", "toiletries", false, 1},
		{"moisturiser", "Moisturiser", "toiletries", false, 1},
		{"sunscreen", "Sunscreen", "toiletries", false, 1},
		// Clothing (per-night)
		{"underwear", "Underwear", "clothing", true, 1},
		{"socks", "Socks", "clothing", true, 1},
		{"t-shirt", "T-shirt", "clothing", true, 1},
		// Clothing (fixed)
		{"trousers", "Trousers", "clothing", false, 1},
		{"shorts", "Shorts", "clothing", false, 1},
		{"jumper", "Jumper", "clothing", false, 1},
		{"waterproof", "Waterproof jacket", "clothing", false, 1},
		{"pyjamas", "Pyjamas", "clothing", false, 1},
		{"hat", "Hat", "clothing", false, 1},
		// Electronics
		{"phone-charger", "Phone charger", "electronics", false, 1},
		{"laptop", "Laptop", "electronics", false, 1},
		{"laptop-charger", "Laptop charger", "electronics", false, 1},
		{"power-bank", "Power bank", "electronics", false, 1},
		{"adapter", "Travel adapter", "electronics", false, 1},
		{"headphones", "Headphones", "electronics", false, 1},
		{"e-reader", "E-reader", "electronics", false, 1},
		// Swim / beach
		{"swimsuit", "Swimsuit", "swim", false, 1},
		{"goggles", "Swim goggles", "swim", false, 1},
		{"swim-cap", "Swim cap", "swim", false, 1},
		{"wetsuit", "Wetsuit", "swim", false, 1},
		{"flip-flops", "Flip-flops", "swim", false, 1},
		{"towel-quickdry", "Quick-dry towel", "swim", false, 1},
		// Running
		{"running-shoes-road", "Running shoes (road)", "running", false, 1},
		{"running-shoes-trail", "Running shoes (trail)", "running", false, 1},
		{"running-shorts", "Running shorts", "running", false, 1},
		{"running-top", "Running top", "running", false, 1},
		{"headtorch", "Head torch", "running", false, 1},
		// Documents / misc
		{"passport", "Passport", "documents", false, 1},
		{"wallet", "Wallet", "documents", false, 1},
		{"sunglasses", "Sunglasses", "general", false, 1},
		{"book", "Book", "leisure", false, 1},
	}
}

func starterBundles() []bundle {
	return []bundle{
		{key: "washbag-basic", name: "washbag-basic", description: "Minimal toiletries",
			items: []bundleItem{
				{itemKey: "toothbrush"}, {itemKey: "toothpaste"}, {itemKey: "deodorant"},
				{itemKey: "shower-gel"}, {itemKey: "razor"},
			}},
		{key: "washbag-full", name: "washbag-full", description: "Full toiletries set",
			items: []bundleItem{
				{itemKey: "toothbrush"}, {itemKey: "toothpaste"}, {itemKey: "floss"},
				{itemKey: "shampoo"}, {itemKey: "conditioner"}, {itemKey: "shower-gel"},
				{itemKey: "deodorant"}, {itemKey: "razor"}, {itemKey: "moisturiser"},
				{itemKey: "sunscreen"},
			}},
		{key: "electronics-day", name: "electronics-day", description: "Short-trip electronics",
			items: []bundleItem{{itemKey: "phone-charger"}, {itemKey: "headphones"}}},
		{key: "electronics-week", name: "electronics-week", description: "Week-long electronics",
			items: []bundleItem{
				{itemKey: "phone-charger"}, {itemKey: "laptop"}, {itemKey: "laptop-charger"},
				{itemKey: "power-bank"}, {itemKey: "adapter"}, {itemKey: "headphones"},
				{itemKey: "e-reader"},
			}},
		{key: "swimming-pool", name: "swimming-pool", description: "Pool kit",
			items: []bundleItem{
				{itemKey: "swimsuit"}, {itemKey: "goggles"}, {itemKey: "swim-cap"},
				{itemKey: "flip-flops"}, {itemKey: "towel-quickdry"},
			}},
		{key: "swimming-coldsea", name: "swimming-coldsea", description: "Cold open-water swim",
			items: []bundleItem{
				{itemKey: "swimsuit"}, {itemKey: "wetsuit"}, {itemKey: "goggles"},
				{itemKey: "swim-cap"}, {itemKey: "towel-quickdry"},
			}},
		{key: "running-road", name: "running-road", description: "Road running",
			items: []bundleItem{
				{itemKey: "running-shoes-road"}, {itemKey: "running-shorts"}, {itemKey: "running-top"},
			}},
		{key: "running-trail", name: "running-trail", description: "Trail running",
			items: []bundleItem{
				{itemKey: "running-shoes-trail"}, {itemKey: "running-shorts"}, {itemKey: "running-top"},
				{itemKey: "headtorch"},
			}},
		{key: "basic-clothing", name: "basic-clothing", description: "Standard clothing",
			items: []bundleItem{
				{itemKey: "underwear"}, {itemKey: "socks"}, {itemKey: "t-shirt"},
				{itemKey: "trousers"}, {itemKey: "jumper"}, {itemKey: "pyjamas"},
			}},
		{key: "weekend-city", name: "weekend-city", description: "City weekend",
			children: []string{"washbag-basic", "electronics-day", "basic-clothing"}},
		{key: "week-beach", name: "week-beach", description: "Beach week",
			items: []bundleItem{{itemKey: "shorts"}, {itemKey: "sunglasses"}},
			children: []string{"washbag-full", "electronics-week", "basic-clothing", "swimming-pool"}},
	}
}
