package notification

import (
	"crypto/tls"
	"errors"
	"strings"

	"story-monitor/log"
)

type Client struct {
	dialer *gomail.Dialer
}

func NewClient(host string, port int, username string, password string) *Client {

	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	logger.Info("mail client established")
	return &Client{
		dialer: d,
	}
}

func (c *Client) SendMail(from, to, subject, body string) error {

	m := gomail.NewMessage()

	m.SetHeader("From", from)

	toArray := strings.Split(to, ",")
	if len(toArray) == 0 {
		return errors.New("failed get notice mail")
	}
	m.SetHeader("To", toArray[0])
	if len(toArray) > 0 {
		m.SetHeader("Cc", toArray[1:]...)
	}

	m.SetHeader("Subject", subject)

	m.SetBody("text/plain", body)

	if err := c.dialer.DialAndSend(m); err != nil {
		logger.Error("send error", "error", err)
		return err
	}

	return nil

}

var logger = log.MailLogger.WithField("module", "notification")
