package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Generate a new pair of migration files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		migrationName := args[0]

		// 1. Sanitize the migration name (replace spaces and dashes with underscores)
		migrationName = strings.ReplaceAll(migrationName, " ", "_")
		migrationName = strings.ReplaceAll(migrationName, "-", "_")

		// 2. Generate timestamp for the version format
		version := time.Now().Format("20060102150405")

		// 3. Prepare the full file paths
		upFilePath := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.up.sql", version, migrationName))
		downFilePath := filepath.Join(migrationDir, fmt.Sprintf("%s_%s.down.sql", version, migrationName))

		// 4. Ensure the target directory exists
		if err := os.MkdirAll(migrationDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create migration directory: %w", err)
		}

		// 5. Create the physical files with default templates
		upTemplate := []byte("-- Write your UP migration statements here\n")
		if err := os.WriteFile(upFilePath, upTemplate, 0644); err != nil {
			return fmt.Errorf("failed to create up migration file: %w", err)
		}

		downTemplate := []byte("-- Write your DOWN (rollback) migration statements here\n")
		if err := os.WriteFile(downFilePath, downTemplate, 0644); err != nil {
			return fmt.Errorf("failed to create down migration file: %w", err)
		}

		fmt.Printf("✅ Created: %s\n", upFilePath)
		fmt.Printf("✅ Created: %s\n", downFilePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&migrationDir, "dir", "p", "migrations", "Directory to store migration files")
}
