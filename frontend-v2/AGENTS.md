<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# frontend-v2 instructions

## routing

- Next.js has `basePath: '/v2'` configured (`next.config.ts`), so it automatically prepends `/v2` to all routes. Next.js navigation (`<Link href>`, `router.push()`, `redirect()`, and route definitions) must therefore omit the `/v2` prefix — Next.js adds it. For example, `<Link href="/activists/123">`, not `<Link href="/v2/activists/123">`.
- The Go API lives at the same origin. The `ApiClient` in `src/lib/api.ts` calls paths like `api/activists`, `event/get`, etc. directly without a `/v2` prefix.

## formatting

- After modifying files, run `pnpx prettier <filename> --write`
- Example: `pnpx prettier 'frontend-v2/src/app/login/page.tsx' --write`

## typecheck and lint

- `tsc` and `eslint` must be run from inside `./frontend-v2` (not the repo root) — that's where `tsconfig.json` and `eslint.config.mjs` live.
- Typecheck: `cd frontend-v2 && pnpm exec tsc --noEmit` (do not use `pnpx tsc` — it hits a shim that refuses to run)
- Lint: `cd frontend-v2 && pnpm lint` (equivalent to `eslint .`)

## chapter-scoped data

- Admins switch chapters via `/auth/switch_chapter` (full page load; the session
  cookie's chapter changes).
- Data that depends on the active chapter should take
  `user.ChapterID` (from `useAuthedPageContext()` or `session.user`) as an
  argument and include it in the TanStack Query key, including any server prefetch.
