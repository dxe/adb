module github.com/dxe/adb/cli

go 1.23.0

require (
	github.com/dxe/adb/pkg v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.7.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

replace github.com/dxe/adb/pkg => ../pkg
