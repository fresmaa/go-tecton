package parser

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/fresmaa/go-tecton/internal/driver"
)

// filenameRegex validates and extracts parts of the migration filename.
// Expected format: <version>_<name>.<direction>.sql
var filenameRegex = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

// ParseFile reads an SQL file from any fs.FS (supports both physical directories and embed.FS)
// and converts it into a MigrationPayload.
func ParseFile(fileSystem fs.FS, filePath string) (driver.MigrationPayload, error) {
	filename := filepath.Base(filePath)
	matches := filenameRegex.FindStringSubmatch(filename)

	if len(matches) != 4 {
		return driver.MigrationPayload{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	// Extract version
	version, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return driver.MigrationPayload{}, fmt.Errorf("invalid version number in filename: %w", err)
	}

	name := matches[2]

	// Read file content using the universal fs.FS interface
	contentBytes, err := fs.ReadFile(fileSystem, filePath)
	if err != nil {
		return driver.MigrationPayload{}, fmt.Errorf("failed to read migration file: %w", err)
	}

	rawSQL := string(contentBytes)

	// Build the payload
	return driver.MigrationPayload{
		Version:    version,
		Name:       name,
		RawSQL:     rawSQL,
		Statements: SplitStatements(rawSQL), // Calls the safe splitter we created earlier
	}, nil
}
