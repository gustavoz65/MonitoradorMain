package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleOnline   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	StyleOffline  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	StyleWarn     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	StyleTitle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	StylePrompt   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	StyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	StyleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleSuccess  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleHeader   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Underline(true)
)

const (
	IconOnline  = "●"
	IconOffline = "✗"
	IconWarn    = "⚠"
)
