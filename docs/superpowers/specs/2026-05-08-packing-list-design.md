# Packing List — Design Spec

**Date:** 2026-05-08
**Status:** Approved (brainstorm), ready for implementation plan
**Owner:** bejl

## 1. Overview

A small, self-hostable web app for assembling trip packing lists out of reusable
"bundles" of items (e.g. `washbag-basic`, `swimming-pool`, `running-trail`).
Users compose a trip from one or more bundles plus ad-hoc extras, then check
items off as they pack. Designed for a single household (~5 trusted users,
no public exposure).

The core value is composition: the same `washbag-basic` bundle is reused across
every trip, edited once, and changes propagate to all trips that use it.

## 2. Goals / non-goals

### Goals
- Compose trip lists from reusable bundles + ad-hoc items.
- Per-night quantity scaling for clothing-type items.
- Interactive check-off while packing, with persisted state.
- Multi-user trip sharing (owner + invited editors).
- Soft-delete with a trash/restore view for bundles, items, trips.
- Self-hostable on any Docker-capable host. Single binary + SQLite file.
- Lightweight: no SPA framework, no Node build pipeline, minimal runtime deps.

### Non-goals (v1)
- Real-time collaboration. Last-write-wins is fine.
- Push notifications, mobile apps, offline mode.
- Conditional / rule-driven bundle logic ("if cold, add fleece").
- Per-user private catalogs. Catalog is global to the household.
- Public sharing of bundles outside the instance.
- Bundle variant selectors. Variants are modelled as separate flat bundles
  (`swimming-pool`, `swimming-coldsea`).

## 3. Locked decisions (from brainstorm)

| # | Decision | Choice |
|---|---|---|
| Q1 | Variant modelling | Separate flat bundles per variant. |
| Q2 | Quantity logic | Per-item `per_night` flag scales by trip nights; fixed otherwise. |
| Q3 | Bundle-edit semantics | Live reference: trip stores bundle IDs; edits propagate. Per-trip overrides allowed. |
| Q4 | Duplicate-item merging | Global item catalog. Bundles + extras reference item IDs. Merge by ID: per-night sums × nights, fixed takes max. |
| Q4b | Bundle composition | Bundles may contain items AND nest other bundles. Cycle-checked on insert. Render expands recursively, de-duped via merge. |
| Q5 | Storage / hosting | Self-hosted, portable single container. SQLite file on a mounted volume. |
| Q6 | Stack | Go + `html/template` + HTMX + Pico.css + SQLite (modernc, cgo-free). |
| Q7 | Auth | Magic-link (email). If `SMTP_HOST` unset, link is logged to stdout. |
| Q8 | Trip sharing | Trip has owner + invited editors via `trip_members`. |
| Q9 | Catalog scope | Global to instance. Soft-delete with restore protects against accidents. |
| Q10 | Pack tracking | Per-trip checked state, persisted. Progress count shown. |
| Q11 | Visual style | Pico.css (classless, ~10KB). Minimal hand-rolled overrides. |

## 4. Architecture & stack

- **Language:** Go 1.23+.
- **HTTP:** stdlib `net/http`. No web framework.
- **Templates:** `html/template`. Pages and HTMX fragment partials.
- **DB:** SQLite via `modernc.org/sqlite` (pure Go, no cgo). Single `data.db` file.
- **Frontend assets:** `htmx.min.js` (~14KB) and `pico.min.css` (~10KB) checked into
  `internal/web/static/`. No build step. Plus a small `app.css` for overrides.
- **Auth:** magic-link sessions. Cookie holds session ID, looked up in DB.
- **IDs:** ULIDs (`github.com/oklog/ulid/v2`). Sortable, URL-safe.
- **Logs:** stdlib `slog` JSON to stdout.
- **Deploy:** single static Go binary in a scratch-based Docker image (~20MB).
  Volume-mount `/data` for DB + backups.

### Runtime configuration (env vars)

