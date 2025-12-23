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
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			logger.L().Info("[DRY RUN MODE] No actions will be executed")
		}

		logger.L().Infof("Attempting to run workflow from file: %s", workflowFile)
		logger.L().Infow("Workflow processing initiated",
			"workflow_file", workflowFile,
			"source_type", "cli_argument",
			"status", "starting",
			"dry_run", dryRun,
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

		// In dry-run mode, show what would be executed
		if dryRun {
			logger.L().Infof("[DRY RUN] Would start workflow: %s", wf.Name)
			logger.L().Infof("[DRY RUN] Trigger: %s", wf.Trigger.Type)

			switch wf.Trigger.Type {
			case workflow.TriggerTypeCron:
				logger.L().Infof("[DRY RUN] Schedule: %s", wf.Trigger.Schedule)
			case workflow.TriggerTypeFileWatch:
				logger.L().Infof("[DRY RUN] Watch path: %s", wf.Trigger.Path)
				logger.L().Infof("[DRY RUN] Events: %v", wf.Trigger.Events)
			}

			logger.L().Infof("[DRY RUN] Would execute %d actions:", len(wf.Actions))
			for i, action := range wf.Actions {
				logger.L().Infof("[DRY RUN]   %d. [%s] %s", i+1, action.Type, action.Name)
				switch action.Type {
				case workflow.ActionTypeBash:
					logger.L().Infof("[DRY RUN]      Command: %s", action.Command)
				case workflow.ActionTypeHTTP:
					logger.L().Infof("[DRY RUN]      %s %s", action.Method, action.URL)
				case workflow.ActionTypeCustom:
					logger.L().Infof("[DRY RUN]      Function: %s", action.FunctionName)
				}
			}

			logger.L().Info("[DRY RUN] Dry run complete. No actions were executed.")
			return
		}

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

	// Add flags
	runCmd.Flags().Bool("dry-run", false, "Show what would be executed without running actions")
}
