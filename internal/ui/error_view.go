package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgconn"
)

// Define UI Styles using Lipgloss
var (
	errorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF5F87")).
			Padding(1, 2).
			Margin(1, 0)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			Bold(true)

	fileNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Underline(true)

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#767676")).
			Width(4).
			Align(lipgloss.Right)

	errorTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)
)

// PrintSQLError analyzes the error, extracts the line, and prints a visual stack trace.
func PrintSQLError(err error, rawSQL string, fileName string) {
	// Normalize rawSQL for consistent processing
	rawSQL = strings.ReplaceAll(rawSQL, "\r", "")     // Normalize line endings to just \n
	rawSQL = strings.ReplaceAll(rawSQL, "\t", "    ") // Replace tabs with spaces for consistent formatting

	// 1. Unpack the error to see if it's a PostgreSQL error
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// Not a Postgres error, do nothing visually specific
		return
	}

	// 2. Extract Byte Position
	// PostgreSQL returns the error position in bytes, not lines.
	pos := pgErr.Position
	if pos == 0 {
		return // No position data available
	}

	// Convert 1-based position to 0-based index
	bytePos := int(pos) - 1
	if bytePos < 0 {
		bytePos = 0
	}
	if bytePos > len(rawSQL) {
		bytePos = len(rawSQL)
	}

	// 3. Calculate Line Number
	prefix := rawSQL[:bytePos]
	targetLine := strings.Count(prefix, "\n")

	// 4. Extract snippet (e.g., 2 lines before, the error line, 2 lines after)
	lines := strings.Split(rawSQL, "\n")

	var snippetBuilder strings.Builder
	snippetBuilder.WriteString(fmt.Sprintf("\n%s %s\n\n", titleStyle.Render("ERROR"), fileNameStyle.Render(fileName)))
	snippetBuilder.WriteString(fmt.Sprintf("Message: %s\n\n", errorTextStyle.Render(pgErr.Message)))

	startLine := targetLine - 2
	if startLine < 0 {
		startLine = 0
	}
	endLine := targetLine + 2
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	for i := startLine; i <= endLine; i++ {
		lineStr := lines[i]

		// Format line number
		numStr := lineNumStyle.Render(fmt.Sprintf("%d |", i+1))

		if i == targetLine {
			// Highlight the error line
			highlightedLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#E82424")).Render(lineStr)
			snippetBuilder.WriteString(fmt.Sprintf("%s %s  <-- 🐛\n", numStr, highlightedLine))
		} else {
			// Normal line
			normalLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#A8CC8C")).Render(lineStr)
			snippetBuilder.WriteString(fmt.Sprintf("%s %s\n", numStr, normalLine))
		}
	}

	// 5. Render the final box to terminal
	fmt.Println(errorBoxStyle.Render(snippetBuilder.String()))
}
