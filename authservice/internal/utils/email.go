package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	gomail "gopkg.in/gomail.v2"
)

type EmailClient struct {
	dialer *gomail.Dialer
	from   string
}

var emailClient *EmailClient

func InitEmail() {
	if os.Getenv("EMAIL_ENABLED") != "true" {
		log.Println("[email] disabled – letters are logged only")
		return
	}
	host, portStr := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")
	user, pass := os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	var port int
	_, _ = fmt.Sscan(portStr, &port)
	d := gomail.NewDialer(host, port, user, pass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	emailClient = &EmailClient{dialer: d, from: from}
}

func Send(to, subj, body string) error {
	if os.Getenv("EMAIL_ENABLED") != "true" || emailClient == nil {
		log.Printf("[email mock] to=%s subj=%q body=%s", to, subj, body)
		return nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", emailClient.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", body)
	return emailClient.dialer.DialAndSend(m)
}
