.PHONY: run_all run watch test clean prod_build deps dev_db

# Runs the application.
run_all:
	npm run dev-build
	go run main.go

# Just start the go program without recompiling the JS.
run:
	go run main.go

# Builds the frontend JS.
js:
	npm run dev-build

# Automatically rebuild the JS when you edit a JS file. This is more
# convenient then manually running `make run_all` every time you
# update the JS. You'll need to do this in a separate terminal.
watch:
	npm run watch

# Wipe and re-create the dev databases. See the readme for more
# details.
dev_db:
	go run ./scripts/create_db.go

# Install all deps for this project.
deps:
	npm install
	go get -t github.com/dxe/adb/...

# Run all tests
test:
	go test ./...

# Clean all built outputs
clean:
	rm -f adb
	rm -rf dist

# Set git hooks
set_git_hooks:
	if [ ! -h .git/hooks/pre-push ] ; then ln -s hooks/pre-push .git/hooks/pre-push ; fi

# Build the project for production.
prod_build: clean set_git_hooks
	./scripts/pull_adb_config.sh
	npm run build
	env GOOS=linux GOARCH=amd64 go build
