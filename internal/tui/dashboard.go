package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/monitor"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type dashTickMsg time.Time
type dashResultMsg []shared.SiteResult

type DashboardModel struct {
	app       AppFacade
	results   []shared.SiteResult
	status    monitor.Status
	resultsCh <-chan []shared.SiteResult
	interval  time.Duration
}

func NewDashboardModel(app AppFacade, interval time.Duration) DashboardModel {
	return DashboardModel{app: app, resultsCh: app.SubscribeResults(), interval: interval}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(m.interval, func(t time.Time) tea.Msg { return dashTickMsg(t) }),
		m.waitResults(),
	)
}

func (m DashboardModel) waitResults() tea.Cmd {
	ch := m.resultsCh
	return func() tea.Msg { return dashResultMsg(<-ch) }
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case dashTickMsg:
		m.status = m.app.MonitorStatus()
		return m, tea.Tick(m.interval, func(t time.Time) tea.Msg { return dashTickMsg(t) })
	case dashResultMsg:
		m.results = []shared.SiteResult(msg)
		return m, m.waitResults()
	}
	return m, nil
}

func (m DashboardModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleTitle.Render("MoniMaster Dashboard") + "\n")

	online, offline, warn := 0, 0, 0
	for _, r := range m.results {
		switch {
		case !r.Online:
			offline++
		case r.LatencyWarn || r.CertWarn:
			warn++
		default:
			online++
		}
	}
	sb.WriteString(fmt.Sprintf("  %s %d online  %s %d offline  %s %d warn  ciclos: %d\n\n",
		StyleOnline.Render(IconOnline), online,
		StyleOffline.Render(IconOffline), offline,
		StyleWarn.Render(IconWarn), warn,
		m.status.CycleCount,
	))

	if len(m.results) == 0 {
		sb.WriteString(StyleInfo.Render("Aguardando resultados... (use monitor start primeiro)") + "\n")
	} else {
		t := Table{Headers: []string{"Site", "Status", "Cód", "Latência", "Cert", "Checado"}}
		for _, r := range m.results {
			status := StyleOnline.Render(IconOnline + " online")
			if !r.Online {
				status = StyleOffline.Render(IconOffline + " offline")
			}
			if r.LatencyWarn {
				status = StyleWarn.Render(IconWarn + " lento")
			}
			certInfo := "-"
			if !r.CertExpiry.IsZero() {
				days := int(time.Until(r.CertExpiry).Hours() / 24)
				if r.CertWarn {
					certInfo = StyleWarn.Render(fmt.Sprintf("%dd", days))
				} else {
					certInfo = StyleOnline.Render(fmt.Sprintf("%dd", days))
				}
			}
			t.Rows = append(t.Rows, TableRow{
				r.Site, status, fmt.Sprintf("%d", r.StatusCode),
				r.Latency.Round(time.Millisecond).String(),
				certInfo, r.CheckedAt.Format("15:04:05"),
			})
		}
		sb.WriteString(t.Render())
	}
	sb.WriteString("\n" + StyleInfo.Render("q para voltar ao shell") + "\n")
	return sb.String()
}

func RunDashboard(app AppFacade, interval time.Duration) error {
	_, err := tea.NewProgram(NewDashboardModel(app, interval), tea.WithAltScreen()).Run()
	return err
}

var _ tea.Model = DashboardModel{}