| Var | Required | Default | Purpose |
|---|---|---|---|
| `PORT` | no | `8080` | listen port |
| `BASE_URL` | yes | — | used in magic-link emails (e.g. `https://pack.example.com`) |
| `DATA_DIR` | no | `/data` | dir for `data.db` + `.session_secret` |
| `SESSION_SECRET` | no | generated to `$DATA_DIR/.session_secret` | HMAC for cookie values |
| `BOOTSTRAP_EMAIL` | no | — | on first boot, if `users` is empty, create this user |
| `SMTP_HOST` | no | unset → log links to stdout | SMTP server host |
| `SMTP_PORT` | no | `587` | |
| `SMTP_USER` | no | — | |
| `SMTP_PASS` | no | — | |
| `SMTP_FROM` | no | `no-reply@<host of BASE_URL>` | |

Boot fails fast on: missing `BASE_URL`, unwritable `DATA_DIR`, migration failure.

## 5. Data model

```sql
-- Users
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE COLLATE NOCASE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

-- Sessions: cookie value = id (HMAC-signed)
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Magic-link tokens (short-lived, single-use)
CREATE TABLE magic_tokens (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL COLLATE NOCASE,
  token_hash BLOB NOT NULL,            -- sha256 of raw token
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Item catalog (global)
CREATE TABLE items (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'general',
  per_night INTEGER NOT NULL DEFAULT 0, -- bool
  default_qty INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_items_active ON items(deleted_at) WHERE deleted_at IS NULL;

-- Bundle catalog (global)
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

-- Bundle composition: items
CREATE TABLE bundle_items (
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  qty       INTEGER,                    -- NULL = use items.default_qty
  PRIMARY KEY (bundle_id, item_id)
);

-- Bundle composition: nested bundles. Cycle-checked on insert.
CREATE TABLE bundle_children (
  parent_id TEXT NOT NULL REFERENCES bundles(id),
  child_id  TEXT NOT NULL REFERENCES bundles(id),
  PRIMARY KEY (parent_id, child_id),
  CHECK (parent_id <> child_id)
);
CREATE INDEX idx_bundle_children_child ON bundle_children(child_id);

-- Trips
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

-- Trip members (owner row also lives here for uniform queries)
CREATE TABLE trip_members (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  role TEXT NOT NULL CHECK (role IN ('owner','editor')),
  added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, user_id)
);

-- Bundles attached to a trip (live reference)
CREATE TABLE trip_bundles (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  added_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, bundle_id)
);

-- Ad-hoc items added directly to a trip (not via a bundle)
CREATE TABLE trip_extras (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  item_id TEXT NOT NULL REFERENCES items(id),
  qty     INTEGER,                      -- NULL = items.default_qty
  PRIMARY KEY (trip_id, item_id)
);

-- Per-trip overrides on bundle items
CREATE TABLE trip_overrides (
  trip_id      TEXT NOT NULL REFERENCES trips(id),
  item_id      TEXT NOT NULL REFERENCES items(id),
  removed      INTEGER NOT NULL DEFAULT 0,  -- bool, drop from rendered list
  qty_override INTEGER,                     -- NULL = no override
  PRIMARY KEY (trip_id, item_id)
);

-- Pack state (checkbox)
CREATE TABLE trip_pack_state (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  packed    INTEGER NOT NULL DEFAULT 0,
  packed_at TIMESTAMP,
  PRIMARY KEY (trip_id, item_id)
);
```

`PRAGMA foreign_keys = ON` set on every connection.

## 6. Render rule (trip → packing list)

Pure function. Inputs from DB; no mutation. Pseudocode:

```
1. sources = []
   for each tb in trip_bundles where bundles.deleted_at is null:
     expand_bundle(tb.bundle_id, sources)

   for each te in trip_extras where items.deleted_at is null:
     qty = te.qty ?? items.default_qty
     sources.append({ item_id, qty, source: "extras" })

   --- where ---
   expand_bundle(b_id, out, visited = {}):
     if b_id in visited: return        // de-dup, also makes diamond inheritance safe
     visited.add(b_id)
     if bundles[b_id].deleted_at: return
     for each bi in bundle_items[b_id] where items.deleted_at is null:
       qty = bi.qty ?? items.default_qty
       out.append({ item_id, qty, source: bundles[b_id].name })
     for each child in bundle_children[b_id] where bundles[child].deleted_at is null:
       expand_bundle(child, out, visited)

2. drop sources whose item_id has trip_overrides.removed = 1

3. apply qty_override: for each override with non-null qty_override,
   replace all sources for that item_id with one synthetic source of
   { qty: qty_override, source: "override" }.
   In the rendered row, the displayed source label is "override" plus the
   names of the original contributing bundles in parentheses, e.g.
   "override (was: washbag-basic + swimming-pool)". This preserves provenance
   while making the override visible.

4. group sources by item_id. For each group:
     if items.per_night:
       qty = sum(group.qty) * trips.nights
     else:
       qty = max(group.qty)
     emit { item, qty, sources: group.sources, packed: pack_state.packed }

5. group results by items.category for display order:
   stable sort by (category, item.name).
```

