package core

import (
	"context"
	"myt/internal/constant"
	"myt/internal/database"
	"myt/internal/render"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if msg.String() == "ctrl+c" {
			m.Close()
			return m, tea.Quit
		}

		if m.State == constant.AppStateConfig {
			switch msg.String() {
			case "q", "esc":
				return m, tea.Quit
			case "up", "j":
				if m.ConfigCursor > 0 {
					m.ConfigCursor--
				}
			case "down", "k":
				if m.ConfigCursor < len(m.Configs)-1 {
					m.ConfigCursor++
				}
			case "enter":
				if m.Configs[m.ConfigCursor].SSHUser != "" && m.Configs[m.ConfigCursor].SSHKey == "" && m.Configs[m.ConfigCursor].SSHPass == "" {
					m.SetPasswordSubmit("SSH")
					return m, textinput.Blink
				}

				if m.Configs[m.ConfigCursor].Pass == "" {
					m.SetPasswordSubmit("DB")
					return m, textinput.Blink
				}

				netType, err := m.GetNetType()
				if err != nil {
					m.ErrorMsg = err.Error()
					return m, nil
				}
				err = m.InitConnection(m.Configs[m.ConfigCursor], netType)
				if err != nil {
					m.ErrorMsg = err.Error()
				} else {
					m.State = constant.AppStateMain
				}
			}
			return m, nil
		}

		if m.State == constant.AppStatePassword {
			m.PasswordInput, cmd = m.PasswordInput.Update(msg)

			switch msg.String() {
			case "enter":
				m.ErrorMsg = ""
				v := m.PasswordInput.Value()

				switch m.PromptTarget {
				case "SSH":
					m.Configs[m.ConfigCursor].SSHPass = v
					if m.Configs[m.ConfigCursor].Pass == "" {
						m.SetPasswordSubmit("DB")
						return m, textinput.Blink
					}
					netType, err := m.GetNetType()
					if err != nil {
						m.ErrorMsg = err.Error()
						m.Configs[m.ConfigCursor].SSHPass = ""
						m.SetPasswordSubmit("SSH")
						return m, textinput.Blink
					}
					err = m.InitConnection(m.Configs[m.ConfigCursor], netType)
					if err != nil {
						m.ErrorMsg = err.Error()
						m.Configs[m.ConfigCursor].Pass = ""
						m.SetPasswordSubmit("DB")
						return m, textinput.Blink
					} else {
						m.ErrorMsg = ""
						m.State = constant.AppStateMain
					}
					return m, nil

				case "DB":
					m.Configs[m.ConfigCursor].Pass = v
					netType, err := m.GetNetType()
					if err != nil {
						m.ErrorMsg = err.Error()
						m.Configs[m.ConfigCursor].SSHPass = ""
						m.Configs[m.ConfigCursor].Pass = ""
						m.SetPasswordSubmit("SSH")
						return m, textinput.Blink
					}
					err = m.InitConnection(m.Configs[m.ConfigCursor], netType)
					if err != nil {
						m.ErrorMsg = err.Error()
						m.Configs[m.ConfigCursor].Pass = ""
						m.SetPasswordSubmit("DB")
						return m, textinput.Blink
					} else {
						m.State = constant.AppStateMain
					}
					return m, nil
				}

			case "esc":
				if m.ConnectionSelect {
					m.State = constant.AppStateConfig
					m.ErrorMsg = ""
					return m, nil
				} else {
					m.Close()
					return m, tea.Quit
				}
			}
			return m, cmd
		}

		if msg.String() == "ctrl+l" {
			return m, tea.ClearScreen
		}

		if m.ShowHelp {
			if msg.String() == "ctrl+h" || msg.String() == "esc" || msg.String() == "enter" {
				m.ShowHelp = false
			}
			return m, nil
		}
		if msg.String() == "ctrl+h" {
			m.ShowHelp = true
			return m, nil
		}

		if msg.String() == "tab" {
			m.FocusPanel = (m.FocusPanel + 1) % 4
			if m.FocusPanel == constant.FocusEditor {
				m.SqlInput.Focus()
			} else {
				m.SqlInput.Blur()
			}
			return m, nil
		}

		if msg.String() == "shift+tab" {
			m.FocusPanel = (m.FocusPanel - 1 + 4) % 4
			if m.FocusPanel == constant.FocusEditor {
				m.SqlInput.Focus()
			} else {
				m.SqlInput.Blur()
			}
			return m, nil
		}

		if msg.String() == "ctrl+f" {
			m.OutputFormat = (m.OutputFormat + 1) % render.OutputFormat(len(render.FormatNames))
			return m, nil
		}
		if msg.String() == "ctrl+e" {
			return m, m.ExecuteSQL()
		}
		if msg.String() == "ctrl+u" {
			m.SqlInput.SetValue("")
			return m, nil
		}
		if msg.String() == "ctrl+r" {
			m.Connect(m.Configs[m.ConfigCursor])
			return m, nil
		}
		if msg.String() == "esc" {
			if m.ConnectionSelect {
				m.Close()
				m.DBCursor = 0
				m.TableCursor = 0
				m.ColumnCursor = 0
				m.State = constant.AppStateConfig
				m.FocusPanel = 0
				m.ErrorMsg = ""
				return m, nil
			}
		}

		if m.FocusPanel == constant.FocusEditor {
			if msg.String() == "ctrl+space" || msg.String() == "ctrl+@" || msg.String() == "ctrl+n" {
				m.Autocomplete()
				return m, nil
			}
			m.TabMatches = nil
			m.SqlInput, cmd = m.SqlInput.Update(msg)
			return m, cmd
		} else {
			switch msg.String() {
			case "up", "j":
				if m.FocusPanel == constant.FocusDB && m.DBCursor > 0 {
					m.DBCursor--
					m.Connect(m.Configs[m.ConfigCursor])
				} else if m.FocusPanel == constant.FocusTable && m.TableCursor > 0 {
					m.TableCursor--
					m.UpdateColumns()
				} else if m.FocusPanel == constant.FocusColumn && m.ColumnCursor > 0 {
					m.ColumnCursor--
				}
			case "down", "k":
				if m.FocusPanel == constant.FocusDB && m.DBCursor < len(m.Databases)-1 {
					m.DBCursor++
					m.Connect(m.Configs[m.ConfigCursor])
				} else if m.FocusPanel == constant.FocusTable && m.TableCursor < len(m.Tables)-1 {
					m.TableCursor++
					m.UpdateColumns()
				} else if m.FocusPanel == constant.FocusColumn && m.ColumnCursor < len(m.Columns)-1 {
					m.ColumnCursor++
				}
			case "enter":
				if m.FocusPanel == constant.FocusDB && m.DBCursor < len(m.Databases) {
					m.Connect(m.Configs[m.ConfigCursor])
					if len(m.Tables) > 0 {
						m.FocusPanel = constant.FocusTable
					}
				} else if m.FocusPanel == constant.FocusTable && m.TableCursor < len(m.Tables) {
					m.InsertStringToSQL(m.Tables[m.TableCursor] + " ")
					m.FocusPanel = constant.FocusEditor
					m.SqlInput.Focus()
				} else if m.FocusPanel == constant.FocusColumn && m.ColumnCursor < len(m.Columns) {
					m.InsertStringToSQL(m.Columns[m.ColumnCursor] + " ")
					m.FocusPanel = constant.FocusEditor
					m.SqlInput.Focus()
				}
			}
		}

	default:
		if m.FocusPanel == constant.FocusEditor {
			m.SqlInput, cmd = m.SqlInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) Autocomplete() {
	m.SqlInput.InsertString("\uF000")
	v := m.SqlInput.Value()

	m.SqlInput, _ = m.SqlInput.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	r := []rune(v)
	pos := -1
	for i, r := range r {
		if r == '\uF000' {
			pos = i
			break
		}
	}

	if pos == -1 {
		return
	}

	text := m.SqlInput.Value()
	runes := []rune(text)
	if pos > len(runes) {
		pos = len(runes)
	}

	textUpToCursor := string(runes[:pos])
	textAfterCursor := string(runes[pos:])

	startIdx := strings.LastIndex(textUpToCursor, ";")
	currentQueryStart := textUpToCursor
	if startIdx != -1 {
		currentQueryStart = textUpToCursor[startIdx+1:]
	}

	endIdx := strings.Index(textAfterCursor, ";")
	currentQueryEnd := textAfterCursor
	if endIdx != -1 {
		currentQueryEnd = textAfterCursor[:endIdx]
	}

	currentQueryText := currentQueryStart + currentQueryEnd

	if len(m.TabMatches) > 0 {
		currentMatch := m.TabMatches[m.TabMatchIdx]
		if strings.HasSuffix(strings.ToUpper(textUpToCursor), strings.ToUpper(currentMatch)) {
			matchRunesLen := len([]rune(currentMatch))
			for i := 0; i < matchRunesLen; i++ {
				m.SqlInput, _ = m.SqlInput.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			}

			m.TabMatchIdx = (m.TabMatchIdx + 1) % len(m.TabMatches)
			nextMatch := m.TabMatches[m.TabMatchIdx]

			m.SqlInput.InsertString(nextMatch)
			return
		}
	}

	m.TabMatches = nil

	isDelimiter := func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == ',' ||
			r == '(' || r == ')' || r == ';' || r == '`' || r == '"' ||
			r == ' ' || r == '、' || r == '。' ||
			r == '=' || r == '<' || r == '>' || r == '+' || r == '-' || r == '*' || r == '/' || r == '!'
	}

	idx := strings.LastIndexFunc(textUpToCursor, isDelimiter)
	var lastWord string
	if idx == -1 {
		lastWord = textUpToCursor
	} else {
		_, size := utf8.DecodeRuneInString(textUpToCursor[idx:])
		lastWord = textUpToCursor[idx+size:]
	}

	if strings.Contains(lastWord, ".") {
		parts := strings.LastIndex(lastWord, ".")
		aliasOrTable := lastWord[:parts]
		colPrefix := lastWord[parts+1:]

		findCols := func(tableName string) bool {
			if cols, ok := m.TableColumns[strings.ToUpper(tableName)]; ok {
				if colPrefix == "*" {
					var allCols []string
					for _, c := range cols {
						allCols = append(allCols, aliasOrTable+"."+c)
					}
					m.TabMatches = append(m.TabMatches, strings.Join(allCols, ", "))
					return true
				}
				for _, c := range cols {
					if strings.HasPrefix(strings.ToUpper(c), strings.ToUpper(colPrefix)) {
						m.TabMatches = append(m.TabMatches, aliasOrTable+"."+c)
					}
				}
				return len(m.TabMatches) > 0
			}
			return false
		}

		if !findCols(aliasOrTable) {
			cleanText := strings.ReplaceAll(currentQueryText, "`", "")
			pattern := `(?i)(?:FROM|JOIN|,)\s+([^\s,()]+)\s+(?:AS\s+)?` + regexp.QuoteMeta(aliasOrTable) + `(?:\s|$|,|\))`
			re := regexp.MustCompile(pattern)

			matches := re.FindAllStringSubmatch(cleanText, -1)
			for i := len(matches) - 1; i >= 0; i-- {
				match := matches[i]
				if len(match) >= 2 {
					tableName := match[1]
					if strings.Contains(tableName, ".") {
						tParts := strings.Split(tableName, ".")
						tableName = tParts[len(tParts)-1]
					}
					if findCols(tableName) {
						break
					}
				}
			}
		}
	} else {
		prefix := strings.ToUpper(lastWord)
		for _, word := range m.AutocompleteDict {
			if strings.HasPrefix(strings.ToUpper(word), prefix) {
				m.TabMatches = append(m.TabMatches, word)
			}
		}
	}

	if len(m.TabMatches) > 0 {
		m.TabMatchIdx = 0

		l := len([]rune(lastWord))
		for i := 0; i < l; i++ {
			m.SqlInput, _ = m.SqlInput.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		}
		m.SqlInput.InsertString(m.TabMatches[0])
	}
}

