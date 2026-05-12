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
