.PHONY: run_all run watch test clean prod_build deps dev_db fmt

# When not using devcontainer, NVM initialization script may be located in home
# directory. In the devcontainer, it is in /usr/local/share/nvm/.
NVM_SCRIPT := $(shell \
    if [ -s "$(HOME)/.nvm/nvm.sh" ]; then \
      echo "$(HOME)/.nvm/nvm.sh"; \
    elif [ -s "/usr/local/share/nvm/nvm.sh" ]; then \
      echo "/usr/local/share/nvm/nvm.sh"; \
    else \
      echo "Error: nvm.sh not found in either location." >&2; \
      exit 1; \
    fi)

# Runs the application (builds Vue.js files, starts Next.js dev server, starts Go server).
# As of January 2025, upgrading past Node 16 breaks old Vue dependencies, and Node 18 is required to use the latest version of React.
run_all:
	. $(NVM_SCRIPT) && \
	export NEXT_PUBLIC_API_BASE_URL=http://localhost:8080; \
    (cd frontend && nvm use 16 && npm run dev-build); \
	(cd frontend-v2 && nvm use 18 && pnpm dev) &
	$(MAKE) run

# Just start the go program without recompiling the JS.
run:
	cd server/src && go install # Install first so that we keep cached build objects around.

	set -a && . server/debug.env && set +a && \
	cd server/src && \
	go run main.go

# Builds the frontend Vue JS files.
js:
	. $(NVM_SCRIPT) && nvm use 16 && cd frontend && npm run dev-build

# Automatically rebuilds the Vue JS app when you edit a file. This is
# more convenient then manually running `make run_all` every time you
# update the JS. You'll need to do this in a separate terminal.
watch:
	cd frontend && npm run watch

# Wipe and re-create the dev databases. See the readme for more
# details.
dev_db:
	export DXE_DEV_EMAIL=test-dev@directactioneverywhere.com && \
	cd server/scripts/create_db_wrapper && ./create_db_wrapper.sh

# Install all deps for this project.
# Note: PNPM must be installed separately for each version of NPM used, since it is installed within each NPM installation.
# Note: `go tool` cannot yet be used to install golang-migrate: https://github.com/golang-migrate/migrate/issues/1232
deps:
	. $(NVM_SCRIPT) && nvm i 22 && npm i -g pnpm && pnpm i
	. $(NVM_SCRIPT) && cd frontend && nvm i 16 && npm i --legacy-peer-deps
	. $(NVM_SCRIPT) && cd frontend-v2 && nvm i 18 && npm i -g pnpm && pnpm i
	cd server/src && go get -t github.com/dxe/adb/...
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run all tests
test:
	cd server/src && go test ./...

# Clean all built outputs
clean:
	rm -f server/adb
	rm -rf frontend/dist
	rm -rf frontend-v2/out

# Set git hooks
set_git_hooks:
	if [ ! -h .git/hooks/pre-commit ] ; then ln -s ../../hooks/pre-commit .git/hooks/pre-commit ; fi
	if [ ! -h .git/hooks/pre-push ] ; then ln -s ../../hooks/pre-push .git/hooks/pre-push ; fi


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

# Reformat source files.
# Keep in sync with hooks/pre-commit.
fmt:
	cd server && gofmt -w `find . -name '*.go'`
	. $(NVM_SCRIPT) && nvm use 22 && pnpx prettier --write .
