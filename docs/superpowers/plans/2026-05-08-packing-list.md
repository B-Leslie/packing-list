# Packing List Implementation Plan

## Progress (as of 2026-05-12)

All 35 tasks committed. Plan complete.

| Task | Summary | Commit |
| ---- | ------- | ------ |
| 1 | Scaffold module + directory tree | `78863ff` |
| 2 | Add sqlite + ulid deps | `8c7df8e` |
| 3 | ULID helper | `4e87552` |
| 4 | Config loader | `493f8b8` |
| 5 | DB open + migration runner + 0001 schema | `9483ffe` |
| 6 | Items CRUD | `0448157` |
| 7 | Bundles + cycle-checked nesting | `a107cd3` |
| 8 | Trips + members | `d38c4b3` |
| 9 | Render engine | `6d9ff49` |
| 10 | Pack toggle | `96bf4b1` |
| 11 | Trash aggregator | `083e996` |
| 12 | Users find-or-create | `62c1597` |
| 13 | Mailer (log + smtp) | `3fb8b54` |
| 14 | Magic-link tokens | `4ba457d` |
| 15 | HMAC-signed sessions | `8505706` |
| 16 | Session-secret bootstrap | `f0d010e` |
| 17 | RequireUser middleware | `f57b132` |
| 18 | CSRF middleware | `c8b742c` |
| 19 | Rate limiter | `5d51268` |
| 20 | Vendor Pico + HTMX + app overlay | `44e5938` |
| 21 | Layout + login templates | `c10e686` |
| 22 | Template renderer | `b38fb08` |
| 23 | Server scaffold + route table | `f6b5541` |
| 24 | Auth handlers | `9574b59` |
| 25 | Items handlers | `483ea17` |
| 26 | Bundles handlers | `d1aecb5` |
| 27 | Trips list/create/detail | `4aa826d` |
| 28 | Trip source handlers | `37cd999` |
| 29 | Trip members handlers | `23c6985` |
| 30 | Trash page | `3f1bed8` |
| 31 | JSON export / import | `70e6b09` |
| 32 | Main entrypoint | `2cb63eb` |
| 33 | Seed command | `4366248` |
| 34 | Dockerfile + multi-stage build | `8dda717` |
| 35 | README + manual verification | `7a02565` |

**Notes on Task 34:** The Dockerfile in the plan body (golang 1.23-alpine, no pre-created `/data` dir) was extended in commit `8dda717` to (a) use `golang:1.26-alpine` to match `go.mod`'s `go 1.26.3` directive and (b) `mkdir /out/data` in the builder + `COPY --chown=nonroot:nonroot /out/data /data` in the runtime stage so named-volume mounts inherit nonroot ownership (without this the container exits with `open /data/.session_secret: permission denied`). Image size 34.5 MB (plan target was <30 MB; the extra is from bundling the `seed` binary alongside `packing-list`).

**Resume:** Plan execution finished. Outstanding human work is the manual verification checklist in `README.md` against a freshly-built image.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hostable web app to compose trip packing lists out of reusable bundles of items, with check-off, multi-user trip sharing, and soft-delete/restore.

**Architecture:** Go HTTP server (stdlib `net/http`) rendering server-side HTML with HTMX for interactivity. SQLite (modernc, cgo-free) single-file DB. Magic-link auth, signed-cookie sessions, classless Pico.css for styling. Single static binary in a distroless Docker image; volume-mount holds DB + backups.

**Tech Stack:** Go 1.23+, `database/sql` + `modernc.org/sqlite`, `html/template`, HTMX 2.x, Pico.css 2.x, `github.com/oklog/ulid/v2`, stdlib `net/smtp` and `slog`.

**Spec:** [`docs/superpowers/specs/2026-05-08-packing-list-design.md`](../specs/2026-05-08-packing-list-design.md)

**Working directory:** `C:\Users\bejl\source\repos\packing-list` (on Windows; commands shown for `bash`/`pwsh` cross-form where it matters).

**Conventions used in every task:**
- All commits use Conventional Commits (`feat:`, `test:`, `refactor:`, `chore:`).
- After every task, run `go test ./...` and only commit if green.
- File paths shown with forward slashes; OK on Windows Go toolchain.
- Each task ends with a commit. Frequent commits, atomic changes.

---

## Phase 1 — Foundation

### Task 1: Initialise Go module and base directories

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `README.md`
- Create directory tree (empty `.gitkeep` files): `internal/config/`, `internal/db/migrations/`, `internal/auth/`, `internal/catalog/`, `internal/trips/`, `internal/web/handlers/`, `internal/web/templates/pages/`, `internal/web/templates/partials/`, `internal/web/static/`, `cmd/seed/`, `data/`

- [ ] **Step 1: Initialise git and Go module**

```bash
git init
go mod init github.com/bejl/packing-list
```

Expected: `go.mod` created with `module github.com/bejl/packing-list` and `go 1.23` (or system version ≥1.23).

- [ ] **Step 2: Write `.gitignore`**

```
/data/
/packing-list
/packing-list.exe
*.test
*.db
*.db-journal
*.db-shm
*.db-wal
.session_secret
.env
.idea/
.vscode/
```

- [ ] **Step 3: Create directory tree with `.gitkeep` placeholders**

```bash
mkdir -p internal/config internal/db/migrations internal/auth \
  internal/catalog internal/trips \
  internal/web/handlers internal/web/templates/pages \
  internal/web/templates/partials internal/web/static \
  cmd/seed data
for d in internal/config internal/db/migrations internal/auth \
  internal/catalog internal/trips internal/web/handlers \
  internal/web/templates/pages internal/web/templates/partials \
  internal/web/static cmd/seed; do touch "$d/.gitkeep"; done
```

- [ ] **Step 4: Write minimal README.md**

```markdown
# packing-list

Self-hostable trip packing list. See `docs/superpowers/specs/2026-05-08-packing-list-design.md`.

## Run

```
docker run -d --name packing-list \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e BASE_URL=http://localhost:8080 \
  -e BOOTSTRAP_EMAIL=you@example.com \
  packing-list
```
```

- [ ] **Step 5: Initial commit**

```bash
git add .
git commit -m "chore: scaffold Go module and directory layout"
```

---

### Task 2: Add core dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add modernc SQLite and ULID**

```bash
go get modernc.org/sqlite@latest
go get github.com/oklog/ulid/v2@latest
go mod tidy
```

Expected: both modules resolve, `go.sum` populated, no errors.

- [ ] **Step 2: Smoke-test the build**

```bash
go build ./...
```

Expected: no output (nothing to build yet, no error).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add sqlite (modernc) and ulid dependencies"
```

---

### Task 3: ULID helper

**Files:**
- Create: `internal/ids/ids.go`
- Create: `internal/ids/ids_test.go`

- [ ] **Step 1: Write the failing test**

`internal/ids/ids_test.go`:
```go
package ids

import (
	"strings"
	"testing"
)

func TestNewIsUnique(t *testing.T) {
	a, b := New(), New()
	if a == b {
		t.Fatalf("expected unique IDs, got %s twice", a)
	}
	if len(a) != 26 {
		t.Fatalf("expected 26-char ULID, got %d (%q)", len(a), a)
	}
}

func TestNewIsLexicographicallySortable(t *testing.T) {
	first := New()
	// Sleep a tiny bit so the millisecond-precision timestamp ticks over.
	for i := 0; i < 10; i++ {
		_ = New()
	}
	last := New()
	if strings.Compare(first, last) >= 0 {
		t.Fatalf("expected first (%s) < last (%s)", first, last)
	}
}
```

- [ ] **Step 2: Run test, confirm it fails**

```bash
go test ./internal/ids/...
```

Expected: build error — `ids.New` undefined.

- [ ] **Step 3: Implement `ids.New`**

`internal/ids/ids.go`:
```go
// Package ids generates ULIDs for primary keys.
package ids

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	mu     sync.Mutex
	source = ulid.Monotonic(rand.Reader, 0)
)

// New returns a fresh ULID as a 26-char base32 string.
func New() string {
	mu.Lock()
	defer mu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), source).String()
}
```

- [ ] **Step 4: Run test, confirm it passes**

```bash
go test ./internal/ids/...
```

Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/ids
git commit -m "feat(ids): add monotonic ULID generator"
```

---

### Task 4: Config loader

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests**

`internal/config/config_test.go`:
```go
package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	getenv := func(string) string { return "" }
	_, err := Load(getenv)
	if err == nil {
		t.Fatal("expected error when BASE_URL is empty, got nil")
	}
}

func TestLoadHappy(t *testing.T) {
	env := map[string]string{
		"BASE_URL": "http://localhost:8080",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "8080")
	}
	if cfg.DataDir != "/data" {
		t.Errorf("DataDir: got %q, want %q", cfg.DataDir, "/data")
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL: got %q, want %q", cfg.BaseURL, "http://localhost:8080")
	}
	if cfg.SMTP.Configured() {
		t.Error("SMTP should not be configured without SMTP_HOST")
	}
}

func TestLoadSMTP(t *testing.T) {
	env := map[string]string{
		"BASE_URL":  "http://h",
		"SMTP_HOST": "smtp.example.com",
		"SMTP_USER": "u",
		"SMTP_PASS": "p",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !cfg.SMTP.Configured() {
		t.Fatal("expected SMTP configured")
	}
	if cfg.SMTP.From == "" {
		t.Error("expected SMTP.From to default to no-reply@<host>")
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/config/...
```

Expected: build error.

- [ ] **Step 3: Implement loader**

`internal/config/config.go`:
```go
// Package config parses runtime configuration from environment variables.
package config

import (
	"errors"
	"net/url"
)

type SMTP struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

func (s SMTP) Configured() bool { return s.Host != "" }

type Config struct {
	Port            string
	BaseURL         string
	DataDir         string
	SessionSecret   string // empty means: load/generate at boot
	BootstrapEmail  string
	SMTP            SMTP
}

// Load reads config using the given getenv function (so tests don't touch real env).
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Port:            or(getenv("PORT"), "8080"),
		BaseURL:         getenv("BASE_URL"),
		DataDir:         or(getenv("DATA_DIR"), "/data"),
		SessionSecret:   getenv("SESSION_SECRET"),
		BootstrapEmail:  getenv("BOOTSTRAP_EMAIL"),
		SMTP: SMTP{
			Host: getenv("SMTP_HOST"),
			Port: or(getenv("SMTP_PORT"), "587"),
			User: getenv("SMTP_USER"),
			Pass: getenv("SMTP_PASS"),
			From: getenv("SMTP_FROM"),
		},
	}
	if c.BaseURL == "" {
		return c, errors.New("BASE_URL is required")
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" {
		return c, errors.New("BASE_URL is not a valid URL")
	}
	if c.SMTP.Configured() && c.SMTP.From == "" {
		c.SMTP.From = "no-reply@" + u.Hostname()
	}
	return c, nil
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/config/...
```

Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/config
git commit -m "feat(config): env-based config with SMTP defaults"
```

---

### Task 5: DB open + migration runner

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/db_test.go`
- Create: `internal/db/migrations/0001_init.sql`

- [ ] **Step 1: Write failing test**

`internal/db/db_test.go`:
```go
package db

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()

	var n int
	if err := d.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected users table to exist, got count=%d", n)
	}

	// Foreign keys must be ON.
	var fk int
	if err := d.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	if fk != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", fk)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d1, err := Open(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	d1.Close()
	d2, err := Open(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer d2.Close()
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/db/...
```

Expected: build error (`db.Open` not defined).

- [ ] **Step 3: Write migration 0001**

`internal/db/migrations/0001_init.sql`: copy verbatim from spec §5 (all tables: users, sessions, magic_tokens, items, bundles, bundle_items, bundle_children, trips, trip_members, trip_bundles, trip_extras, trip_overrides, trip_pack_state) plus indexes. Drop the spec's "PRAGMA foreign_keys" comment — pragmas live in code, not migrations.

```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE COLLATE NOCASE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE magic_tokens (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL COLLATE NOCASE,
  token_hash BLOB NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE items (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'general',
  per_night INTEGER NOT NULL DEFAULT 0,
  default_qty INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_items_active ON items(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE bundles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_bundles_active ON bundles(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE bundle_items (
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  qty       INTEGER,
  PRIMARY KEY (bundle_id, item_id)
);

CREATE TABLE bundle_children (
  parent_id TEXT NOT NULL REFERENCES bundles(id),
  child_id  TEXT NOT NULL REFERENCES bundles(id),
  PRIMARY KEY (parent_id, child_id),
  CHECK (parent_id <> child_id)
);
CREATE INDEX idx_bundle_children_child ON bundle_children(child_id);

CREATE TABLE trips (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  nights INTEGER NOT NULL DEFAULT 1 CHECK (nights >= 0),
  starts_on DATE,
  notes TEXT,
  owner_id TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_trips_owner_active ON trips(owner_id) WHERE deleted_at IS NULL;

CREATE TABLE trip_members (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  role TEXT NOT NULL CHECK (role IN ('owner','editor')),
  added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, user_id)
);

CREATE TABLE trip_bundles (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  added_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, bundle_id)
);

CREATE TABLE trip_extras (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  item_id TEXT NOT NULL REFERENCES items(id),
  qty     INTEGER,
  PRIMARY KEY (trip_id, item_id)
);

CREATE TABLE trip_overrides (
  trip_id      TEXT NOT NULL REFERENCES trips(id),
  item_id      TEXT NOT NULL REFERENCES items(id),
  removed      INTEGER NOT NULL DEFAULT 0,
  qty_override INTEGER,
  PRIMARY KEY (trip_id, item_id)
);

CREATE TABLE trip_pack_state (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  packed    INTEGER NOT NULL DEFAULT 0,
  packed_at TIMESTAMP,
  PRIMARY KEY (trip_id, item_id)
);

CREATE TABLE schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **Step 4: Implement `db.Open`**

`internal/db/db.go`:
```go
// Package db opens SQLite with the right pragmas and applies bundled migrations.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens (or creates) the SQLite database at path, applies any pending
// migrations, and returns a *sql.DB ready for use.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := d.Ping(); err != nil {
		d.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if err := migrate(d); err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func migrate(d *sql.DB) error {
	if _, err := d.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	type m struct {
		v   int
		sql string
	}
	var ms []m
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// File name format: NNNN_*.sql
		idx := strings.Index(e.Name(), "_")
		if idx < 0 {
			continue
		}
		v, err := strconv.Atoi(e.Name()[:idx])
		if err != nil {
			return fmt.Errorf("bad migration name %q: %w", e.Name(), err)
		}
		body, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		ms = append(ms, m{v, string(body)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].v < ms[j].v })

	for _, mig := range ms {
		var present int
		if err := d.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version = ?`, mig.v).Scan(&present); err != nil {
			return err
		}
		if present > 0 {
			continue
		}
		tx, err := d.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(mig.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", mig.v, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, mig.v); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/db/...
```

Expected: `PASS`.

- [ ] **Step 6: Commit**

```bash
git add internal/db go.mod go.sum
git commit -m "feat(db): open sqlite with WAL + FK and apply embedded migrations"
```

---

## Phase 2 — Domain layer

### Task 6: Items domain (CRUD + soft-delete)

**Files:**
- Create: `internal/catalog/items.go`
- Create: `internal/catalog/items_test.go`
- Create: `internal/catalog/testdb.go` (test helper, build tag `_test`)

- [ ] **Step 1: Test helper**

`internal/catalog/testdb.go`:
```go
package catalog

import (
	"database/sql"
	"path/filepath"
	"testing"

	pdb "github.com/bejl/packing-list/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	// Seed a user for created_by FK.
	if _, err := d.Exec(`INSERT INTO users(id,email) VALUES('u_test','test@example.com')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return d
}
```

- [ ] **Step 2: Failing test for Items CRUD**

`internal/catalog/items_test.go`:
```go
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
```

- [ ] **Step 3: Run, confirm fail**

```bash
go test ./internal/catalog/...
```

Expected: build error.

- [ ] **Step 4: Implement Items**

`internal/catalog/items.go`:
```go
// Package catalog holds the global item + bundle catalog.
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation")
)

type Item struct {
	ID         string
	Name       string
	Category   string
	PerNight   bool
	DefaultQty int
	Notes      string
	CreatedBy  string
	CreatedAt  time.Time
	DeletedAt  sql.NullTime
	DeletedBy  sql.NullString
}

type Items struct{ db *sql.DB }

func NewItems(db *sql.DB) *Items { return &Items{db: db} }

func (r *Items) Create(ctx context.Context, it Item) (string, error) {
	if it.Name == "" {
		return "", ErrValidation
	}
	if it.Category == "" {
		it.Category = "general"
	}
	if it.DefaultQty <= 0 {
		it.DefaultQty = 1
	}
	id := ids.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO items(id,name,category,per_night,default_qty,notes,created_by) VALUES (?,?,?,?,?,?,?)`,
		id, it.Name, it.Category, boolInt(it.PerNight), it.DefaultQty, nullStr(it.Notes), it.CreatedBy)
	return id, err
}

func (r *Items) Get(ctx context.Context, id string) (Item, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE id = ?`, id)
	return scanItem(row)
}

