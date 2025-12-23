package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/parser"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [workflow_files...]",
	Short: "Validate workflow files without executing them",
	Long: `Validate checks workflow YAML files for syntax errors and configuration issues
without executing them. This is useful for CI/CD pipelines and pre-deployment checks.

The command validates:
- YAML syntax
- Required fields (name, trigger, actions)
- Trigger type and configuration
- Action types and required fields
- Cron schedule syntax (if using cron trigger)

Examples:
  autozap validate ./workflows/backup.yaml
  autozap validate ./workflows/*.yaml
  autozap validate ./workflows/backup.yaml ./workflows/monitor.yaml
  autozap validate ./workflows/*.yaml --strict`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strict, _ := cmd.Flags().GetBool("strict")

		// Expand glob patterns
		var workflowFiles []string
		for _, pattern := range args {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				logger.L().Errorw("Invalid file pattern",
					"pattern", pattern,
					"error", err,
				)
				continue
			}
			workflowFiles = append(workflowFiles, matches...)
		}

		if len(workflowFiles) == 0 {
			logger.L().Error("No workflow files found to validate")
			os.Exit(1)
		}

		validCount := 0
		invalidCount := 0
		warnings := 0

		fmt.Println("🔍 Validating workflow files...\n")

		for _, file := range workflowFiles {
			fmt.Printf("Validating: %s\n", file)

			// Check if file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				fmt.Printf("  ✗ File does not exist\n\n")
				invalidCount++
				continue
			}

			// Parse and validate workflow
			wf, err := parser.ParseWorkflowFile(file)
			if err != nil {
				fmt.Printf("  ✗ Validation failed: %v\n\n", err)
				invalidCount++
				continue
			}

			// Print validation details
			fmt.Printf("  ✓ YAML syntax valid\n")
			fmt.Printf("  ✓ Workflow name: '%s'\n", wf.Name)
			fmt.Printf("  ✓ Trigger type: '%s'\n", wf.Trigger.Type)

			// Validate trigger configuration
			switch wf.Trigger.Type.String() {
			case "cron":
				if wf.Trigger.Schedule != "" {
					fmt.Printf("  ✓ Cron schedule: '%s'\n", wf.Trigger.Schedule)
				}
				// Warn if filewatch fields are present
				if wf.Trigger.Path != "" || len(wf.Trigger.Events) > 0 {
					fmt.Printf("  ⚠ Warning: filewatch fields present in cron trigger (will be ignored)\n")
					warnings++
					if strict {
						invalidCount++
						fmt.Printf("  ✗ Strict mode: warnings treated as errors\n\n")
						continue
					}
				}
			case "filewatch":
				if wf.Trigger.Path != "" {
					fmt.Printf("  ✓ Watch path: '%s'\n", wf.Trigger.Path)
				}
				if len(wf.Trigger.Events) > 0 {
					fmt.Printf("  ✓ Events: %v\n", wf.Trigger.Events)
				}
				// Warn if cron schedule is present
				if wf.Trigger.Schedule != "" {
					fmt.Printf("  ⚠ Warning: schedule field present in filewatch trigger (will be ignored)\n")
					warnings++
					if strict {
						invalidCount++
						fmt.Printf("  ✗ Strict mode: warnings treated as errors\n\n")
						continue
					}
				}
			}

			// Validate actions
			fmt.Printf("  ✓ Actions count: %d\n", len(wf.Actions))
			for i, action := range wf.Actions {
				actionType := action.Type.String()
				fmt.Printf("    [%d] %s (%s)\n", i+1, action.Name, actionType)

				// Validate action-specific fields
				switch actionType {
				case "bash":
					if action.Command == "" {
						fmt.Printf("      ✗ Missing required field: command\n")
						invalidCount++
						fmt.Printf("\n")
						continue
					}
				case "http":
					if action.URL == "" {
						fmt.Printf("      ✗ Missing required field: url\n")
						invalidCount++
						fmt.Printf("\n")
						continue
					}
					if action.Method == "" {
						fmt.Printf("      ✗ Missing required field: method\n")
						invalidCount++
						fmt.Printf("\n")
						continue
					}
				case "custom":
					if action.FunctionName == "" {
						fmt.Printf("      ✗ Missing required field: function_name\n")
						invalidCount++
						fmt.Printf("\n")
						continue
					}
				}
			}

			fmt.Printf("  ✓ Ready to deploy\n\n")
			validCount++
		}

		// Print summary
		fmt.Println("─────────────────────────────────────")
		fmt.Printf("Validation Summary:\n")
		fmt.Printf("  Total files: %d\n", len(workflowFiles))
		fmt.Printf("  ✓ Valid: %d\n", validCount)
		fmt.Printf("  ✗ Invalid: %d\n", invalidCount)
		if warnings > 0 {
			fmt.Printf("  ⚠ Warnings: %d\n", warnings)
		}
		fmt.Println("─────────────────────────────────────")

		// Exit with appropriate code
		if invalidCount > 0 {
			fmt.Println("\n❌ Validation failed")
			os.Exit(1)
		} else if warnings > 0 && strict {
			fmt.Println("\n❌ Validation failed (strict mode)")
			os.Exit(1)
		} else {
			fmt.Println("\n✅ All workflows valid")
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Add flags
	validateCmd.Flags().Bool("strict", false, "Treat warnings as errors")
}
