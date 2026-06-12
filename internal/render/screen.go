package render

import (
	"myt/internal/config"
	"myt/internal/constant"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var highlightColor = lipgloss.Color("62")
var inactiveColor = lipgloss.Color("240")
var dangerColor = lipgloss.Color("9")
var safeColor = lipgloss.Color("10")

func Help() string {
	helpContent := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== Global Shortcuts =="),
		"  [Tab]          Focus Next Panel",
		"  [Shift + Tab]  Focus Previous Panel",
		"  [Ctrl+L]       Clear Result",
		"  [Ctrl+R]       Reload Schema Panel",
		"  [Ctrl+H]       Help",
		"  [Ctrl+C]       Exit",
		"",
		lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== Schema Panels =="),
		"  [↑ / ↓]         Move Cursor",
		"  [Enter]         Select item / Insert to SQL",
		"",
		lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render("== SQL Panel =="),
		"  [Ctrl+N/Space]  Auto Complete",
		"  [Ctrl+F]        Change Output Format",
		"  [Ctrl+E]        Run SQL",
		"  [Ctrl+U]        Clear SQL",
	)

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(highlightColor).
		Width(72).
		Height(16).
		Padding(0, 3).
		Render(helpContent)

	return helpBox
}

func Config(configs []config.Config, configCursor int, errorMsg string) string {
	var s MyStringBuilder
	s.WriteStrings("\n", lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(" Select Connection"), "\n\n")

	if errorMsg != "" {
		s.WriteStrings(lipgloss.NewStyle().Foreground(dangerColor).Render("Error: "+errorMsg), "\n\n")
	}

	maxEndpoint := 8

	type rowData struct {
		name     string
		endpoint string
		mode     string
		network  string
	}

	var rows []rowData
	for _, cfg := range configs {
		endpoint := cfg.Host + ":" + strconv.Itoa(cfg.Port)

		mode := "[Read Only]"
		if cfg.ReadWrite {
			mode = "[Read-Write]"
		}

		network := ""
		if cfg.SSHHost != "" {
			network = "[SSH]"
		}

		if lipgloss.Width(endpoint) > maxEndpoint {
			maxEndpoint = lipgloss.Width(endpoint)
		}

		rows = append(rows, rowData{truncateText(cfg.Name, 28), endpoint, mode, network})
	}

	nameStyle := lipgloss.NewStyle().Width(30)
	endpointStyle := lipgloss.NewStyle().Width(maxEndpoint + 2)
	modeStyle := lipgloss.NewStyle().Width(14)

	headerStyle := lipgloss.NewStyle().Foreground(inactiveColor).Bold(true)
	s.WriteStrings(
		"   ",
		nameStyle.Render(headerStyle.Render("NAME")),
		endpointStyle.Render(headerStyle.Render("ENDPOINT")),
		modeStyle.Render(headerStyle.Render("MODE")),
		headerStyle.Render("NETWORK"),
		"\n",
	)

	total := 33 + (maxEndpoint + 2) + 14 + 7
	s.WriteStrings(" "+lipgloss.NewStyle().Foreground(inactiveColor).Render(strings.Repeat("─", total)), "\n")

	for i, r := range rows {
		cursor := "  "
		var rowStyle lipgloss.Style

		if configCursor == i {
			cursor = lipgloss.NewStyle().Foreground(highlightColor).Render("> ")
			rowStyle = lipgloss.NewStyle().Foreground(highlightColor).Bold(true)
		} else {
			rowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		var renderedMode string
		if configCursor == i {
			renderedMode = rowStyle.Render(r.mode)
		} else {
			modeColor := safeColor
			if r.mode == "[Read-Write]" {
				modeColor = dangerColor
			}
			renderedMode = lipgloss.NewStyle().Foreground(modeColor).Render(r.mode)
		}

		ssh := ""
		if r.network != "" {
			if configCursor == i {
				ssh = rowStyle.Render(r.network)
			} else {
				ssh = lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(r.network)
			}
		}

		s.WriteStrings(
			cursor,
			nameStyle.Render(rowStyle.Render(r.name)),
			endpointStyle.Render(rowStyle.Render(r.endpoint)),
			modeStyle.Render(renderedMode),
			ssh,
			"\n",
		)
	}

	s.WriteString("\n   [Enter] Connect | [Q/Esc] Quit\n")
	return s.String()
}

func PasswordPrompt(target string, inputView string, errorMsg string, conName string) string {

	var s MyStringBuilder

	if errorMsg != "" {
		s.WriteStrings(lipgloss.NewStyle().Foreground(dangerColor).Render("Error: "+errorMsg), "\n\n")
	}

	label := "🔐 MySQL Password Required"
	if target == "SSH" {
		label = "🔑 SSH Password Required"
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true).Render(conName),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render(label),
		"",
		inputView,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" [Enter] Submit | [Esc] Cancel | [Ctrl+C] Quit"),
	)

	s.WriteString(content)

	return lipgloss.NewStyle().
		Render(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).Padding(1, 3).Width(72).Render(s.String()))

}

