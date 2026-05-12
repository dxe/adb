# Agents

Note: CLAUDE.md is a symlink to this file.

## architecture

Background context on how the stack fits together.

### frontend/backend routing

- The Go server reverse-proxies `/v2/*` to the Next.js app in `./frontend-v2` (see `proxyHandler` in `server/src/main.go`). Next.js has `basePath: '/v2'` configured (`next.config.ts`), so it automatically prepends `/v2` to all routes. Next.js navigation (`<Link href>`, `router.push()`, `redirect()`, and route definitions) must therefore omit the `/v2` prefix — Next.js adds it. For example, `<Link href="/activists/123">`, not `<Link href="/v2/activists/123">`.
- The Go API lives at the same origin. The `ApiClient` in `frontend-v2/src/lib/api.ts` calls paths like `api/activists`, `event/get`, etc. directly without a `/v2` prefix.

## frontend

Rules regarding the Next.js app in ./frontend-v2 and legacy Vue app in ./frontend

### frontend-file-formatting

- Scope: ./frontend and ./frontend-v2
- Rule: After modifying files, run `pnpx prettier <filename> --write`
- Example: `pnpx prettier 'frontend-v2/src/app/login/page.tsx' --write`

### frontend-v2-typecheck-and-lint

- Scope: ./frontend-v2
- Rule: `tsc` and `eslint` must be run from inside `./frontend-v2` (not `/workspace`) — that's where `tsconfig.json` and `eslint.config.mjs` live.
- Typecheck: `cd frontend-v2 && pnpm exec tsc --noEmit` (do not use `pnpx tsc` — it hits a shim that refuses to run)
- Lint: `cd frontend-v2 && pnpm lint` (equivalent to `eslint .`)

## backend

Rules regarding Go server in ./server

### go-backend-build

- Scope: `server/**/*.go`
- Rule: Build the Go server with `cd server/src && go build ./...`

### go-deps-sync

- Scope: `**/go.mod`, `**/go.sum`, `go.work`, `go.work.sum`
- Rule: After adding or updating Go dependencies in any module (`pkg`, `cli`, `server/src`), run `make go_mod_sync`. This ensures all transitive dependency hashes are fully resolved at both the individual module and workspace levels.

### go-backend-error-wrapping

- Scope: `server/**/*.go`
- Rule: Do not use `github.com/pkg/errors` for new code.

### go-test-runner

- Scope: Go test execution tasks
- Rule: Delegate `go test` runs to a subagent and return only a brief summary to avoid noisy MySQL testcontainer/migration logs in the main context.

## database

Rules regarding the MySql database.

### connect-to-mysql

- Scope: database schema questions and live database manipulation
- Rule: query the live MySQL schema via the `db` shell function which wraps the `mysql` with the correct database and dev credentials. It is already sourced for you from `~/.bash_profile`. Fall back to migration files only as needed.
- Example: `db -e "SHOW COLUMNS FROM activists"`.

## CLI

Context for using the CLI, located in ./cli

### cli-usage

- The CLI is available to you as a shell function called `adb`. It is already sourced for you from `~/.bash_profile` and environment variables are already configured to connect to the dev database.
- You can run `adb -h` for usage instructions.

## preferences

- After making and verifying changes, do not needlessly start summaries with
  "Clean." or "Clean build." as this gets repetitive.
