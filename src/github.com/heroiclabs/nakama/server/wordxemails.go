package server

import (
	"crypto/tls"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

const DEFAULT_EMAIL_FROM = "darrell@murdockimagineeringstudio.com"

func (wx *WordX) SendPasswordEmail(email string, code string) {
	const PASSWORD_EMAIL_BODY = `We heard that you lost your WordX password. Sorry about that!

You can use this code to restore password in game:

*CODE*
`

	SendEmail(DEFAULT_EMAIL_FROM, email, "[WordX] Please reset your password", strings.Replace(PASSWORD_EMAIL_BODY, "*CODE*", code, -1), wx.Logger)
}

func SendEmail(fromStr string, toStr string, subject string, body string, logger *zap.Logger) {
	from := mail.Address{Name: "", Address: fromStr}
	to := mail.Address{Name: "", Address: toStr}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	servername := "email-smtp.us-east-1.amazonaws.com:465"

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", "AKIAIDMWDSD34JJO6LQA", "BMakwaDBotCMbRgUKkS5JE+aobUkQ3hMsOekXs+Ghlog", host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	if err = c.Rcpt(to.Address); err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	// Data
	w, err := c.Data()
	if err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	err = w.Close()
	if err != nil {
		logger.Error("failed send email", zap.Error(err))
		return
	}

	c.Quit()
}
