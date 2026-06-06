package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dxe/adb/jobs/internal/db"
	"github.com/dxe/adb/jobs/internal/secrets"
	"github.com/dxe/adb/pkg/activists"
)

const (
	dbHostParam     = "mysql_lambda_host"
	dbUserParam     = "mysql_lambda_user"
	dbPasswordParam = "mysql_lambda_password"
)

func handler(ctx context.Context) (string, error) {
	sec, err := secrets.New(ctx)
	if err != nil {
		return "", err
	}
	host, err := sec.Get(ctx, dbHostParam)
	if err != nil {
		return "", err
	}
	user, err := sec.Get(ctx, dbUserParam)
	if err != nil {
		return "", err
	}
	password, err := sec.Get(ctx, dbPasswordParam)
	if err != nil {
		return "", err
	}

	conn, err := db.Open(ctx, host, user, password)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	repo := activists.NewRepository(conn)

	// Placeholder for future community-report logic.
	count, err := repo.CountActivists(activists.QueryActivistFilters{})
	if err != nil {
		return "", fmt.Errorf("counting activists: %w", err)
	}

	log.Printf("community-reports: counted %d activists", count)
	return fmt.Sprintf("counted %d activists", count), nil
}

func main() {
	lambda.Start(handler)
}
