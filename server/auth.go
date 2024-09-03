package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendAuthEmail(recipient string, code string) error {
	smtpHost := os.Getenv("EMAIL_HOST")
	smtpPortStr := os.Getenv("EMAIL_PORT")
	smtpUser := os.Getenv("EMAIL_USER")
	smtpPass := os.Getenv("EMAIL_PASS")
	smtpPort, _ := strconv.Atoi(smtpPortStr)
	subject := os.Getenv("EMAIL_SUBJECT")
	body := os.Getenv("EMAIL_BODY")

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body+code)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
