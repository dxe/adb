package main

import (
	"context"

	"github.com/dxe/adb/jobs/internal/secrets"
)

const (
	dbHostParam     = "mysql_lambda_host"
	dbUserParam     = "mysql_lambda_user"
	dbPasswordParam = "mysql_lambda_password"
)

// dbConfig holds the database connection parameters fetched from the secrets
// store.
type dbConfig struct {
	host     string
	user     string
	password string
}

// loadDBConfig retrieves the database connection parameters from the secrets
// store.
func loadDBConfig(ctx context.Context, sec *secrets.Client) (dbConfig, error) {
	host, err := sec.Get(ctx, dbHostParam)
	if err != nil {
		return dbConfig{}, err
	}
	user, err := sec.Get(ctx, dbUserParam)
	if err != nil {
		return dbConfig{}, err
	}
	password, err := sec.Get(ctx, dbPasswordParam)
	if err != nil {
		return dbConfig{}, err
	}
	return dbConfig{host: host, user: user, password: password}, nil
}
