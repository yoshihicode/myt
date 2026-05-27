package core

import (
	"myt/internal/render"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func truncateText(name string, maxWidth int) string {
	if lipgloss.Width(name) <= maxWidth {
		return name
	}
	var w int
	var sb strings.Builder
	for _, r := range name {
		rw := lipgloss.Width(string(r))
		if w+rw > maxWidth-2 { // ".."
			break
		}
		sb.WriteRune(r)
		w += rw
	}
	return sb.String() + ".."
}

func (m *Model) View() string {
	var s render.MyStringBuilder

	if m.State == SelectConfig {
		return render.Config(m.Configs, m.ConfigCursor, m.ErrorMsg)
	}

	if m.ShowHelp {
		return render.Help()
	}

	schema := render.SchemaPanel(!m.FocusSQL, m.SchemaPane, m.Databases, m.Tables, m.Columns, m.DBCursor, m.TableCursor, m.ColumnCursor)
	s.WriteStrings(schema, "\n")

	query := render.QueryPanel(m.FocusSQL, m.OutputFormat, m.SqlInput.View(), m.Configs[m.ConfigCursor].ReadWrite)

	s.WriteStrings(query, "\n")
	s.WriteStrings(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] Help | [Tab] Switch Panel | [Ctrl+E] Run Query"), "\n")

	return s.String()
}
