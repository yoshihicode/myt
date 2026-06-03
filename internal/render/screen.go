package render

import (
	"myt/internal/config"
	"myt/internal/constant"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var highlightColor = lipgloss.Color("62")
var inactiveColor = lipgloss.Color("240")
var dangerColor = lipgloss.Color("9")
var safeColor = lipgloss.Color("10")

func Help() string {
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

	return helpBox
}

func Config(configs []config.Config, configCursor int, errorMsg string) string {
	var s MyStringBuilder
	s.WriteStrings("\n", lipgloss.NewStyle().Foreground(highlightColor).Bold(true).Render(" Select Connection"), "\n\n")

	if errorMsg != "" {
		s.WriteStrings(lipgloss.NewStyle().Foreground(dangerColor).Render("Error: "+errorMsg), "\n\n")
	}

	maxName := 4
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

		if lipgloss.Width(cfg.Name) > maxName {
			maxName = lipgloss.Width(cfg.Name)
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

func SchemaPanel(focusPanel constant.Focus, databases []string, tables []string, columns []string, dbCursor int, tblCursor int, colCursor int) string {
	schemaBorderColor := inactiveColor
	if focusPanel == constant.FocusDB {
		schemaBorderColor = highlightColor
	}
	borderStyle := lipgloss.NewStyle().Foreground(schemaBorderColor)

	var dStr, tStr, cStr MyStringBuilder

	// Databases
	dStr.WriteStrings(borderStyle.Render("┌─ Databases ─"+strings.Repeat("─", 9)+"┐"), "\n")

	startD := dbCursor - 2
	if startD < 0 {
		startD = 0
	}
	for i := 0; i < 5; i++ {
		idx := startD + i
		lineText := ""
		isSelected := false

		if idx < len(databases) {
			isSelected = (dbCursor == idx)
			lineText = truncateText(databases[idx], 18)
		}

		formattedLine := padRight(lineText, isSelected, 20)

		if isSelected && focusPanel == constant.FocusDB {
			formattedLine = lipgloss.NewStyle().Foreground(highlightColor).Render(formattedLine)
		}

		dStr.WriteStrings(borderStyle.Render("│")+formattedLine+borderStyle.Render("│"), "\n")
	}
	dStr.WriteString(borderStyle.Render("└" + strings.Repeat("─", 22) + "┘"))

	// Tables
	schemaBorderColor = inactiveColor
	if focusPanel == constant.FocusTable {
		schemaBorderColor = highlightColor
	}
	borderStyle = lipgloss.NewStyle().Foreground(schemaBorderColor)
	tStr.WriteStrings(borderStyle.Render("┌─ Tables ─"+strings.Repeat("─", 14)+"┐"), "\n")

	startT := tblCursor - 2
	if startT < 0 {
		startT = 0
	}
	for i := 0; i < 5; i++ {
		idx := startT + i
		lineText := ""
		isSelected := false

		if idx < len(tables) {
			isSelected = (tblCursor == idx)
			lineText = truncateText(tables[idx], 20)
		}

		formattedLine := padRight(lineText, isSelected, 22)

		if isSelected && constant.FocusTable == focusPanel {
			formattedLine = lipgloss.NewStyle().Foreground(highlightColor).Render(formattedLine)
		}

		tStr.WriteStrings(borderStyle.Render("│")+formattedLine+borderStyle.Render("│"), "\n")
	}
	tStr.WriteString(borderStyle.Render("└" + strings.Repeat("─", 24) + "┘"))

	// Columns
	schemaBorderColor = inactiveColor
	if focusPanel == constant.FocusColumn {
		schemaBorderColor = highlightColor
	}
	borderStyle = lipgloss.NewStyle().Foreground(schemaBorderColor)
	cStr.WriteStrings(borderStyle.Render("┌─ Columns ─"+strings.Repeat("─", 11)+"┐"), "\n")

	startC := colCursor - 2
	if startC < 0 {
		startC = 0
	}
	for i := 0; i < 5; i++ {
		idx := startC + i
		lineText := ""
		isSelected := false

		if idx < len(columns) {
			isSelected = (colCursor == idx)
			lineText = truncateText(columns[idx], 18)
		}

		formattedLine := padRight(lineText, isSelected, 20)

		if isSelected && focusPanel == constant.FocusColumn {
			formattedLine = lipgloss.NewStyle().Foreground(highlightColor).Render(formattedLine)
		}

		cStr.WriteStrings(borderStyle.Render("│")+formattedLine+borderStyle.Render("│"), "\n")
	}
	cStr.WriteString(borderStyle.Render("└" + strings.Repeat("─", 22) + "┘"))

	leftPane := dStr.String()
	middlePane := tStr.String()
	rightPane := cStr.String()

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, middlePane, rightPane)
}

func padRight(text string, isSelected bool, targetLen int) string {
	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	fullText := prefix + text

	currentLen := len([]rune(fullText))

	padLen := targetLen - currentLen
	if padLen < 0 {
		padLen = 0
	}

	return " " + fullText + strings.Repeat(" ", padLen) + " "
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

	envInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true).Render(modeStr + truncateText(connName, 22))

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

	sqlContent := lipgloss.JoinVertical(lipgloss.Left, text, "", statusBar)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(sqlBorderColor).
		Width(72).
		Render(sqlContent)
}
