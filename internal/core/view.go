package core

import (
	"myt/internal/render"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var s render.MyStringBuilder

	if m.State == SelectConfig {
		return render.Config(m.Configs, m.ConfigCursor, m.ErrorMsg)
	}

	if m.State == PasswordPrompt {
		return render.PasswordPrompt(m.PromptTarget, m.PasswordInput.View(), m.ErrorMsg)
	}

	if m.ShowHelp {
		return render.Help()
	}

	schema := render.SchemaPanel(!m.FocusSQL, m.SchemaPane, m.Databases, m.Tables, m.Columns, m.DBCursor, m.TableCursor, m.ColumnCursor)
	s.WriteStrings(schema, "\n")

	query := render.QueryPanel(m.FocusSQL, m.OutputFormat, m.SqlInput.View(), m.Configs[m.ConfigCursor].ReadWrite, m.TxPending)

	s.WriteStrings(query, "\n")
	s.WriteStrings(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] Help | [Tab] Switch Panel | [Ctrl+E] Run Query"), "\n")

	return s.String()
}
