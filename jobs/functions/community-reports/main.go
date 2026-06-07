package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dxe/adb/jobs/internal/db"
	"github.com/dxe/adb/jobs/internal/mailer"
	"github.com/dxe/adb/jobs/internal/secrets"
	"github.com/dxe/adb/pkg/activists"
)

const (
	toAddress    = "community@directactioneverywhere.com"
	emailSubject = "Community Report: New Activists"

	reportChapterID = 47
)

func handler(ctx context.Context) (string, error) {
	sec, err := secrets.New(ctx)
	if err != nil {
		return "", err
	}

	cfg, err := loadDBConfig(ctx, sec)
	if err != nil {
		return "", err
	}

	conn, err := db.Open(ctx, cfg.host, cfg.user, cfg.password)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	repo := activists.NewRepository(conn)

	now := time.Now().UTC()
	sections := []reportSection{
		newFirstEventLastWeekSection(now),
		newSingleEventSection(now),
	}

	var body strings.Builder
	total := 0
	for _, section := range sections {
		rows, err := runSection(repo, section)
		if err != nil {
			return "", fmt.Errorf("section %q: %w", section.title, err)
		}
		total += len(rows)
		body.WriteString(renderSection(section, rows))
	}

	if err := mailer.Send(ctx, sec, mailer.Message{
		To:       toAddress,
		Subject:  emailSubject,
		BodyHTML: body.String(),
	}); err != nil {
		return "", err
	}

	log.Printf("community-reports: emailed %d activists across %d sections to %s", total, len(sections), toAddress)
	return fmt.Sprintf("emailed %d activists", total), nil
}

func main() {
	lambda.Start(handler)
}