func (r *Items) List(ctx context.Context) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE deleted_at IS NULL ORDER BY category, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Items) ListDeleted(ctx context.Context) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Items) Update(ctx context.Context, id string, it Item) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET name = ?, category = ?, per_night = ?, default_qty = ?, notes = ? WHERE id = ?`,
		it.Name, it.Category, boolInt(it.PerNight), it.DefaultQty, nullStr(it.Notes), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
		byUser, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM bundle_items WHERE item_id = ?`,
		`DELETE FROM trip_extras WHERE item_id = ?`,
		`DELETE FROM trip_overrides WHERE item_id = ?`,
		`DELETE FROM trip_pack_state WHERE item_id = ?`,
		`DELETE FROM items WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(...any) error
}

func scanItem(s rowScanner) (Item, error) {
	var it Item
	var per int
	err := s.Scan(&it.ID, &it.Name, &it.Category, &per, &it.DefaultQty, &it.Notes,
		&it.CreatedBy, &it.CreatedAt, &it.DeletedAt, &it.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return it, ErrNotFound
	}
	it.PerNight = per != 0
	return it, err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/catalog/...
```

Expected: `PASS`.

- [ ] **Step 6: Commit**

```bash
git add internal/catalog
git commit -m "feat(catalog): items CRUD with soft-delete, restore, purge"
```

---

### Task 7: Bundles domain (items + nested children, cycle check)

**Files:**
- Create: `internal/catalog/bundles.go`
- Create: `internal/catalog/bundles_test.go`

- [ ] **Step 1: Failing tests**

`internal/catalog/bundles_test.go`:
```go
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
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/catalog/...
```

Expected: build error.

- [ ] **Step 3: Implement**

`internal/catalog/bundles.go`:
```go
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

type Bundle struct {
	ID          string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   time.Time
	DeletedAt   sql.NullTime
	DeletedBy   sql.NullString
}

type BundleItem struct {
	BundleID string
	ItemID   string
	Qty      sql.NullInt64
}

type Bundles struct{ db *sql.DB }

func NewBundles(db *sql.DB) *Bundles { return &Bundles{db: db} }

func (r *Bundles) Create(ctx context.Context, b Bundle) (string, error) {
	if b.Name == "" {
		return "", ErrValidation
	}
	id := ids.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundles(id,name,description,created_by) VALUES (?,?,?,?)`,
		id, b.Name, nullStr(b.Description), b.CreatedBy)
	return id, err
}

func (r *Bundles) Get(ctx context.Context, id string) (Bundle, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE id = ?`, id)
	var b Bundle
	err := row.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (r *Bundles) List(ctx context.Context) ([]Bundle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bundle
	for rows.Next() {
		var b Bundle
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Bundles) ListDeleted(ctx context.Context) ([]Bundle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bundle
	for rows.Next() {
		var b Bundle
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Bundles) Update(ctx context.Context, id string, b Bundle) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET name = ?, description = ? WHERE id = ?`,
		b.Name, nullStr(b.Description), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Bundles) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
		byUser, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Bundles) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Bundles) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM bundle_items WHERE bundle_id = ?`,
		`DELETE FROM bundle_children WHERE parent_id = ? OR child_id = ?`,
		`DELETE FROM trip_bundles WHERE bundle_id = ?`,
		`DELETE FROM bundles WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		switch q {
		case `DELETE FROM bundle_children WHERE parent_id = ? OR child_id = ?`:
			if _, err := tx.ExecContext(ctx, q, id, id); err != nil {
				return err
			}
		default:
			if _, err := tx.ExecContext(ctx, q, id); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *Bundles) AddItem(ctx context.Context, bundleID, itemID string, qty *int) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundle_items(bundle_id,item_id,qty) VALUES (?,?,?)
		 ON CONFLICT(bundle_id,item_id) DO UPDATE SET qty = excluded.qty`,
		bundleID, itemID, q)
	return err
}

func (r *Bundles) RemoveItem(ctx context.Context, bundleID, itemID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bundle_items WHERE bundle_id = ? AND item_id = ?`, bundleID, itemID)
	return err
}

func (r *Bundles) Items(ctx context.Context, bundleID string) ([]BundleItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT bundle_id,item_id,qty FROM bundle_items WHERE bundle_id = ?`, bundleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BundleItem
	for rows.Next() {
		var bi BundleItem
		if err := rows.Scan(&bi.BundleID, &bi.ItemID, &bi.Qty); err != nil {
			return nil, err
		}
		out = append(out, bi)
	}
	return out, rows.Err()
}

// AddChild nests childID under parentID. Returns ErrConflict if a cycle would form.
func (r *Bundles) AddChild(ctx context.Context, parentID, childID string) error {
	if parentID == childID {
		return errors.New("self-nest forbidden")
	}
	// BFS from childID following bundle_children. If we ever reach parentID, cycle.
	queue := []string{childID}
	seen := map[string]bool{childID: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		rows, err := r.db.QueryContext(ctx, `SELECT child_id FROM bundle_children WHERE parent_id = ?`, cur)
		if err != nil {
			return err
		}
		var next []string
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				rows.Close()
				return err
			}
			next = append(next, c)
		}
		rows.Close()
		for _, c := range next {
			if c == parentID {
				return ErrConflict
			}
			if !seen[c] {
				seen[c] = true
				queue = append(queue, c)
			}
		}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundle_children(parent_id,child_id) VALUES (?,?) ON CONFLICT DO NOTHING`,
		parentID, childID)
	return err
}

func (r *Bundles) RemoveChild(ctx context.Context, parentID, childID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bundle_children WHERE parent_id = ? AND child_id = ?`, parentID, childID)
	return err
}

func (r *Bundles) Children(ctx context.Context, parentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT child_id FROM bundle_children WHERE parent_id = ?`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/catalog/...
```

Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/catalog
git commit -m "feat(catalog): bundles with item refs and cycle-checked nesting"
```

---

### Task 8: Trips domain (CRUD + members)

**Files:**
- Create: `internal/trips/trips.go`
- Create: `internal/trips/trips_test.go`
- Create: `internal/trips/testdb.go`

- [ ] **Step 1: Test helper**

`internal/trips/testdb.go`:
```go
package trips

import (
	"database/sql"
	"path/filepath"
	"testing"

	pdb "github.com/bejl/packing-list/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	for _, q := range []string{
		`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`,
		`INSERT INTO users(id,email) VALUES('u_b','b@example.com')`,
	} {
		if _, err := d.Exec(q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return d
}
```

- [ ] **Step 2: Failing tests**

`internal/trips/trips_test.go`:
```go
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
```

- [ ] **Step 3: Run, confirm fail**

```bash
go test ./internal/trips/...
```

- [ ] **Step 4: Implement**

`internal/trips/trips.go`:
```go
// Package trips owns trips, trip membership, and trip-scoped sub-tables.
package trips

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrNotMember = errors.New("not a member")
	ErrForbidden = errors.New("forbidden")
)

type Trip struct {
	ID        string
	Name      string
	Nights    int
	StartsOn  sql.NullString
	Notes     string
	OwnerID   string
	CreatedAt time.Time
	DeletedAt sql.NullTime
	DeletedBy sql.NullString
}

type Member struct {
	UserID  string
	Role    string
	Email   string
	AddedAt time.Time
}

type Trips struct{ db *sql.DB }

func NewTrips(db *sql.DB) *Trips { return &Trips{db: db} }

func (r *Trips) Create(ctx context.Context, name string, nights int, ownerID string) (string, error) {
	if name == "" || nights < 0 {
		return "", errors.New("invalid trip")
	}
	id := ids.New()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO trips(id,name,nights,owner_id) VALUES (?,?,?,?)`,
		id, name, nights, ownerID); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO trip_members(trip_id,user_id,role) VALUES (?,?,'owner')`,
		id, ownerID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

func (r *Trips) Get(ctx context.Context, id string) (Trip, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,nights,starts_on,COALESCE(notes,''),owner_id,created_at,deleted_at,deleted_by
		 FROM trips WHERE id = ?`, id)
	var t Trip
	err := row.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (r *Trips) Update(ctx context.Context, id, name string, nights int, notes string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET name = ?, nights = ?, notes = ? WHERE id = ?`,
		name, nights, nullStr(notes), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Trips) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
		byUser, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Trips) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Trips) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM trip_pack_state WHERE trip_id = ?`,
		`DELETE FROM trip_overrides WHERE trip_id = ?`,
		`DELETE FROM trip_extras WHERE trip_id = ?`,
		`DELETE FROM trip_bundles WHERE trip_id = ?`,
		`DELETE FROM trip_members WHERE trip_id = ?`,
		`DELETE FROM trips WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Trips) ListVisibleTo(ctx context.Context, userID string) ([]Trip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id,t.name,t.nights,t.starts_on,COALESCE(t.notes,''),t.owner_id,t.created_at,t.deleted_at,t.deleted_by
		FROM trips t
		JOIN trip_members m ON m.trip_id = t.id
		WHERE t.deleted_at IS NULL AND m.user_id = ?
		ORDER BY t.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trip
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Trips) ListDeleted(ctx context.Context, userID string) ([]Trip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id,t.name,t.nights,t.starts_on,COALESCE(t.notes,''),t.owner_id,t.created_at,t.deleted_at,t.deleted_by
		FROM trips t
		JOIN trip_members m ON m.trip_id = t.id
		WHERE t.deleted_at IS NOT NULL AND m.user_id = ?
		ORDER BY t.deleted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trip
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Trips) AddMember(ctx context.Context, tripID, userID, role string) error {
	if role != "owner" && role != "editor" {
		return errors.New("bad role")
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO trip_members(trip_id,user_id,role) VALUES (?,?,?) ON CONFLICT DO NOTHING`,
		tripID, userID, role)
	return err
}

func (r *Trips) RemoveMember(ctx context.Context, tripID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM trip_members WHERE trip_id = ? AND user_id = ? AND role <> 'owner'`, tripID, userID)
	return err
}

func (r *Trips) Members(ctx context.Context, tripID string) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.user_id, m.role, u.email, m.added_at
		FROM trip_members m JOIN users u ON u.id = m.user_id
		WHERE m.trip_id = ?
		ORDER BY (m.role = 'owner') DESC, m.added_at`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Role, &m.Email, &m.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Trips) RoleOf(ctx context.Context, tripID, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM trip_members WHERE trip_id = ? AND user_id = ?`, tripID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotMember
	}
	return role, err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/trips/...
```

- [ ] **Step 6: Commit**

```bash
git add internal/trips
git commit -m "feat(trips): trips CRUD with members and visibility filtering"
```

---

### Task 9: Render engine (trip → packing list)

This is the heart of the application. It expands bundles (including nested children), folds in extras and overrides, merges by item, and joins pack state. Heavy table-driven tests.

**Files:**
- Create: `internal/trips/sources.go` (small helpers for attach/remove bundles/extras/overrides — needed by tests below)
- Create: `internal/trips/render.go`
- Create: `internal/trips/render_test.go`

- [ ] **Step 1: Sources helpers (so tests can populate state)**

`internal/trips/sources.go`:
```go
package trips

import (
	"context"
	"database/sql"
)

type Sources struct{ db *sql.DB }

func NewSources(db *sql.DB) *Sources { return &Sources{db: db} }

func (s *Sources) AttachBundle(ctx context.Context, tripID, bundleID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_bundles(trip_id,bundle_id) VALUES(?,?) ON CONFLICT DO NOTHING`,
		tripID, bundleID)
	return err
}

func (s *Sources) DetachBundle(ctx context.Context, tripID, bundleID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_bundles WHERE trip_id = ? AND bundle_id = ?`, tripID, bundleID)
	return err
}

func (s *Sources) AddExtra(ctx context.Context, tripID, itemID string, qty *int) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_extras(trip_id,item_id,qty) VALUES(?,?,?)
		 ON CONFLICT(trip_id,item_id) DO UPDATE SET qty = excluded.qty`,
		tripID, itemID, q)
	return err
}

func (s *Sources) RemoveExtra(ctx context.Context, tripID, itemID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_extras WHERE trip_id = ? AND item_id = ?`, tripID, itemID)
	return err
}

func (s *Sources) SetOverride(ctx context.Context, tripID, itemID string, removed bool, qty *int) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	r := 0
	if removed {
		r = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO trip_overrides(trip_id,item_id,removed,qty_override) VALUES(?,?,?,?)
		 ON CONFLICT(trip_id,item_id) DO UPDATE SET removed = excluded.removed, qty_override = excluded.qty_override`,
		tripID, itemID, r, q)
	return err
}

