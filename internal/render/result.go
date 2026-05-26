package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"myt/internal/database"

	"github.com/charmbracelet/lipgloss"
)

type OutputFormat int

const (
	FormatGrid OutputFormat = iota
	FormatMarkdown
	FormatCSV
	FormatJSON
)

var FormatNames = []string{"GRID", "MARKDOWN", "CSV", "JSON"}

func FormatResult(res *database.QueryResult, format OutputFormat) string {
	if res.Message != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(res.Message + "\n")
	}

	var sb MyStringBuilder
	cols := res.Columns
	results := res.Rows

	switch format {
	case FormatMarkdown:
		sb.WriteStrings("| ", strings.Join(cols, " | "), " |\n|")
		for range cols {
			sb.WriteString("---|")
		}
		sb.WriteString("\n")
		for _, row := range results {
			var rowStrs []string
			for _, col := range cols {
				str := getStringValue(row[col])
				str = strings.ReplaceAll(str, "|", "\\|")
				rowStrs = append(rowStrs, str)
			}
			sb.WriteStrings("| ", strings.Join(rowStrs, " | "), " |\n")
		}

	case FormatCSV:
		sb.WriteStrings(strings.Join(cols, ","), "\n")
		for _, row := range results {
			var rowStrs []string
			for _, col := range cols {
				str := getStringValue(row[col])
				if row[col] == nil {
					rowStrs = append(rowStrs, "")
				} else {
					rowStrs = append(rowStrs, fmt.Sprintf("\"%s\"", strings.ReplaceAll(str, "\"", "\"\"")))
				}
			}
			sb.WriteStrings(strings.Join(rowStrs, ","), "\n")
		}

	case FormatJSON:
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		sb.WriteStrings(string(jsonData), "\n")

	case FormatGrid:
		colWidths := make([]int, len(cols))
		for i, col := range cols {
			colWidths[i] = lipgloss.Width(col)
		}
		for _, row := range results {
			for i, col := range cols {
				str := getStringValue(row[col])
				w := lipgloss.Width(str)
				if w > colWidths[i] {
					colWidths[i] = w
				}
			}
		}

		makeSeparator := func() string {
			var sep strings.Builder
			sep.WriteString("+")
			for _, w := range colWidths {
				sep.WriteString(strings.Repeat("-", w+2))
				sep.WriteString("+")
			}
			return sep.String()
		}

		sep := makeSeparator()
		sb.WriteStrings(sep, "\n|")
		for i, col := range cols {
			pad := colWidths[i] - lipgloss.Width(col)
			sb.WriteString(fmt.Sprintf(" %s%s |", col, strings.Repeat(" ", pad)))
		}
		sb.WriteStrings("\n", sep, "\n")

		for _, row := range results {
			sb.WriteString("|")
			for i, col := range cols {
				str := getStringValue(row[col])
				pad := colWidths[i] - lipgloss.Width(str)
				sb.WriteString(fmt.Sprintf(" %s%s |", str, strings.Repeat(" ", pad)))
			}
			sb.WriteString("\n")
		}
		if len(results) > 0 {
			sb.WriteStrings(sep, "\n")
		}
	}

	return sb.String()
}

func getStringValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " ")
}
