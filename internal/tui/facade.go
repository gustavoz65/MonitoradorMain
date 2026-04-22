package tui

import (
	"github.com/gustavoz65/MoniMaster/internal/cli"
	"github.com/gustavoz65/MoniMaster/internal/monitor"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type EntryResult struct {
	Proceed  bool
	Identity *shared.Identity
	Mode     string
}

// AppFacade desacopla os modelos TUI da implementacao concreta de App.
type AppFacade interface {
	Login(username, password string) (*shared.Identity, error)
	Register(username, email, password string) (*shared.Identity, error)
	ConfigDB(driver, dsn string) error
	SetupWizard(useDB bool, driver, dsn, defaultEmail string) error
	Execute(cmd cli.Command) (exit bool, output string, err error)
	MonitorStatus() monitor.Status
	SubscribeResults() <-chan []shared.SiteResult
}
