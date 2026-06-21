package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/engine"
	"github.com/spf13/cobra"
)

var migrationDir string

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Long:  `Scans the migration directory, detects unapplied scripts, and executes them safely inside transactions.`,
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
		// os.DirFS(".") sets the base directory to the current working directory,
		// allowing the engine to navigate into the specific migrationDir folder path.
		fileSys := os.DirFS(".")

		// 3. Instantiate the engine and run migrations
		migrator := engine.New(pgDriver, fileSys, migrationDir)
		if err := migrator.Up(ctx); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	// Register upCmd as a subcommand of rootCmd
	rootCmd.AddCommand(upCmd)

	// Local flags specific to the up command
	upCmd.Flags().StringVarP(&migrationDir, "dir", "p", "migrations", "Path to the directory containing migration files")

	// Silence usage output on error to avoid cluttering the console with unnecessary information
	upCmd.SilenceUsage = true
	upCmd.SilenceErrors = true
}