This logic lives in `internal/trips/render.go` and is the single source of
truth for "what does this trip contain". Tested heavily.

## 7. Trash / restore

- Page `/trash`. Tabs: **Items**, **Bundles**, **Trips**.
- Lists rows where `deleted_at IS NOT NULL`, newest first, with deleted-by/when.
- **Restore** sets `deleted_at = NULL`, `deleted_by = NULL`. Cascade is implicit:
  `bundle_items` and `bundle_children` rows are never deleted on bundle soft-delete;
  restoring the bundle resurrects its full composition (own items + nested children)
  automatically.
- **Purge** is a hard delete. For bundles, also deletes `bundle_items`,
  `bundle_children` (rows where the purged bundle is parent OR child), and
  `trip_bundles` referencing it. For items, also deletes `bundle_items`,
  `trip_extras`, `trip_overrides`, `trip_pack_state` referencing it. For trips,
  cascade-deletes everything trip-scoped. Purge is per-row, manual only.
- Auto-purge: not in v1. Manual.
- Anyone authenticated can restore or purge. Audit lives in `deleted_by`.

## 8. Code structure

```
packing-list/
  main.go                          -- entry: config, DB open, migrations, router, listen
  go.mod / go.sum
  Dockerfile
  README.md
  internal/
    config/                        -- env parsing, defaults, validation
    db/
      db.go                        -- open, pragmas, migrate
      migrations/
        0001_init.sql              -- all tables above
        000N_*.sql                 -- future
    auth/
      magic.go                     -- token gen, hash, email send / log
      session.go                   -- cookie, HMAC, middleware
      mailer.go                    -- SMTP impl + log impl behind interface
    catalog/
      items.go                     -- CRUD items
      bundles.go                   -- CRUD bundles + bundle_items
      trash.go                     -- list / restore / purge
    trips/
      trips.go                     -- CRUD trips, members
      render.go                    -- merge engine (Section 6)
      pack.go                      -- check-off state
    web/
      router.go                    -- routes + middleware wiring
      handlers/
        auth.go items.go bundles.go trips.go pack.go trash.go health.go
      templates/
        layout.html
        partials/
          item-row.html bundle-pill.html pack-checkbox.html progress.html ...
        pages/
          login.html trips-list.html trip-detail.html
          items.html bundles.html bundle-edit.html trash.html
      static/
        pico.min.css htmx.min.js app.css favicon.ico
  cmd/
    seed/main.go                   -- starter catalog populator
  data/                            -- gitignored, runtime DB + backups
```

Boundaries:
- `catalog/`, `trips/`, `auth/` know nothing of HTTP. Pure domain over `*sql.DB`.
- `web/handlers/` is thin: parse → call domain → render template.
- `render.go` is pure: takes loaded rows, returns `[]RenderedItem`. No DB access
  in the merge step itself; the loader is a separate function.
- All handlers take a single `*App` struct (DB, mailer, templates, config) — no
  globals.

## 9. Routes & UX

### Route table

