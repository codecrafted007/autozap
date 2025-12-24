package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/codecrafted007/autozap/internal/database"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/spf13/cobra"
)

var failuresCmd = &cobra.Command{
	Use:   "failures",
	Short: "Show recent failed workflow executions",
	Long:  `Display recently failed workflow executions with error details.`,
	Run: func(cmd *cobra.Command, args []string) {
		hours, _ := cmd.Flags().GetInt("hours")
		limit, _ := cmd.Flags().GetInt("limit")

		// Initialize database
		dbPath, _ := cmd.Flags().GetString("db")
		if err := database.InitDB(dbPath); err != nil {
			logger.L().Errorw("Failed to initialize database", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
			return
		}
		defer database.CloseDB()

		since := time.Now().Add(-time.Duration(hours) * time.Hour)
		failures, err := database.GetFailedExecutions(since, limit)
		if err != nil {
			logger.L().Errorw("Failed to get failed executions", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to get failed executions: %v\n", err)
			return
		}

		if len(failures) == 0 {
			fmt.Printf("✓ No failures found in the last %d hours.\n", hours)
			return
		}

		fmt.Printf("\n✗ Failed Executions (Last %d hours)\n\n", hours)

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKFLOW\tTRIGGER\tSTARTED\tERROR")
		fmt.Fprintln(w, "---\t--------\t-------\t-------\t-----")

		for _, exec := range failures {
			errorMsg := "unknown error"
			if exec.Error != nil {
				errorMsg = truncateFailure(*exec.Error, 80)
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				exec.ID,
				exec.WorkflowName,
				exec.TriggerType,
				exec.StartedAt.Format("2006-01-02 15:04:05"),
				errorMsg,
			)
		}
		w.Flush()
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(failuresCmd)

	failuresCmd.Flags().Int("hours", 24, "Show failures from last N hours")
	failuresCmd.Flags().Int("limit", 50, "Maximum number of failures to show")
	failuresCmd.Flags().String("db", "./data/autozap.db", "Database file path")
}

func truncateFailure(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
