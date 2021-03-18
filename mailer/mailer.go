package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/dxe/adb/config"

	"github.com/pkg/errors"
)

type Message struct {
	FromName    string
	FromAddress string
	ToName      string
	ToEmail     string
	Subject     string
	BodyHTML    string
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

	host := config.SMTPHost
	port := config.SMTPPort
	user := config.SMTPUser
	pass := config.SMTPPassword

	auth := smtp.PlainAuth("", user, pass, host)

	message := fmt.Sprintf(`To: "%v" <%v>
From: "%v" <%v>
Subject: %v
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

<!DOCTYPE html>
<html lang="en">
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		%v
	</body>
</html>
`, e.ToName, e.ToEmail, e.FromName, e.FromAddress, e.Subject, e.BodyHTML)

	if err := smtp.SendMail(host+":"+port, auth, e.FromAddress, []string{e.ToEmail}, []byte(message)); err != nil {
		return errors.Wrap(err, "failed to send email")
	}
	return nil
}
