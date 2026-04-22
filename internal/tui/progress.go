package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressMsg struct{ Done, Total int }

type ProgressModel struct {
	bar   progress.Model
	done  int
	total int
	label string
}

func NewProgressModel(label string, total int) ProgressModel {
	return ProgressModel{bar: progress.New(progress.WithDefaultGradient()), label: label, total: total}
}

func (m ProgressModel) Init() tea.Cmd { return nil }

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		m.done, m.total = msg.Done, msg.Total
		pct := 0.0
		if m.total > 0 {
			pct = float64(m.done) / float64(m.total)
		}
		return m, m.bar.SetPercent(pct)
	case progress.FrameMsg:
		bar, cmd := m.bar.Update(msg)
		m.bar = bar.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m ProgressModel) View() string {
	return fmt.Sprintf("%s\n%s  %d/%d\n", m.label, m.bar.View(), m.done, m.total)
}
