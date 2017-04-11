.PHONY: run deps test dev_db samer_deploy jake_deploy


run:
	go run main.go

dev_db:
	go run ./scripts/create_db.go

deps:
	go get -t github.com/directactioneverywhere/adb

test:
	go test ./...

samer_deploy:
	rm -f adb
	go build
	ssh samer@adb.dxetech.org "sudo svc -d /etc/service/adb"
	scp adb samer@adb.dxetech.org:/opt/adb/
	scp run samer@adb.dxetech.org:/opt/adb/
	scp -r templates samer@adb.dxetech.org:/opt/adb/
	scp -r static samer@adb.dxetech.org:/opt/adb/
	ssh samer@adb.dxetech.org "sudo svc -u /etc/service/adb"

jake_deploy:
	rm -f adb
	env GOOS=linux GOARCH=amd64 go build
	ssh ubuntu@adb.dxetech.org "sudo svc -d /etc/service/adb"
	scp adb ubuntu@adb.dxetech.org:/opt/adb/
	scp run ubuntu@adb.dxetech.org:/opt/adb/
	scp -r templates ubuntu@adb.dxetech.org:/opt/adb/
	scp -r static ubuntu@adb.dxetech.org:/opt/adb/
	ssh ubuntu@adb.dxetech.org "sudo svc -u /etc/service/adb"
