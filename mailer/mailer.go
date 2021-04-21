package mailer

import (
	"net/smtp"
	"strings"

	"github.com/dxe/adb/config"

	"github.com/pkg/errors"
)

type Message struct {
	FromName       string
	FromAddress    string
	ToName         string
	ToAddress      string
	Subject        string
	BodyHTML       string
	ReplyToAddress string
	CC             []string
}

func smtpConfigSet() bool {
	if config.SMTPHost == "" || config.SMTPUser == "" || config.SMTPPassword == "" || config.SMTPPort == "" {
		return false
	}
	return true
}

func Send(e Message) error {
	if !smtpConfigSet() {
		return errors.New("failed to send email due to missing SMTP config")
	}

	requiredFieldsSet := e.FromName != "" && e.FromAddress != "" && e.ToName != "" && e.ToAddress != "" && e.Subject != "" && e.BodyHTML != ""
	if !requiredFieldsSet {
		return errors.New("failed to send email due to missing sender, recipient, subject, or body")
	}

	host := config.SMTPHost
	port := config.SMTPPort
	user := config.SMTPUser
	pass := config.SMTPPassword

	auth := smtp.PlainAuth("", user, pass, host)

	headers := `To: "` + e.ToName + `" <` + e.ToAddress + ">\n"
	headers += `From: "` + e.FromName + `" <` + e.FromAddress + ">\n"
	if len(e.CC) > 0 {
		headers += strings.Join(e.CC, ", ") + "\n"
	}
	if e.ReplyToAddress != "" {
		headers += `Reply-To: ` + e.ReplyToAddress + "\n"
	}
	headers += `Subject: ` + e.Subject + "\n"
	headers += "MIME-version: 1.0;\n"
	headers += `Content-Type: text/html; charset="UTF-8";` + "\n"

	body := `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		` + e.BodyHTML + `
	</body>
</html>
`

	message := headers + "\n" + body

	if err := smtp.SendMail(host+":"+port, auth, e.FromAddress, []string{e.ToAddress}, []byte(message)); err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	return nil
}
