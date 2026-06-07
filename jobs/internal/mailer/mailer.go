// Package mailer sends email from the jobs lambdas over SES SMTP. It resolves
// credentials from SSM and delegates transport to the shared pkg/mailer engine.
package mailer

import (
	"context"

	"github.com/dxe/adb/jobs/internal/secrets"
	"github.com/dxe/adb/pkg/mailer"
)

const (
	smtpHost = "email-smtp.us-west-2.amazonaws.com"
	smtpPort = "465"

	fromAddress = "tech-noreply@directactioneverywhere.com"
	fromName    = "DxE Reporting"

	smtpUserParam = "smtp_user"
	smtpPassParam = "smtp_pass"
)

// Message is a single HTML email. The sender (from address/name) is fixed by
// this package; only the recipient and content vary per call.
type Message struct {
	To       string
	Subject  string
	BodyHTML string
}

// Send fetches SMTP credentials from SSM and delivers msg via the shared mailer.
func Send(ctx context.Context, sec *secrets.Client, msg Message) error {
	user, err := sec.Get(ctx, smtpUserParam)
	if err != nil {
		return err
	}
	pass, err := sec.Get(ctx, smtpPassParam)
	if err != nil {
		return err
	}

	return mailer.Send(ctx, mailer.Config{
		Host: smtpHost,
		Port: smtpPort,
		User: user,
		Pass: pass,
	}, mailer.Message{
		FromName:    fromName,
		FromAddress: fromAddress,
		ToAddress:   msg.To,
		Subject:     msg.Subject,
		BodyHTML:    msg.BodyHTML,
	})
}
