.PHONY: run


run: adb.db
	go run main.go

adb.db:
	./scripts/create_db.sh
