package core

import (
	"fmt"
	"myt/internal/render"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type MyStringBuilder struct {
	strings.Builder
}

func (b *MyStringBuilder) WriteStrings(text ...string) {
	for _, t := range text {
		b.WriteString(t)
	}
}

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
	var s MyStringBuilder

	highlightColor := lipgloss.Color("62")
	inactiveColor := lipgloss.Color("240")
	dangerColor := lipgloss.Color("9")
	safeColor := lipgloss.Color("10")

	if m.State == SelectConfig {
		s.WriteStrings(lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("=== Select Connection ==="), "\n\n")

		if m.ErrorMsg != "" {
			s.WriteStrings(lipgloss.NewStyle().Foreground(dangerColor).Render("Error: "+m.ErrorMsg), "\n\n")
		}

		for i, cfg := range m.Configs {
			cursor := "  "
			if m.ConfigCursor == i {
				cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("▶ ")
			}

			// SSHが設定されていれば表示にマークをつける
			sshTag := ""
			if cfg.SSHHost != "" {
				sshTag = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render("[SSH]")
			}
			s.WriteString(fmt.Sprintf("%s %s %s (%s:%d)\n", cursor, cfg.Name, sshTag, cfg.Host, cfg.Port))
		}
		s.WriteString("\n   [Enter] Connect | [Q/Esc] Quit\n")
		return s.String()
	}

	if m.ShowHelp {
		helpContent := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== Global Shortcuts =="),
			"  [Tab]          Switch Panel (Schema <-> SQL)",
			"  [Ctrl+L]       Clear Result Screen",
			"  [Ctrl+R]       Reload Schema Panel",
			"  [Ctrl+C]       Exit Application",
			"  [Ctrl+H]       Help",
			"",
			lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== Schema Panel =="),
			"  [← / →]        Switch Schema(DB -> Table -> Column)",
			"  [↑ / ↓]        Move Cursor",
			"  [Enter]        Select item / Insert to SQL",
			"",
			lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== SQL Panel =="),
			"  [Ctrl+N/Space] Auto Complete",
			"  [Ctrl+F]       Change Output Format",
			"  [Ctrl+E]       Run SQL",
			"  [Ctrl+U]       Clear SQL Text",
		)

		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(highlightColor).
			Width(72).
			Height(16).
			Padding(0, 3).
			Render(helpContent)

		s.WriteStrings(helpBox, "\n")
		s.WriteStrings(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] / [Esc] Close Help"), "\n")
		return s.String()
	}

	schemaBorderColor := inactiveColor
	sqlBorderColor := inactiveColor
	if !m.FocusSQL {
		schemaBorderColor = highlightColor
	} else {
		sqlBorderColor = highlightColor
	}

	var dStr, tStr, cStr MyStringBuilder

	const maxNameLen = 18

	// [1. Databases]
	startD := m.DBCursor - 2
	if startD < 0 {
		startD = 0
	}
	dTitle := "--- Databases ---"
	if !m.FocusSQL && m.SchemaPane == 0 {
		dTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + dTitle)
	} else {
		dTitle = "  " + dTitle
	}
	dStr.WriteStrings(dTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startD + i
		if idx < len(m.Databases) {
			cursor := "  "
			if m.DBCursor == idx {
				if !m.FocusSQL && m.SchemaPane == 0 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(m.Databases[idx], maxNameLen)
			dStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			dStr.WriteString("\n")
		}
	}

	// [2. Tables]
	startT := m.TableCursor - 2
	if startT < 0 {
		startT = 0
	}
	tTitle := "--- Tables ---"
	if !m.FocusSQL && m.SchemaPane == 1 {
		tTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + tTitle)
	} else {
		tTitle = "  " + tTitle
	}
	tStr.WriteStrings(tTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startT + i
		if idx < len(m.Tables) {
			cursor := "  "
			if m.TableCursor == idx {
				if !m.FocusSQL && m.SchemaPane == 1 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(m.Tables[idx], maxNameLen)
			tStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			tStr.WriteString("\n")
		}
	}

	// [3. Columns]
	startC := m.ColumnCursor - 2
	if startC < 0 {
		startC = 0
	}
	cTitle := "--- Columns ---"
	if !m.FocusSQL && m.SchemaPane == 2 {
		cTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + cTitle)
	} else {
		cTitle = "  " + cTitle
	}
	cStr.WriteStrings(cTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startC + i
		if idx < len(m.Columns) {
			cursor := "  "
			if m.ColumnCursor == idx {
				if !m.FocusSQL && m.SchemaPane == 2 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(m.Columns[idx], maxNameLen)
			cStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			cStr.WriteString("\n")
		}
	}

	leftPane := lipgloss.NewStyle().Width(23).Render(dStr.String())
	middlePane := lipgloss.NewStyle().Width(23).Border(lipgloss.NormalBorder(), false, false, false, true).PaddingLeft(1).Render(tStr.String())
	rightPane := lipgloss.NewStyle().Width(24).Border(lipgloss.NormalBorder(), false, false, false, true).PaddingLeft(1).Render(cStr.String())
	schemaContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, middlePane, rightPane)

	schemaBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(schemaBorderColor).
		Width(72).
		Render(schemaContent)

	s.WriteStrings(schemaBox, "\n")

	var formats []string
	for i, name := range render.FormatNames {
		style := lipgloss.NewStyle().Padding(0, 1)
		if int(m.OutputFormat) == i {
			style = style.Background(highlightColor).Foreground(lipgloss.Color("230")).Bold(true)
		} else {
			style = style.Foreground(inactiveColor)
		}
		formats = append(formats, style.Render(name))
	}
	formatBar := lipgloss.JoinHorizontal(lipgloss.Top, formats...)

	modeStr := lipgloss.NewStyle().Foreground(safeColor).Render("[Read Only]")
	if m.ReadWrite {
		modeStr = lipgloss.NewStyle().Foreground(dangerColor).Bold(true).Render("[Read-Write]")
	}

	statusBar := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode: "),
		modeStr,
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("  |  Format: "),
		formatBar,
	)

	sqlContent := lipgloss.JoinVertical(lipgloss.Left, m.SqlInput.View(), "", statusBar)

	sqlBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(sqlBorderColor).
		Width(72).
		Render(sqlContent)

	s.WriteStrings(sqlBox, "\n")
	s.WriteStrings(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] Help | [Tab] Switch Panel | [Ctrl+E] Run Query"), "\n")

	return s.String()
}
