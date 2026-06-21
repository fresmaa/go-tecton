package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var createSeedCmd = &cobra.Command{
	Use:   "create-seed [name]",
	Short: "Generate a new seeder file with an idempotent template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		seederName := args[0]

		// 1. Sanitize the seeder name
		seederName = strings.ReplaceAll(seederName, " ", "_")
		seederName = strings.ReplaceAll(seederName, "-", "_")

		// 2. Generate a timestamp prefix for ordered execution
		version := time.Now().Format("20060102150405")
		fileName := fmt.Sprintf("%s_%s.sql", version, seederName)
		filePath := filepath.Join(seederDir, fileName)

		// 3. Ensure the target directory exists
		if err := os.MkdirAll(seederDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create seeder directory: %w", err)
		}

		// 4. Create the file with an educational template
		template := []byte(`-- Write your data seeding statements here
-- BEST PRACTICE: Make it idempotent so it can be safely executed multiple times.
-- Use "ON CONFLICT DO NOTHING" or "ON CONFLICT (...) DO UPDATE".

-- Example:
-- INSERT INTO users (id, username) 
-- VALUES (1, 'superadmin') 
-- ON CONFLICT (id) DO NOTHING;

`)
		if err := os.WriteFile(filePath, template, 0644); err != nil {
			return fmt.Errorf("failed to create seeder file: %w", err)
		}

		fmt.Printf("✅ Created Seeder: %s\n", filePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createSeedCmd)

	// Reuse the seederDir variable defined in seed.go
	createSeedCmd.Flags().StringVarP(&seederDir, "dir", "s", "seeders", "Directory to store seeder files")
}
