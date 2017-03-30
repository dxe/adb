.PHONY: run deps


run: adb.db
	go run main.go

adb.db: scripts/create_db.sh scripts/create_sqlite_db.sql
	./scripts/create_db.sh

deps:
	go get github.com/directactioneverywhere/adb
