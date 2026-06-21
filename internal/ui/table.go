package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// Status styles
	appliedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true) // Green
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true) // Yellow
	missingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true) // Red

	// Row style
	rowStyle = lipgloss.NewStyle().Padding(0, 1)
)

// PrintStatusTable renders a beautiful table for migration statuses.
func PrintStatusTable(records [][]string) {
	fmt.Println()

	// Print Header
	fmt.Printf("%-15s %-20s %-35s %-25s\n",
		headerStyle.Render("STATUS"),
		headerStyle.Render("VERSION"),
		headerStyle.Render("MIGRATION NAME"),
		headerStyle.Render("APPLIED AT"),
	)
	fmt.Println(strings.Repeat("─", 95))

	// Print Rows
	for _, row := range records {
		status := row[0]
		version := row[1]
		name := row[2]
		appliedAt := row[3]

		var renderedStatus string
		switch status {
		case "Applied":
			renderedStatus = appliedStyle.Render("🟢 Applied")
		case "Pending":
			renderedStatus = pendingStyle.Render("🟡 Pending")
		case "Missing":
			renderedStatus = missingStyle.Render("🔴 Missing") // In DB, but file deleted
		default:
			renderedStatus = status
		}

		fmt.Printf("%-24s %-20s %-35s %-25s\n",
			renderedStatus,
			rowStyle.Render(version),
			rowStyle.Render(name),
			rowStyle.Render(appliedAt),
		)
	}
	fmt.Println()
}
