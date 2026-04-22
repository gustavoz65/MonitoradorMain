package doctor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gustavoz65/MoniMaster/internal/config"
	"github.com/gustavoz65/MoniMaster/internal/storage"
)

type Check struct {
	Name    string
	Healthy bool
	Details string
}

func Run(manager *config.Manager, cfg config.AppConfig, store storage.Store) []Check {
	var checks []Check
	if _, err := os.Stat(manager.HomeDir()); err == nil {
		checks = append(checks, Check{Name: "workspace", Healthy: true, Details: manager.HomeDir()})
	} else {
		checks = append(checks, Check{Name: "workspace", Healthy: false, Details: err.Error()})
	}

	if _, err := os.Stat(manager.SitesPath()); err == nil {
		checks = append(checks, Check{Name: "sites", Healthy: true, Details: manager.SitesPath()})
	} else {
		checks = append(checks, Check{Name: "sites", Healthy: false, Details: err.Error()})
	}

	if cfg.Storage.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := store.Ping(ctx)
		checks = append(checks, Check{Name: "database", Healthy: err == nil, Details: healthText(err, cfg.Storage.Driver)})
	} else {
		checks = append(checks, Check{Name: "database", Healthy: true, Details: "modo anonimo/local"})
	}

	smtpReady := strings.TrimSpace(cfg.SMTP.Host) != "" && cfg.SMTP.Port > 0 && strings.TrimSpace(cfg.SMTP.User) != "" && strings.TrimSpace(cfg.SMTP.Password) != ""
	checks = append(checks, Check{Name: "smtp", Healthy: smtpReady, Details: smtpDetails(smtpReady, cfg)})

	return checks
}

func healthText(err error, driver string) string {
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("conectado com %s", driver)
}

func smtpDetails(healthy bool, cfg config.AppConfig) string {
	if healthy {
		return fmt.Sprintf("%s:%d como %s", cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User)
	}
	return "smtp incompleto"
}
