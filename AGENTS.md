# Agents

Note: CLAUDE.md is a symlink to this file.

## environments

This repo runs in the devcontainer (local dev machine) and in the Claude Code cloud VM (no devcontainer; `$CLAUDE_CODE_REMOTE` is `true`).
Cloud VM setup is managed in the repo; see [.claude/cloud/README.md](.claude/cloud/README.md) for details.

- The default branch is `main`.

### first-time setup and running

Tool versions (Go, Node, pnpm, golangci-lint, Delve, `migrate`) and dev tasks are defined in [mise.toml](mise.toml). Run `mise tasks` to list tasks.

Run these in order on a fresh checkout.

1. `mise run deps` — installs Go/Node/pnpm and other tools, creates `server/.env`, downloads Go and node deps.
2. MySQL must be running. In the devcontainer it already is; elsewhere: `mise run db:up`.
3. `mise run db:init` — creates the dev database and dev user if the database has no tables yet. To wipe and re-create them, run `mise run db:reset`; ask first, as it deletes all dev data.

In the cloud VM, the SessionStart hook runs `.claude/cloud/start.sh` to do this instead (note hooks are disabled in multi-repo environments)
and tasks that need the docker engine will start it automatically.

To run the app:

- `mise run dev:backend` — Go server (port 8080)
- `mise run dev:frontend` — Next.js dev server, proxied by Go server

Then access in browser at `http://localhost:8080/v2/`.

## architecture

Background context on how the stack fits together.

### frontend/backend routing

- The Go server reverse-proxies `/v2/*` to the Next.js app in `./frontend-v2` (see `proxyHandler` in `server/src/main.go`). The Go API lives at the same origin, without the `/v2` prefix.

## frontend

Rules regarding the Next.js app in ./frontend-v2

### UI vocabulary

- "The X page": most pages are found under `frontend-v2/src/app/(authed)/`, e.g. `activists`, `chapters`, `users`.
- "The navigation" is defined in `frontend-v2/src/components/nav-data.ts`.
- "The member portal" is a separate server-rendered Go app in `server/src/members/` (own port), not part of `frontend-v2`.

### frontend-v2-debugging

- Scope: ./frontend-v2
- Rule: Do not try to fix devserver issues such as live reloading issues without first running `mise run clean`, then `mise run deps` and asking user to restart devcontainer.

### playwright-mcp

- If running in the devcontainer, `mise run install-playwright` before invoking Playwright MCP to install the browser.
- In the Claude Code cloud VM, Playwright is already set up.

## backend

Rules regarding Go server in ./server

### go-backend-build

- Scope: `server/**/*.go`
- Rule: Build the Go server with `cd server/src && go build ./...`

### gopls-in-cloud-vm

- Scope: Go tasks in the cloud VM
- Rule: gopls isn't installed in the cloud VM. If you need it, install it with `go install golang.org/x/tools/gopls@latest` (~1 min).

### go-deps-sync

- Scope: `**/go.mod`, `**/go.sum`, `go.work`, `go.work.sum`
- Rule: After adding or updating Go dependencies in any module (`pkg`, `cli`, `server/src`), run `mise run go-mod-sync`. This ensures all transitive dependency hashes are fully resolved at both the individual module and workspace levels.

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
- Rule: query the live MySQL schema via the `db` mise task which wraps `mysql` with the correct database and dev credentials. Fall back to migration files only as needed.
- Example: `mise db -e "SHOW COLUMNS FROM activists"`.

### db-migrations

- Scope: creating, running, or modifying database migrations
- Rule: follow the conventions in [pkg/shared/db-migrations/README.md](pkg/shared/db-migrations/README.md).

## CLI

Context for using the CLI, located in ./cli

### cli-usage

- The CLI is available to you as a mise task called `adb` (e.g. `mise adb activists count`), which runs the CLI from source. Environment variables are already configured to connect to the dev database.
- You can run `mise adb -- -h` for usage instructions. Flags mise itself understands (`-h`, `--help`) must come after `--`.

## preferences

- After making and verifying changes, do not needlessly start summaries with
  "Clean." or "Clean build." as this gets repetitive.
