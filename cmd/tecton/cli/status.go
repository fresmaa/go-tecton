package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fresmaa/go-tecton/internal/driver/postgres"
	"github.com/fresmaa/go-tecton/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all migrations",
	Long:  `Displays a visual table comparing local migration files against applied migrations in the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if databaseURL == "" {
			return fmt.Errorf("database connection URL is required")
		}

		ctx := context.Background()
		pgDriver := postgres.New()
		if err := pgDriver.Initialize(ctx, databaseURL); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer pgDriver.Close(ctx)

		// 1. Get applied migrations from database
		appliedRecords, err := pgDriver.GetAppliedMigrations(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch applied migrations: %w", err)
		}

		// Create a map for quick DB lookup
		dbMap := make(map[string]string) // Version -> AppliedAt
		for _, rec := range appliedRecords {
			dbMap[fmt.Sprintf("%d", rec.Version)] = rec.AppliedAt.Format("2006-01-02 15:04:05")
		}

		// 2. Get local migration files
		files, err := os.ReadDir(migrationDir)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to read migration directory: %w", err)
		}

		localVersions := make(map[string]string) // Version -> Name
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
				parts := strings.SplitN(f.Name(), "_", 2)
				if len(parts) == 2 {
					version := parts[0]
					name := strings.TrimSuffix(parts[1], ".up.sql")
					localVersions[version] = name
				}
			}
		}

		// 3. Reconcile both lists
		var allVersions []string
		for v := range dbMap {
			allVersions = append(allVersions, v)
		}
		for v := range localVersions {
			if _, exists := dbMap[v]; !exists {
				allVersions = append(allVersions, v)
			}
		}

		// Sort versions chronologically
		sort.Strings(allVersions)

		// 4. Prepare data for UI
		var tableData [][]string
		for _, version := range allVersions {
			name := localVersions[version]
			appliedAt, inDB := dbMap[version]
			_, inLocal := localVersions[version]

			if name == "" {
				name = "(Unknown / File Deleted)"
			}

			status := ""
			if inDB && inLocal {
				status = "Applied"
			} else if !inDB && inLocal {
				status = "Pending"
				appliedAt = "Not applied yet"
			} else if inDB && !inLocal {
				status = "Missing"
			}

			tableData = append(tableData, []string{status, version, name, appliedAt})
		}

		// 5. Render
		ui.PrintStatusTable(tableData)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVarP(&migrationDir, "dir", "m", "migrations", "Path to migration files")
}
