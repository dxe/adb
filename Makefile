.PHONY: run_all run watch test clean prod_build deps dev_db samer_deploy jake_deploy

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

# Build the project for production.
prod_build:
	npm run build
	env GOOS=linux GOARCH=amd64 go build

samer_deploy: clean prod_build
	ssh samer@adb.dxetech.org "sudo svc -d /etc/service/adb"
	rsync --chmod=ug+w --groupmap="*:adb" -azPO --delete adb run templates static dist samer@adb.dxetech.org:/opt/adb/
	ssh samer@adb.dxetech.org "sudo svc -u /etc/service/adb"

jake_deploy: clean prod_build
	ssh ubuntu@adb.dxetech.org "sudo svc -d /etc/service/adb"
	rsync --chmod=ug+w --groupmap="*:adb" -azPO --delete adb run templates static dist ubuntu@adb.dxetech.org:/opt/adb/
	ssh ubuntu@adb.dxetech.org "sudo svc -u /etc/service/adb"