```
GET    /login                              login form
POST   /login                              issue magic token
GET    /auth/verify?t=...                  consume token, set session, redirect /
POST   /logout

GET    /                                   trips list (mine + shared with me)
GET    /trips/new                          new-trip form
POST   /trips                              create
GET    /trips/{id}                         trip detail
PATCH  /trips/{id}                         inline edit name/nights/notes
DELETE /trips/{id}                         soft-delete

POST   /trips/{id}/bundles                 attach bundle
DELETE /trips/{id}/bundles/{bid}           detach bundle

POST   /trips/{id}/extras                  add ad-hoc item
PATCH  /trips/{id}/items/{iid}             override qty / mark removed
POST   /trips/{id}/pack/{iid}              toggle packed

POST   /trips/{id}/members                 invite editor by email
DELETE /trips/{id}/members/{uid}           remove member (owner only)

GET    /items                              catalog page
POST   /items                              create
PATCH  /items/{id}                         edit
DELETE /items/{id}                         soft-delete

GET    /bundles                            list
GET    /bundles/{id}                       editor
POST   /bundles                            create
PATCH  /bundles/{id}                       edit metadata
DELETE /bundles/{id}                       soft-delete
POST   /bundles/{id}/items                 add item to bundle
DELETE /bundles/{id}/items/{iid}           remove item from bundle
POST   /bundles/{id}/children              nest a child bundle (cycle-checked)
DELETE /bundles/{id}/children/{cid}        unnest a child bundle

GET    /trash                              trash view
POST   /trash/{kind}/{id}/restore          restore (kind = items|bundles|trips)
DELETE /trash/{kind}/{id}                  hard purge

GET    /export                             JSON dump (catalog + trips visible to caller)
POST   /import                             restore from JSON

GET    /healthz                            liveness
```

Most mutating routes return an HTMX fragment (e.g. updated item row, refreshed
progress bar). Whole-page reloads happen only on explicit navigation.

### Trip detail page layout

Three panels stacked on mobile, columns on desktop:

1. **Header:** trip name (inline-editable), nights, starts-on, members, progress
   bar (`12 / 30 packed`).
2. **Bundles attached:** chip row with attached bundle names. Each chip has a
   small remove (×) action. A "+ Add bundle" button opens a searchable picker.
3. **Final list:** rendered, grouped by category. Each row:
   - checkbox (toggles `trip_pack_state.packed`)
   - item name
   - qty (inline-editable for override; shows computed value otherwise)
   - source label ("from washbag-basic" or "from washbag-basic + swimming-pool")
   - row menu: override qty, remove (sets `trip_overrides.removed = 1`)
4. **Extras:** quick-add form (item picker + qty), then list of extras with
   remove action.
5. **Members:** list, role badges, "invite by email" form (owner only).

### Catalog pages

- `/items`: table. Columns: name, category, per-night, default qty, actions.
  Inline create at the top. Inline edit per row.
- `/bundles`: list with name, description, item count.
- `/bundles/{id}`: name + description editable; two lists side by side:
  - **Items in this bundle:** add via item picker with qty; remove per row.
  - **Nested bundles:** add via bundle picker (excludes self + transitive
    descendants to prevent cycles); remove per row.

### Mobile

- Single column under 768px (Pico default).
- Sticky header with trip name + progress on `/trips/{id}`.
- Tap targets stay above 44px (Pico default checkboxes are fine).

## 10. Auth & security

### Magic-link login (invite-only)

The instance is invite-only. New users are never created from the public login
form. They only enter the system via:

- The `BOOTSTRAP_EMAIL` env var on first boot (creates one user row if the
  table is empty), or
- A trip-member invite from an existing user (creates the user row and the
  `trip_members` row in one transaction).

`POST /login` flow:

1. Read email. Lower-case it.
2. Look up `users` by email. If absent, return the same generic
   "check your email" page and do not send anything (constant-time response;
   this prevents email enumeration without leaking the allowlist).
3. Generate 32 random bytes, base64-url encode → raw token.
4. Insert into `magic_tokens` with `token_hash = sha256(raw)`, expiry `now + 15m`.
5. Send `${BASE_URL}/auth/verify?t={raw}` to the email. If `SMTP_HOST` is empty,
   log the URL at INFO level and return the same generic page (which mentions
   the log fallback).

Trip-member invite flow:

1. Owner submits an email on `POST /trips/{id}/members`.
2. Server upserts `users` (creates row if absent), inserts `trip_members`
   row with `role = 'editor'`, then issues a magic-link as in steps 3-5 above.
3. The invited user clicks the link, lands on `/`, and immediately sees the
   shared trip in their list.

