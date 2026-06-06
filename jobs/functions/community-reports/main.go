package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dxe/adb/jobs/internal/db"
	"github.com/dxe/adb/pkg/activists"
)

func handler(ctx context.Context) (string, error) {
	conn, err := db.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	repo := activists.NewRepository(conn)

	// Placeholder for future community-report logic. This natively runs the
	// shared activist query code (no HTTP call to the server) and logs a count.
	// Future work can build a richer activists.QueryActivistShape and call
	// repo.StreamActivists to produce the actual report.
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
