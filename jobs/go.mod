module github.com/dxe/adb/jobs

// Keep in sync with server/src/go.mod
go 1.25.8

require (
	github.com/aws/aws-lambda-go v1.50.0
	github.com/dxe/adb/pkg v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.10.0
	github.com/jmoiron/sqlx v1.4.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

// The shared activist query code lives in the pkg module.
replace github.com/dxe/adb/pkg => ../pkg
