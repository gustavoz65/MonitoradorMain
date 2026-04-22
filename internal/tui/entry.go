package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gustavoz65/MoniMaster/internal/shared"
)

type entryState int

const (
	stateMenu entryState = iota
	stateLogin
	stateRegister
	stateConfigDB
)

var menuItems = []string{"Login", "Cadastro", "Continuar anônimo", "Configurar banco", "Assistente inicial", "Sair"}

type loginDoneMsg struct {
	identity *shared.Identity
	err      error
}

type registerDoneMsg struct {
	identity *shared.Identity
	err      error
}

type configDBDoneMsg struct{ err error }

type EntryModel struct {
	app      AppFacade
	state    entryState
	cursor   int
	inputs   []textinput.Model
	inputIdx int
	err      error
	Result   EntryResult
	Done     bool
}

func NewEntryModel(app AppFacade) EntryModel { return EntryModel{app: app} }

func (m EntryModel) Init() tea.Cmd { return nil }

func (m EntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == stateMenu {
			return m.updateMenu(msg)
		}
		return m.updateForm(msg)
	case loginDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateMenu
			return m, nil
		}
		m.Result = EntryResult{Proceed: true, Identity: msg.identity, Mode: shared.SessionModeAuthenticated}
		m.Done = true
		return m, tea.Quit
	case registerDoneMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateMenu
			return m, nil
		}
		m.Result = EntryResult{Proceed: true, Identity: msg.identity, Mode: shared.SessionModeAuthenticated}
		m.Done = true
		return m, tea.Quit
	case configDBDoneMsg:
		m.err = msg.err
		m.state = stateMenu
		return m, nil
	}
	return m, nil
}

func (m EntryModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
	case "enter":
		m.err = nil
		switch m.cursor {
		case 0:
			m.state = stateLogin
			m.inputs = makeInputs([]inputSpec{{"Usuário", textinput.EchoNormal}, {"Senha", textinput.EchoPassword}})
			m.inputIdx = 0
		case 1:
			m.state = stateRegister
			m.inputs = makeInputs([]inputSpec{{"Usuário", textinput.EchoNormal}, {"Email", textinput.EchoNormal}, {"Senha", textinput.EchoPassword}})
			m.inputIdx = 0
		case 2:
			m.Result = EntryResult{Proceed: true, Mode: shared.SessionModeAnonymous}
			m.Done = true
			return m, tea.Quit
		case 3, 4:
			m.state = stateConfigDB
			m.inputs = makeInputs([]inputSpec{{"Driver (postgres/mysql/sqlite/oracle)", textinput.EchoNormal}, {"DSN", textinput.EchoNormal}})
			m.inputIdx = 0
		case 5:
			m.Result = EntryResult{Proceed: false}
			m.Done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m EntryModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		return m, nil
	case "tab", "down":
		if m.inputIdx < len(m.inputs)-1 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx++
			m.inputs[m.inputIdx].Focus()
		}
	case "shift+tab", "up":
		if m.inputIdx > 0 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx--
			m.inputs[m.inputIdx].Focus()
		}
	case "enter":
		if m.inputIdx < len(m.inputs)-1 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx++
			m.inputs[m.inputIdx].Focus()
		} else {
			return m, m.submitForm()
		}
	default:
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
	}
	return m, cmd
}

func (m EntryModel) submitForm() tea.Cmd {
	app := m.app
	switch m.state {
	case stateLogin:
		u, p := m.inputs[0].Value(), m.inputs[1].Value()
		return func() tea.Msg {
			id, err := app.Login(u, p)
			return loginDoneMsg{id, err}
		}
	case stateRegister:
		u, e, p := m.inputs[0].Value(), m.inputs[1].Value(), m.inputs[2].Value()
		return func() tea.Msg {
			id, err := app.Register(u, e, p)
			return registerDoneMsg{id, err}
		}
	case stateConfigDB:
		d, dsn := m.inputs[0].Value(), m.inputs[1].Value()
		return func() tea.Msg { return configDBDoneMsg{app.ConfigDB(d, dsn)} }
	}
	return nil
}

func (m EntryModel) View() string {
	var sb strings.Builder
	sb.WriteString(StyleTitle.Render("MoniMaster CLI") + "\n\n")
	if m.err != nil {
		sb.WriteString(StyleError.Render("Erro: "+m.err.Error()) + "\n\n")
	}
	if m.state == stateMenu {
		for i, item := range menuItems {
			if i == m.cursor {
				sb.WriteString("▶ " + StyleSelected.Render(item) + "\n")
			} else {
				sb.WriteString("  " + item + "\n")
			}
		}
		sb.WriteString("\n" + StyleInfo.Render("↑/↓ navegar  Enter confirmar") + "\n")
		return sb.String()
	}
	titles := map[entryState]string{stateLogin: "Login", stateRegister: "Cadastro", stateConfigDB: "Configurar banco"}
	sb.WriteString(StyleHeader.Render(titles[m.state]) + "\n\n")
	for i, inp := range m.inputs {
		prefix := "  "
		if i == m.inputIdx {
			prefix = StylePrompt.Render("▶ ")
		}
		sb.WriteString(prefix + inp.View() + "\n")
	}
	sb.WriteString("\n" + StyleInfo.Render("Tab navegar  Enter confirmar  Esc voltar") + "\n")
	return sb.String()
}

type inputSpec struct {
	placeholder string
	echo        textinput.EchoMode
}

func makeInputs(specs []inputSpec) []textinput.Model {
	inputs := make([]textinput.Model, len(specs))
	for i, spec := range specs {
		ti := textinput.New()
		ti.Placeholder = spec.placeholder
		ti.EchoMode = spec.echo
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return inputs
}

func RunEntry(app AppFacade) (EntryResult, error) {
	p := tea.NewProgram(NewEntryModel(app), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return EntryResult{}, err
	}
	return final.(EntryModel).Result, nil
}

var _ tea.Model = EntryModel{}
