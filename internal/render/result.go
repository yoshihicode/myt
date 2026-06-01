package render

import (
	"encoding/json"
	"strconv"
	"strings"

	"myt/internal/database"

	"github.com/charmbracelet/lipgloss"
)

type OutputFormat int

const (
	Grid OutputFormat = iota
	Markdown
	CSV
	JSON
)

var FormatNames = []string{"GRID", "MARKDOWN", "CSV", "JSON"}

func Format(res *database.QueryResult, format OutputFormat) string {
	if res.Message != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(res.Message + "\n")
	}

	var sb MyStringBuilder
	cols := res.Columns
	results := res.Rows

	switch format {
	case Markdown:
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

	case CSV:
		sb.WriteStrings(strings.Join(cols, ","), "\n")
		for _, row := range results {
			var rowStrs []string
			for _, col := range cols {
				str := getStringValue(row[col])
				if row[col] == nil {
					rowStrs = append(rowStrs, "")
				} else {
					rowStrs = append(rowStrs, "\""+strings.ReplaceAll(str, "\"", "\"\"")+"\"")
				}
			}
			sb.WriteStrings(strings.Join(rowStrs, ","), "\n")
		}

	case JSON:
		jsonData, _ := json.MarshalIndent(results, "", "  ")
		sb.WriteStrings(string(jsonData), "\n")

	case Grid:
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
			sb.WriteStrings(" ", col, strings.Repeat(" ", pad), " |")

		}
		sb.WriteStrings("\n", sep, "\n")

		for _, row := range results {
			sb.WriteString("|")
			for i, col := range cols {
				str := getStringValue(row[col])
				pad := colWidths[i] - lipgloss.Width(str)
				sb.WriteStrings(" ", str, strings.Repeat(" ", pad), " |")
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
	var res string
	switch val := v.(type) {
	case string:
		res = val
	case int64:
		res = strconv.FormatInt(val, 10)
	case float64:
		res = strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		res = strconv.FormatBool(val)
	default:
		if stringer, ok := v.(interface{ String() string }); ok {
			res = stringer.String()
		} else {
			res = ""
		}
	}

	return strings.ReplaceAll(res, "\n", " ")
}
