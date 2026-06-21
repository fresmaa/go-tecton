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

	// Create a map for fast lookup of applied versions AND find the maximum batch number
	appliedMap := make(map[int64]bool)
	maxBatch := 0

	for _, record := range appliedRecords {
		appliedMap[record.Version] = true

		// Get the highest batch number from the applied migrations
		if record.Batch > maxBatch {
			maxBatch = record.Batch
		}
	}

	// The current execution batch is the highest batch number + 1
	nextBatch := maxBatch + 1

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

		// Set the batch number to the current batch
		payload.Batch = nextBatch

		// 4. Execute the migration via the driver
		fmt.Printf("Applying migration: %s (Batch %d)... ", fileName, nextBatch)
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
		fmt.Printf("Successfully applied %d migration(s) in Batch %d.\n", executionCount, nextBatch)
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

	// 3. Find the highest batch number
	maxBatch := 0
	for _, record := range appliedRecords {
		if record.Batch > maxBatch {
			maxBatch = record.Batch
		}
	}

	// 4. Collect all migrations belonging to the highest batch
	var batchToRevert []driver.MigrationRecord
	for _, record := range appliedRecords {
		if record.Batch == maxBatch {
			batchToRevert = append(batchToRevert, record)
		}
	}

	// 5. Sort descending (LIFO - Last In, First Out for the batch)
	// We want to revert the newest migrations first
	sort.Slice(batchToRevert, func(i, j int) bool {
		return batchToRevert[i].Version > batchToRevert[j].Version
	})

	fmt.Printf("⏪ Reverting %d migration(s) from Batch %d...\n", len(batchToRevert), maxBatch)

	// 6. Execute the revert for each file in the batch
	executionCount := 0
	for _, record := range batchToRevert {
		// Reconstruct the expected filename for the down migration
		downFileName := fmt.Sprintf("%d_%s.down.sql", record.Version, record.Name)
		filePath := path.Join(m.dirPath, downFileName)

		// Parse the .down.sql file
		payload, err := parser.ParseFile(m.fileSys, filePath)
		if err != nil {
			return fmt.Errorf("failed to parse rollback file %s: %w", downFileName, err)
		}

		// Ensure the payload has the exact version to tell the driver which record to delete
		payload.Version = record.Version

		// Execute the revert via the driver
		fmt.Printf("Reverting migration: %s... ", downFileName)
		if err := m.driver.Revert(ctx, payload); err != nil {
			fmt.Println("❌ FAILED")
			return fmt.Errorf("error reverting %s: %w", downFileName, err)
		}

		fmt.Println("✅ OK")
		executionCount++
	}

	fmt.Printf("Successfully reverted %d migration(s) (Batch %d reversed).\n", executionCount, maxBatch)

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
