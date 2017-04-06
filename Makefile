.PHONY: run deps test dev_db samer_deploy


run:
	go run main.go

dev_db:
	go run ./scripts/create_db.go

deps:
	go get github.com/directactioneverywhere/adb

test:
	go test ./...

samer_deploy:
	rm -f adb
	go build
	scp adb samer@adb.dxetech.org:~/adb
	scp -r templates samer@adb.dxetech.org:~/
	scp -r static samer@adb.dxetech.org:~/
	@echo "\nTo deploy, log onto the server and run:"
	@echo "pkill adb # to kill the running server"
	@echo "nohup ./adb -prod & # run adb in the background"
