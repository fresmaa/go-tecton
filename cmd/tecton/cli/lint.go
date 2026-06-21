package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/fresmaa/go-tecton/internal/linter"
	"github.com/fresmaa/go-tecton/internal/parser"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Perform static analysis on migration files",
	Long:  `Scans the SQL syntax of your pending and applied migrations to detect dangerous operations like table drops or locking queries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileSys := os.DirFS(".")
		entries, err := fs.ReadDir(fileSys, migrationDir)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		lintEngine := linter.New()
		totalViolations := 0

		fmt.Println("🔎 Scanning migration files for anti-patterns...")
		fmt.Println("--------------------------------------------------")

		for _, entry := range entries {
			if entry.IsDir() || path.Ext(entry.Name()) != ".sql" {
				continue
			}

			filePath := path.Join(migrationDir, entry.Name())
			payload, err := parser.ParseFile(fileSys, filePath)
			if err != nil {
				return fmt.Errorf("failed to parse file %s: %w", entry.Name(), err)
			}

			violations := lintEngine.Analyze(payload.RawSQL)

			if len(violations) > 0 {
				fmt.Printf("📄 File: %s\n", entry.Name())
				for _, v := range violations {
					totalViolations++
					// Color code the output based on level (using basic ANSI codes for simplicity here)
					colorCode := "\033[33m" // Yellow for warning
					if v.Level == linter.LevelError {
						colorCode = "\033[31m" // Red for error
					}
					resetCode := "\033[0m"

					fmt.Printf("   %s[%s] %s: %s%s\n", colorCode, v.Level, v.RuleCode, v.Description, resetCode)
					fmt.Printf("   %sL%d%s: %s\n\n", "\033[90m", v.LineNumber, resetCode, v.LineText)
				}
			}
		}

		if totalViolations == 0 {
			fmt.Println("✅ All clear! No violations found.")
		} else {
			fmt.Printf("⚠️  Found %d violation(s) across your migration files.\n", totalViolations)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.Flags().StringVarP(&migrationDir, "dir", "p", "migrations", "Path to the directory containing migration files")
}
