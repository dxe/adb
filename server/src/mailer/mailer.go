// Package mailer sends email from the server using SMTP credentials supplied
// via environment-backed config. It is a thin adapter over pkg/mailer, which
// owns the transport, TLS, and header construction.
package mailer

import (
	"context"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/pkg/mailer"
)

// Message re-exports the shared message type so existing callers keep using
// mailer.Message unchanged.
type Message = mailer.Message

// SendContext delivers e using the server's env-configured SMTP credentials,
// honoring cancellation and deadlines from ctx.
func SendContext(ctx context.Context, e Message) error {
	cfg := mailer.Config{
		Host: config.SMTPHost,
		Port: config.SMTPPort,
		User: config.SMTPUser,
		Pass: config.SMTPPassword,
	}
	return mailer.Send(ctx, cfg, e)
}

// Send delivers e using the server's env-configured SMTP credentials. It is a
// thin wrapper over SendContext for callers without a context.
func Send(e Message) error {
	return SendContext(context.Background(), e)
}
