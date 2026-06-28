// Package mailer is the shared SMTP engine used by both the long-running server
// and the jobs lambdas. Callers resolve their own SMTP credentials (env vars on
// the server, SSM in lambdas) and hand them in via Config; this package owns the
// transport, TLS, header construction, and injection-safety concerns.
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds the SMTP endpoint and credentials. The transport is derived from
// Port: 465 uses implicit TLS (the connection is TLS from the first byte), any
// other port uses STARTTLS. Both modes require TLS 1.3.
type Config struct {
	Host string
	Port string
	User string
	Pass string
}

// Message is a single HTML email. From* and To* are required; the remaining
// fields are optional and omitted from the headers when empty.
type Message struct {
	FromName         string
	FromAddress      string
	ToName           string
	ToAddress        string
	Subject          string
	BodyHTML         string
	ReplyToAddress   string
	ReplyToAddresses []string
	CC               []string
	BCC              []string
}

// dialTimeout bounds the TCP/TLS handshake; sendDeadline bounds the whole SMTP
// conversation. A shorter context deadline always wins over sendDeadline.
const (
	dialTimeout  = 10 * time.Second
	sendDeadline = 30 * time.Second
)

// Send delivers msg using cfg. It validates inputs, opens a TLS-protected SMTP
// connection, authenticates, and writes the message. The context bounds the
// dial and, if its deadline is sooner than sendDeadline, the conversation.
func Send(ctx context.Context, cfg Config, msg Message) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Pass == "" {
		return fmt.Errorf("incomplete SMTP config")
	}
	if msg.FromAddress == "" || msg.FromName == "" || msg.ToAddress == "" || msg.Subject == "" || msg.BodyHTML == "" {
		return fmt.Errorf("missing sender, recipient, subject, or body")
	}

	headers, err := buildHeaders(msg)
	if err != nil {
		return err
	}
	raw := []byte(headers + "\r\n" + wrapBody(msg.BodyHTML))

	conn, err := dial(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	deadline := time.Now().Add(sendDeadline)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Quit()

	// On STARTTLS ports the connection is still plaintext here; upgrade it
	// before authenticating so credentials are never sent in the clear.
	if cfg.Port != "465" {
		if err := client.StartTLS(tlsConfig(cfg.Host)); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if err := client.Auth(smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(msg.FromAddress); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, rcpt := range recipients(msg) {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %q: %w", rcpt, err)
		}
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

// dial opens the underlying connection: a TLS connection on 465 (implicit TLS),
// or a plain TCP connection elsewhere that Send later upgrades via STARTTLS.
func dial(ctx context.Context, cfg Config) (net.Conn, error) {
	addr := cfg.Host + ":" + cfg.Port
	if cfg.Port == "465" {
		dialer := &tls.Dialer{
			NetDialer: &net.Dialer{Timeout: dialTimeout},
			Config:    tlsConfig(cfg.Host),
		}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("tls dial: %w", err)
		}
		return conn, nil
	}
	conn, err := (&net.Dialer{Timeout: dialTimeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	return conn, nil
}

func tlsConfig(host string) *tls.Config {
	return &tls.Config{ServerName: host, MinVersion: tls.VersionTLS13}
}

// recipients is the SMTP envelope recipient list: the primary recipient plus all
// CC and BCC addresses. BCC appears here but is intentionally absent from the
// headers built by buildHeaders.
func recipients(msg Message) []string {
	out := []string{msg.ToAddress}
	out = append(out, msg.CC...)
	out = append(out, msg.BCC...)
	return out
}

// buildHeaders renders the message headers, rejecting any header field that
// contains CR or LF to prevent header injection.
func buildHeaders(msg Message) (string, error) {
	fields := []string{msg.FromName, msg.FromAddress, msg.ToName, msg.ToAddress, msg.Subject, msg.ReplyToAddress}
	fields = append(fields, msg.ReplyToAddresses...)
	fields = append(fields, msg.CC...)
	fields = append(fields, msg.BCC...)
	for _, f := range fields {
		if strings.ContainsAny(f, "\r\n") {
			return "", fmt.Errorf("invalid header value: contains CR or LF")
		}
	}

	to := "To: " + msg.ToAddress
	if msg.ToName != "" {
		to = `To: "` + msg.ToName + `" <` + msg.ToAddress + ">"
	}
	h := []string{
		to,
		`From: "` + msg.FromName + `" <` + msg.FromAddress + ">",
	}
	if len(msg.CC) > 0 {
		h = append(h, "CC: "+strings.Join(msg.CC, ", "))
	}
	if replyTo := replyToList(msg); len(replyTo) > 0 {
		h = append(h, "Reply-To: "+strings.Join(replyTo, ", "))
	}
	h = append(h,
		"Subject: "+msg.Subject,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
	)
	return strings.Join(h, "\r\n") + "\r\n", nil
}

func replyToList(msg Message) []string {
	out := append([]string(nil), msg.ReplyToAddresses...)
	if msg.ReplyToAddress != "" {
		out = append(out, msg.ReplyToAddress)
	}
	return out
}

func wrapBody(bodyHTML string) string {
	return `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		` + bodyHTML + `
	</body>
</html>
`
}