### Verification

1. `GET /auth/verify?t=...`: hash, look up by hash, ensure `used_at IS NULL`
   and not expired.
2. Mark `used_at = now`.
3. Create a session row, expiry `now + 30d`.
4. Set cookie `sid` = `session_id . hmac(session_id, SESSION_SECRET)`,
   `HttpOnly`, `SameSite=Lax`, `Secure` when `BASE_URL` is https, `Path=/`.
5. Redirect to `/`.

### Sessions

- Cookie format: `<id>.<hex_hmac>`. Server splits, verifies HMAC, then DB lookup.
- Sliding renewal on each request: if expiry within 7 days, push expiry to
  `now + 30d`.
- Logout deletes the row and clears the cookie.
- `SESSION_SECRET`: read from env; if absent, read from
  `$DATA_DIR/.session_secret`; if file absent, generate 32 bytes and write
  the file with `0600` perms.

### Authorization

- Middleware loads session → user. On miss:
  - HTML request: 302 to `/login`.
  - HTMX request (`HX-Request: true`): 401 plus `HX-Redirect: /login`.
- Trip routes: caller must be in `trip_members` for the trip.
- Owner-only actions: delete trip, manage members. Enforced at handler level.
- Catalog routes: any authenticated user.

### Rate limiting

- In-memory token bucket on `/login`, keyed by lowercased email and by client
  IP. Limit: 10 requests per hour per key. State is process-local; resets on
  restart, which is acceptable.

### CSRF

- Per-session CSRF token stored alongside session cookie (separate cookie
  `csrf`, not `HttpOnly` so JS can read it for HTMX).
- All mutating handlers require header `X-CSRF-Token` matching the cookie.
  HTMX configured globally via `hx-headers` to inject it.
- Plain HTML forms include a hidden input.

### Input handling

- Parameterised queries everywhere. No string interpolation into SQL.
- All handler inputs validated: length caps (names ≤200, notes ≤4000),
  integer ranges (`nights` 0–365, qty 0–999), allowed characters where relevant.
- On 400, return an HTMX-friendly inline error fragment, not a full page.

### Errors

- Domain errors: `ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrValidation`.
- HTTP layer maps to 404 / 403 / 409 / 400.
- Other errors → 500, logged with full detail and a request ID; user sees a
  generic page citing the request ID.
- Logging: stdlib `slog` JSON to stdout. Container logs are the audit trail.

### Backups

- Optional shell script in image: `sqlite3 /data/data.db ".backup /data/backups/$(date +%F).db"`.
- README documents how to wire it as a host cron, or via a sidecar.
- Keep last 14 by default in the script.

## 11. Testing strategy

- **Render engine** (`internal/trips/render.go`): table-driven unit tests.
  Cases: per-night × nights, fixed-item max across bundles, override removed,
  override qty, multi-bundle dedup by item ID, deleted source item excluded,
  deleted bundle excluded, nested bundle expansion (parent → child → items),
  diamond inheritance (two parents share a grandchild — items appear once),
  deleted nested child excluded from parent expansion.
- **Bundle cycle check** (`internal/catalog/bundles.go`): tests for direct
  cycle (A→B→A), indirect cycle (A→B→C→A), and self-loop rejection. Plus
  successful nesting up to several levels.
- **DB-touching tests:** real SQLite in `t.TempDir()`. Cover migrations, soft
  delete + restore + purge cascades, FK enforcement, unique constraints.
- **HTTP tests:** `net/http/httptest`. One happy path per handler plus auth /
  authorization checks (unauthed → redirect, non-member → 403, non-owner →
  403 on owner-only routes). Avoid duplicating render tests at the HTTP layer.
- **Auth tests:** mailer behind an interface; tests use an in-memory
  capture mailer to assert link generation and single-use token consumption.
- **No mocking of SQLite.** Real DB, temp file. Suite target: <2s.
- CI: GitHub Actions running `go test ./...` plus `go vet`. Optional but
  recommended.

## 12. Seed data

Provided via `cmd/seed/main.go` (also runnable from main with `--seed` flag).
Idempotent: skips if `items` already non-empty.