func SchemaPanels(focusPanel constant.Focus, databases []string, tables []string, columns []string, dbCursor int, tblCursor int, colCursor int) string {

	leftPane := schemaPanel(dbCursor, databases, "Databases", 22, focusPanel == constant.FocusDB)
	middlePane := schemaPanel(tblCursor, tables, "Tables", 24, focusPanel == constant.FocusTable)
	rightPane := schemaPanel(colCursor, columns, "Columns", 22, focusPanel == constant.FocusColumn)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, middlePane, rightPane)
}

func schemaPanel(cursor int, items []string, title string, width int, isFocused bool) string {
	var str MyStringBuilder

	borderColor := inactiveColor
	if isFocused {
		borderColor = highlightColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	str.WriteStrings(borderStyle.Render("┌─ "+title+" ─"+strings.Repeat("─", width-4-len(title))+"┐"), "\n")

	st := cursor - 2
	if st < 0 {
		st = 0
	}
	for i := 0; i < 5; i++ {
		idx := st + i
		line := ""
		isSelected := false

		if idx < len(items) {
			isSelected = (cursor == idx)
			line = truncateTextWithPrfx(items[idx], isSelected, width)
		} else {
			line = truncateTextWithPrfx("", false, width)
		}

		if isSelected && isFocused {
			line = lipgloss.NewStyle().Foreground(highlightColor).Render(line)
		}

		str.WriteStrings(borderStyle.Render("│")+line+borderStyle.Render("│"), "\n")
	}
	str.WriteString(borderStyle.Render("└" + strings.Repeat("─", width) + "┘"))

	return str.String()
}

func truncateTextWithPrfx(name string, isSelected bool, with int) string {
	prefix := "  "
	if isSelected {
		prefix = "> "
	}
	return truncateText(prefix+name, with)
}

func truncateText(name string, maxWidth int) string {
	currentWidth := lipgloss.Width(name)
	if currentWidth <= maxWidth {
		return name + strings.Repeat(" ", maxWidth-currentWidth)
	}

	var w int
	var sb strings.Builder
	for _, r := range name {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-2 {
			break
		}
		sb.WriteRune(r)
		w += rw
	}

	result := sb.String() + ".."
	finalWidth := w + 2
	if finalWidth < maxWidth {
		result += strings.Repeat(" ", maxWidth-finalWidth)
	}

	return result
}

func QueryPanel(isFocused bool, format OutputFormat, text string, rw bool, txPending bool, connName string) string {
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

	modeStr := lipgloss.NewStyle().Foreground(safeColor).Render("[Read Only] ")
	if rw {
		modeStr = lipgloss.NewStyle().Foreground(dangerColor).Bold(true).Render("[Read-Write] ")
	}

	envInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true).Render(modeStr + truncateText(connName, 34-lipgloss.Width(modeStr)))

	metaInfo := lipgloss.JoinHorizontal(lipgloss.Top,
		envInfo,
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(" | Format: "),
		formatBar,
	)

	var statusBar string
	if rw && txPending {
		txAlert := lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			Bold(true).
			Render("⚠️ Uncommitted changes! Please run COMMIT or ROLLBACK.")

		statusBar = lipgloss.JoinVertical(lipgloss.Left, metaInfo, txAlert)
	} else {
		statusBar = lipgloss.JoinVertical(lipgloss.Left, metaInfo)
	}

	var sb MyStringBuilder
	borderStyle := lipgloss.NewStyle().Foreground(sqlBorderColor)
	sb.WriteStrings(borderStyle.Render("┌─ SQL Editor ─"+strings.Repeat("─", 58)+"┐"), "\n")

	sqlContent := lipgloss.JoinVertical(lipgloss.Left, text, "", statusBar)
	lines := strings.Split(sqlContent, "\n")

	for _, line := range lines {
		w := lipgloss.Width(line)
		if w < 70 {
			line += strings.Repeat(" ", 70-w)
		}
		sb.WriteStrings(borderStyle.Render("│")+line+borderStyle.Render("│"), "\n")
	}

	sb.WriteString(borderStyle.Render("└" + strings.Repeat("─", 72) + "┘"))

	return sb.String()
}
