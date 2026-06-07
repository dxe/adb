// Package mailer sends email from the jobs lambdas over SES SMTP.
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/dxe/adb/jobs/internal/secrets"
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

// Send fetches SMTP credentials from SSM and delivers msg over a TLS SMTP
// connection.
func Send(ctx context.Context, sec *secrets.Client, msg Message) error {
	if strings.ContainsAny(msg.To, "\r\n") {
		return fmt.Errorf("invalid recipient address: contains CR or LF")
	}
	if strings.ContainsAny(msg.Subject, "\r\n") {
		return fmt.Errorf("invalid subject: contains CR or LF")
	}

	user, err := sec.Get(ctx, smtpUserParam)
	if err != nil {
		return err
	}
	pass, err := sec.Get(ctx, smtpPassParam)
	if err != nil {
		return err
	}

	raw := []byte(
		"To: " + msg.To + "\r\n" +
			"From: \"" + fromName + "\" <" + fromAddress + ">\r\n" +
			"Subject: " + msg.Subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			msg.BodyHTML + "\r\n",
	)

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{ServerName: smtpHost, MinVersion: tls.VersionTLS13},
	}
	conn, err := dialer.DialContext(ctx, "tcp", smtpHost+":"+smtpPort)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", user, pass, smtpHost)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(fromAddress); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return nil
}
