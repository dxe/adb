.PHONY: run_all run test test-server test-frontend lint clean prod_build deps dev_db install_playwright fmt go_mod_sync _go_mod_sync cloud_setup cloud_start

# When not using devcontainer, NVM initialization script may be located in home
# directory. In the devcontainer, it is in /usr/local/share/nvm/. In the Claude
# Code cloud VM, it is in /opt/nvm/. $NVM_DIR takes precedence if set.
NVM_SCRIPT := $(shell \
    if [ -n "$(NVM_DIR)" ] && [ -s "$(NVM_DIR)/nvm.sh" ]; then \
      echo "$(NVM_DIR)/nvm.sh"; \
    elif [ -s "$(HOME)/.nvm/nvm.sh" ]; then \
      echo "$(HOME)/.nvm/nvm.sh"; \
    elif [ -s "/opt/nvm/nvm.sh" ]; then \
      echo "/opt/nvm/nvm.sh"; \
    elif [ -s "/usr/local/share/nvm/nvm.sh" ]; then \
      echo "/usr/local/share/nvm/nvm.sh"; \
    else \
      echo "Error: nvm.sh not found in either location." >&2; \
      exit 1; \
    fi)

# Please keep versions in sync
# * package.json
# * all Dockerfiles
# * .github/workflows/main.yml
# * devcontainer.json `ghcr.io/devcontainers/features/node` feature
REACT_FRONTEND_NODE_VERSION := 25
PNPM_VERSION := 10.32.1

# Port the Go server listens on. Defaults to 8080; override with `make run_all
# PORT=NNNN` or a PORT environment variable so parallel checkouts can use
# distinct ports. The frontend's browser code uses relative URLs, so only
# server-side rendering reads NEXT_PUBLIC_API_BASE_URL.
PORT ?= 8080

# Runs the application (starts Next.js dev server, starts Go server).
run_all:
	. $(NVM_SCRIPT) && \
	export NEXT_PUBLIC_API_BASE_URL=http://localhost:$(PORT); \
	(cd frontend-v2 && nvm use $(REACT_FRONTEND_NODE_VERSION) && env -u PORT pnpm dev) &
	$(MAKE) run

# Just start the go program.
run:
	@[ -f server/.env ] || { echo "Error: server/.env not found. Run 'make deps' first."; exit 1; }

	cd server/src && go install # Install first so that we keep cached build objects around.

	set -a && \
	. server/debug.env && \
	. server/.env && \
	set +a && \
	cd server/src && \
	PORT=$(PORT) go run main.go

# Wipe and re-create the dev databases. See the readme for more
# details.
dev_db:
	cd cli && go run . db create --dev-email="${DXE_DEV_EMAIL:-test-dev@directactioneverywhere.com}"

# Install the browser binary and system libraries used by the Playwright MCP.
install_playwright:
	. $(NVM_SCRIPT) && nvm i $(REACT_FRONTEND_NODE_VERSION) && bash .devcontainer/install-playwright.sh

# One-time setup for the Claude Code cloud VM (no devcontainer). Call from the
# environment's setup script. See .claude/cloud/.
cloud_setup:
	.claude/cloud/setup.sh

# Per-session startup for the Claude Code cloud VM: dockerd + MySQL + dev data.
# Run automatically by the SessionStart hook in .claude/settings.json.
cloud_start:
	.claude/cloud/start.sh

# Install all deps for this project.
# Note: PNPM must be installed separately for each version of NPM used, since it is installed within each NPM installation.
# Note: `go tool` cannot yet be used to install golang-migrate: https://github.com/golang-migrate/migrate/issues/1232
deps:
	@# Ensure server/.env exists. Both `make run` and the VS Code "Go Server"
	@# launch config load it unconditionally (layered on server/debug.env) and
	@# error if it's missing, so deps (the required first step) creates it.
	touch server/.env
	. $(NVM_SCRIPT) && nvm i 22 && npm i -g pnpm@$(PNPM_VERSION) && pnpm i --config.confirmModulesPurge=false
	. $(NVM_SCRIPT) && cd frontend-v2 && nvm i $(REACT_FRONTEND_NODE_VERSION) && npm i -g pnpm@$(PNPM_VERSION) && pnpm i --config.confirmModulesPurge=false
	cd pkg && go mod download
	cd server/src && go mod download
	cd cli && go mod download
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(MAKE) go_mod_sync

# Normalize module/workspace metadata across all Go modules. Runs twice
# because workspace-level resolution can surface new transitive deps that
# require a second pass to stabilize go.mod/go.sum.
go_mod_sync:
	$(MAKE) _go_mod_sync
	$(MAKE) _go_mod_sync

# Normalize Go module metadata across the workspace.
#
# Tidies each module with GOWORK=off so each go.mod/go.sum reflects only
# its own dependency graph (independent of the workspace), then `go work
# sync` reconciles selected versions across the workspace.
#
# The trailing `go list -m all` resolves the workspace module graph so
# go.work.sum picks up the /go.mod hashes Go's lazy loader records for
# transitive deps it had to consider but didn't select. Without it, the
# first workspace-mode invocation afterward (another `go list -m all`,
# gopls in the IDE, etc.) writes those entries and leaves go.work.sum
# dirty.
_go_mod_sync:
	cd pkg && GOWORK=off go mod tidy
	cd cli && GOWORK=off go mod tidy
	cd server/src && GOWORK=off go mod tidy
	go work sync
	cd server/src && go list -m all >/dev/null

# Run all tests
test: test-server test-frontend

test-server:
	cd server/src && go test ./...

test-frontend:
	. $(NVM_SCRIPT) && cd frontend-v2 && nvm use $(REACT_FRONTEND_NODE_VERSION) && pnpm test --run

# Run golangci-lint on all Go modules.
# TODO: Run linter automatically once existing lint errors are fixed.
lint:
	for mod in pkg cli server/src; do \
		(cd $$mod && golangci-lint run ./...) || exit $$?; \
	done

# Clean all built outputs
clean:
	rm -f cli/adb
	rm -f server/adb-server
	rm -rf frontend-v2/out
	rm -rf frontend-v2/.next

# Set git hooks
set_git_hooks:
	if [ ! -h .git/hooks/pre-commit ] ; then ln -s ../../.githooks/pre-commit .git/hooks/pre-commit ; fi
	if [ ! -h .git/hooks/pre-push ] ; then ln -s ../../.githooks/pre-push .git/hooks/pre-push ; fi


# Test docker image
# (Note, --net=host runs the container in the same network as the devcontainer created by the devcontainer's docker-compose config)
docker_run:
	docker build . -t dxe/adb
	docker container run --rm -p 8080:8080 --net=host -it --name adbtest --env-file docker-debug.env dxe/adb

# Open shell inside docker container while it's running
docker_shell:
	docker exec -it adbtest /bin/ash

# Build the project for production.
prod_build:
	docker build . -t dxe/adb
	docker build . -f Dockerfile.frontend-v2 -t dxe/adb-next
	docker build . -f Dockerfile.cli -t dxe/adb-cli

# Reformat source files.
# Keep in sync with .githooks/pre-commit.
fmt:
	cd server && gofmt -w .
	. $(NVM_SCRIPT) && nvm use 22 && pnpm exec prettier --write --cache --cache-strategy metadata .