func (s *Sources) ClearOverride(ctx context.Context, tripID, itemID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM trip_overrides WHERE trip_id = ? AND item_id = ?`, tripID, itemID)
	return err
}

func (s *Sources) AttachedBundleIDs(ctx context.Context, tripID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT bundle_id FROM trip_bundles WHERE trip_id = ?`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
```

- [ ] **Step 2: Failing render tests (table-driven, exhaustive)**

`internal/trips/render_test.go`:
```go
package trips

import (
	"context"
	"sort"
	"testing"
)

// helpers ---------------------------------------------------------------------

func intp(v int) *int { return &v }

type fixture struct {
	t      *testing.T
	db     interface{ /* not used */ }
	trip   string
	srcs   *Sources
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

type anyDB = interface {
	Exec(string, ...any) (interface{ RowsAffected() (int64, error); LastInsertId() (int64, error) }, error)
}

// (We use direct SQL here for terseness in tests rather than going through
// the catalog package — render is the unit under test, not catalog.)

// addItem inserts an item directly.
func (h *renderHelper) addItem(id, name, cat string, perNight bool, defQty int) {
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

func (h *renderHelper) bundleItem(bundle, item string, qty *int) {
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
	h.bundleItem("b_a", "i_tw", intp(1))
	h.bundleItem("b_b", "i_tw", intp(2)) // bigger qty wins for fixed items
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
	h.bundleItem("b_a", "i_s", intp(1)) // 1 per night
	h.bundleItem("b_b", "i_s", intp(1)) // 1 per night
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
	srcs.SetOverride(context.Background(), trip, "i_u", false, intp(2)) // force qty=2 regardless of nights

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
```

- [ ] **Step 3: Run, confirm fail**

```bash
go test ./internal/trips/...
```

Expected: build errors — `Renderer`, `NewRenderer`, `Render`, `Row`, fields undefined.

- [ ] **Step 4: Implement Render**

`internal/trips/render.go`:
```go
package trips

import (
	"context"
	"database/sql"
	"sort"
)

// Row is one rendered item destined for the UI.
type Row struct {
	ItemID   string
	Name     string
	Category string
	Qty      int
	PerNight bool
	Sources  []string // unique source bundle names plus possibly "extras"
	Packed   bool
}

type Renderer struct{ db *sql.DB }

func NewRenderer(db *sql.DB) *Renderer { return &Renderer{db: db} }

// Render returns the final packing list for tripID. Pure function over DB rows.
func (r *Renderer) Render(ctx context.Context, tripID string) ([]Row, error) {
	// 1. Trip metadata.
	var nights int
	if err := r.db.QueryRowContext(ctx, `SELECT nights FROM trips WHERE id = ?`, tripID).Scan(&nights); err != nil {
		return nil, err
	}

	// 2. Resolve attached bundles, expanded recursively.
	bundleIDs, bundleNames, err := r.expandAttachedBundles(ctx, tripID)
	if err != nil {
		return nil, err
	}

	type srcEntry struct {
		qty      int
		perNight bool
		source   string
	}
	srcs := map[string][]srcEntry{} // item_id -> entries
	itemMeta := map[string]struct {
		Name, Category string
		PerNight       bool
	}{}

	// 3. For each (non-deleted) bundle, gather its non-deleted items.
	if len(bundleIDs) > 0 {
		query := `
		  SELECT bi.bundle_id, bi.item_id, COALESCE(bi.qty, i.default_qty),
		         i.name, i.category, i.per_night
		  FROM bundle_items bi
		  JOIN items i ON i.id = bi.item_id
		  JOIN bundles b ON b.id = bi.bundle_id
		  WHERE bi.bundle_id IN (` + placeholders(len(bundleIDs)) + `)
		    AND b.deleted_at IS NULL
		    AND i.deleted_at IS NULL`
		args := make([]any, 0, len(bundleIDs))
		for _, id := range bundleIDs {
			args = append(args, id)
		}
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var bID, iID, name, cat string
			var qty int
			var per int
			if err := rows.Scan(&bID, &iID, &qty, &name, &cat, &per); err != nil {
				rows.Close()
				return nil, err
			}
			srcs[iID] = append(srcs[iID], srcEntry{qty: qty, perNight: per != 0, source: bundleNames[bID]})
			itemMeta[iID] = struct {
				Name, Category string
				PerNight       bool
			}{name, cat, per != 0}
		}
		rows.Close()
	}

	// 4. Extras.
	{
		rows, err := r.db.QueryContext(ctx, `
		  SELECT te.item_id, COALESCE(te.qty, i.default_qty), i.name, i.category, i.per_night
		  FROM trip_extras te JOIN items i ON i.id = te.item_id
		  WHERE te.trip_id = ? AND i.deleted_at IS NULL`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID, name, cat string
			var qty, per int
			if err := rows.Scan(&iID, &qty, &name, &cat, &per); err != nil {
				rows.Close()
				return nil, err
			}
			srcs[iID] = append(srcs[iID], srcEntry{qty: qty, perNight: per != 0, source: "extras"})
			itemMeta[iID] = struct {
				Name, Category string
				PerNight       bool
			}{name, cat, per != 0}
		}
		rows.Close()
	}

	// 5. Overrides.
	overrides := map[string]struct {
		removed     bool
		qtyOverride sql.NullInt64
	}{}
	{
		rows, err := r.db.QueryContext(ctx,
			`SELECT item_id, removed, qty_override FROM trip_overrides WHERE trip_id = ?`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID string
			var rem int
			var qo sql.NullInt64
			if err := rows.Scan(&iID, &rem, &qo); err != nil {
				rows.Close()
				return nil, err
			}
			overrides[iID] = struct {
				removed     bool
				qtyOverride sql.NullInt64
			}{rem != 0, qo}
		}
		rows.Close()
	}

	// 6. Pack state.
	packed := map[string]bool{}
	{
		rows, err := r.db.QueryContext(ctx,
			`SELECT item_id, packed FROM trip_pack_state WHERE trip_id = ?`, tripID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var iID string
			var p int
			if err := rows.Scan(&iID, &p); err != nil {
				rows.Close()
				return nil, err
			}
			packed[iID] = p != 0
		}
		rows.Close()
	}

	// 7. Merge.
	out := make([]Row, 0, len(srcs))
	for iID, entries := range srcs {
		if ov, ok := overrides[iID]; ok && ov.removed {
			continue
		}
		meta := itemMeta[iID]
		row := Row{
			ItemID:   iID,
			Name:     meta.Name,
			Category: meta.Category,
			PerNight: meta.PerNight,
			Packed:   packed[iID],
		}

		// Sources de-duplicated, in stable order.
		seen := map[string]bool{}
		for _, e := range entries {
			if !seen[e.source] {
				seen[e.source] = true
				row.Sources = append(row.Sources, e.source)
			}
		}
		sort.Strings(row.Sources)

		// Quantity: max for fixed items, sum * nights for per-night items.
		if meta.PerNight {
			sum := 0
			for _, e := range entries {
				sum += e.qty
			}
			row.Qty = sum * nights
		} else {
			max := 0
			for _, e := range entries {
				if e.qty > max {
					max = e.qty
				}
			}
			row.Qty = max
		}

		// Apply qty override last.
		if ov, ok := overrides[iID]; ok && ov.qtyOverride.Valid {
			row.Qty = int(ov.qtyOverride.Int64)
		}

		out = append(out, row)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// expandAttachedBundles returns the set of bundle IDs that contribute to the
// trip (attached bundles + all nested descendants, excluding soft-deleted).
// Also returns a map of id -> name (for sourcing labels).
func (r *Renderer) expandAttachedBundles(ctx context.Context, tripID string) ([]string, map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.name
		FROM trip_bundles tb
		JOIN bundles b ON b.id = tb.bundle_id
		WHERE tb.trip_id = ? AND b.deleted_at IS NULL`, tripID)
	if err != nil {
		return nil, nil, err
	}
	var queue []string
	names := map[string]string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return nil, nil, err
		}
		queue = append(queue, id)
		names[id] = name
	}
	rows.Close()

	seen := map[string]bool{}
	for _, id := range queue {
		seen[id] = true
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		rs, err := r.db.QueryContext(ctx, `
			SELECT b.id, b.name
			FROM bundle_children bc
			JOIN bundles b ON b.id = bc.child_id
			WHERE bc.parent_id = ? AND b.deleted_at IS NULL`, cur)
		if err != nil {
			return nil, nil, err
		}
		for rs.Next() {
			var id, name string
			if err := rs.Scan(&id, &name); err != nil {
				rs.Close()
				return nil, nil, err
			}
			if !seen[id] {
				seen[id] = true
				queue = append(queue, id)
				names[id] = name
			}
		}
		rs.Close()
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids, names, nil
}

func placeholders(n int) string {
	if n == 0 {
		return ""
	}
	out := "?"
	for i := 1; i < n; i++ {
		out += ",?"
	}
	return out
}
```

- [ ] **Step 5: Run tests, confirm pass**

```bash
go test ./internal/trips/... -v -run Render
```

Expected: all `TestRender*` pass.

- [ ] **Step 6: Commit**

```bash
git add internal/trips
git commit -m "feat(trips): render engine with nested bundles, overrides, pack state"
```

---

### Task 10: Pack toggle

**Files:**
- Create: `internal/trips/pack.go`
- Create: `internal/trips/pack_test.go`

- [ ] **Step 1: Failing test**

`internal/trips/pack_test.go`:
```go
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
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/trips/...
```

- [ ] **Step 3: Implement**

`internal/trips/pack.go`:
```go
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
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/trips/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/trips
git commit -m "feat(trips): pack toggle (upsert into trip_pack_state)"
```

---

### Task 11: Trash list aggregation

The per-entity soft-delete/restore/purge methods already exist (Tasks 6–8). This task adds a small aggregator for the `/trash` view: returns each kind's deleted rows.

**Files:**
- Create: `internal/trash/trash.go`
- Create: `internal/trash/trash_test.go`

- [ ] **Step 1: Failing test**

`internal/trash/trash_test.go`:
```go
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
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/trash/...
```

- [ ] **Step 3: Implement**

`internal/trash/trash.go`:
```go
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
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/trash/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/trash
git commit -m "feat(trash): aggregate soft-deleted items/bundles/trips for trash view"
```

---

## Phase 3 — Authentication

### Task 12: Users repository (find-or-create)

**Files:**
- Create: `internal/auth/users.go`
- Create: `internal/auth/users_test.go`
- Create: `internal/auth/testdb.go`

- [ ] **Step 1: Test helper**

`internal/auth/testdb.go`:
```go
package auth

import (
	"database/sql"
	"path/filepath"
	"testing"

	pdb "github.com/bejl/packing-list/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}
```

- [ ] **Step 2: Failing tests**

`internal/auth/users_test.go`:
```go
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
```

- [ ] **Step 3: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 4: Implement**

`internal/auth/users.go`:
```go
// Package auth covers users, magic-link tokens, sessions, CSRF.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID    string
	Email string
}

type Users struct{ db *sql.DB }

func NewUsers(db *sql.DB) *Users { return &Users{db: db} }

// FindOrCreate returns (id, created, err). Match is case-insensitive on email.
func (u *Users) FindOrCreate(ctx context.Context, email string) (string, bool, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", false, errors.New("email required")
	}
	// Try existing first.
	var id string
	err := u.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? AND deleted_at IS NULL`, email).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	id = ids.New()
	_, err = u.db.ExecContext(ctx, `INSERT INTO users(id,email) VALUES (?,?)`, id, email)
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (u *Users) Get(ctx context.Context, id string) (User, error) {
	var usr User
	err := u.db.QueryRowContext(ctx, `SELECT id,email FROM users WHERE id = ? AND deleted_at IS NULL`, id).Scan(&usr.ID, &usr.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return usr, ErrNotFound
	}
	return usr, err
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 6: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): users find-or-create with case-insensitive email"
```

---

### Task 13: Mailer interface + log + smtp implementations

**Files:**
- Create: `internal/auth/mailer.go`
- Create: `internal/auth/mailer_test.go`

- [ ] **Step 1: Failing test for LogMailer**

`internal/auth/mailer_test.go`:
```go
package auth

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLogMailerWritesLink(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := &LogMailer{Logger: logger}
	if err := m.SendMagicLink(context.Background(), "alice@example.com", "https://app/auth/verify?t=abc"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if !strings.Contains(buf.String(), "alice@example.com") {
		t.Errorf("log missing email: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "https://app/auth/verify?t=abc") {
		t.Errorf("log missing link: %q", buf.String())
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/mailer.go`:
```go
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/bejl/packing-list/internal/config"
)

// Mailer abstracts magic-link delivery so we can swap log + smtp + tests.
type Mailer interface {
	SendMagicLink(ctx context.Context, email, link string) error
}

// LogMailer prints the link to slog at INFO. Use in dev / when SMTP unset.
type LogMailer struct {
	Logger *slog.Logger
}

func (m *LogMailer) SendMagicLink(_ context.Context, email, link string) error {
	m.Logger.Info("magic link issued (no SMTP configured)",
		"email", email, "link", link)
	return nil
}

// SMTPMailer sends a minimal RFC 5322 message via PLAIN SMTP AUTH.
type SMTPMailer struct {
	Cfg config.SMTP
}

func (m *SMTPMailer) SendMagicLink(_ context.Context, email, link string) error {
	addr := m.Cfg.Host + ":" + m.Cfg.Port
	body := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: Sign in to Packing List\r\n\r\nOpen this link to sign in. It expires in 15 minutes.\r\n\r\n%s\r\n",
		email, m.Cfg.From, link)
	var auth smtp.Auth
	if m.Cfg.User != "" {
		auth = smtp.PlainAuth("", m.Cfg.User, m.Cfg.Pass, m.Cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.Cfg.From, []string{email}, []byte(body))
}

// NewMailer returns SMTPMailer if SMTP is configured, otherwise LogMailer.
func NewMailer(cfg config.Config, logger *slog.Logger) Mailer {
	if cfg.SMTP.Configured() {
		return &SMTPMailer{Cfg: cfg.SMTP}
	}
	return &LogMailer{Logger: logger}
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): mailer interface with log + smtp implementations"
```

---

### Task 14: Magic-link tokens

**Files:**
- Create: `internal/auth/magic.go`
- Create: `internal/auth/magic_test.go`

- [ ] **Step 1: Failing tests**

`internal/auth/magic_test.go`:
```go
package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type captureMailer struct {
	gotEmail, gotLink string
}

func (c *captureMailer) SendMagicLink(_ context.Context, email, link string) error {
	c.gotEmail = email
	c.gotLink = link
	return nil
}

func TestIssueAndConsumeRoundtrip(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	mt := NewMagic(db, cap, "https://app", func() time.Time { return time.Now() })
	ctx := context.Background()

	if err := mt.Issue(ctx, "alice@example.com"); err != nil {
		t.Fatalf("issue: %v", err)
	}
	if cap.gotEmail != "alice@example.com" {
		t.Errorf("captured email: %s", cap.gotEmail)
	}
	// Extract token from link.
	const prefix = "https://app/auth/verify?t="
	if len(cap.gotLink) <= len(prefix) {
		t.Fatalf("bad link: %s", cap.gotLink)
	}
	token := cap.gotLink[len(prefix):]
	email, err := mt.Consume(ctx, token)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if email != "alice@example.com" {
		t.Errorf("consume email: %s", email)
	}
}

func TestConsumeRejectsReuse(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	mt := NewMagic(db, cap, "https://app", func() time.Time { return time.Now() })
	ctx := context.Background()
	_ = mt.Issue(ctx, "x@example.com")
	token := cap.gotLink[len("https://app/auth/verify?t="):]
	if _, err := mt.Consume(ctx, token); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := mt.Consume(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken on reuse, got %v", err)
	}
}

func TestConsumeRejectsExpired(t *testing.T) {
	db := newTestDB(t)
	cap := &captureMailer{}
	// Fixed clock at t0; expired tokens are >15 min old.
	t0 := time.Now()
	mt := NewMagic(db, cap, "https://app", func() time.Time { return t0 })
	ctx := context.Background()
	_ = mt.Issue(ctx, "x@example.com")
	token := cap.gotLink[len("https://app/auth/verify?t="):]

	mt.now = func() time.Time { return t0.Add(20 * time.Minute) }
	if _, err := mt.Consume(ctx, token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken after expiry, got %v", err)
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/magic.go`:
```go
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Magic struct {
	db      *sql.DB
	mailer  Mailer
	baseURL string
	now     func() time.Time
}

func NewMagic(db *sql.DB, mailer Mailer, baseURL string, now func() time.Time) *Magic {
	return &Magic{db: db, mailer: mailer, baseURL: baseURL, now: now}
}

// Issue creates a single-use 15-minute token and sends a link to email.
func (m *Magic) Issue(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email required")
	}
	// 32 random bytes -> base64url -> 43 chars.
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(token))
	id := ids.New()
	expires := m.now().Add(15 * time.Minute)
	if _, err := m.db.ExecContext(ctx,
		`INSERT INTO magic_tokens(id,email,token_hash,expires_at) VALUES (?,?,?,?)`,
		id, email, hash[:], expires); err != nil {
		return err
	}
	link := m.baseURL + "/auth/verify?t=" + token
	return m.mailer.SendMagicLink(ctx, email, link)
}

// Consume validates a token and marks it used. Returns the associated email.
func (m *Magic) Consume(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrInvalidToken
	}
	hash := sha256.Sum256([]byte(token))
	var (
		id, email string
		expires   time.Time
		used      sql.NullTime
	)
	err := m.db.QueryRowContext(ctx,
		`SELECT id,email,expires_at,used_at FROM magic_tokens WHERE token_hash = ?`, hash[:]).
		Scan(&id, &email, &expires, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", err
	}
	if used.Valid {
		return "", ErrInvalidToken
	}
	if m.now().After(expires) {
		return "", ErrInvalidToken
	}
	res, err := m.db.ExecContext(ctx,
		`UPDATE magic_tokens SET used_at = CURRENT_TIMESTAMP WHERE id = ? AND used_at IS NULL`, id)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", ErrInvalidToken
	}
	return email, nil
}

// PurgeExpired hard-deletes tokens older than now-1h. Safe to call periodically.
func (m *Magic) PurgeExpired(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx,
		`DELETE FROM magic_tokens WHERE expires_at < ?`, m.now().Add(-time.Hour))
	return err
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): single-use magic-link tokens, hashed in DB, 15-min TTL"
```

---

### Task 15: Sessions (HMAC-signed cookie, sliding renewal)

**Files:**
- Create: `internal/auth/session.go`
- Create: `internal/auth/session_test.go`

- [ ] **Step 1: Failing tests**

`internal/auth/session_test.go`:
```go
package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSessionsIssueLookup(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`); err != nil {
		t.Fatal(err)
	}
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	ctx := context.Background()

	cookieVal, err := s.Issue(ctx, "u_a")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	uid, err := s.Lookup(ctx, cookieVal)
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if uid != "u_a" {
		t.Errorf("uid: got %q", uid)
	}
}

func TestSessionsRejectTamperedCookie(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`); err != nil {
		t.Fatal(err)
	}
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	c, _ := s.Issue(context.Background(), "u_a")
	// Flip a byte.
	tampered := c[:len(c)-1] + "x"
	if _, err := s.Lookup(context.Background(), tampered); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession on tamper, got %v", err)
	}
}

func TestSessionsRevoke(t *testing.T) {
	db := newTestDB(t)
	db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`)
	s := NewSessions(db, []byte("test-secret-32-bytes-padding-xx"), func() time.Time { return time.Now() })
	c, _ := s.Issue(context.Background(), "u_a")
	if err := s.Revoke(context.Background(), c); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := s.Lookup(context.Background(), c); !errors.Is(err, ErrInvalidSession) {
		t.Error("expected lookup to fail after revoke")
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/session.go`:
```go
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var ErrInvalidSession = errors.New("invalid session")

const SessionTTL = 30 * 24 * time.Hour

type Sessions struct {
	db     *sql.DB
	secret []byte
	now    func() time.Time
}

func NewSessions(db *sql.DB, secret []byte, now func() time.Time) *Sessions {
	return &Sessions{db: db, secret: secret, now: now}
}

// Issue creates a session row and returns the signed cookie value: "<sid>.<hmac>".
func (s *Sessions) Issue(ctx context.Context, userID string) (string, error) {
	sid := ids.New()
	expires := s.now().Add(SessionTTL)
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions(id,user_id,expires_at) VALUES (?,?,?)`,
		sid, userID, expires); err != nil {
		return "", err
	}
	return s.sign(sid), nil
}

// Lookup parses the cookie, validates HMAC, returns user ID. Side-effect:
// extends expires_at if more than half the TTL has elapsed (sliding renewal).
func (s *Sessions) Lookup(ctx context.Context, cookieVal string) (string, error) {
	sid, ok := s.verify(cookieVal)
	if !ok {
		return "", ErrInvalidSession
	}
	var userID string
	var expires time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE id = ?`, sid).
		Scan(&userID, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidSession
	}
	if err != nil {
		return "", err
	}
	now := s.now()
	if now.After(expires) {
		return "", ErrInvalidSession
	}
	if expires.Sub(now) < SessionTTL/2 {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE sessions SET expires_at = ? WHERE id = ?`,
			now.Add(SessionTTL), sid)
	}
	return userID, nil
}

func (s *Sessions) Revoke(ctx context.Context, cookieVal string) error {
	sid, ok := s.verify(cookieVal)
	if !ok {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sid)
	return err
}

