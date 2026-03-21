module github.com/dxe/adb/cli

// Keep in sync with /workspace/server/src/go.mod.
go 1.25.0

require (
	github.com/dxe/adb/pkg v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.7.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opentelemetry.io/otel v1.41.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

replace github.com/dxe/adb/pkg => ../pkg
