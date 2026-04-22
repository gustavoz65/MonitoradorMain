package notify

import (
	"fmt"
	"strings"

	"github.com/gustavoz65/MoniMaster/internal/config"
	"gopkg.in/gomail.v2"
)

type SMTPProvider struct{}

func (p *SMTPProvider) Name() string { return "smtp" }

func (p *SMTPProvider) Send(cfg config.AppConfig, to, subject, body string) error {
	c := cfg.SMTP
	if strings.TrimSpace(c.Host) == "" || c.Port == 0 || strings.TrimSpace(c.User) == "" || strings.TrimSpace(c.Password) == "" {
		return fmt.Errorf("smtp incompleto; use config smtp set")
	}
	from := c.From
	if strings.TrimSpace(from) == "" {
		from = c.User
	}
	msg := gomail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)
	return gomail.NewDialer(c.Host, c.Port, c.User, c.Password).DialAndSend(msg)
}
