package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codecrafted007/autozap/internal/database"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show workflow execution history",
	Long:  `Display the execution history of workflows stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowName, _ := cmd.Flags().GetString("workflow")
		limit, _ := cmd.Flags().GetInt("limit")

		// Initialize database
		dbPath, _ := cmd.Flags().GetString("db")
		if err := database.InitDB(dbPath); err != nil {
			logger.L().Errorw("Failed to initialize database", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
			return
		}
		defer database.CloseDB()

		var executions []database.WorkflowExecution
		var err error

		if workflowName != "" {
			executions, err = database.GetWorkflowHistory(workflowName, limit)
		} else {
			executions, err = database.GetAllWorkflowHistory(limit)
		}

		if err != nil {
			logger.L().Errorw("Failed to get workflow history", "error", err)
			fmt.Fprintf(os.Stderr, "Error: Failed to get workflow history: %v\n", err)
			return
		}

		if len(executions) == 0 {
			fmt.Println("No execution history found.")
			return
		}

		// Print table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKFLOW\tSTATUS\tTRIGGER\tSTARTED\tDURATION\tERROR")
		fmt.Fprintln(w, "---\t--------\t------\t-------\t-------\t--------\t-----")

		for _, exec := range executions {
			duration := "-"
			if exec.DurationMs != nil {
				duration = fmt.Sprintf("%dms", *exec.DurationMs)
			}

			errorMsg := "-"
			if exec.Error != nil {
				errorMsg = truncate(*exec.Error, 50)
			}

			status := exec.Status
			if status == "success" {
				status = "✓ " + status
			} else if status == "failed" {
				status = "✗ " + status
			}

			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
				exec.ID,
				exec.WorkflowName,
				status,
				exec.TriggerType,
				exec.StartedAt.Format("2006-01-02 15:04:05"),
				duration,
				errorMsg,
			)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().String("workflow", "", "Filter by workflow name")
	historyCmd.Flags().Int("limit", 20, "Maximum number of records to show")
	historyCmd.Flags().String("db", "./data/autozap.db", "Database file path")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