func (m *Model) ExecuteSQL() tea.Cmd {
	rawInput := m.SqlInput.Value()
	queries := strings.Split(rawInput, ";")
	var output render.MyStringBuilder
	cnt := 0

	for _, q := range queries {
		query := cleanQuery(q)
		if query == "" {
			continue
		}
		cnt++
		qUpper := strings.ToUpper(query)

		isAllowed := m.Configs[m.ConfigCursor].ReadWrite
		if !isAllowed {
			allowedCommands := []string{"SELECT", "SHOW", "DESC", "EXPLAIN", "WITH"}
			for _, cmd := range allowedCommands {
				if strings.HasPrefix(qUpper, cmd) {
					isAllowed = true
					break
				}
			}
		}
		nowStr := time.Now().Format("2006-01-02 15:04:05.000")
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render("--- Result [" + nowStr + "]: [" + strconv.Itoa(cnt) + "] " + query + " ---")
		output.WriteStrings("\n", header, "\n")

		if !isAllowed {
			output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: Read-only mode. Restart with '-rw' flag to enable modifications.\n"))
			continue
		}

		res, err := database.ExecuteQuery(context.Background(), m.Conn, query)
		if err != nil {
			output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: " + err.Error() + "\n"))
			continue
		}
		if isAllowed {
			if strings.HasPrefix(qUpper, "UPDATE") ||
				strings.HasPrefix(qUpper, "INSERT") ||
				strings.HasPrefix(qUpper, "DELETE") ||
				strings.HasPrefix(qUpper, "REPLACE") {
				m.TxPending = true
			}

			if strings.HasPrefix(qUpper, "COMMIT") ||
				strings.HasPrefix(qUpper, "ROLLBACK") ||
				strings.HasPrefix(qUpper, "CREATE") || // Implicit Commit
				strings.HasPrefix(qUpper, "DROP") || // Implicit Commit
				strings.HasPrefix(qUpper, "ALTER") || // Implicit Commit
				strings.HasPrefix(qUpper, "TRUNCATE") { // Implicit Commit
				m.TxPending = false
			}
		}
		formatted := render.Format(res, m.OutputFormat)
		output.WriteString(formatted)
	}

	if output.Len() == 0 {
		return nil
	}

	if m.Configs[m.ConfigCursor].Tee != "" {
		f, err := os.OpenFile(m.Configs[m.ConfigCursor].Tee, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return tea.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: Failed to open output file: " + err.Error()))
		}
		defer f.Close()

		if _, err := f.WriteString(ansiRegex.ReplaceAllString(output.String(), "")); err != nil {
			return tea.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: Failed to write to file: " + err.Error()))
		}
	}

	var cmds []tea.Cmd
	lines := strings.Split(output.String(), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		cmds = append(cmds, tea.Println(line))
	}
	return tea.Sequence(cmds...)
}

var (
	ansiRegex         = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	blockCommentRegex = regexp.MustCompile(`(?s)/\*[^*]*\*+(?:[^/*][^*]*\*+)*/`)
	lineCommentRegex  = regexp.MustCompile(`(?m)(?:--|#).*$`)
)

func cleanQuery(q string) string {
	q = blockCommentRegex.ReplaceAllString(q, "")
	q = lineCommentRegex.ReplaceAllString(q, "")

	return strings.TrimSpace(q)
}
