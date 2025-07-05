module github.com/dxe/adb/scripts/create_db_wrapper

go 1.23.0

toolchain go1.24.3

require github.com/dxe/adb v0.0.0

require (
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

replace github.com/dxe/adb => ../../src
