package components

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends an email using SMTP settings from environment variables
func SendEmail(to, subject, body string) error {
	host := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")
	email := os.Getenv("EMAIL")
	password := os.Getenv("EMAIL_PASSWORD")
	if host == "" || port == "" || email == "" || password == "" {
		return fmt.Errorf("email configuration is missing")
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", email, password, host)

	headers := make(map[string]string)
	headers["From"] = email
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	return smtp.SendMail(addr, auth, email, []string{to}, []byte(msg))
}
