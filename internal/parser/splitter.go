package parser

import "strings"

// SplitStatements safely splits a raw SQL script into individual statements
// by semicolon (;), ignoring semicolons inside strings or comments.
func SplitStatements(rawSQL string) []string {
	var statements []string
	var currentStmt strings.Builder

	inSingleQuote := false
	inDoubleQuote := false
	inInlineComment := false
	inBlockComment := false

	runes := []rune(rawSQL)
	length := len(runes)

	for i := 0; i < length; i++ {
		char := runes[i]
		var nextChar rune
		if i+1 < length {
			nextChar = runes[i+1]
		}

		// Handle state toggles
		if !inInlineComment && !inBlockComment {
			if char == '\'' && !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else if char == '"' && !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		}

		// Handle comments
		if !inSingleQuote && !inDoubleQuote {
			// Start inline comment (--)
			if char == '-' && nextChar == '-' && !inBlockComment {
				inInlineComment = true
			}
			// End inline comment (\n)
			if char == '\n' && inInlineComment {
				inInlineComment = false
			}
			// Start block comment (/*)
			if char == '/' && nextChar == '*' && !inInlineComment {
				inBlockComment = true
			}
			// End block comment (*/)
			if char == '*' && nextChar == '/' && inBlockComment {
				inBlockComment = false
				currentStmt.WriteRune(char)
				currentStmt.WriteRune(nextChar)
				i++ // Skip the '/'
				continue
			}
		}

		// Split on semicolon if we are not inside any quotes or comments
		if char == ';' && !inSingleQuote && !inDoubleQuote && !inInlineComment && !inBlockComment {
			stmt := strings.TrimSpace(currentStmt.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStmt.Reset()
			continue
		}

		currentStmt.WriteRune(char)
	}

	// Append any remaining text as the last statement
	lastStmt := strings.TrimSpace(currentStmt.String())
	if lastStmt != "" {
		statements = append(statements, lastStmt)
	}

	return statements
}
