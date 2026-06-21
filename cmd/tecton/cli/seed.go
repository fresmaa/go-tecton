package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/engine"
	"github.com/spf13/cobra"
)

var seederDir string

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Insert dummy or initial data into the database",
	Long:  `Scans the specified seeder directory and executes all valid SQL files to populate the database with initial records.`,
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

		// 3. Instantiate the seeder engine and run
		seeder := engine.NewSeeder(pgDriver, fileSys, seederDir)
		if err := seeder.Run(ctx); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)

	// Local flags specific to the seed command
	seedCmd.Flags().StringVarP(&seederDir, "dir", "s", "seeders", "Path to the directory containing seeder files")

	seedCmd.SilenceUsage = true
	seedCmd.SilenceErrors = true
}
