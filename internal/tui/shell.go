package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/cli"
)

type execDoneMsg struct {
	exit   bool
	output string
	err    error
}

type ShellModel struct {
	app     AppFacade
	input   textinput.Model
	vp      viewport.Model
	history []string
	histIdx int
	ready   bool
	Exit    bool
}

func NewShellModel(app AppFacade) ShellModel {
	ti := textinput.New()
	ti.Placeholder = "digite um comando..."
	ti.Prompt = StylePrompt.Render("monimaster> ")
	ti.Focus()
	return ShellModel{app: app, input: ti, histIdx: -1}
}

func (m ShellModel) Init() tea.Cmd { return textinput.Blink }

func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd, vpCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.vp = viewport.New(msg.Width, msg.Height-3)
			m.vp.SetContent(StyleInfo.Render("Sessão pronta. Digite `help` para ver os comandos."))
			m.ready = true
		} else {
			m.vp.Width, m.vp.Height = msg.Width, msg.Height-3
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if len(m.history) > 0 {
				if m.histIdx < len(m.history)-1 {
					m.histIdx++
				}
				m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
			}
			return m, nil
		case "down":
			if m.histIdx > 0 {
				m.histIdx--
				m.input.SetValue(m.history[len(m.history)-1-m.histIdx])
			} else {
				m.histIdx = -1
				m.input.SetValue("")
			}
			return m, nil
		case "tab":
			m.input.SetValue(autoComplete(m.input.Value()))
			return m, nil
		case "enter":
			line := strings.TrimSpace(m.input.Value())
			if line == "" {
				return m, nil
			}
			m.history = append(m.history, line)
			m.histIdx = -1
			m.input.SetValue("")
			cmd, err := cli.Parse(line)
			if err != nil {
				if m.ready {
					m.vp.SetContent(StyleError.Render("Erro: " + err.Error()))
				}
				return m, nil
			}
			app := m.app
			return m, func() tea.Msg {
				exit, output, execErr := app.Execute(cmd)
				return execDoneMsg{exit, output, execErr}
			}
		}
	case execDoneMsg:
		if msg.exit {
			m.Exit = true
			return m, tea.Quit
		}
		content := msg.output
		if msg.err != nil {
			content = StyleError.Render("Erro: " + msg.err.Error())
		}
		if m.ready {
			m.vp.SetContent(content)
			m.vp.GotoBottom()
		}
	}
	m.input, tiCmd = m.input.Update(msg)
	if m.ready {
		m.vp, vpCmd = m.vp.Update(msg)
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m ShellModel) View() string {
	if !m.ready {
		return ""
	}
	return m.vp.View() + "\n" + m.input.View() + "\n"
}

var completions = []string{
	"auth login", "auth register", "auth logout",
	"cleanup interval",
	"config db", "config notify", "config show", "config smtp", "config wizard",
	"doctor run",
	"exit", "help",
	"history show",
	"logs clear", "logs export", "logs show",
	"monitor alert", "monitor dashboard", "monitor once", "monitor start", "monitor status", "monitor stop",
	"notify email",
	"portscan run",
	"profile",
	"report ports", "report uptime",
	"sites add", "sites import", "sites list", "sites remove", "sites update",
	"version",
}

func autoComplete(input string) string {
	if input == "" {
		return input
	}
	for _, c := range completions {
		if strings.HasPrefix(c, input) {
			return c
		}
	}
	return input
}

func RunShell(app AppFacade) error {
	_, err := tea.NewProgram(NewShellModel(app), tea.WithAltScreen()).Run()
	return err
}

var _ tea.Model = ShellModel{}
