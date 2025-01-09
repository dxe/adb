.PHONY: run_all run watch test clean prod_build deps dev_db fmt

# Runs the application (builds Vue.js files, starts Next.js dev server, starts Go server).
run_all:
	. ~/.nvm/nvm.sh && \
    (cd frontend && nvm install 16 && npm run dev-build) && \
	(cd frontend-v2 && nvm install 18 && nvm use 18 && pnpm dev) &
	$(MAKE) run

# Just start the go program without recompiling the JS.
run:
	cd server/src && go install # Install first so that we keep cached build objects around.

	cd server/src; \
	export TEMPLATES_DIRECTORY=../templates; \
	export STATIC_DIRECTORY=../../frontend/static; \
	export DIST_DIRECTORY=../../frontend/dist; \
	export JS_V2_DIRECTORY=../../frontend-v2/out; \
	export NEXT_JS_PROXY_URL=http://localhost:3000; \
	go run main.go

# Builds the frontend Vue JS files.
js:
	. ~/.nvm/nvm.sh && nvm use 16 && cd frontend && npm run dev-build

# Automatically rebuilds the Vue JS app when you edit a file. This is
# more convenient then manually running `make run_all` every time you
# update the JS. You'll need to do this in a separate terminal.
watch:
	cd frontend && npm run watch

# Wipe and re-create the dev databases. See the readme for more
# details.
dev_db:
	cd server/scripts/create_db_wrapper && ./create_db_wrapper.sh

# Install all deps for this project.
deps:
	cd frontend && npm install --legacy-peer-deps
	cd frontend-v2 && pnpm i
	cd server/src && go get -t github.com/dxe/adb/...

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
docker_run:
	docker build . -t dxe/adb
	docker container run --rm -p 8080:8080 -it --name adbtest dxe/adb

# Open shell inside docker container while it's running
docker_shell:
	docker exec -it adbtest /bin/ash

# Build the project for production.
prod_build: clean set_git_hooks
	cd server && ./scripts/pull_adb_config.sh
	cd frontend && npm run build
	cd server/src && env GOOS=linux GOARCH=amd64 go build

# Reformat source files.
# Keep in sync with hooks/pre-commit.
fmt:
	cd server && gofmt -w `find . -name '*.go'`
	cd frontend && npx prettier --write *.{ts,vue,js}
	cd frontend-v2 && pnpm fmt
