package render

import (
	"fmt"
	"myt/internal/types"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var highlightColor = lipgloss.Color("62")
var inactiveColor = lipgloss.Color("240")
var dangerColor = lipgloss.Color("9")
var safeColor = lipgloss.Color("10")

func Help() string {
	//var s MyStringBuilder
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

	//s.WriteStrings(helpBox, "\n", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" [Ctrl+H] / [Esc] Close Help"), "\n")

	return helpBox
}

func Config(configs []types.Config, configCursor int, errorMsg string) string {
	var s MyStringBuilder
	s.WriteStrings(lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("=== Select Connection ==="), "\n\n")

	if errorMsg != "" {
		s.WriteStrings(lipgloss.NewStyle().Foreground(dangerColor).Render("Error: "+errorMsg), "\n\n")
	}

	for i, cfg := range configs {
		cursor := "  "
		if configCursor == i {
			cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("▶ ")
		}

		sshTag := ""
		if cfg.SSHHost != "" {
			sshTag = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render("[SSH]")
		}
		s.WriteString(fmt.Sprintf("%s %s %s (%s:%d)\n", cursor, cfg.Name, sshTag, cfg.Host, cfg.Port))
	}
	s.WriteString("\n   [Enter] Connect | [Q/Esc] Quit\n")
	return s.String()
}

func SchemaPanel(isFocused bool, schemaPane int, databases []string, tables []string, columns []string, dbCursor int, tblCursor int, colCursor int) string {
	schemaBorderColor := inactiveColor
	if isFocused {
		schemaBorderColor = highlightColor
	}

	var dStr, tStr, cStr MyStringBuilder

	const maxNameLen = 18

	// [1. Databases]
	startD := dbCursor - 2
	if startD < 0 {
		startD = 0
	}
	dTitle := "--- Databases ---"
	if isFocused && schemaPane == 0 {
		dTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + dTitle)
	} else {
		dTitle = "  " + dTitle
	}
	dStr.WriteStrings(dTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startD + i
		if idx < len(databases) {
			cursor := "  "
			if dbCursor == idx {
				if isFocused && schemaPane == 0 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(databases[idx], maxNameLen)
			dStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			dStr.WriteString("\n")
		}
	}

	// [2. Tables]
	startT := tblCursor - 2
	if startT < 0 {
		startT = 0
	}
	tTitle := "--- Tables ---"
	if isFocused && schemaPane == 1 {
		tTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + tTitle)
	} else {
		tTitle = "  " + tTitle
	}
	tStr.WriteStrings(tTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startT + i
		if idx < len(tables) {
			cursor := "  "
			if tblCursor == idx {
				if isFocused && schemaPane == 1 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(tables[idx], maxNameLen)
			tStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			tStr.WriteString("\n")
		}
	}

	// [3. Columns]
	startC := colCursor - 2
	if startC < 0 {
		startC = 0
	}
	cTitle := "--- Columns ---"
	if isFocused && schemaPane == 2 {
		cTitle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("> " + cTitle)
	} else {
		cTitle = "  " + cTitle
	}
	cStr.WriteStrings(cTitle, "\n")
	for i := 0; i < 5; i++ {
		idx := startC + i
		if idx < len(columns) {
			cursor := "  "
			if colCursor == idx {
				if isFocused && schemaPane == 2 {
					cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
				} else {
					cursor = "> "
				}
			}
			safeName := truncateText(columns[idx], maxNameLen)
			cStr.WriteString(fmt.Sprintf("%s%s\n", cursor, safeName))
		} else {
			cStr.WriteString("\n")
		}
	}

	leftPane := lipgloss.NewStyle().Width(23).Render(dStr.String())
	middlePane := lipgloss.NewStyle().Width(23).Border(lipgloss.NormalBorder(), false, false, false, true).PaddingLeft(1).Render(tStr.String())
	rightPane := lipgloss.NewStyle().Width(24).Border(lipgloss.NormalBorder(), false, false, false, true).PaddingLeft(1).Render(cStr.String())
	schemaContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, middlePane, rightPane)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(schemaBorderColor).
		Width(72).
		Render(schemaContent)
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

func QueryPanel(isFocused bool, format OutputFormat, text string, rw bool) string {
	sqlBorderColor := inactiveColor
	if isFocused {
		sqlBorderColor = highlightColor
	}

	var formats []string
	for i, name := range FormatNames {
		style := lipgloss.NewStyle().Padding(0, 1)
		if int(format) == i {
			style = style.Background(highlightColor).Foreground(lipgloss.Color("230")).Bold(true)
		} else {
			style = style.Foreground(inactiveColor)
		}
		formats = append(formats, style.Render(name))
	}
	formatBar := lipgloss.JoinHorizontal(lipgloss.Top, formats...)

	modeStr := lipgloss.NewStyle().Foreground(safeColor).Render("[Read Only]")
	if rw {
		modeStr = lipgloss.NewStyle().Foreground(dangerColor).Bold(true).Render("[Read-Write]")
	}

	statusBar := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode: "),
		modeStr,
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("  |  Format: "),
		formatBar,
	)

	sqlContent := lipgloss.JoinVertical(lipgloss.Left, text, "", statusBar)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(sqlBorderColor).
		Width(72).
		Render(sqlContent)
}
