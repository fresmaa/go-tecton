package engine

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/fresmaa/go-tecton/internal/driver"
	"github.com/fresmaa/go-tecton/internal/ui"
)

// Seeder handles the execution of database seed files.
type Seeder struct {
	driver  driver.DatabaseDriver
	fileSys fs.FS
	dirPath string
}

// NewSeeder creates a new Seeder instance.
func NewSeeder(dbDriver driver.DatabaseDriver, fileSys fs.FS, dirPath string) *Seeder {
	return &Seeder{
		driver:  dbDriver,
		fileSys: fileSys,
		dirPath: dirPath,
	}
}

// Run executes all SQL files found in the specified seeder directory.
func (s *Seeder) Run(ctx context.Context) error {
	entries, err := fs.ReadDir(s.fileSys, s.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read seeder directory: %w", err)
	}

	var seedFiles []string
	for _, entry := range entries {
		// Only target standard .sql files (not .up.sql or .down.sql)
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") && !strings.Contains(entry.Name(), ".up.") && !strings.Contains(entry.Name(), ".down.") {
			seedFiles = append(seedFiles, entry.Name())
		}
	}

	// Sort alphabetically to ensure a predictable execution order (e.g., 01_users.sql, 02_posts.sql)
	sort.Strings(seedFiles)

	if len(seedFiles) == 0 {
		fmt.Println("No seeder files found in directory.")
		return nil
	}

	fmt.Printf("🌱 Found %d seeder file(s). Starting data insertion...\n", len(seedFiles))

	executionCount := 0
	for _, fileName := range seedFiles {
		filePath := path.Join(s.dirPath, fileName)

		fmt.Printf("Seeding: %s... ", fileName)

		rawBytes, err := fs.ReadFile(s.fileSys, filePath)
		if err != nil {
			fmt.Println("❌ FAILED")
			return fmt.Errorf("failed to read seeder file %s: %w", fileName, err)
		}
		rawSQL := string(rawBytes)

		if err := s.driver.ExecuteSeed(ctx, rawSQL); err != nil {
			fmt.Println("❌ FAILED")

			// Print error using UI
			ui.PrintSQLError(err, rawSQL, fileName)

			// Return error to stop the process
			return fmt.Errorf("error executing seeder %s: %w", fileName, err)
		}

		fmt.Println("✅ OK")
		executionCount++
	}

	fmt.Printf("Successfully executed %d seeder(s).\n", executionCount)
	return nil
}
