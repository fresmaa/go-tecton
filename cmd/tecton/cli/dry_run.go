package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/engine"
	"github.com/spf13/cobra"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Simulate pending migrations without saving changes",
	Long:  `Executes all pending migrations inside a database transaction and forces a rollback at the end. This safely validates SQL syntax and constraints.`,
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

		// 3. Instantiate the engine and run the dry-run validation
		migrator := engine.New(pgDriver, fileSys, migrationDir)
		if err := migrator.DryRun(ctx); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Register dryRunCmd as a subcommand of rootCmd
	rootCmd.AddCommand(dryRunCmd)

	// Local flags specific to the dry-run command
	dryRunCmd.Flags().StringVarP(&migrationDir, "dir", "p", "migrations", "Path to the directory containing migration files")

	// Keep the terminal clean when an error occurs
	dryRunCmd.SilenceUsage = true
	dryRunCmd.SilenceErrors = true
}
