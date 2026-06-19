package core

import (
	"myt/internal/constant"
	"myt/internal/render"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	var s render.MyStringBuilder

	if m.State == constant.AppStateConfig {
		return render.Config(m.Configs, m.ConfigCursor, m.ErrorMsg)
	}

	if m.State == constant.AppStatePassword {
		return render.PasswordPrompt(m.PromptTarget, m.PasswordInput.View(), m.ErrorMsg, m.Configs[m.ConfigCursor].Name)
	}

	if m.State == constant.AppStateDBSelect {
		return render.DbSelect(m.Databases, m.DBCursor, m.ErrorMsg)
	}

	if m.ShowHelp {
		return render.Help()
	}

	currentDB := ""
	if m.DBCursor < len(m.Databases) {
		currentDB = m.Databases[m.DBCursor]
	}
	s.WriteStrings(render.HeaderBar(m.Configs[m.ConfigCursor].Name, currentDB, m.Configs[m.ConfigCursor].ReadWrite), "\n")
	schema := render.SchemaPanels(m.FocusPanel, m.Databases, m.Tables, m.Columns, m.DBCursor, m.TableCursor, m.ColumnCursor)
	s.WriteStrings(schema, "\n")

	query := render.QueryPanel(m.FocusPanel == constant.FocusEditor, m.OutputFormat, m.SqlInput.View(), m.Configs[m.ConfigCursor].ReadWrite, m.TxPending, m.Configs[m.ConfigCursor].Name)

	s.WriteStrings(query, "\n")
	s.WriteStrings(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] Help | [Tab] Switch Panel | [Ctrl+E] Run Query | [ESC] Back"), "\n")

	return s.String()
}
