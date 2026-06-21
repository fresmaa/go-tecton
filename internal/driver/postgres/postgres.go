package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fresmaa/go-tecton/internal/driver"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx as the standard sql driver
)

// PostgresDriver implements the driver.DatabaseDriver interface for PostgreSQL.
type PostgresDriver struct {
	db *sql.DB
}

// Compile-time check to ensure PostgresDriver implements driver.DatabaseDriver
var _ driver.DatabaseDriver = (*PostgresDriver)(nil)

// New creates a new instance of PostgresDriver.
func New() *PostgresDriver {
	return &PostgresDriver{}
}

// Initialize connects to the database and ensures the migration metadata table exists.
func (p *PostgresDriver) Initialize(ctx context.Context, connectionURL string) error {
	db, err := sql.Open("pgx", connectionURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db

	// Create the tecton_migrations table if it doesn't exist
	schemaQuery := `
		CREATE TABLE IF NOT EXISTS tecton_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			execution_time_ms BIGINT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err = p.db.ExecContext(ctx, schemaQuery)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	return nil
}

// Apply executes the UP migration in a single transaction to prevent dirty states.
func (p *PostgresDriver) Apply(ctx context.Context, payload driver.MigrationPayload) error {
	// 1. Begin Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Safely rollback if anything fails or panics.
	// If tx is already committed, Rollback() is a no-op (does nothing).
	defer tx.Rollback()

	startTime := time.Now()

	// 2. Execute the entire SQL payload.
	// We use RawSQL because PostgreSQL handles multi-statement blocks inside a transaction perfectly.
	_, err = tx.ExecContext(ctx, payload.RawSQL)
	if err != nil {
		// In the future, we will extract the exact line number from this error for the UI.
		return fmt.Errorf("migration execution failed: %w", err)
	}

	// 3. Record the migration in the metadata table
	executionTime := time.Since(startTime).Milliseconds()
	insertQuery := `
		INSERT INTO tecton_migrations (version, name, status, execution_time_ms) 
		VALUES ($1, $2, 'applied', $3)
	`
	_, err = tx.ExecContext(ctx, insertQuery, payload.Version, payload.Name, executionTime)
	if err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}

	// 4. Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close gracefully closes the database connection.
func (p *PostgresDriver) Close(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetAppliedMigrations fetches the history of successful migrations from the database.
// This is used by the Engine to determine which files in the directory are pending.
func (p *PostgresDriver) GetAppliedMigrations(ctx context.Context) ([]driver.MigrationRecord, error) {
	query := `
		SELECT version, name, status, execution_time_ms, applied_at 
		FROM tecton_migrations 
		ORDER BY version ASC
	`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var records []driver.MigrationRecord
	for rows.Next() {
		var record driver.MigrationRecord
		if err := rows.Scan(&record.Version, &record.Name, &record.Status, &record.ExecutionTimeMs, &record.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over migration records: %w", err)
	}

	return records, nil
}

// Revert executes the DOWN migration in a single transaction.
// It removes the schema changes and deletes the metadata record simultaneously.
func (p *PostgresDriver) Revert(ctx context.Context, payload driver.MigrationPayload) error {
	// 1. Begin Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Safely rollback if anything fails or panics.
	defer tx.Rollback()

	// 2. Execute the entire SQL payload (the .down.sql content)
	_, err = tx.ExecContext(ctx, payload.RawSQL)
	if err != nil {
		return fmt.Errorf("revert execution failed: %w", err)
	}

	// 3. Remove the migration record from the metadata table
	deleteQuery := `DELETE FROM tecton_migrations WHERE version = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, payload.Version)
	if err != nil {
		return fmt.Errorf("failed to delete metadata record: %w", err)
	}

	// 4. Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit revert transaction: %w", err)
	}

	return nil
}

// DryRun executes the migration inside a transaction but forces a rollback at the end.
// This validates the SQL syntax and constraints without altering the database state.
func (p *PostgresDriver) DryRun(ctx context.Context, payload driver.MigrationPayload) error {
	// 1. Begin Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for dry-run: %w", err)
	}

	// 2. FORCE ROLLBACK: This guarantees no changes are saved, regardless of success or failure.
	defer tx.Rollback()

	// 3. Execute the SQL payload
	_, err = tx.ExecContext(ctx, payload.RawSQL)
	if err != nil {
		return fmt.Errorf("dry-run syntax/logic error: %w", err)
	}

	// No commit here! The defer block will safely discard the transaction.
	return nil
}

// ExecuteSeed runs raw SQL within a transaction without tracking its state.
// Ideal for idempotent data seeding operations.
func (p *PostgresDriver) ExecuteSeed(ctx context.Context, rawSQL string) error {
	// 1. Begin Transaction for safety
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Safely rollback if anything panics or errors out before Commit
	defer tx.Rollback()

	// 2. Execute the raw SQL
	_, err = tx.ExecContext(ctx, rawSQL)
	if err != nil {
		return fmt.Errorf("failed to execute seeder query: %w", err)
	}

	// 3. Commit the transaction
	return tx.Commit()
}

// DropAll completely wipes the database schema. USE WITH EXTREME CAUTION.
func (p *PostgresDriver) DropAll(ctx context.Context) error {
	// 1. Drop the public schema and everything inside it (CASCADE)
	// 2. Recreate the clean public schema
	// 3. Restore default permissions
	query := `
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO public;

		-- 4. Recreate the tracking tables matching the EXACT original schema
		CREATE TABLE IF NOT EXISTS tecton_migrations (
			version BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			execution_time_ms BIGINT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop database schema: %w", err)
	}

	return nil
}
