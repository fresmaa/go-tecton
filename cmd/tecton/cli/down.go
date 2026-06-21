package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/engine"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last applied migration",
	Long:  `Identifies the most recent successful migration from the database and safely executes its .down.sql pair.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if databaseURL == "" {
			return fmt.Errorf("database connection URL is required, specify with --db or -d")
		}

		ctx := context.Background()

		// 1. Initialize the PostgreSQL driver
		pgDriver := postgres.New()
		if err := pgDriver.Initialize(ctx, databaseURL); err != nil {
			return fmt.Errorf("failed to initialize database driver: %w", err)
		}
		defer pgDriver.Close(ctx)

		// 2. Prepare the local filesystem target
		fileSys := os.DirFS(".")

		// 3. Instantiate the engine and rollback the migration
		migrator := engine.New(pgDriver, fileSys, migrationDir)
		if err := migrator.Down(ctx); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Register downCmd as a subcommand of rootCmd
	rootCmd.AddCommand(downCmd)

	// Local flags specific to the down command
	downCmd.Flags().StringVarP(&migrationDir, "dir", "p", "migrations", "Path to the directory containing migration files")

	// Silence usage output on error to avoid cluttering the console with unnecessary information
	downCmd.SilenceUsage = true
	downCmd.SilenceErrors = true
}
