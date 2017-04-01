.PHONY: run deps test


run: adb.db
	go run main.go

adb.db: ./model/db.go ./scripts/create_db.go
	go run ./scripts/create_db.go

deps:
	go get github.com/directactioneverywhere/adb

test:
	go test ./...
