# Agents

Note: CLAUDE.md is a symlink to this file.

## architecture

Background context on how the stack fits together.

### frontend/backend routing

- The Go server reverse-proxies `/v2/*` to the Next.js app in `./frontend-v2` (see `proxyHandler` in `server/src/main.go`). Next.js itself is unaware of this prefix — its routes start at `/`, so internal links and route definitions should never include `/v2`. For example, link to `/activists/123`, not `/v2/activists/123`.
- The Go API lives at the same origin. The `ApiClient` in `frontend-v2/src/lib/api.ts` calls paths like `api/activists`, `event/get`, etc. directly without a `/v2` prefix.

## frontend

Rules regarding the Next.js app in ./frontend-v2 and legacy Vue app in ./frontend

### frontend-file-formatting

- Scope: ./frontend and ./frontend-v2
- Rule: After modifying files, run `pnpx prettier <filename> --write`

### frontend-v2-typecheck-and-lint

- Scope: ./frontend-v2
- Rule: `tsc` and `eslint` must be run from inside `./frontend-v2` (not `/workspace`) — that's where `tsconfig.json` and `eslint.config.mjs` live.
- Typecheck: `cd frontend-v2 && pnpm exec tsc --noEmit` (do not use `pnpx tsc` — it hits a shim that refuses to run)
- Lint: `cd frontend-v2 && pnpm lint` (equivalent to `eslint .`)

## backend

Rules regarding Go server in ./server

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
