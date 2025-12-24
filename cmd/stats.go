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

var statsCmd = &cobra.Command{
	Use:   "stats [workflow-name]",
	Short: "Show workflow execution statistics",
	Long:  `Display statistics for workflow executions including success rate and average duration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowName := args[0]
		days, _ := cmd.Flags().GetInt("days")

		// Initialize database
		dbPath, _ := cmd.Flags().GetString("db")
		if err := database.InitDB(dbPath); err != nil {
			logger.L().Errorw("Failed to initialize database", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
			return
		}
		defer database.CloseDB()

		since := time.Now().AddDate(0, 0, -days)
		stats, err := database.GetWorkflowStats(workflowName, since)
		if err != nil {
			logger.L().Errorw("Failed to get workflow stats", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to get workflow stats: %v\n", err)
			return
		}

		if stats.TotalExecutions == 0 {
			fmt.Printf("No executions found for workflow '%s' in the last %d days.\n", workflowName, days)
			return
		}

		// Print stats
		fmt.Printf("\n📊 Statistics for workflow: %s (Last %d days)\n\n", workflowName, days)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "METRIC\tVALUE")
		fmt.Fprintln(w, "------\t-----")
		fmt.Fprintf(w, "Total Executions\t%d\n", stats.TotalExecutions)
		fmt.Fprintf(w, "Successful\t%d (✓)\n", stats.SuccessCount)
		fmt.Fprintf(w, "Failed\t%d (✗)\n", stats.FailedCount)
		fmt.Fprintf(w, "Success Rate\t%.2f%%\n", stats.SuccessRate)

		if stats.AvgDurationMs > 0 {
			if stats.AvgDurationMs < 1000 {
				fmt.Fprintf(w, "Avg Duration\t%.2fms\n", stats.AvgDurationMs)
			} else {
				fmt.Fprintf(w, "Avg Duration\t%.2fs\n", stats.AvgDurationMs/1000)
			}
		} else {
			fmt.Fprintln(w, "Avg Duration\t-")
		}

		w.Flush()
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().Int("days", 7, "Number of days to analyze")
	statsCmd.Flags().String("db", "./data/autozap.db", "Database file path")
}
