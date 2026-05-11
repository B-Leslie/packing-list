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
