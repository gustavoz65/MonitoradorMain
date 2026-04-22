package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type TableRow []string

type Table struct {
	Headers []string
	Rows    []TableRow
}

func (t Table) Render() string {
	if len(t.Headers) == 0 {
		return ""
	}
	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	headerCells := make([]string, len(t.Headers))
	for i, h := range t.Headers {
		headerCells[i] = StyleHeader.Render(pad(h, widths[i]))
	}

	sep := make([]string, len(t.Headers))
	for i, w := range widths {
		sep[i] = strings.Repeat("─", w)
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(headerCells, "  ") + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Join(sep, "  ")) + "\n")
	for _, row := range t.Rows {
		cells := make([]string, len(t.Headers))
		for i := range t.Headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cells[i] = pad(cell, widths[i])
		}
		sb.WriteString(strings.Join(cells, "  ") + "\n")
	}
	return sb.String()
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
