package linter

import (
	"regexp"
	"strings"
)

// Level represents the severity of a linter violation.
type Level string

const (
	LevelWarning Level = "WARNING"
	LevelError   Level = "ERROR"
)

// Rule defines a static analysis check for SQL statements.
type Rule struct {
	Code        string
	Description string
	Level       Level
	// Check evaluates a single SQL line. Returns true if a violation is found.
	Check func(line string) bool
}

// Violation represents an instance where a SQL line broke a Rule.
type Violation struct {
	RuleCode    string
	Description string
	Level       Level
	LineNumber  int
	LineText    string
}

// DefaultRules contains the standard set of safety checks.
var DefaultRules = []Rule{
	{
		Code:        "L001",
		Description: "Destructive operation detected (DROP TABLE). Ensure you have a proper backup.",
		Level:       LevelWarning,
		Check: func(line string) bool {
			// Matches "DROP TABLE" case-insensitively
			re := regexp.MustCompile(`(?i)\bDROP\s+TABLE\b`)
			return re.MatchString(line)
		},
	},
	{
		Code:        "L002",
		Description: "Adding a column with a DEFAULT value can cause a full table lock in large tables.",
		Level:       LevelWarning,
		Check: func(line string) bool {
			// Matches "ADD COLUMN ... DEFAULT" case-insensitively
			re := regexp.MustCompile(`(?i)\bADD\s+COLUMN\b.*\bDEFAULT\b`)
			return re.MatchString(line)
		},
	},
}

// Engine is responsible for running a set of rules against SQL payloads.
type Engine struct {
	rules []Rule
}

// New creates a new Linter Engine with the default ruleset.
func New() *Engine {
	return &Engine{
		rules: DefaultRules,
	}
}

// Analyze scans the entire SQL payload line-by-line and returns any violations.
func (e *Engine) Analyze(sqlPayload string) []Violation {
	var violations []Violation

	// Normalize line endings to standard newline
	normalizedSQL := strings.ReplaceAll(sqlPayload, "\r\n", "\n")
	lines := strings.Split(normalizedSQL, "\n")

	for i, line := range lines {
		// Skip empty lines or pure comments (basic check)
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		for _, rule := range e.rules {
			if rule.Check(line) {
				violations = append(violations, Violation{
					RuleCode:    rule.Code,
					Description: rule.Description,
					Level:       rule.Level,
					LineNumber:  i + 1, // 1-based line number
					LineText:    strings.TrimSpace(line),
				})
			}
		}
	}

	return violations
}
