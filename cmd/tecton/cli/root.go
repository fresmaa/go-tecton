package cli

import (
	"github.com/spf13/cobra"
)

// Global flags configuration
var databaseURL string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tecton",
	Short: "A robust database migration and seeding tool",
	Long: `go-tecton ensures safe database migrations by preventing dirty states,
providing visual error stack traces, and offering built-in data seeding tools.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Define global flags that apply to all subcommands
	rootCmd.PersistentFlags().StringVarP(&databaseURL, "db", "d", "", "Database connection URL (e.g., postgres://user:pass@localhost:5432/dbname)")

	// Ensure the database URL is provided for commands that need it
	// (We will handle specific validation in subcommands like 'migrate' or 'seed')
}
