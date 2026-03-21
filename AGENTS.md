# Agent Workflow Notes

- After writing to files in frontend or frontend-v2, run `pnpx prettier <filename> --write`

- For database schema questions, query the live MySQL schema via the `db` shell function, falling back to migration files only as needed.

## frontend-v2 (Next.js)

- **`/v2` prefix**: The Go server reverse-proxies `/v2/*` to the Next.js app (see `proxyHandler` in `server/src/main.go`). Next.js itself is unaware of this prefix — its routes start at `/`. So internal links and route definitions in frontend-v2 should never include `/v2`. For example, link to `/activists/123`, not `/v2/activists/123`.

- **API endpoints**: The Go API lives at the same origin (no `/v2` prefix). The `ApiClient` in `frontend-v2/src/lib/api.ts` calls paths like `api/activists`, `event/get`, etc. directly.

## Go backend

- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` in new or changed code. Do not use `github.com/pkg/errors` — it is deprecated. Existing code still uses it but should not be copied.
