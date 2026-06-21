package engine

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/fresmaa/go-tecton/internal/driver"
	"github.com/fresmaa/go-tecton/internal/parser"
	"github.com/fresmaa/go-tecton/internal/ui"
)

// Migrator is the core engine that orchestrates the migration process.
type Migrator struct {
	driver  driver.DatabaseDriver
	fileSys fs.FS
	dirPath string
}

// New creates a new Migrator instance.
func New(dbDriver driver.DatabaseDriver, fileSys fs.FS, dirPath string) *Migrator {
	return &Migrator{
		driver:  dbDriver,
		fileSys: fileSys,
		dirPath: dirPath,
	}
}

// Up executes all pending 'up' migrations.
func (m *Migrator) Up(ctx context.Context) error {
	// 1. Get the list of already applied migrations from the database
	appliedRecords, err := m.driver.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create a map for fast lookup of applied versions
	appliedMap := make(map[int64]bool)
	for _, record := range appliedRecords {
		appliedMap[record.Version] = true
	}

	// 2. Read migration files from the file system
	entries, err := fs.ReadDir(m.fileSys, m.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var pendingFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		// We only care about .up.sql files for the Up command
		if strings.HasSuffix(fileName, ".up.sql") {
			pendingFiles = append(pendingFiles, fileName)
		}
	}

	// Sort files by name (which starts with version number) to ensure correct order
	sort.Strings(pendingFiles)

	// 3. Filter and Execute pending migrations
	executionCount := 0
	for _, fileName := range pendingFiles {
		filePath := path.Join(m.dirPath, fileName)

		// Parse the file to get its version and payload
		payload, err := parser.ParseFile(m.fileSys, filePath)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", fileName, err)
		}

		// Check if this version is already applied
		if appliedMap[payload.Version] {
			continue // Skip, already applied
		}

		// 4. Execute the migration via the driver
		fmt.Printf("Applying migration: %s... ", fileName)
		if err := m.driver.Apply(ctx, payload); err != nil {
			fmt.Println("❌ FAILED")

			// Print SQLError
			ui.PrintSQLError(err, payload.RawSQL, fileName)

			// Return immediately on first error to halt the process
			return fmt.Errorf("error applying %s: %w", fileName, err)
		}

		fmt.Println("✅ OK")
		executionCount++
	}

	if executionCount == 0 {
		fmt.Println("No pending migrations. Database is up to date!")
	} else {
		fmt.Printf("Successfully applied %d migration(s).\n", executionCount)
	}

	return nil
}

// Down reverts the most recently applied migration (1 step rollback).
func (m *Migrator) Down(ctx context.Context) error {
	// 1. Get the list of already applied migrations from the database
	appliedRecords, err := m.driver.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 2. Check if there is anything to revert
	if len(appliedRecords) == 0 {
		fmt.Println("No applied migrations found. Database is already empty.")
		return nil
	}

	// 3. Identify the latest applied migration (Last In, First Out)
	lastRecord := appliedRecords[len(appliedRecords)-1]

	// Reconstruct the expected filename for the down migration
	// Note: Assuming version is a timestamp or unpadded integer (e.g., 20260621120000)
	downFileName := fmt.Sprintf("%d_%s.down.sql", lastRecord.Version, lastRecord.Name)
	filePath := path.Join(m.dirPath, downFileName)

	// 4. Parse the .down.sql file
	payload, err := parser.ParseFile(m.fileSys, filePath)
	if err != nil {
		return fmt.Errorf("failed to parse rollback file %s: %w", downFileName, err)
	}

	// 5. Execute the revert via the driver
	fmt.Printf("Reverting migration: %s... ", downFileName)
	if err := m.driver.Revert(ctx, payload); err != nil {
		fmt.Println("❌ FAILED")
		return fmt.Errorf("error reverting %s: %w", downFileName, err)
	}

	fmt.Println("✅ OK")
	fmt.Printf("Successfully reverted migration version: %d\n", lastRecord.Version)

	return nil
}

// DryRun simulates applying all pending migrations to catch errors before actual deployment.
func (m *Migrator) DryRun(ctx context.Context) error {
	appliedRecords, err := m.driver.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[int64]bool)
	for _, record := range appliedRecords {
		appliedMap[record.Version] = true
	}

	entries, err := fs.ReadDir(m.fileSys, m.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var pendingFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			pendingFiles = append(pendingFiles, entry.Name())
		}
	}
	sort.Strings(pendingFiles)

	fmt.Println("🔍 Starting Dry-Run Validation...")

	validationCount := 0
	for _, fileName := range pendingFiles {
		filePath := path.Join(m.dirPath, fileName)
		payload, err := parser.ParseFile(m.fileSys, filePath)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", fileName, err)
		}

		if appliedMap[payload.Version] {
			continue
		}

		fmt.Printf("Validating: %s... ", fileName)
		if err := m.driver.DryRun(ctx, payload); err != nil {
			fmt.Println("❌ FAILED")

			// Print SQLError
			ui.PrintSQLError(err, payload.RawSQL, fileName)

			// Return immediately on first error to halt the process
			return fmt.Errorf("validation error in %s: %w", fileName, err)
		}

		fmt.Println("✅ PASSED")
		validationCount++
	}

	fmt.Printf("Dry-Run complete. %d file(s) are safe to deploy.\n", validationCount)
	return nil
}
