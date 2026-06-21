package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending database migrations",
	Long: `Applies all pending up migrations to the database. 
If an error occurs in a transactional database like PostgreSQL, 
it will automatically rollback to prevent a dirty state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Initialize the engine.Migrator and driver.Postgres
		// TODO: Execute the migration process
		fmt.Println("Migration command triggered with DB:", databaseURL)
		return nil
	},
}

func init() {
	// Register migrateCmd as a child of the root command
	rootCmd.AddCommand(migrateCmd)

	// Add flags specific to the migrate command
	migrateCmd.Flags().StringP("dir", "p", "migrations", "Directory containing migration files")
}