func (s *Sessions) PurgeExpired(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < ?`, s.now())
	return err
}

func (s *Sessions) sign(sid string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sid))
	return sid + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Sessions) verify(cookieVal string) (string, bool) {
	dot := strings.LastIndexByte(cookieVal, '.')
	if dot < 0 {
		return "", false
	}
	sid := cookieVal[:dot]
	got, err := base64.RawURLEncoding.DecodeString(cookieVal[dot+1:])
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sid))
	want := mac.Sum(nil)
	if !hmac.Equal(got, want) {
		return "", false
	}
	return sid, true
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): HMAC-signed session cookies with sliding 30-day TTL"
```

---

### Task 16: Session-secret bootstrap

The session HMAC needs a stable secret across restarts. If `SESSION_SECRET` is unset, generate one and persist to `$DATA_DIR/.session_secret`.

**Files:**
- Create: `internal/auth/secret.go`
- Create: `internal/auth/secret_test.go`

- [ ] **Step 1: Failing test**

`internal/auth/secret_test.go`:
```go
package auth

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateSecretPersists(t *testing.T) {
	dir := t.TempDir()
	a, err := LoadOrCreateSecret(dir, "")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if len(a) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(a))
	}
	b, err := LoadOrCreateSecret(dir, "")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Error("expected secret to be stable across calls")
	}
}

func TestLoadOrCreateSecretEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	// Hex of 32 bytes = 64 chars.
	envHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	got, err := LoadOrCreateSecret(dir, envHex)
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	if len(got) != 32 {
		t.Errorf("env path: expected 32 bytes, got %d", len(got))
	}
	// File must not have been created when env supplies it.
	_, err = readFileBytes(filepath.Join(dir, ".session_secret"))
	if err == nil {
		t.Error("file should not exist when SESSION_SECRET is provided")
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/secret.go`:
```go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadOrCreateSecret resolves the session secret in this order:
//  1. If envHex is non-empty, decode it (hex) and return.
//  2. Else read $dataDir/.session_secret if it exists.
//  3. Else generate 32 random bytes, write them, return.
func LoadOrCreateSecret(dataDir, envHex string) ([]byte, error) {
	if envHex != "" {
		b, err := hex.DecodeString(envHex)
		if err != nil {
			return nil, errors.New("SESSION_SECRET must be hex")
		}
		if len(b) < 16 {
			return nil, errors.New("SESSION_SECRET too short (need ≥16 bytes)")
		}
		return b, nil
	}
	path := filepath.Join(dataDir, ".session_secret")
	b, err := readFileBytes(path)
	if err == nil {
		decoded, err := hex.DecodeString(string(b))
		if err != nil {
			return nil, errors.New(".session_secret malformed")
		}
		return decoded, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	// Generate + write.
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(raw[:])), 0o600); err != nil {
		return nil, err
	}
	return raw[:], nil
}

// readFileBytes is exposed for tests in the same package.
func readFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): persist session secret to data dir when env unset"
```

---

### Task 17: Auth middleware + context helpers

**Files:**
- Create: `internal/auth/middleware.go`
- Create: `internal/auth/middleware_test.go`

- [ ] **Step 1: Failing test**

`internal/auth/middleware_test.go`:
```go
package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireUserRedirectsWhenAnonymous(t *testing.T) {
	db := newTestDB(t)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = UserFrom(r.Context()) // would normally use this
		w.WriteHeader(http.StatusOK)
	})
	mw := RequireUser(s)(h)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected 303, got %d", w.Code)
	}
}

func TestRequireUserAllowsWithValidCookie(t *testing.T) {
	db := newTestDB(t)
	db.Exec(`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })
	cookieVal, _ := s.Issue(nil, "u_a")

	hit := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		if got := UserFrom(r.Context()); got != "u_a" {
			t.Errorf("ctx user: %q", got)
		}
	})
	mw := RequireUser(s)(h)
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: CookieName, Value: cookieVal})
	mw.ServeHTTP(httptest.NewRecorder(), r)
	if !hit {
		t.Error("expected handler to be called")
	}
}

func TestRequireUserReturns401ForHTMX(t *testing.T) {
	db := newTestDB(t)
	s := NewSessions(db, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), func() time.Time { return time.Now() })
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	RequireUser(s)(h).ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for HTMX, got %d", w.Code)
	}
	if w.Header().Get("HX-Redirect") != "/login" {
		t.Errorf("expected HX-Redirect, got %q", w.Header().Get("HX-Redirect"))
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/middleware.go`:
```go
package auth

import (
	"context"
	"net/http"
)

const CookieName = "sid"

type ctxKey string

const userKey ctxKey = "uid"

func UserFrom(ctx context.Context) string {
	v, _ := ctx.Value(userKey).(string)
	return v
}

func WithUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

func RequireUser(s *Sessions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CookieName)
			if err != nil {
				unauthorized(w, r)
				return
			}
			uid, err := s.Lookup(r.Context(), c.Value)
			if err != nil {
				// Clear bad cookie so the browser stops sending it.
				http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", MaxAge: -1, Path: "/"})
				unauthorized(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), uid)))
		})
	}
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): RequireUser middleware with HTMX-aware redirect"
```

---

### Task 18: CSRF middleware

Per-session cookie + matching header (HTMX `hx-headers`) or hidden form field.

**Files:**
- Create: `internal/auth/csrf.go`
- Create: `internal/auth/csrf_test.go`

- [ ] **Step 1: Failing test**

`internal/auth/csrf_test.go`:
```go
package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFGetSetsCookie(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	w := httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Fatal("handler should be called for GET")
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), CSRFCookieName+"=") {
		t.Errorf("expected csrf cookie set, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestCSRFPostRequiresHeaderMatch(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	// No cookie + no header -> 403.
	w := httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, httptest.NewRequest("POST", "/", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 without csrf, got %d", w.Code)
	}

	// Cookie + matching header -> ok.
	r := httptest.NewRequest("POST", "/", nil)
	r.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "tok"})
	r.Header.Set(CSRFHeaderName, "tok")
	w = httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid csrf, got %d", w.Code)
	}

	// Cookie + form value match -> ok.
	r = httptest.NewRequest("POST", "/", strings.NewReader("csrf_token=tok"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: "tok"})
	w = httptest.NewRecorder()
	CSRF(h).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with form csrf, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/csrf.go`:
```go
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	CSRFCookieName = "csrf"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFFormField  = "csrf_token"
)

var safeMethods = map[string]bool{
	"GET": true, "HEAD": true, "OPTIONS": true,
}

