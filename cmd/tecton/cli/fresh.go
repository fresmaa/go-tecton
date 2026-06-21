package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/engine"
	"github.com/spf13/cobra"
)

var (
	runSeed   bool
	forceWipe bool
)

func askForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

var freshCmd = &cobra.Command{
	Use:   "fresh",
	Short: "Drop all tables and re-run all migrations",
	Long:  `Wipes the entire database schema clean, then runs all available migrations from scratch. You can also optionally run seeders afterward.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if databaseURL == "" {
			return fmt.Errorf("database connection URL is required")
		}

		if !forceWipe {
			warningBox := lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#EF4444")).
				Foreground(lipgloss.Color("#EF4444")).
				Padding(1, 0).
				Width(42).
				Align(lipgloss.Center).
				Bold(true).
				Render("APPLICATION IN PRODUCTION?")

			descText := lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3")).Render("You are about to ") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true).Render("DROP ALL TABLES") +
				lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3")).Render(" in your database.")

			fmt.Println("\n" + warningBox)
			fmt.Println(descText)

			promptMsg := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Bold(true).
				Render("\nAre you sure you want to execute this destructive operation? [y/N]: ")

			if !askForConfirmation(promptMsg) {
				cancelMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("🛑 Operation cancelled.")
				fmt.Println(cancelMsg)
				return nil
			}
		}

		ctx := context.Background()
		pgDriver := postgres.New()
		if err := pgDriver.Initialize(ctx, databaseURL); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer pgDriver.Close(ctx)

		fmt.Println("⚠️  WARNING: Wiping database schema...")
		if err := pgDriver.DropAll(ctx); err != nil {
			return fmt.Errorf("failed to wipe database: %w", err)
		}
		fmt.Println("✅ Database wiped successfully.")

		// Re-run migrations
		fmt.Println("\n🚀 Starting fresh migrations...")
		fileSys := os.DirFS(".")
		migrator := engine.New(pgDriver, fileSys, migrationDir)
		if err := migrator.Up(ctx); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		// Run seeders if flag is provided
		if runSeed {
			fmt.Println("\n🌱 Seeding database...")
			seeder := engine.NewSeeder(pgDriver, fileSys, seederDir)
			if err := seeder.Run(ctx); err != nil {
				return fmt.Errorf("seeding failed: %w", err)
			}
		}

		fmt.Println("\n🎉 Database is fresh and ready to go!")
		return nil
	},
}

func init() {
	// Register freshCmd as a subcommand of rootCmd
	rootCmd.AddCommand(freshCmd)

	// Register flags
	freshCmd.Flags().StringVarP(&migrationDir, "dir", "m", "migrations", "Path to migration files")
	freshCmd.Flags().StringVarP(&seederDir, "seeder-dir", "s", "seeders", "Path to seeder files")
	freshCmd.Flags().BoolVar(&runSeed, "seed", false, "Run seeders after migrations complete")
	freshCmd.Flags().BoolVar(&forceWipe, "force", false, "Force the operation to run without confirmation prompts")

	// Silence usage output on error to avoid cluttering the console with unnecessary information
	freshCmd.SilenceUsage = true
	freshCmd.SilenceErrors = true
}