### Items (~40, illustrative)
General: passport, wallet, keys, glasses, sunglasses, charger-phone,
charger-laptop, headphones, book, water-bottle.
Toiletries: toothbrush, toothpaste, shampoo, conditioner, soap, deodorant,
razor, floss.
Clothing (per-night flagged): underwear, socks, t-shirt. Fixed: trousers,
shorts, jumper, waterproof, swimsuit, pyjamas, hat, gloves.
Sport: goggles, swim-cap, towel-quickdry, towel-cotton, running-shoes,
trail-shoes, running-shorts, headtorch, gels.

### Bundles
- `washbag-basic`: toothbrush, toothpaste, deodorant, shampoo, soap.
- `washbag-full`: above + conditioner, razor, floss.
- `electronics-day`: charger-phone, headphones.
- `electronics-week`: above + charger-laptop.
- `swimming-pool`: swimsuit, goggles, towel-quickdry.
- `swimming-coldsea`: swimsuit, towel-cotton, hat (warm), waterproof.
- `running-road`: running-shoes, running-shorts, t-shirt.
- `running-trail`: trail-shoes, running-shorts, t-shirt, headtorch, gels.
- `weekend-uk`: underwear, socks, t-shirt, trousers, jumper, waterproof.
  Nests `washbag-basic` and `electronics-day`.
- `beach-week`: underwear, socks, t-shirt, shorts, sunglasses, swimsuit,
  towel-quickdry. Nests `washbag-basic` and `electronics-week`.

Nested bundles are live references: editing `washbag-basic` updates every
parent and every trip using either parent.

(Final composition can be tuned during implementation; this is a starter, not a contract.)

## 13. Deployment

### Dockerfile (sketch)

```dockerfile
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/packing-list .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/packing-list /packing-list
USER nonroot:nonroot
ENV DATA_DIR=/data
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["/packing-list"]
```

### Run

```
docker run -d --name packing-list \
  -p 8080:8080 \
  -v /srv/packing-list/data:/data \
  -e BASE_URL=https://pack.example.com \
  -e SMTP_HOST=smtp.example.com \
  -e SMTP_USER=apikey \
  -e SMTP_PASS=*** \
  -e SMTP_FROM=no-reply@example.com \
  packing-list:latest
```

A reverse proxy (Caddy / nginx / Traefik / Cloudflare Tunnel) terminates TLS.
The app trusts `X-Forwarded-Proto` only when behind a proxy; configurable via
env if needed.

## 14. Manual verification checklist

Used to gate release after implementation:

- Create trip → add `washbag-basic` and `running-road` → final list shows
  merged items grouped by category with correct quantities for trip nights.
- Tick items off → progress bar updates → restart container → check state
  persists.
- Edit `washbag-basic` (add floss) → existing trip's list now includes floss
  (live reference confirmed).
- Attach `weekend-uk` (which nests `washbag-basic`) to a trip → trip list
  contains items from both bundles. Toothbrush from washbag appears once
  even if `washbag-basic` is also attached directly (de-dup verified).
- Try to nest `weekend-uk` inside `washbag-basic` → server rejects with a
  cycle error (edge: weekend-uk already nests washbag-basic).
- Override qty on a single trip item → bundle untouched.
- Remove an item via override → item disappears from rendered list.
- Invite member by email; SMTP unset → magic link printed to log; open in
  private window → land on shared trip and edit.
- Delete bundle → trash shows it → restore → trip composition returns.
- Purge a deleted item → all references gone, no orphans.
- Export → wipe DB → import → fully restored.
- Run on a Raspberry Pi (ARM64) using the same image.

## 15. Open questions / future work

These are explicitly out of scope but worth noting:

- Conditional bundle logic ("if cold add fleece"). Triggered by Q2 option C.
- Per-trip "skip without delete" flag (Q10 option C) — already partly satisfied
  by `trip_overrides.removed`; could surface as a UI toggle distinct from
  remove.
- Multi-instance sync (e.g. holiday cottage offline copy). Probably JSON
  export/import remains the answer.
- Bundle categories / tags for quicker discovery as the catalog grows.
- Real-time co-editing. Likely never needed at this scale.
- Auto-purge of trash after N days.