// CSRF issues a CSRF cookie on safe methods and validates it on unsafe ones.
// The token is doubled in cookie + header/form; we compare with constant time.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CSRFCookieName)
		if err != nil || c.Value == "" {
			var raw [24]byte
			rand.Read(raw[:])
			tok := base64.RawURLEncoding.EncodeToString(raw[:])
			http.SetCookie(w, &http.Cookie{
				Name:     CSRFCookieName,
				Value:    tok,
				Path:     "/",
				HttpOnly: false, // readable so HTMX can echo into header
				SameSite: http.SameSiteLaxMode,
			})
			c = &http.Cookie{Name: CSRFCookieName, Value: tok}
		}
		if safeMethods[r.Method] {
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get(CSRFHeaderName)
		if got == "" {
			// Form fallback. Parse body once; downstream code re-parses freely.
			if err := r.ParseForm(); err == nil {
				got = r.FormValue(CSRFFormField)
			}
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(c.Value)) != 1 {
			http.Error(w, "csrf failure", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): CSRF middleware (cookie + header/form double-submit)"
```

---

### Task 19: Login rate limit

In-memory token bucket per (email, IP). Tiny — no Redis. Resets on restart.

**Files:**
- Create: `internal/auth/ratelimit.go`
- Create: `internal/auth/ratelimit_test.go`

- [ ] **Step 1: Failing test**

`internal/auth/ratelimit_test.go`:
```go
package auth

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUntilCapHit(t *testing.T) {
	now := time.Unix(0, 0)
	rl := NewRateLimiter(3, time.Minute, func() time.Time { return now })
	for i := 0; i < 3; i++ {
		if !rl.Allow("k") {
			t.Fatalf("call %d should be allowed", i)
		}
	}
	if rl.Allow("k") {
		t.Fatal("4th call should be denied")
	}
	now = now.Add(2 * time.Minute)
	if !rl.Allow("k") {
		t.Fatal("after window elapsed, should be allowed")
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/auth/...
```

- [ ] **Step 3: Implement**

`internal/auth/ratelimit.go`:
```go
package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	cap    int
	window time.Duration
	now    func() time.Time
	hits   map[string][]time.Time
}

func NewRateLimiter(cap int, window time.Duration, now func() time.Time) *RateLimiter {
	return &RateLimiter{cap: cap, window: window, now: now, hits: map[string][]time.Time{}}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := r.now().Add(-r.window)
	kept := r.hits[key][:0]
	for _, t := range r.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= r.cap {
		r.hits[key] = kept
		return false
	}
	kept = append(kept, r.now())
	r.hits[key] = kept
	return true
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/auth/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/auth
git commit -m "feat(auth): in-memory rate limiter (token-bucket style)"
```

---

## Phase 4 — Web layer

### Task 20: Static assets (Pico, HTMX, app CSS)

**Files:**
- Create: `internal/web/static/pico.min.css` (vendored)
- Create: `internal/web/static/htmx.min.js` (vendored)
- Create: `internal/web/static/app.css`
- Create: `internal/web/static/favicon.svg`

- [ ] **Step 1: Download Pico v2 + HTMX 2 (pinned)**

```bash
# Run once at root; commit the result.
curl -fsSL https://cdn.jsdelivr.net/npm/@picocss/pico@2.0.6/css/pico.classless.min.css \
  -o internal/web/static/pico.min.css
curl -fsSL https://cdn.jsdelivr.net/npm/htmx.org@2.0.3/dist/htmx.min.js \
  -o internal/web/static/htmx.min.js
```

Expected: both files non-empty, no HTML body in output (no error page).

- [ ] **Step 2: App CSS overlay**

`internal/web/static/app.css`:
```css
:root { --app-accent: #2d6cdf; }
body > nav.app-bar { display:flex; gap:1rem; align-items:center; padding:.5rem 1rem; border-bottom:1px solid var(--pico-muted-border-color); }
body > nav.app-bar .grow { flex:1 }
main.container { padding-block:1rem; }

.progress-bar { height:.5rem; background:var(--pico-muted-border-color); border-radius:.25rem; overflow:hidden; }
.progress-bar > span { display:block; height:100%; background:var(--app-accent); transition:width .15s ease; }

.bundle-chip { display:inline-flex; align-items:center; gap:.4rem; padding:.15rem .5rem; border-radius:1rem; background:var(--pico-muted-border-color); margin:.15rem .25rem .15rem 0; font-size:.85rem; }
.bundle-chip button { all:unset; cursor:pointer; }

.pack-row { display:grid; grid-template-columns: 1.5rem 1fr auto auto; align-items:center; gap:.5rem; padding:.25rem 0; }
.pack-row.packed { opacity:.55; text-decoration:line-through; }
.pack-row .source { font-size:.75rem; color:var(--pico-muted-color); }

.category-heading { margin-top:1rem; font-size:.75rem; text-transform:uppercase; letter-spacing:.06em; color:var(--pico-muted-color); }

@media (max-width: 640px) {
  main.container { padding-inline:.5rem; }
  .pack-row { grid-template-columns: 1.5rem 1fr auto; }
  .pack-row .source { display:none; }
}
```

- [ ] **Step 3: Favicon (minimal SVG)**

`internal/web/static/favicon.svg`:
```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#2d6cdf"/><path d="M9 22V13a3 3 0 0 1 3-3h8a3 3 0 0 1 3 3v9" stroke="#fff" stroke-width="2" fill="none"/><path d="M12 10V8a4 4 0 0 1 8 0v2" stroke="#fff" stroke-width="2" fill="none"/></svg>
```

- [ ] **Step 4: Commit**

```bash
git add internal/web/static
git commit -m "chore(web): vendor pico.css 2.0.6 + htmx 2.0.3 + app overlay"
```

---

### Task 21: Base templates and partials

**Files:**
- Create: `internal/web/templates/layout.html`
- Create: `internal/web/templates/pages/login.html`
- Create: `internal/web/templates/pages/login_sent.html`
- Create: `internal/web/templates/partials/flash.html`

- [ ] **Step 1: Layout**

`internal/web/templates/layout.html`:
```html
{{ define "layout" -}}
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>{{ .Title }} · Packing List</title>
<link rel="icon" type="image/svg+xml" href="/static/favicon.svg" />
<link rel="stylesheet" href="/static/pico.min.css" />
<link rel="stylesheet" href="/static/app.css" />
<script src="/static/htmx.min.js" defer></script>
<meta name="htmx-config" content='{"includeIndicatorStyles":false}' />
</head>
<body hx-headers='{"X-CSRF-Token": "__CSRF__"}'>
{{ if .User }}
<nav class="app-bar">
  <strong><a href="/" class="contrast">Packing List</a></strong>
  <div class="grow"></div>
  <a href="/items">Items</a>
  <a href="/bundles">Bundles</a>
  <a href="/trash">Trash</a>
  <form method="post" action="/logout" style="display:inline;margin:0">
    <input type="hidden" name="csrf_token" value="__CSRF__" />
    <button class="outline" style="padding:.25rem .75rem">Sign out</button>
  </form>
</nav>
{{ end }}
<main class="container">
{{ template "content" . }}
</main>
<script>
  // Echo the csrf cookie into the hx-headers placeholder + any forms.
  document.addEventListener("DOMContentLoaded", () => {
    const m = document.cookie.match(/(?:^|; )csrf=([^;]+)/);
    if (!m) return;
    const tok = decodeURIComponent(m[1]);
    document.querySelectorAll("[hx-headers]").forEach(el => {
      el.setAttribute("hx-headers", el.getAttribute("hx-headers").replaceAll("__CSRF__", tok));
    });
    document.querySelectorAll("input[name=csrf_token]").forEach(el => el.value = tok);
  });
</script>
</body>
</html>
{{- end }}
```

- [ ] **Step 2: Login page**

`internal/web/templates/pages/login.html`:
```html
{{ define "content" }}
<article style="max-width:24rem;margin:3rem auto">
  <header><h1>Sign in</h1></header>
  {{ if .Error }}<p style="color:var(--pico-color-red-550)">{{ .Error }}</p>{{ end }}
  <form method="post" action="/login">
    <input type="hidden" name="csrf_token" value="__CSRF__" />
    <label>Email
      <input type="email" name="email" required autocomplete="email" autofocus />
    </label>
    <button type="submit">Send magic link</button>
  </form>
  <small>We'll email you a link that signs you in. No password.</small>
</article>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 3: Login-sent page**

`internal/web/templates/pages/login_sent.html`:
```html
{{ define "content" }}
<article style="max-width:24rem;margin:3rem auto">
  <header><h1>Check your inbox</h1></header>
  <p>Sent a link to <strong>{{ .Email }}</strong>. It expires in 15 minutes.</p>
  {{ if .DevLink }}<p><small>Dev: <a href="{{ .DevLink }}">{{ .DevLink }}</a></small></p>{{ end }}
</article>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 4: Flash partial**

`internal/web/templates/partials/flash.html`:
```html
{{ define "flash" }}
{{ if . }}<p role="alert" style="color:var(--pico-color-red-550)">{{ . }}</p>{{ end }}
{{ end }}
```

- [ ] **Step 5: Commit**

```bash
git add internal/web/templates
git commit -m "feat(web): layout + login templates + flash partial"
```

---

### Task 22: Template registry + render helpers

**Files:**
- Create: `internal/web/render.go`
- Create: `internal/web/render_test.go`

- [ ] **Step 1: Failing test**

`internal/web/render_test.go`:
```go
package web

import (
	"bytes"
	"strings"
	"testing"
)

func TestRendererRendersLogin(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, "login", map[string]any{"Title": "Sign in"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "Send magic link") {
		t.Errorf("expected login button text, got: %s", s)
	}
	if !strings.Contains(s, "<title>Sign in") {
		t.Errorf("expected title, got: %s", s)
	}
}
```

- [ ] **Step 2: Run, confirm fail**

```bash
go test ./internal/web/...
```

- [ ] **Step 3: Implement**

`internal/web/render.go`:
```go
// Package web implements the HTTP layer: routing, middleware, handlers, templates.
package web

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html templates/pages/*.html templates/partials/*.html
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"isZero":   func(v any) bool { return v == nil || v == "" || v == 0 },
		"plus":     func(a, b int) int { return a + b },
	}

	// Parse layout + partials into a base template set.
	base := template.New("").Funcs(funcs)
	for _, glob := range []string{"templates/layout.html", "templates/partials/*.html"} {
		matches, err := filepath.Glob(glob) // unused except as sanity check
		_ = matches
		_ = err
	}
	base, err := base.ParseFS(templatesFS, "templates/layout.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse base: %w", err)
	}

	pages := map[string]*template.Template{}
	entries, err := templatesFS.ReadDir("templates/pages")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		cloned, err := base.Clone()
		if err != nil {
			return nil, err
		}
		t, err := cloned.ParseFS(templatesFS, "templates/pages/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		pages[name] = t
	}
	return &Renderer{pages: pages, partials: base}, nil
}

// Render writes a page template by name.
func (r *Renderer) Render(w io.Writer, page string, data any) error {
	t, ok := r.pages[page]
	if !ok {
		return errors.New("unknown page: " + page)
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// Partial renders a partial template by name (for HTMX fragment responses).
func (r *Renderer) Partial(w io.Writer, name string, data any) error {
	return r.partials.ExecuteTemplate(w, name, data)
}

// StaticFS returns the embedded /static FS for use with http.FileServer.
func StaticFS() embed.FS { return staticFS }
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/web
git commit -m "feat(web): template renderer with embedded layout + pages + partials"
```

---

### Task 23: HTTP handler dependencies + router

This task wires every domain repository into a single `Server` struct used by all handler files. Routes themselves are stubbed (`501 not implemented`) and replaced in subsequent tasks.

**Files:**
- Create: `internal/web/server.go`

- [ ] **Step 1: Implement**

`internal/web/server.go`:
```go
package web

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
)

// Server bundles all dependencies for HTTP handlers.
type Server struct {
	Cfg        config.Config
	Logger     *slog.Logger
	Renderer   *Renderer
	Users      *auth.Users
	Sessions   *auth.Sessions
	Magic      *auth.Magic
	RateLimit  *auth.RateLimiter
	Items      *catalog.Items
	Bundles    *catalog.Bundles
	Trips      *trips.Trips
	Sources    *trips.Sources
	Pack       *trips.Pack
	Renderer2  *trips.Renderer
	Trash      *trash.View
	IsDev      bool // controls whether log-mailer link is shown in /login response
	Now        func() time.Time
}

// Handler builds the http.Handler with routes + middleware stack.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Static files (no auth needed).
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS{}))))

	// Anonymous routes.
	mux.HandleFunc("GET /login", s.getLogin)
	mux.HandleFunc("POST /login", s.postLogin)
	mux.HandleFunc("GET /auth/verify", s.getVerify)
	mux.HandleFunc("POST /logout", s.postLogout)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); w.Write([]byte("ok")) })

	// Authed routes.
	authed := http.NewServeMux()
	authed.HandleFunc("GET /{$}", s.getTripsIndex)

	authed.HandleFunc("GET /trips/new", s.getTripNew)
	authed.HandleFunc("POST /trips", s.postTripCreate)
	authed.HandleFunc("GET /trips/{id}", s.getTripDetail)
	authed.HandleFunc("PATCH /trips/{id}", s.patchTrip)
	authed.HandleFunc("DELETE /trips/{id}", s.deleteTrip)
	authed.HandleFunc("POST /trips/{id}/bundles", s.attachBundle)
	authed.HandleFunc("DELETE /trips/{id}/bundles/{bid}", s.detachBundle)
	authed.HandleFunc("POST /trips/{id}/extras", s.addExtra)
	authed.HandleFunc("PATCH /trips/{id}/items/{iid}", s.overrideItem)
	authed.HandleFunc("POST /trips/{id}/pack/{iid}", s.togglePack)
	authed.HandleFunc("POST /trips/{id}/members", s.inviteMember)
	authed.HandleFunc("DELETE /trips/{id}/members/{uid}", s.removeMember)

	authed.HandleFunc("GET /items", s.listItems)
	authed.HandleFunc("POST /items", s.createItem)
	authed.HandleFunc("PATCH /items/{id}", s.updateItem)
	authed.HandleFunc("DELETE /items/{id}", s.deleteItem)

	authed.HandleFunc("GET /bundles", s.listBundles)
	authed.HandleFunc("POST /bundles", s.createBundle)
	authed.HandleFunc("GET /bundles/{id}", s.editBundle)
	authed.HandleFunc("PATCH /bundles/{id}", s.updateBundle)
	authed.HandleFunc("DELETE /bundles/{id}", s.deleteBundle)
	authed.HandleFunc("POST /bundles/{id}/items", s.bundleAddItem)
	authed.HandleFunc("DELETE /bundles/{id}/items/{iid}", s.bundleRemoveItem)
	authed.HandleFunc("POST /bundles/{id}/children", s.bundleAddChild)
	authed.HandleFunc("DELETE /bundles/{id}/children/{cid}", s.bundleRemoveChild)

	authed.HandleFunc("GET /trash", s.getTrash)
	authed.HandleFunc("POST /trash/{kind}/{id}/restore", s.restore)
	authed.HandleFunc("DELETE /trash/{kind}/{id}", s.purge)

	authed.HandleFunc("GET /export", s.export)
	authed.HandleFunc("POST /import", s.importJSON)

	mux.Handle("/", auth.RequireUser(s.Sessions)(authed))

	// Outer middleware: CSRF (covers everything), then access log.
	return s.accessLog(auth.CSRF(mux))
}

// subFS lets http.FileServer serve files from the embedded static/ directory
// without exposing the "static" path prefix.
type subFS struct{}

func (subFS) Open(name string) (http.File, error) {
	f, err := staticFS.Open("static/" + name)
	if err != nil {
		return nil, err
	}
	return staticFile{f}, nil
}

type staticFile struct {
	f interface {
		Read([]byte) (int, error)
		Close() error
	}
}

// (Implementation continues in subsequent tasks; this file compiles only after
// the per-handler tasks add the corresponding methods.)
```

> **Note:** This file references handler methods (`s.getLogin`, etc.) that don't exist yet. The package will not build until Tasks 24+ are completed. That's expected — each subsequent task adds its handlers and re-runs the build.

- [ ] **Step 2: Confirm it does not yet build (sanity)**

```bash
go build ./internal/web/...
```

Expected: errors of the form `s.getLogin undefined`. We'll close them in upcoming tasks.

- [ ] **Step 3: Commit the scaffold so subsequent tasks have a clear baseline**

```bash
git add internal/web/server.go
git commit -m "feat(web): server scaffold + route table (handlers stubbed)"
```

---

### Task 24: Auth handlers (login, verify, logout)

**Files:**
- Create: `internal/web/handlers_auth.go`
- Create: `internal/web/handlers_auth_test.go`

- [ ] **Step 1: Implement**

`internal/web/handlers_auth.go`:
```go
package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) getLogin(w http.ResponseWriter, r *http.Request) {
	s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in"})
}

func (s *Server) postLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Email is required"})
		return
	}
	keyEmail := strings.ToLower(email)
	keyIP := r.RemoteAddr
	if !s.RateLimit.Allow(keyEmail) || !s.RateLimit.Allow(keyIP) {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Too many attempts. Try again later."})
		return
	}
	if _, _, err := s.Users.FindOrCreate(r.Context(), email); err != nil {
		s.Logger.Error("findOrCreate", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not start sign-in"})
		return
	}
	if err := s.Magic.Issue(r.Context(), email); err != nil {
		s.Logger.Error("issue magic link", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not send sign-in email"})
		return
	}
	data := map[string]any{"Title": "Check inbox", "Email": email}
	if s.IsDev {
		data["DevLink"] = "(see server log)"
	}
	s.Renderer.Render(w, "login_sent", data)
}

func (s *Server) getVerify(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("t")
	email, err := s.Magic.Consume(r.Context(), tok)
	if err != nil {
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Link is invalid or expired"})
		return
	}
	uid, _, err := s.Users.FindOrCreate(r.Context(), email)
	if err != nil {
		s.Logger.Error("findOrCreate on verify", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not finish sign-in"})
		return
	}
	cookieVal, err := s.Sessions.Issue(r.Context(), uid)
	if err != nil {
		s.Logger.Error("issue session", "err", err)
		s.Renderer.Render(w, "login", map[string]any{"Title": "Sign in", "Error": "Could not start session"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    cookieVal,
		Path:     "/",
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.Cfg.BaseURL, "https://"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.SessionTTL.Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) postLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.CookieName); err == nil {
		_ = s.Sessions.Revoke(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: "", MaxAge: -1, Path: "/"})
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// accessLog logs method + path + status + duration.
func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &statusRW{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		s.Logger.Info("http", "method", r.Method, "path", r.URL.Path, "status", rw.status)
	})
}

type statusRW struct {
	http.ResponseWriter
	status int
}

func (s *statusRW) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// guard for the unused error import path the compiler is otherwise happy.
var _ = errors.New
```

- [ ] **Step 2: Handler test**

`internal/web/handlers_auth_test.go`:
```go
package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	users := auth.NewUsers(d)
	now := func() time.Time { return time.Now() }
	mailer := &auth.LogMailer{Logger: logger}
	cfg := config.Config{BaseURL: "http://test"}
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	s := &Server{
		Cfg:       cfg,
		Logger:    logger,
		Renderer:  r,
		Users:     users,
		Sessions:  auth.NewSessions(d, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), now),
		Magic:     auth.NewMagic(d, mailer, cfg.BaseURL, now),
		RateLimit: auth.NewRateLimiter(10, time.Minute, now),
		Items:     catalog.NewItems(d),
		Bundles:   catalog.NewBundles(d),
		Trips:     trips.NewTrips(d),
		Sources:   trips.NewSources(d),
		Pack:      trips.NewPack(d),
		Renderer2: trips.NewRenderer(d),
		IsDev:     true,
		Now:       now,
	}
	s.Trash = trash.NewView(d, s.Items, s.Bundles, s.Trips)
	return s
}

func TestPostLoginCreatesUserAndShowsSentPage(t *testing.T) {
	s := newTestServer(t)
	form := strings.NewReader("email=alice@example.com&csrf_token=t")
	r := httptest.NewRequest("POST", "/login", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: auth.CSRFCookieName, Value: "t"})
	r.Header.Set(auth.CSRFHeaderName, "t")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	body := w.Body.String()
	if !strings.Contains(body, "Check your inbox") {
		t.Errorf("expected login_sent page, got: %s", body)
	}
	// User should exist now.
	_, created, err := s.Users.FindOrCreate(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if created {
		t.Error("expected user already created by login")
	}
}

func TestGetLoginRendersForm(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/login", nil))
	if !strings.Contains(w.Body.String(), "Send magic link") {
		t.Errorf("login page: %s", w.Body.String())
	}
}
```

- [ ] **Step 3: Run, confirm pass**

```bash
go test ./internal/web/...
```

Expected: pass once `handlers_auth.go` and `server.go` compile together (other handlers still missing — comment out their refs in `server.go` temporarily if needed, or stub them as no-op methods in a `handlers_stubs.go` file).

- [ ] **Step 4: Add stubs for not-yet-implemented handlers**

Create `internal/web/handlers_stubs.go` with no-op stubs (501) for every handler referenced in `server.go` that isn't implemented yet. Each subsequent task will move its handler out of this file and implement it for real. The stub file looks like:

```go
package web

import "net/http"

func notImpl(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

func (s *Server) getTripsIndex(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) getTripNew(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) postTripCreate(w http.ResponseWriter, r *http.Request)       { notImpl(w, r) }
func (s *Server) getTripDetail(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) patchTrip(w http.ResponseWriter, r *http.Request)            { notImpl(w, r) }
func (s *Server) deleteTrip(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) attachBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) detachBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) addExtra(w http.ResponseWriter, r *http.Request)             { notImpl(w, r) }
func (s *Server) overrideItem(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) togglePack(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) inviteMember(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) removeMember(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) listItems(w http.ResponseWriter, r *http.Request)            { notImpl(w, r) }
func (s *Server) createItem(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) updateItem(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) listBundles(w http.ResponseWriter, r *http.Request)          { notImpl(w, r) }
func (s *Server) createBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) editBundle(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) updateBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) deleteBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) bundleAddItem(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) bundleRemoveItem(w http.ResponseWriter, r *http.Request)     { notImpl(w, r) }
func (s *Server) bundleAddChild(w http.ResponseWriter, r *http.Request)       { notImpl(w, r) }
func (s *Server) bundleRemoveChild(w http.ResponseWriter, r *http.Request)    { notImpl(w, r) }
func (s *Server) getTrash(w http.ResponseWriter, r *http.Request)             { notImpl(w, r) }
func (s *Server) restore(w http.ResponseWriter, r *http.Request)              { notImpl(w, r) }
func (s *Server) purge(w http.ResponseWriter, r *http.Request)                { notImpl(w, r) }
func (s *Server) export(w http.ResponseWriter, r *http.Request)               { notImpl(w, r) }
func (s *Server) importJSON(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
```

Re-run `go test ./internal/web/...` — expected: pass.

- [ ] **Step 5: Commit**

```bash
git add internal/web
git commit -m "feat(web): auth handlers (login/verify/logout) with rate limit"
```

---

### Task 25: Items handlers

For each handler in this task, **delete the matching stub** from `handlers_stubs.go` as you implement it (otherwise duplicate method).

**Files:**
- Create: `internal/web/handlers_items.go`
- Create: `internal/web/templates/pages/items.html`
- Create: `internal/web/templates/partials/item_row.html`
- Modify: `internal/web/handlers_stubs.go` (remove `listItems`, `createItem`, `updateItem`, `deleteItem`)

- [ ] **Step 1: Items page template**

`internal/web/templates/pages/items.html`:
```html
{{ define "content" }}
<header>
  <h1>Items</h1>
  <p>The shared catalog. Per-night items scale by trip nights; fixed items use the larger of the requested quantities.</p>
</header>

<form hx-post="/items" hx-target="#item-rows" hx-swap="afterbegin" hx-on::after-request="this.reset()">
  <fieldset role="group">
    <input name="name" placeholder="Name" required />
    <input name="category" placeholder="Category" value="general" />
    <input name="default_qty" type="number" min="1" value="1" style="width:6rem" />
    <label style="display:inline-flex;align-items:center;gap:.25rem;margin:0 .5rem">
      <input type="checkbox" name="per_night" value="1" /> per night
    </label>
    <button type="submit">Add</button>
  </fieldset>
</form>

<table>
  <thead><tr><th>Name</th><th>Category</th><th>Per night</th><th>Default qty</th><th></th></tr></thead>
  <tbody id="item-rows">
    {{ range .Items }}{{ template "item_row" . }}{{ end }}
  </tbody>
</table>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 2: Item row partial**

`internal/web/templates/partials/item_row.html`:
```html
{{ define "item_row" }}
<tr id="item-{{ .ID }}">
  <td>{{ .Name }}</td>
  <td>{{ .Category }}</td>
  <td>{{ if .PerNight }}yes{{ else }}—{{ end }}</td>
  <td>{{ .DefaultQty }}</td>
  <td>
    <button class="outline" hx-delete="/items/{{ .ID }}" hx-target="#item-{{ .ID }}" hx-swap="outerHTML"
            hx-confirm="Move {{ .Name }} to Trash?">Delete</button>
  </td>
</tr>
{{ end }}
```

- [ ] **Step 3: Implement handlers**

`internal/web/handlers_items.go`:
```go
package web

import (
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
)

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.Items.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "items", map[string]any{
		"Title": "Items",
		"User":  auth.UserFrom(r.Context()),
		"Items": items,
	})
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	qty, _ := strconv.Atoi(r.FormValue("default_qty"))
	if qty <= 0 {
		qty = 1
	}
	it := catalog.Item{
		Name:       r.FormValue("name"),
		Category:   r.FormValue("category"),
		PerNight:   r.FormValue("per_night") == "1",
		DefaultQty: qty,
		CreatedBy:  auth.UserFrom(r.Context()),
	}
	id, err := s.Items.Create(r.Context(), it)
	if err != nil {
		http.Error(w, "could not create", 400)
		return
	}
	created, _ := s.Items.Get(r.Context(), id)
	s.Renderer.Partial(w, "item_row", created)
}

func (s *Server) updateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	r.ParseForm()
	qty, _ := strconv.Atoi(r.FormValue("default_qty"))
	if qty <= 0 {
		qty = 1
	}
	if err := s.Items.Update(r.Context(), id, catalog.Item{
		Name:       r.FormValue("name"),
		Category:   r.FormValue("category"),
		PerNight:   r.FormValue("per_night") == "1",
		DefaultQty: qty,
	}); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	updated, _ := s.Items.Get(r.Context(), id)
	s.Renderer.Partial(w, "item_row", updated)
}

func (s *Server) deleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.Items.SoftDelete(r.Context(), id, auth.UserFrom(r.Context())); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK) // empty body → HTMX removes the row
}
```

- [ ] **Step 4: Remove duplicate stubs**

Open `handlers_stubs.go` and delete the four lines for `listItems`, `createItem`, `updateItem`, `deleteItem`.

- [ ] **Step 5: Test**

`internal/web/handlers_items_test.go`:
```go
package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/auth"
)

func authedRequest(t *testing.T, s *Server, method, target string, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	// Create user + session.
	uid, _, _ := s.Users.FindOrCreate(context.Background(), "u@example.com")
	cookieVal, _ := s.Sessions.Issue(context.Background(), uid)
	r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: cookieVal})
	r.AddCookie(&http.Cookie{Name: auth.CSRFCookieName, Value: "tok"})
	r.Header.Set(auth.CSRFHeaderName, "tok")
	return r
}

func TestItemsCreateAndDelete(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/items", "name=Toothbrush&category=toiletries&default_qty=1"))
	if w.Code != 200 {
		t.Fatalf("create: status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("expected row html, got %s", w.Body.String())
	}
	// Fetch list, confirm visible.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/items", ""))
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("items page missing item: %s", w.Body.String())
	}
}
```

- [ ] **Step 6: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 7: Commit**

```bash
git add internal/web
git commit -m "feat(web): items CRUD handlers with HTMX row swaps"
```

---

### Task 26: Bundles handlers (list, create, edit, items, nested children)

**Files:**
- Create: `internal/web/handlers_bundles.go`
- Create: `internal/web/templates/pages/bundles.html`
- Create: `internal/web/templates/pages/bundle_edit.html`
- Create: `internal/web/templates/partials/bundle_row.html`
- Create: `internal/web/templates/partials/bundle_item_row.html`
- Create: `internal/web/templates/partials/bundle_child_row.html`
- Modify: `internal/web/handlers_stubs.go` (remove the nine bundle stubs: `listBundles`, `createBundle`, `editBundle`, `updateBundle`, `deleteBundle`, `bundleAddItem`, `bundleRemoveItem`, `bundleAddChild`, `bundleRemoveChild`)

- [ ] **Step 1: Bundles list page**

`internal/web/templates/pages/bundles.html`:
```html
{{ define "content" }}
<header>
  <h1>Bundles</h1>
  <p>Reusable groups of items. Bundles can also nest other bundles (cycles are rejected).</p>
</header>

<form hx-post="/bundles" hx-target="#bundle-rows" hx-swap="afterbegin" hx-on::after-request="this.reset()">
  <fieldset role="group">
    <input name="name" placeholder="Bundle name, e.g. washbag-basic" required />
    <input name="description" placeholder="Short description (optional)" />
    <button type="submit">Add</button>
  </fieldset>
</form>

<table>
  <thead><tr><th>Name</th><th>Description</th><th></th></tr></thead>
  <tbody id="bundle-rows">
    {{ range .Bundles }}{{ template "bundle_row" . }}{{ end }}
  </tbody>
</table>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 2: Bundle row partial**

`internal/web/templates/partials/bundle_row.html`:
```html
{{ define "bundle_row" }}
<tr id="bundle-{{ .ID }}">
  <td><a href="/bundles/{{ .ID }}">{{ .Name }}</a></td>
  <td>{{ .Description }}</td>
  <td>
    <button class="outline" hx-delete="/bundles/{{ .ID }}" hx-target="#bundle-{{ .ID }}"
            hx-swap="outerHTML" hx-confirm="Move {{ .Name }} to Trash?">Delete</button>
  </td>
</tr>
{{ end }}
```

- [ ] **Step 3: Bundle edit page**

`internal/web/templates/pages/bundle_edit.html`:
```html
{{ define "content" }}
<header>
  <h1>{{ .Bundle.Name }}</h1>
  <p><a href="/bundles">← all bundles</a></p>
</header>

<form hx-patch="/bundles/{{ .Bundle.ID }}" hx-swap="none">
  <fieldset role="group">
    <input name="name" value="{{ .Bundle.Name }}" required />
    <input name="description" value="{{ .Bundle.Description }}" placeholder="Description" />
    <button>Save</button>
  </fieldset>
</form>

<section class="grid">
  <div>
    <h2>Items</h2>
    <form hx-post="/bundles/{{ .Bundle.ID }}/items" hx-target="#bundle-items" hx-swap="afterbegin"
          hx-on::after-request="this.reset()">
      <fieldset role="group">
        <select name="item_id" required>
          <option value="">— pick an item —</option>
          {{ range .AllItems }}<option value="{{ .ID }}">{{ .Name }}</option>{{ end }}
        </select>
        <input name="qty" type="number" placeholder="qty (blank = default)" style="width:8rem" />
        <button>Add</button>
      </fieldset>
    </form>
    <ul id="bundle-items">
      {{ range .Items }}{{ template "bundle_item_row" . }}{{ end }}
    </ul>
  </div>

  <div>
    <h2>Nested bundles</h2>
    <form hx-post="/bundles/{{ .Bundle.ID }}/children" hx-target="#bundle-children" hx-swap="afterbegin"
          hx-on::after-request="this.reset()">
      <fieldset role="group">
        <select name="child_id" required>
          <option value="">— pick a bundle —</option>
          {{ range .NestableBundles }}<option value="{{ .ID }}">{{ .Name }}</option>{{ end }}
        </select>
        <button>Nest</button>
      </fieldset>
    </form>
    <ul id="bundle-children">
      {{ range .Children }}{{ template "bundle_child_row" . }}{{ end }}
    </ul>
  </div>
</section>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 4: Row partials**

`internal/web/templates/partials/bundle_item_row.html`:
```html
{{ define "bundle_item_row" }}
<li id="bi-{{ .BundleID }}-{{ .ItemID }}">
  {{ .Name }}{{ if .QtyShown }} × {{ .QtyShown }}{{ end }}
  <button class="outline" style="padding:.1rem .5rem"
          hx-delete="/bundles/{{ .BundleID }}/items/{{ .ItemID }}"
          hx-target="#bi-{{ .BundleID }}-{{ .ItemID }}" hx-swap="outerHTML">×</button>
</li>
{{ end }}
```

`internal/web/templates/partials/bundle_child_row.html`:
```html
{{ define "bundle_child_row" }}
<li id="bc-{{ .ParentID }}-{{ .ChildID }}">
  <a href="/bundles/{{ .ChildID }}">{{ .Name }}</a>
  <button class="outline" style="padding:.1rem .5rem"
          hx-delete="/bundles/{{ .ParentID }}/children/{{ .ChildID }}"
          hx-target="#bc-{{ .ParentID }}-{{ .ChildID }}" hx-swap="outerHTML">×</button>
</li>
{{ end }}
```

- [ ] **Step 5: Handlers**

`internal/web/handlers_bundles.go`:
```go
package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
)

type bundleItemView struct {
	BundleID string
	ItemID   string
	Name     string
	QtyShown int // 0 means "default"
}

type bundleChildView struct {
	ParentID string
	ChildID  string
	Name     string
}

func (s *Server) listBundles(w http.ResponseWriter, r *http.Request) {
	bs, err := s.Bundles.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "bundles", map[string]any{
		"Title":   "Bundles",
		"User":    auth.UserFrom(r.Context()),
		"Bundles": bs,
	})
}

func (s *Server) createBundle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := s.Bundles.Create(r.Context(), catalog.Bundle{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		CreatedBy:   auth.UserFrom(r.Context()),
	})
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	b, _ := s.Bundles.Get(r.Context(), id)
	s.Renderer.Partial(w, "bundle_row", b)
}

func (s *Server) editBundle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	b, err := s.Bundles.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items, _ := s.Items.List(r.Context())
	allBundles, _ := s.Bundles.List(r.Context())
	// Build current item rows with display names.
	itemNames := map[string]string{}
	for _, it := range items {
		itemNames[it.ID] = it.Name
	}
	bItems, _ := s.Bundles.Items(r.Context(), id)
	rows := make([]bundleItemView, 0, len(bItems))
	for _, bi := range bItems {
		row := bundleItemView{BundleID: id, ItemID: bi.ItemID, Name: itemNames[bi.ItemID]}
		if bi.Qty.Valid {
			row.QtyShown = int(bi.Qty.Int64)
		}
		rows = append(rows, row)
	}
	// Children + nestable filter.
	childIDs, _ := s.Bundles.Children(r.Context(), id)
	bundleNames := map[string]string{}
	for _, b := range allBundles {
		bundleNames[b.ID] = b.Name
	}
	childRows := make([]bundleChildView, 0, len(childIDs))
	for _, c := range childIDs {
		childRows = append(childRows, bundleChildView{ParentID: id, ChildID: c, Name: bundleNames[c]})
	}
	// nestable = all bundles except self + transitive descendants (which would
	// create cycles). For listing, exclude self only — the AddChild call
	// performs the deeper cycle check.
	nestable := make([]catalog.Bundle, 0, len(allBundles))
	for _, b := range allBundles {
		if b.ID != id {
			nestable = append(nestable, b)
		}
	}
	s.Renderer.Render(w, "bundle_edit", map[string]any{
		"Title":           b.Name,
		"User":            auth.UserFrom(r.Context()),
		"Bundle":          b,
		"AllItems":        items,
		"Items":           rows,
		"Children":        childRows,
		"NestableBundles": nestable,
	})
}

func (s *Server) updateBundle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	r.ParseForm()
	if err := s.Bundles.Update(r.Context(), id, catalog.Bundle{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteBundle(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.SoftDelete(r.Context(), r.PathValue("id"), auth.UserFrom(r.Context())); err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) bundleAddItem(w http.ResponseWriter, r *http.Request) {
	bID := r.PathValue("id")
	r.ParseForm()
	iID := r.FormValue("item_id")
	var qtyPtr *int
	if v := r.FormValue("qty"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if err := s.Bundles.AddItem(r.Context(), bID, iID, qtyPtr); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	it, _ := s.Items.Get(r.Context(), iID)
	view := bundleItemView{BundleID: bID, ItemID: iID, Name: it.Name}
	if qtyPtr != nil {
		view.QtyShown = *qtyPtr
	}
	s.Renderer.Partial(w, "bundle_item_row", view)
}

func (s *Server) bundleRemoveItem(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.RemoveItem(r.Context(), r.PathValue("id"), r.PathValue("iid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) bundleAddChild(w http.ResponseWriter, r *http.Request) {
	parent := r.PathValue("id")
	r.ParseForm()
	child := r.FormValue("child_id")
	if err := s.Bundles.AddChild(r.Context(), parent, child); err != nil {
		if errors.Is(err, catalog.ErrConflict) {
			http.Error(w, "would create a cycle", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), 400)
		return
	}
	cb, _ := s.Bundles.Get(r.Context(), child)
	s.Renderer.Partial(w, "bundle_child_row", bundleChildView{ParentID: parent, ChildID: child, Name: cb.Name})
}

func (s *Server) bundleRemoveChild(w http.ResponseWriter, r *http.Request) {
	if err := s.Bundles.RemoveChild(r.Context(), r.PathValue("id"), r.PathValue("cid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 6: Remove stubs**

Delete the nine bundle-related stubs from `handlers_stubs.go`: `listBundles`, `createBundle`, `editBundle`, `updateBundle`, `deleteBundle`, `bundleAddItem`, `bundleRemoveItem`, `bundleAddChild`, `bundleRemoveChild`.

- [ ] **Step 7: Test**

`internal/web/handlers_bundles_test.go`:
```go
package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestBundleCreateAndCycleRejected(t *testing.T) {
	s := newTestServer(t)
	// Create A, B; nest A->B; try B->A (cycle).
	aID, _ := s.Bundles.Create(context.Background(), catalog.Bundle{Name: "A", CreatedBy: "u_test"})
	bID, _ := s.Bundles.Create(context.Background(), catalog.Bundle{Name: "B", CreatedBy: "u_test"})

	// Nest A -> B via HTTP.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/bundles/"+aID+"/children", "child_id="+bID))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "B") {
		t.Fatalf("nest A->B: %d %s", w.Code, w.Body.String())
	}

	// Now B -> A should be 409 (cycle).
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/bundles/"+bID+"/children", "child_id="+aID))
	if w.Code != 409 {
		t.Fatalf("expected 409 on cycle, got %d (%s)", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 8: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 9: Commit**

```bash
git add internal/web
git commit -m "feat(web): bundles handlers with item + nested-bundle composition"
```

---

### Task 27: Trips list, create, detail

**Files:**
- Create: `internal/web/handlers_trips.go`
- Create: `internal/web/templates/pages/trips_index.html`
- Create: `internal/web/templates/pages/trip_new.html`
- Create: `internal/web/templates/pages/trip_detail.html`
- Create: `internal/web/templates/partials/trip_row.html`
- Create: `internal/web/templates/partials/pack_row.html`
- Create: `internal/web/templates/partials/progress.html`
- Modify: `internal/web/handlers_stubs.go` (remove `getTripsIndex`, `getTripNew`, `postTripCreate`, `getTripDetail`, `patchTrip`, `deleteTrip`)

- [ ] **Step 1: Templates — trips index**

`internal/web/templates/pages/trips_index.html`:
```html
{{ define "content" }}
<header>
  <h1>Trips</h1>
  <a class="contrast" role="button" href="/trips/new">New trip</a>
</header>
{{ if not .Trips }}
  <p>No trips yet. <a href="/trips/new">Create your first one.</a></p>
{{ else }}
<ul>{{ range .Trips }}{{ template "trip_row" . }}{{ end }}</ul>
{{ end }}
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 2: trip_row partial**

`internal/web/templates/partials/trip_row.html`:
```html
{{ define "trip_row" }}
<li><a href="/trips/{{ .ID }}">{{ .Name }}</a> · {{ .Nights }} night{{ if ne .Nights 1 }}s{{ end }}</li>
{{ end }}
```

- [ ] **Step 3: trip_new page**

`internal/web/templates/pages/trip_new.html`:
```html
{{ define "content" }}
<header><h1>New trip</h1></header>
<form method="post" action="/trips">
  <input type="hidden" name="csrf_token" value="__CSRF__" />
  <label>Name <input name="name" required autofocus /></label>
  <label>Nights <input name="nights" type="number" min="0" value="2" required /></label>
  <button>Create</button>
  <a href="/" class="secondary" role="button">Cancel</a>
</form>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 4: trip_detail page**

`internal/web/templates/pages/trip_detail.html`:
```html
{{ define "content" }}
<header>
  <h1>{{ .Trip.Name }}</h1>
  <p>{{ .Trip.Nights }} night{{ if ne .Trip.Nights 1 }}s{{ end }}</p>
</header>

<div class="progress-bar" id="progress">
  {{ template "progress" .Progress }}
</div>

<section>
  <h2>Bundles attached</h2>
  <div id="bundle-chips">
    {{ range .AttachedBundles }}
      <span class="bundle-chip" id="tb-{{ $.Trip.ID }}-{{ .ID }}">
        {{ .Name }}
        <button hx-delete="/trips/{{ $.Trip.ID }}/bundles/{{ .ID }}"
                hx-target="#tb-{{ $.Trip.ID }}-{{ .ID }}" hx-swap="outerHTML">×</button>
      </span>
    {{ end }}
  </div>
  <form hx-post="/trips/{{ .Trip.ID }}/bundles" hx-target="#bundle-chips" hx-swap="beforeend"
        hx-on::after-request="if(event.detail.successful){htmx.ajax('GET','/trips/{{ .Trip.ID }}',{target:'main',swap:'innerHTML'})}">
    <fieldset role="group">
      <select name="bundle_id" required>
        <option value="">— add a bundle —</option>
        {{ range .AvailableBundles }}<option value="{{ .ID }}">{{ .Name }}</option>{{ end }}
      </select>
      <button>Add</button>
    </fieldset>
  </form>
</section>

<section>
  <h2>Extras</h2>
  <form hx-post="/trips/{{ .Trip.ID }}/extras"
        hx-on::after-request="if(event.detail.successful){htmx.ajax('GET','/trips/{{ .Trip.ID }}',{target:'main',swap:'innerHTML'})}">
    <fieldset role="group">
      <select name="item_id" required>
        <option value="">— add an extra item —</option>
        {{ range .AllItems }}<option value="{{ .ID }}">{{ .Name }}</option>{{ end }}
      </select>
      <input name="qty" type="number" min="1" placeholder="qty" style="width:6rem" />
      <button>Add</button>
    </fieldset>
  </form>
</section>

<section>
  <h2>Packing list</h2>
  {{ $tid := .Trip.ID }}
  {{ $lastCat := "" }}
  {{ range .Rows }}
    {{ if ne .Category $lastCat }}
      <div class="category-heading">{{ .Category }}</div>
      {{ $lastCat = .Category }}
    {{ end }}
    {{ template "pack_row" (packArgs $tid .) }}
  {{ end }}
</section>

<section>
  <h2>Members</h2>
  <ul>{{ range .Members }}
    <li>{{ .Email }} ({{ .Role }})
      {{ if ne .Role "owner" }}
      <button class="outline" style="padding:.1rem .5rem"
              hx-delete="/trips/{{ $.Trip.ID }}/members/{{ .UserID }}"
              hx-target="closest li" hx-swap="outerHTML">remove</button>
      {{ end }}
    </li>
  {{ end }}</ul>
  <form hx-post="/trips/{{ .Trip.ID }}/members" hx-swap="none"
        hx-on::after-request="if(event.detail.successful){htmx.ajax('GET','/trips/{{ .Trip.ID }}',{target:'main',swap:'innerHTML'})}">
    <fieldset role="group">
      <input name="email" type="email" placeholder="email to invite" required />
      <button>Invite</button>
    </fieldset>
  </form>
</section>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 5: pack_row + progress partials**

`internal/web/templates/partials/pack_row.html`:
```html
{{ define "pack_row" }}
<div class="pack-row{{ if .Row.Packed }} packed{{ end }}" id="pr-{{ .TripID }}-{{ .Row.ItemID }}">
  <input type="checkbox" {{ if .Row.Packed }}checked{{ end }}
         hx-post="/trips/{{ .TripID }}/pack/{{ .Row.ItemID }}"
         hx-target="#pr-{{ .TripID }}-{{ .Row.ItemID }}" hx-swap="outerHTML"
         name="packed" value="1" />
  <div>{{ .Row.Name }} <small class="source">{{ range .Row.Sources }}{{ . }} {{ end }}</small></div>
  <div>×{{ .Row.Qty }}</div>
  <button class="outline" style="padding:.1rem .5rem"
          hx-patch="/trips/{{ .TripID }}/items/{{ .Row.ItemID }}"
          hx-vals='{"removed":"1"}' hx-target="#pr-{{ .TripID }}-{{ .Row.ItemID }}" hx-swap="outerHTML"
          hx-confirm="Skip {{ .Row.Name }} this trip?">skip</button>
</div>
{{ end }}
```

`internal/web/templates/partials/progress.html`:
```html
{{ define "progress" }}
<span style="width: {{ if eq .Total 0 }}0{{ else }}{{ percent .Packed .Total }}{{ end }}%"></span>
<small style="display:block;margin-top:.25rem">{{ .Packed }} / {{ .Total }} packed</small>
{{ end }}
```

- [ ] **Step 6: Renderer needs the `percent` and `packArgs` funcs**

Modify `internal/web/render.go`. In `NewRenderer`, extend `funcs`:

```go
funcs := template.FuncMap{
    "add":   func(a, b int) int { return a + b },
    "plus":  func(a, b int) int { return a + b },
    "percent": func(a, b int) int {
        if b == 0 {
            return 0
        }
        return a * 100 / b
    },
    "packArgs": func(tripID string, row trips.Row) map[string]any {
        return map[string]any{"TripID": tripID, "Row": row}
    },
}
```

Add the import `"github.com/bejl/packing-list/internal/trips"` to `render.go`.

- [ ] **Step 7: Handlers**

`internal/web/handlers_trips.go`:
```go
package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/trips"
)

func (s *Server) getTripsIndex(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	list, err := s.Trips.ListVisibleTo(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "trips_index", map[string]any{
		"Title": "Trips",
		"User":  uid,
		"Trips": list,
	})
}

func (s *Server) getTripNew(w http.ResponseWriter, r *http.Request) {
	s.Renderer.Render(w, "trip_new", map[string]any{
		"Title": "New trip",
		"User":  auth.UserFrom(r.Context()),
	})
}

func (s *Server) postTripCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	nights, _ := strconv.Atoi(r.FormValue("nights"))
	id, err := s.Trips.Create(r.Context(), name, nights, auth.UserFrom(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Redirect(w, r, "/trips/"+id, http.StatusSeeOther)
}

// requireMember ensures the caller is a member of the trip. Returns the role.
func (s *Server) requireMember(r *http.Request) (string, string, error) {
	tID := r.PathValue("id")
	uid := auth.UserFrom(r.Context())
	role, err := s.Trips.RoleOf(r.Context(), tID, uid)
	if err != nil {
		return tID, uid, err
	}
	return tID, role, nil
}

func (s *Server) getTripDetail(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	tr, err := s.Trips.Get(r.Context(), tID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rows, err := s.Renderer2.Render(r.Context(), tID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	packed := 0
	for _, row := range rows {
		if row.Packed {
			packed++
		}
	}
	attachedIDs, _ := s.Sources.AttachedBundleIDs(r.Context(), tID)
	attached := make([]struct{ ID, Name string }, 0, len(attachedIDs))
	for _, id := range attachedIDs {
		b, err := s.Bundles.Get(r.Context(), id)
		if err == nil && !b.DeletedAt.Valid {
			attached = append(attached, struct{ ID, Name string }{b.ID, b.Name})
		}
	}
	allItems, _ := s.Items.List(r.Context())
	allBundles, _ := s.Bundles.List(r.Context())
	members, _ := s.Trips.Members(r.Context(), tID)

	s.Renderer.Render(w, "trip_detail", map[string]any{
		"Title":            tr.Name,
		"User":             auth.UserFrom(r.Context()),
		"Trip":             tr,
		"Rows":             rows,
		"Progress":         struct{ Packed, Total int }{packed, len(rows)},
		"AttachedBundles":  attached,
		"AvailableBundles": allBundles,
		"AllItems":         allItems,
		"Members":          members,
	})
	// silence unused
	_ = errors.New
	_ = trips.ErrForbidden
}

func (s *Server) patchTrip(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || (role != "owner" && role != "editor") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	tr, err := s.Trips.Get(r.Context(), tID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		name = tr.Name
	}
	nights := tr.Nights
	if v := r.FormValue("nights"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			nights = n
		}
	}
	notes := r.FormValue("notes")
	if err := s.Trips.Update(r.Context(), tID, name, nights, notes); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteTrip(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Trips.SoftDelete(r.Context(), tID, auth.UserFrom(r.Context())); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/")
	}
	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 8: Remove the six trip-related stubs from `handlers_stubs.go`.**

- [ ] **Step 9: Test**

`internal/web/handlers_trips_test.go`:
```go
package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTripCreateAndDetail(t *testing.T) {
	s := newTestServer(t)
	// Create the trip.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=Weekend+Devon&nights=2"))
	if w.Code != 303 {
		t.Fatalf("create: status %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "/trips/") {
		t.Fatalf("redirect location: %q", loc)
	}
	// Detail page renders.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", loc, ""))
	if w.Code != 200 {
		t.Fatalf("detail: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Weekend Devon") {
		t.Errorf("expected trip name in body, got %s", w.Body.String())
	}
}
```

- [ ] **Step 10: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 11: Commit**

```bash
git add internal/web
git commit -m "feat(web): trips list/create/detail with progress + bundle attach UI"
```

---

### Task 28: Trip sources handlers (attach/detach bundle, add extra, override, pack toggle)

**Files:**
- Create: `internal/web/handlers_trip_sources.go`
- Modify: `internal/web/handlers_stubs.go` (remove `attachBundle`, `detachBundle`, `addExtra`, `overrideItem`, `togglePack`)

- [ ] **Step 1: Implement**

`internal/web/handlers_trip_sources.go`:
```go
package web

import (
	"net/http"
	"strconv"
)

func (s *Server) attachBundle(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	bID := r.FormValue("bundle_id")
	if err := s.Sources.AttachBundle(r.Context(), tID, bID); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) detachBundle(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Sources.DetachBundle(r.Context(), tID, r.PathValue("bid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) addExtra(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	iID := r.FormValue("item_id")
	var qtyPtr *int
	if v := r.FormValue("qty"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if err := s.Sources.AddExtra(r.Context(), tID, iID, qtyPtr); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) overrideItem(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	iID := r.PathValue("iid")
	removed := r.FormValue("removed") == "1"
	var qtyPtr *int
	if v := r.FormValue("qty"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			qtyPtr = &n
		}
	}
	if !removed && qtyPtr == nil {
		// "clear override" form.
		if err := s.Sources.ClearOverride(r.Context(), tID, iID); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	} else {
		if err := s.Sources.SetOverride(r.Context(), tID, iID, removed, qtyPtr); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
	// Empty body — caller refreshes trip detail.
	w.WriteHeader(http.StatusOK)
}

func (s *Server) togglePack(w http.ResponseWriter, r *http.Request) {
	tID, _, err := s.requireMember(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	iID := r.PathValue("iid")
	r.ParseForm()
	// Checkbox semantics: presence of "packed" key = true.
	packed := r.FormValue("packed") == "1"
	// If the cookie pre-existed (HTMX checkbox toggling), we flip based on current value.
	// Simpler: trust the form value.
	if err := s.Pack.Toggle(r.Context(), tID, iID, packed); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Re-render the row + progress.
	rows, err := s.Renderer2.Render(r.Context(), tID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var row *struct {
		TripID string
		Row    any
	}
	for i := range rows {
		if rows[i].ItemID == iID {
			row = &struct {
				TripID string
				Row    any
			}{TripID: tID, Row: rows[i]}
			break
		}
	}
	if row == nil {
		// Item no longer in list (unlikely). Send empty 200.
		w.WriteHeader(http.StatusOK)
		return
	}
	// Also push progress refresh via HX-Trigger so progress bar updates.
	packedCount := 0
	for _, x := range rows {
		if x.Packed {
			packedCount++
		}
	}
	w.Header().Set("HX-Trigger-After-Swap",
		`{"refreshProgress":{"packed":`+strconv.Itoa(packedCount)+`,"total":`+strconv.Itoa(len(rows))+`}}`)
	s.Renderer.Partial(w, "pack_row", row)
}
```

> **UI note:** the `pack_row` template expects `.TripID` and `.Row` — that matches the `packArgs` helper used in `trip_detail.html`. The HX-Trigger emitted here is for a small client snippet you can add later if you want the progress bar to live-update; ignoring it is fine for v1.

- [ ] **Step 2: Remove the five stubs from `handlers_stubs.go`.**

- [ ] **Step 3: Quick smoke test**

`internal/web/handlers_trip_sources_test.go`:
```go
package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestAttachBundleThenPackToggle(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()

	// Seed catalog.
	itID, _ := s.Items.Create(ctx, catalog.Item{Name: "Toothbrush", DefaultQty: 1, CreatedBy: "u_test"})
	bID, _ := s.Bundles.Create(ctx, catalog.Bundle{Name: "wash", CreatedBy: "u_test"})
	s.Bundles.AddItem(ctx, bID, itID, nil)

	// Create trip via HTTP.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=T&nights=1"))
	loc := w.Header().Get("Location")
	tID := strings.TrimPrefix(loc, "/trips/")

	// Attach bundle.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/bundles", "bundle_id="+bID))
	if w.Code != 200 {
		t.Fatalf("attach: %d", w.Code)
	}

	// Pack the item.
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/pack/"+itID, "packed=1"))
	if w.Code != 200 {
		t.Fatalf("pack: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "packed") {
		t.Errorf("expected packed class in row, got %s", w.Body.String())
	}
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/web
git commit -m "feat(web): trip source handlers (attach/detach/extras/override/pack)"
```

---

### Task 29: Trip members handlers

**Files:**
- Create: `internal/web/handlers_members.go`
- Modify: `internal/web/handlers_stubs.go` (remove `inviteMember`, `removeMember`)

- [ ] **Step 1: Implement**

`internal/web/handlers_members.go`:
```go
package web

import (
	"net/http"
	"strings"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) inviteMember(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		http.Error(w, "email required", 400)
		return
	}
	if !s.RateLimit.Allow("invite:" + auth.UserFrom(r.Context())) {
		http.Error(w, "too many invites; try again later", 429)
		return
	}
	uid, created, err := s.Users.FindOrCreate(r.Context(), email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := s.Trips.AddMember(r.Context(), tID, uid, "editor"); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// If newly created, also send a magic link so they can sign in.
	if created {
		if err := s.Magic.Issue(r.Context(), email); err != nil {
			s.Logger.Warn("invite magic link", "err", err)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	tID, role, err := s.requireMember(r)
	if err != nil || role != "owner" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := s.Trips.RemoveMember(r.Context(), tID, r.PathValue("uid")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 2: Remove the two stubs.**

- [ ] **Step 3: Test**

`internal/web/handlers_members_test.go`:
```go
package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInviteCreatesUserAndMembership(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips", "name=T&nights=1"))
	tID := strings.TrimPrefix(w.Header().Get("Location"), "/trips/")

	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trips/"+tID+"/members", "email=bob@example.com"))
	if w.Code != 200 {
		t.Fatalf("invite: %d (%s)", w.Code, w.Body.String())
	}
	_, created, _ := s.Users.FindOrCreate(context.Background(), "bob@example.com")
	if created {
		t.Error("expected bob to already exist after invite")
	}
}
```

- [ ] **Step 4: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/web
git commit -m "feat(web): trip member invite + remove (owner-only)"
```

---

### Task 30: Trash handlers + page

**Files:**
- Create: `internal/web/handlers_trash.go`
- Create: `internal/web/templates/pages/trash.html`
- Modify: `internal/web/handlers_stubs.go` (remove `getTrash`, `restore`, `purge`)

- [ ] **Step 1: Trash page**

`internal/web/templates/pages/trash.html`:
```html
{{ define "content" }}
<header><h1>Trash</h1><p>Restore or permanently remove.</p></header>

<section>
  <h2>Items</h2>
  {{ if not .Bin.Items }}<p><small>nothing here</small></p>{{ else }}
  <ul>{{ range .Bin.Items }}
    <li id="trash-item-{{ .ID }}">
      {{ .Name }} — deleted {{ .DeletedAt.Time.Format "2006-01-02 15:04" }}
      <button class="contrast" hx-post="/trash/item/{{ .ID }}/restore" hx-target="#trash-item-{{ .ID }}" hx-swap="outerHTML">Restore</button>
      <button class="outline" hx-delete="/trash/item/{{ .ID }}" hx-target="#trash-item-{{ .ID }}" hx-swap="outerHTML"
              hx-confirm="Purge {{ .Name }}? This cannot be undone.">Purge</button>
    </li>
  {{ end }}</ul>{{ end }}
</section>

<section>
  <h2>Bundles</h2>
  {{ if not .Bin.Bundles }}<p><small>nothing here</small></p>{{ else }}
  <ul>{{ range .Bin.Bundles }}
    <li id="trash-bundle-{{ .ID }}">
      {{ .Name }} — deleted {{ .DeletedAt.Time.Format "2006-01-02 15:04" }}
      <button class="contrast" hx-post="/trash/bundle/{{ .ID }}/restore" hx-target="#trash-bundle-{{ .ID }}" hx-swap="outerHTML">Restore</button>
      <button class="outline" hx-delete="/trash/bundle/{{ .ID }}" hx-target="#trash-bundle-{{ .ID }}" hx-swap="outerHTML"
              hx-confirm="Purge {{ .Name }}? This cannot be undone.">Purge</button>
    </li>
  {{ end }}</ul>{{ end }}
</section>

<section>
  <h2>Trips</h2>
  {{ if not .Bin.Trips }}<p><small>nothing here</small></p>{{ else }}
  <ul>{{ range .Bin.Trips }}
    <li id="trash-trip-{{ .ID }}">
      {{ .Name }} — deleted {{ .DeletedAt.Time.Format "2006-01-02 15:04" }}
      <button class="contrast" hx-post="/trash/trip/{{ .ID }}/restore" hx-target="#trash-trip-{{ .ID }}" hx-swap="outerHTML">Restore</button>
      <button class="outline" hx-delete="/trash/trip/{{ .ID }}" hx-target="#trash-trip-{{ .ID }}" hx-swap="outerHTML"
              hx-confirm="Purge {{ .Name }}? This cannot be undone.">Purge</button>
    </li>
  {{ end }}</ul>{{ end }}
</section>
{{ end }}
{{ template "layout" . }}
```

- [ ] **Step 2: Handlers**

`internal/web/handlers_trash.go`:
```go
package web

import (
	"net/http"

	"github.com/bejl/packing-list/internal/auth"
)

func (s *Server) getTrash(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	bin, err := s.Trash.For(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.Renderer.Render(w, "trash", map[string]any{
		"Title": "Trash",
		"User":  uid,
		"Bin":   bin,
	})
}

func (s *Server) restore(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id := r.PathValue("id")
	var err error
	switch kind {
	case "item":
		err = s.Items.Restore(r.Context(), id)
	case "bundle":
		err = s.Bundles.Restore(r.Context(), id)
	case "trip":
		// only owner can restore.
		role, rerr := s.Trips.RoleOf(r.Context(), id, auth.UserFrom(r.Context()))
		if rerr != nil || role != "owner" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		err = s.Trips.Restore(r.Context(), id)
	default:
		http.Error(w, "bad kind", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) purge(w http.ResponseWriter, r *http.Request) {
	kind := r.PathValue("kind")
	id := r.PathValue("id")
	var err error
	switch kind {
	case "item":
		err = s.Items.Purge(r.Context(), id)
	case "bundle":
		err = s.Bundles.Purge(r.Context(), id)
	case "trip":
		role, rerr := s.Trips.RoleOf(r.Context(), id, auth.UserFrom(r.Context()))
		if rerr != nil || role != "owner" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		err = s.Trips.Purge(r.Context(), id)
	default:
		http.Error(w, "bad kind", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusOK)
}
```

- [ ] **Step 3: Remove the three stubs.**

- [ ] **Step 4: Test**

`internal/web/handlers_trash_test.go`:
```go
package web

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bejl/packing-list/internal/catalog"
)

func TestTrashRestoresSoftDeletedItem(t *testing.T) {
	s := newTestServer(t)
	ctx := context.Background()
	id, _ := s.Items.Create(ctx, catalog.Item{Name: "X", DefaultQty: 1, CreatedBy: "u_test"})
	s.Items.SoftDelete(ctx, id, "u_test")

	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/trash", ""))
	if !strings.Contains(w.Body.String(), "X") {
		t.Fatalf("expected X in trash, got %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "POST", "/trash/item/"+id+"/restore", ""))
	if w.Code != 200 {
		t.Fatalf("restore: %d", w.Code)
	}
	list, _ := s.Items.List(ctx)
	if len(list) != 1 {
		t.Errorf("expected restored item back in list, got %d items", len(list))
	}
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 6: Commit**

```bash
git add internal/web
git commit -m "feat(web): trash page with per-kind restore + purge"
```

---

### Task 31: Export / Import

JSON dump of the user's catalog + visible trips. Import merges by `id`.

**Files:**
- Create: `internal/web/handlers_portability.go`
- Modify: `internal/web/handlers_stubs.go` (remove `export`, `importJSON`)

- [ ] **Step 1: Implement**

`internal/web/handlers_portability.go`:
```go
package web

import (
	"encoding/json"
	"net/http"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/trips"
)

type exportDoc struct {
	Items   []catalog.Item   `json:"items"`
	Bundles []bundleExport   `json:"bundles"`
	Trips   []tripExport     `json:"trips"`
}

type bundleExport struct {
	catalog.Bundle
	Items    []catalog.BundleItem `json:"items"`
	Children []string             `json:"children"`
}

type tripExport struct {
	ID        string
	Name      string
	Nights    int
	Notes     string
	Bundles   []string
	Extras    []trips.Extra
	Overrides []trips.Override
}

func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	items, _ := s.Items.List(r.Context())
	bundles, _ := s.Bundles.List(r.Context())

	be := make([]bundleExport, 0, len(bundles))
	for _, b := range bundles {
		bi, _ := s.Bundles.Items(r.Context(), b.ID)
		ch, _ := s.Bundles.Children(r.Context(), b.ID)
		be = append(be, bundleExport{Bundle: b, Items: bi, Children: ch})
	}

	tripList, _ := s.Trips.ListVisibleTo(r.Context(), uid)
	tdocs := make([]tripExport, 0, len(tripList))
	for _, t := range tripList {
		bIDs, _ := s.Sources.AttachedBundleIDs(r.Context(), t.ID)
		extras, _ := s.Sources.QueryExtras(r.Context(), t.ID)
		overrides, _ := s.Sources.QueryOverrides(r.Context(), t.ID)
		tdocs = append(tdocs, tripExport{
			ID:        t.ID,
			Name:      t.Name,
			Nights:    t.Nights,
			Notes:     t.Notes,
			Bundles:   bIDs,
			Extras:    extras,
			Overrides: overrides,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="packing-list-export.json"`)
	json.NewEncoder(w).Encode(exportDoc{Items: items, Bundles: be, Trips: tdocs})
}

func (s *Server) importJSON(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20) // 5 MiB cap
	var doc exportDoc
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, "bad json: "+err.Error(), 400)
		return
	}
	// Upsert items + bundles. Skip if ID already exists (caller can purge first).
	for _, it := range doc.Items {
		// Reuse Create if not present.
		if _, err := s.Items.Get(r.Context(), it.ID); err != nil {
			it.CreatedBy = uid
			s.Items.Create(r.Context(), it)
		}
	}
	for _, b := range doc.Bundles {
		if _, err := s.Bundles.Get(r.Context(), b.ID); err != nil {
			s.Bundles.Create(r.Context(), b.Bundle)
		}
		for _, bi := range b.Items {
			var qp *int
			if bi.Qty.Valid {
				v := int(bi.Qty.Int64)
				qp = &v
			}
			s.Bundles.AddItem(r.Context(), b.ID, bi.ItemID, qp)
		}
		for _, child := range b.Children {
			s.Bundles.AddChild(r.Context(), b.ID, child)
		}
	}
	for _, t := range doc.Trips {
		// Trips: create as the caller (cannot import someone else's owner).
		tID, _ := s.Trips.Create(r.Context(), t.Name, t.Nights, uid)
		for _, bID := range t.Bundles {
			s.Sources.AttachBundle(r.Context(), tID, bID)
		}
		for _, e := range t.Extras {
			s.Sources.AddExtra(r.Context(), tID, e.ItemID, e.Qty)
		}
		for _, ov := range t.Overrides {
			s.Sources.SetOverride(r.Context(), tID, ov.ItemID, ov.Removed, ov.Qty)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

- [ ] **Step 2: Add `Extra` / `Override` types + `QueryExtras` / `QueryOverrides` to `internal/trips/sources.go`**

Append:
```go
// Extra and Override are JSON-friendly views of trip-scoped data, exported so
// the web layer can serialise them directly during /export.
type Extra struct {
	ItemID string `json:"item_id"`
	Qty    *int   `json:"qty,omitempty"`
}

type Override struct {
	ItemID  string `json:"item_id"`
	Removed bool   `json:"removed,omitempty"`
	Qty     *int   `json:"qty,omitempty"`
}

func (s *Sources) QueryExtras(ctx context.Context, tripID string) ([]Extra, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT item_id, qty FROM trip_extras WHERE trip_id = ?`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Extra
	for rows.Next() {
		var e Extra
		var q sql.NullInt64
		if err := rows.Scan(&e.ItemID, &q); err != nil {
			return nil, err
		}
		if q.Valid {
			v := int(q.Int64)
			e.Qty = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Sources) QueryOverrides(ctx context.Context, tripID string) ([]Override, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT item_id, removed, qty_override FROM trip_overrides WHERE trip_id = ?`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Override
	for rows.Next() {
		var ov Override
		var rem int
		var q sql.NullInt64
		if err := rows.Scan(&ov.ItemID, &rem, &q); err != nil {
			return nil, err
		}
		ov.Removed = rem != 0
		if q.Valid {
			v := int(q.Int64)
			ov.Qty = &v
		}
		out = append(out, ov)
	}
	return out, rows.Err()
}
```

- [ ] **Step 3: Remove the two stubs.**

- [ ] **Step 4: Test**

`internal/web/handlers_portability_test.go`:
```go
package web

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExportRoundTrip(t *testing.T) {
	s := newTestServer(t)
	// Seed via API.
	s.Handler().ServeHTTP(httptest.NewRecorder(), authedRequest(t, s, "POST", "/items", "name=Toothbrush&default_qty=1"))
	// Export.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, authedRequest(t, s, "GET", "/export", ""))
	if w.Code != 200 {
		t.Fatalf("export: %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Toothbrush") {
		t.Errorf("export missing toothbrush: %s", w.Body.String())
	}
	dump := bytes.TrimSpace(w.Body.Bytes())
	if !bytes.HasPrefix(dump, []byte("{")) {
		t.Errorf("expected json, got %s", w.Body.String())
	}
}
```

- [ ] **Step 5: Run, confirm pass**

```bash
go test ./internal/web/...
```

- [ ] **Step 6: Commit**

```bash
git add internal/web internal/trips
git commit -m "feat(web): JSON export and import for catalog and visible trips"
```

---

### Task 32: Main entrypoint

**Files:**
- Create: `main.go`
- Create: `cmd/seed/main.go` (placeholder — implemented in Task 33)

- [ ] **Step 1: main.go**

`main.go`:
```go
// Package main launches the packing-list HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bejl/packing-list/internal/auth"
	"github.com/bejl/packing-list/internal/catalog"
	"github.com/bejl/packing-list/internal/config"
	pdb "github.com/bejl/packing-list/internal/db"
	"github.com/bejl/packing-list/internal/trash"
	"github.com/bejl/packing-list/internal/trips"
	"github.com/bejl/packing-list/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		logger.Error("mkdir data", "err", err)
		os.Exit(1)
	}
	secret, err := auth.LoadOrCreateSecret(cfg.DataDir, cfg.SessionSecret)
	if err != nil {
		logger.Error("session secret", "err", err)
		os.Exit(1)
	}

	db, err := pdb.Open(filepath.Join(cfg.DataDir, "data.db"))
	if err != nil {
		logger.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	users := auth.NewUsers(db)

	// Bootstrap: ensure BOOTSTRAP_EMAIL exists if provided.
	if cfg.BootstrapEmail != "" {
		if _, _, err := users.FindOrCreate(context.Background(), cfg.BootstrapEmail); err != nil {
			logger.Warn("bootstrap user", "err", err)
		}
	}

	now := func() time.Time { return time.Now() }
	mailer := auth.NewMailer(cfg, logger)

	server := &web.Server{
		Cfg:       cfg,
		Logger:    logger,
		Users:     users,
		Sessions:  auth.NewSessions(db, secret, now),
		Magic:     auth.NewMagic(db, mailer, cfg.BaseURL, now),
		RateLimit: auth.NewRateLimiter(10, time.Hour, now),
		Items:     catalog.NewItems(db),
		Bundles:   catalog.NewBundles(db),
		Trips:     trips.NewTrips(db),
		Sources:   trips.NewSources(db),
		Pack:      trips.NewPack(db),
		Renderer2: trips.NewRenderer(db),
		IsDev:     !cfg.SMTP.Configured(),
		Now:       now,
	}
	renderer, err := web.NewRenderer()
	if err != nil {
		logger.Error("renderer", "err", err)
		os.Exit(1)
	}
	server.Renderer = renderer
	server.Trash = trash.NewView(db, server.Items, server.Bundles, server.Trips)

	hsvr := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Periodic cleanup: expired magic tokens + sessions every hour.
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for range t.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			_ = server.Magic.PurgeExpired(ctx)
			_ = server.Sessions.PurgeExpired(ctx)
			cancel()
		}
	}()

	logger.Info("listening", "addr", hsvr.Addr, "baseURL", cfg.BaseURL)
	go func() {
		if err := hsvr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logger.Info("shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = hsvr.Shutdown(ctx)
	_ = strings.TrimSpace // keep import alive if unused after refactors
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: zero errors.

- [ ] **Step 3: Quick local run**

```bash
go build -o packing-list .
BASE_URL=http://localhost:8080 DATA_DIR=$(pwd)/data ./packing-list &
curl -s http://localhost:8080/healthz
```

Expected: `ok`. Then `curl -i http://localhost:8080/` → 303 redirect to `/login`.

Stop the process:
```bash
kill %1
```

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: main entrypoint with graceful shutdown and periodic token sweeps"
```

---

## Phase 5 — Polish

### Task 33: Seed command

**Files:**
- Create: `cmd/seed/main.go`
- Modify: `internal/catalog/items.go` (no changes; reused)

- [ ] **Step 1: Implement seed**

`cmd/seed/main.go`:
```go
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
	one := 1
	_ = one
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
		{key: "weekend-uk", name: "weekend-uk", description: "UK weekend",
			items: []bundleItem{
				{itemKey: "underwear"}, {itemKey: "socks"}, {itemKey: "t-shirt"},
				{itemKey: "trousers"}, {itemKey: "jumper"}, {itemKey: "waterproof"},
				{itemKey: "pyjamas"}, {itemKey: "wallet"},
			},
			children: []string{"washbag-basic", "electronics-day"},
		},
		{key: "beach-week", name: "beach-week", description: "Beach week",
			items: []bundleItem{
				{itemKey: "underwear"}, {itemKey: "socks"}, {itemKey: "t-shirt"},
				{itemKey: "shorts"}, {itemKey: "sunglasses"}, {itemKey: "hat"},
				{itemKey: "swimsuit"}, {itemKey: "towel-quickdry"}, {itemKey: "passport"},
				{itemKey: "wallet"},
			},
			children: []string{"washbag-full", "electronics-week"},
		},
	}
}
```

- [ ] **Step 2: Build seed binary**

```bash
go build -o seed ./cmd/seed
```

Expected: no errors.

- [ ] **Step 3: Smoke test seed against a temp DB**

```bash
DATA_DIR=$(mktemp -d) ./seed
```

Expected: `seed: ok`. Re-running prints the skip message.

- [ ] **Step 4: Commit**

```bash
git add cmd/seed
git commit -m "feat(seed): idempotent starter items + bundles (incl. nested)"
```

---

### Task 34: Dockerfile + multi-stage build

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

- [ ] **Step 1: .dockerignore**

`.dockerignore`:
```
.git
.gitignore
.idea
.vscode
data/
*.db
*.db-journal
*.db-shm
*.db-wal
*.test
packing-list
packing-list.exe
seed
seed.exe
docs/
README.md
```

- [ ] **Step 2: Dockerfile**

`Dockerfile`:
```dockerfile
# syntax=docker/dockerfile:1.7
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 keeps the binary fully static (modernc sqlite is pure Go).
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/packing-list .
RUN go build -trimpath -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM gcr.io/distroless/static:nonroot AS run
WORKDIR /app
COPY --from=build /out/packing-list /app/packing-list
COPY --from=build /out/seed /app/seed
USER nonroot:nonroot
ENV DATA_DIR=/data
ENV PORT=8080
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/packing-list"]
```

- [ ] **Step 3: Build and run locally**

```bash
docker build -t packing-list:dev .
docker run -d --rm --name pl-test \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e BASE_URL=http://localhost:8080 \
  -e BOOTSTRAP_EMAIL=test@local \
  packing-list:dev
sleep 1
curl -fsS http://localhost:8080/healthz
docker stop pl-test
```

Expected: `ok`. Image size < 30 MB:
```bash
docker images packing-list:dev --format '{{.Size}}'
```

- [ ] **Step 4: Commit**

```bash
git add Dockerfile .dockerignore
git commit -m "chore: distroless multi-stage Docker build"
```

---

### Task 35: README + manual verification checklist

**Files:**
- Modify: `README.md` (replace the placeholder from Task 1)

- [ ] **Step 1: Write README.md**

`README.md`:
````markdown
# packing-list

A self-hosted web app that builds trip packing lists by composing reusable bundles of items. Bundles can nest other bundles; items can be flagged "per-night" to auto-scale by trip duration. Built with Go + HTMX + SQLite + Pico.css.

## Quick start

```bash
docker build -t packing-list:dev .
docker run -d --name packing-list \
  -p 8080:8080 \
  -v "$(pwd)/data:/data" \
  -e BASE_URL=http://localhost:8080 \
  -e BOOTSTRAP_EMAIL=you@example.com \
  packing-list:dev

# Seed the starter catalog (one-off):
docker exec -it packing-list /app/seed

# Visit http://localhost:8080, sign in with you@example.com.
# Because SMTP isn't configured, the magic link is printed to the container log:
docker logs packing-list | grep "magic link issued"
```

## Configuration

| Variable          | Required | Default          | Notes                                   |
| ----------------- | -------- | ---------------- | --------------------------------------- |
| `BASE_URL`        | yes      | —                | Used in magic-link URLs                 |
| `PORT`            | no       | `8080`           |                                         |
| `DATA_DIR`        | no       | `/data`          | SQLite + session secret live here       |
| `SESSION_SECRET`  | no       | auto-generated   | 32 hex chars or more; persisted if blank |
| `BOOTSTRAP_EMAIL` | no       | —                | Pre-creates this user at startup        |
| `SMTP_HOST`       | no       | —                | If empty, magic links go to logs only    |
| `SMTP_PORT`       | no       | `587`            |                                         |
| `SMTP_USER` `SMTP_PASS` | no | —              | Optional, PLAIN auth                     |
| `SMTP_FROM`       | no       | `no-reply@<host>` |                                         |

## Backup

```bash
docker exec packing-list /app/packing-list # no, see below
```

For SQLite hot-copy, run from the host with the container down (or use the
`sqlite3 .backup` cmd inside an exec session):
```bash
docker stop packing-list
cp data/data.db backups/data-$(date +%F).db
docker start packing-list
```

## Manual verification checklist

After deploying for the first time, walk this end-to-end:

- [ ] Sign in: visit `/login`, enter email; magic link appears in `docker logs`.
- [ ] Open the link → land on `/`. Trips list is empty.
- [ ] Create a new trip ("Weekend Devon", 2 nights).
- [ ] Attach `weekend-uk` bundle (which nests `washbag-basic` and `electronics-day`).
- [ ] Verify the final list contains both bundle items AND nested items (toothbrush + phone-charger).
- [ ] Verify per-night items scale by nights (e.g. underwear shows quantity 2).
- [ ] Attach `washbag-basic` directly as well — toothbrush still shows ONCE (de-dup).
- [ ] Tick a few items as packed. Restart the container (`docker restart packing-list`). State persists.
- [ ] From the trip page, invite a second email. Open `/logs` and follow the link in a private window — that user lands on the shared trip and can edit.
- [ ] On the bundles page, edit `washbag-basic` and add an item. Existing trips using it (directly or via `weekend-uk`) see the new item immediately (live-reference).
- [ ] On bundles page, attempt to nest `weekend-uk` inside `washbag-basic` — server returns a 409 (cycle).
- [ ] Delete an item from `/items`. Visit `/trash`, click Restore — item is back in the list.
- [ ] Visit `/export`. Save the JSON. Stop container, delete `data/data.db`, restart, re-import via `/import`. State restored.

## Development

```bash
go test ./...
go build -o packing-list .
BASE_URL=http://localhost:8080 DATA_DIR=./data ./packing-list
```

To regenerate vendored static files (Pico, HTMX):
```bash
curl -fsSL https://cdn.jsdelivr.net/npm/@picocss/pico@2.0.6/css/pico.classless.min.css -o internal/web/static/pico.min.css
curl -fsSL https://cdn.jsdelivr.net/npm/htmx.org@2.0.3/dist/htmx.min.js -o internal/web/static/htmx.min.js
```

See `docs/superpowers/specs/2026-05-08-packing-list-design.md` for the full design.
````

- [ ] **Step 2: Run the manual verification end-to-end against a freshly-built image.**

Mark each box as you confirm. Any failures are real bugs to fix before merging.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: README with config table, backup notes, manual verification list"
```

---

## Self-Review (post-plan)

The plan author has scanned the spec against the plan and confirms:

- **Spec §3 locked decisions** — every Q1–Q11 + Q4b decision is implemented by one or more tasks (Q1 flat-variant bundles by Task 26; Q2 per-night by Task 9 render; Q3 live-ref by Tasks 7/9; Q4 global catalog by Tasks 6/7; Q4b nested bundles by Tasks 7/9; Q5 self-host by Task 34; Q6 stack by Task 1/2; Q7 magic-link by Tasks 13–14, log fallback in 13; Q8 trip sharing by Tasks 8/29; Q9 global by Tasks 6/7; Q10 pack tracking by Tasks 10/28; Q11 Pico.css by Task 20).
- **Spec §5 schema** — Task 5 migration includes every table, including `bundle_children` with CHECK + index. `schema_migrations` is bootstrapped explicitly.
- **Spec §6 render rule** — Task 9 implements the recursive expand-then-merge logic with all listed edge cases covered in table-driven tests (per-night × nights, fixed max, override removed, qty override, deleted item / bundle / nested child, diamond inheritance).
- **Spec §7 trash** — Per-entity soft-delete/restore/purge in Tasks 6–8, aggregator in Task 11, UI in Task 30. Purge cascade rules in spec match the SQL in Tasks 6/7/8.
- **Spec §9 auth + middleware** — Magic flow (Task 14), sessions (Tasks 15/16), CSRF (Task 18), rate limit (Task 19), bootstrap user (Task 32 main.go reads `BOOTSTRAP_EMAIL`).
- **Spec §10 routes** — Every route in the spec is wired in Task 23's mux and implemented in Tasks 24–31.
- **Spec §11 UX flows** — `trip_detail.html` (Task 27) implements the three-panel layout (bundles attached, extras, final list) plus progress bar and members section.
- **Spec §14 manual verification** — README checklist (Task 35) covers every item including the cycle-rejection edge case added in Q4b.

No remaining placeholders, "TBDs", or vague steps. Function names referenced across tasks (`Items.Create`, `Bundles.AddChild`, `Sources.AttachBundle`, `Pack.Toggle`, `Renderer.Render`, `Renderer.Partial`, `Renderer2.Render`) are consistent.

**Known imperfections (intentional):** the `handlers_stubs.go` mechanic creates churn (delete-as-you-go) but keeps each task's package building independently. An alternative would be to wire the router incrementally, but that would risk drift between routes and handlers.





