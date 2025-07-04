/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/parser"
	"github.com/codecrafted007/autozap/internal/trigger"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [workflow_file]", // This is the subcommand 'run'
	Short: "Executes an AutoZap workflow from a YAML file.",
	Long:  `The run command takes a YAML workflow file as input and executes the defined automation`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowFile := args[0]
		logger.L().Infof("Attempting to run workflow from file: %s", workflowFile)
		logger.L().Infow("Workflow processing initiated",
			"workflow_file", workflowFile,
			"source_type", "cli_argument",
			"status", "starting",
		)
		wf, err := parser.ParseWorkflowFile(workflowFile)
		if err != nil {
			logger.L().Errorf("Failed to parse workflow file: %s, error: %v", workflowFile, err)
			return // Exit the run function on error
		}

		logger.L().Infow("Successfully loaded workflow",
			"workflow_name", wf.Name,
			"workflow_description", wf.Description,
			"trigger_type", wf.Trigger.Type,
			"actions_count", len(wf.Actions),
			"trigger_schedule", wf.Trigger.Schedule,
		)

		for i, action := range wf.Actions {
			logger.L().Infow("Parsed action",
				"action_index", i,
				"action_type", action.Type.String(),
				"action_name", action.Name,
				"action_command", action.Command)
		}
		// Start the cron trigger
		switch wf.Trigger.Type {
		case workflow.TriggerTypeCron:
			if err := trigger.StartCronTrigger(wf); err != nil {
				logger.L().Errorw("Failed to start cron trigger",
					"workflow_name", wf.Name,
					"error", err,
				)
				return // Exit the run function on error
			}
		case workflow.TriggerTypeFileWatch:
			if err := trigger.StartFileWatchTrigger(wf); err != nil {
				logger.L().Errorw("Failed to start file watch trigger",
					"workflow_name", wf.Name,
					"error", err,
				)
				return // Exit the run function on error
			}
		default:
			logger.L().Errorf("Unsupported trigger type '%s' for workflow '%s'. Only 'cron' is supported at this time.", wf.Trigger.Type, wf.Name)
			return // Exit the run function on unsupported trigger type
		}

		logger.L().Info("Autozap is now running in background. Press Ctrl+C to stop.")
		select {} // Block forever to keep the cron trigger running

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
