package core

import (
	"context"
	"fmt"
	"myt/internal/database"
	"myt/internal/render"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.Conn != nil {
				m.Conn.Close()
			}
			if m.DB != nil {
				m.DB.Close()
			}
			return m, tea.Quit
		}

		if m.State == SelectConfig {
			switch msg.String() {
			case "q", "esc", "ctrl+c":
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
				err := m.InitConnection(m.Configs[m.ConfigCursor])
				if err != nil {
					m.ErrorMsg = err.Error()
				} else {
					m.State = Main
				}
			}
			return m, nil
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

		if msg.String() == "tab" || msg.String() == "shift+tab" {
			m.FocusSQL = !m.FocusSQL
			if m.FocusSQL {
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
			m.Connect()
			return m, nil
		}

		if m.FocusSQL {
			if msg.String() == "ctrl+space" || msg.String() == "ctrl+@" || msg.String() == "ctrl+n" {
				m.Autocomplete()
				return m, nil
			}
			m.TabMatches = nil
			m.SqlInput, cmd = m.SqlInput.Update(msg)
			return m, cmd
		} else {
			switch msg.String() {
			case "left", "h":
				if m.SchemaPane > 0 {
					m.SchemaPane--
				}
			case "right", "l":
				if m.SchemaPane == 0 && len(m.Tables) > 0 {
					m.SchemaPane = 1
				} else if m.SchemaPane == 1 && len(m.Columns) > 0 {
					m.SchemaPane = 2
				}
			case "up", "j":
				if m.SchemaPane == 0 && m.DBCursor > 0 {
					m.DBCursor--
					m.Connect()
				} else if m.SchemaPane == 1 && m.TableCursor > 0 {
					m.TableCursor--
					m.UpdateColumns()
				} else if m.SchemaPane == 2 && m.ColumnCursor > 0 {
					m.ColumnCursor--
				}
			case "down", "k":
				if m.SchemaPane == 0 && m.DBCursor < len(m.Databases)-1 {
					m.DBCursor++
					m.Connect()
				} else if m.SchemaPane == 1 && m.TableCursor < len(m.Tables)-1 {
					m.TableCursor++
					m.UpdateColumns()
				} else if m.SchemaPane == 2 && m.ColumnCursor < len(m.Columns)-1 {
					m.ColumnCursor++
				}
			case "enter":
				if m.SchemaPane == 0 && m.DBCursor < len(m.Databases) {
					m.Connect()
					if len(m.Tables) > 0 {
						m.SchemaPane = 1
					}
				} else if m.SchemaPane == 1 && m.TableCursor < len(m.Tables) {
					m.InsertStringToSQL(m.Tables[m.TableCursor] + " ")
					m.FocusSQL = true
					m.SqlInput.Focus()
				} else if m.SchemaPane == 2 && m.ColumnCursor < len(m.Columns) {
					m.InsertStringToSQL(m.Columns[m.ColumnCursor] + " ")
					m.FocusSQL = true
					m.SqlInput.Focus()
				}
			}
		}
	default:
		if m.FocusSQL {
			m.SqlInput, cmd = m.SqlInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) Autocomplete() {
	m.SqlInput.InsertString("\uF000")
	textWithSentinel := m.SqlInput.Value()

	m.SqlInput, _ = m.SqlInput.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	runesWithSentinel := []rune(textWithSentinel)
	absoluteRunePos := -1
	for i, r := range runesWithSentinel {
		if r == '\uF000' {
			absoluteRunePos = i
			break
		}
	}

	if absoluteRunePos == -1 {
		return
	}

	text := m.SqlInput.Value()
	runes := []rune(text)
	if absoluteRunePos > len(runes) {
		absoluteRunePos = len(runes)
	}

	textUpToCursor := string(runes[:absoluteRunePos])

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

	if lastWord == "" {
		return
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
			cleanText := strings.ReplaceAll(text, "`", "")
			pattern := `(?i)(?:FROM|JOIN|,)\s+([^\s,()]+)\s+(?:AS\s+)?` + regexp.QuoteMeta(aliasOrTable) + `(?:\s|$|,|\))`
			re := regexp.MustCompile(pattern)

			matches := re.FindAllStringSubmatch(cleanText, -1)
			for _, match := range matches {
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
		sort.Strings(m.TabMatches)
		m.TabMatchIdx = 0

		lastWordRunesLen := len([]rune(lastWord))
		for i := 0; i < lastWordRunesLen; i++ {
			m.SqlInput, _ = m.SqlInput.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		}
		m.SqlInput.InsertString(m.TabMatches[0])
	}
}

func (m *Model) ExecuteSQL() tea.Cmd {
	rawInput := m.SqlInput.Value()
	queries := strings.Split(rawInput, ";")
	var finalOutput render.MyStringBuilder
	queryCount := 0

	for _, q := range queries {
		query := strings.TrimSpace(q)
		if query == "" {
			continue
		}
		queryCount++
		qUpper := strings.ToUpper(query)

		isAllowed := m.ReadWrite
		if !isAllowed {
			allowedCommands := []string{"SELECT", "SHOW", "DESC", "EXPLAIN", "WITH"}
			for _, cmd := range allowedCommands {
				if strings.HasPrefix(qUpper, cmd) {
					isAllowed = true
					break
				}
			}
		}

		header := lipgloss.NewStyle().Foreground(lipgloss.Color("36")).Render(fmt.Sprintf("--- Result [%d]: %s ---", queryCount, query))
		finalOutput.WriteStrings("\n", header, "\n")

		if !isAllowed {
			finalOutput.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("【ERROR】Read-only mode. Restart with '-rw' flag to enable modifications.\n"))
			continue
		}

		res, err := database.ExecuteQuery(context.Background(), m.Conn, query)
		if err != nil {
			finalOutput.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v\n", err)))
			continue
		}

		formatted := render.FormatResult(res, m.OutputFormat)
		finalOutput.WriteString(formatted)
	}

	if finalOutput.Len() == 0 {
		return nil
	}

	var cmds []tea.Cmd
	lines := strings.Split(finalOutput.String(), "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		cmds = append(cmds, tea.Println(line))
	}
	return tea.Sequence(cmds...)
}
