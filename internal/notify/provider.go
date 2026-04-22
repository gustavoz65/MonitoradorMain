package notify

import "github.com/gustavoz65/MoniMaster/internal/config"

type Provider interface {
	Name() string
	Send(cfg config.AppConfig, to, subject, body string) error
}
