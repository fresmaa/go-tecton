package driver

import (
	"context"
	"time"
)

// MigrationRecord represents a row in the tecton_migrations table.
// This is used to track the state of applied migrations.
type MigrationRecord struct {
	Version         int64
	Name            string
	Status          string // e.g., "applied", "failed" (useful for MySQL later)
	ExecutionTimeMs int64
	Batch           int
	AppliedAt       time.Time
}

// MigrationPayload represents a parsed migration file ready for execution.
// It acts as the contract payload between the Parser and the Driver.
type MigrationPayload struct {
	Version int64
	Name    string
	// Statements contains the SQL queries split by the Parser.
	// Very useful for MySQL later which needs statement-by-statement execution.
	Statements []string
	// RawSQL contains the original full SQL string.
	// Useful for PostgreSQL where we can execute the whole block in one transaction.
	RawSQL string
	// Batch is the number of statements to execute at once.
	// This is used for bulk operations like inserting many rows.
	Batch int
}

// DatabaseDriver is the contract that every supported DBMS must implement.
type DatabaseDriver interface {
	// Initialize establishes the connection and ensures the metadata table exists.
	Initialize(ctx context.Context, connectionURL string) error

	// Apply executes an "up" migration script.
	// For PostgreSQL: It will wrap RawSQL in a transaction.
	// For MySQL: It will iterate over Statements one by one.
	Apply(ctx context.Context, payload MigrationPayload) error

	// Revert executes a "down" migration script for intentional rollbacks.
	Revert(ctx context.Context, payload MigrationPayload) error

	// GetAppliedMigrations fetches the history of successful migrations.
	GetAppliedMigrations(ctx context.Context) ([]MigrationRecord, error)

	// Close terminates the database connection gracefully.
	Close(ctx context.Context) error

	// DryRun executes a "dry run" migration script.
	// It executes the migration but does not persist the changes.
	DryRun(ctx context.Context, payload MigrationPayload) error

	// ExecuteSeed executes a single SQL script (typically for seeding).
	// Unlike Apply, it does not check or update the migrations table.
	ExecuteSeed(ctx context.Context, rawSQL string) error

	// DropAll drops all tables in the database.
	// This is used by the Reset command.
	DropAll(ctx context.Context) error
}
