package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

// TODO: make these configurable via environment / SSM once additional jobs need different SMTP targets.
const (
	smtpHost     = "email-smtp.us-west-2.amazonaws.com"
	smtpPort     = "465"
	toAddress    = "ataylor@directactioneverywhere.com"
	fromAddress  = "tech-noreply@directactioneverywhere.com"
	fromName     = "DxE Reporting"
	emailSubject = "Hello World"
	emailBody    = "Hello World!"
)

func handler(ctx context.Context) (string, error) {
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	if user == "" || pass == "" {
		return "", fmt.Errorf("SMTP_USER and SMTP_PASS must be set")
	}

	msg := []byte(
		"To: " + toAddress + "\r\n" +
			"From: \"" + fromName + "\" <" + fromAddress + ">\r\n" +
			"Subject: " + emailSubject + "\r\n" +
			"\r\n" +
			emailBody + "\r\n",
	)

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 10 * time.Second},
		Config:    &tls.Config{ServerName: smtpHost, MinVersion: tls.VersionTLS13},
	}
	conn, err := dialer.DialContext(ctx, "tcp", smtpHost+":"+smtpPort)
	if err != nil {
		return "", fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return "", fmt.Errorf("smtp client: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(smtp.PlainAuth("", user, pass, smtpHost)); err != nil {
		return "", fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(fromAddress); err != nil {
		return "", fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(toAddress); err != nil {
		return "", fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return "", fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return "", fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("smtp close: %w", err)
	}

	return "sent", nil
}

func main() {
	lambda.Start(handler)
}
