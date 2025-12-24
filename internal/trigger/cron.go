package trigger

import (
	"context"
	"fmt"
	"time"

	"github.com/codecrafted007/autozap/internal/action"
	"github.com/codecrafted007/autozap/internal/database"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/metrics"
	"github.com/codecrafted007/autozap/internal/server"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/robfig/cron/v3"
)

// Helper functions for database operations
func startWorkflowExecutionInDB(workflowName, triggerType string) (int64, error) {
	return database.StartWorkflowExecution(workflowName, triggerType)
}

func completeWorkflowExecutionInDB(id int64, status string, errorMsg *string, duration time.Duration) error {
	return database.CompleteWorkflowExecution(id, status, errorMsg, duration)
}

func StartCronTrigger(ctx context.Context, wf *workflow.Workflow) error {
	// Register workflow in the registry
	server.GetRegistry().RegisterWorkflow(wf)

	c := cron.New()

	entryId, err := c.AddFunc(wf.Trigger.Schedule, func() {
		// Record trigger fire
		metrics.RecordTriggerFire(wf.Name, string(workflow.TriggerTypeCron))

		logger.L().Infow("Cron Trigger fired for workflow",
			"workflow_name", wf.Name,
			"trigger_schedule", wf.Trigger.Schedule,
			"timestamp", time.Now().Format(time.RFC3339))

		// Track workflow execution time
		workflowStartTime := time.Now()
		workflowStatus := "success"
		var workflowError *string

		// Start workflow execution in database
		workflowExecID, err := startWorkflowExecutionInDB(wf.Name, string(workflow.TriggerTypeCron))
		if err != nil {
			logger.L().Errorw("Failed to start workflow execution in database",
				"workflow_name", wf.Name,
				"error", err)
		}

		for i, act := range wf.Actions {
			var actionError error
			switch act.Type {
			case workflow.ActionTypeBash:
				logger.L().Infow("Attempting to execute Bash Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"command", act.Command)
				actionError = action.ExecuteBashAction(&act, wf.Name)
				if actionError != nil {
					logger.L().Errorw("Failed to execute Bash Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", actionError)
					workflowStatus = "failed"
					errMsg := actionError.Error()
					workflowError = &errMsg
				}
			case workflow.ActionTypeHTTP:
				logger.L().Infow("Attempting to execute HTTP Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"url", act.URL,
					"method", act.Method)
				actionError = action.ExecuteHttpAction(&act, wf.Name)
				if actionError != nil {
					logger.L().Errorw("Failed to execute Http Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", actionError)
					workflowStatus = "failed"
					errMsg := actionError.Error()
					workflowError = &errMsg
				}
			case workflow.ActionTypeCustom:
				logger.L().Infow("Attempting to execute Custom Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"action_type", act.Type.String())
				// TODO: Implement Custom action execution
			default:
				logger.L().Errorw("Unknown Action Type",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"action_type", act.Type.String(),
					"error", "unsupported action type")
				workflowStatus = "failed"
				errMsg := "unsupported action type: " + act.Type.String()
				workflowError = &errMsg
			}
		}

		// Record workflow execution metrics
		workflowDuration := time.Since(workflowStartTime)
		metrics.RecordWorkflowExecution(wf.Name, workflowStatus, workflowDuration)

		// Complete workflow execution in database
		if workflowExecID > 0 {
			if err := completeWorkflowExecutionInDB(workflowExecID, workflowStatus, workflowError, workflowDuration); err != nil {
				logger.L().Errorw("Failed to complete workflow execution in database",
					"workflow_name", wf.Name,
					"workflow_exec_id", workflowExecID,
					"error", err)
			}
		}

		// Update registry with execution stats
		errorMsg := ""
		if workflowError != nil {
			errorMsg = *workflowError
		}
		server.GetRegistry().UpdateExecutionStats(wf.Name, workflowStatus == "success", errorMsg)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job for workflow '%s': %w", wf.Name, err)
	}
	logger.L().Infof("Cron Job %s scheduled for workflow '%s' with entry ID %d",
		wf.Trigger.Schedule, wf.Name, entryId)
	c.Start()

	// Get next execution time
	entry := c.Entry(entryId)
	nextRun := entry.Next
	server.GetRegistry().UpdateNextExecution(wf.Name, nextRun)

	logger.L().Infow("Cron Trigger started for workflow",
		"workflow_name", wf.Name,
		"trigger_schedule", wf.Trigger.Schedule,
		"entry_id", entryId,
		"next_run", nextRun)

	// Register workflow info metric
	metrics.RegisterWorkflow(wf.Name, string(workflow.TriggerTypeCron), wf.Trigger.Schedule)

	// Update next execution time after each run
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.L().Infow("Stopping next execution updater for workflow",
					"workflow_name", wf.Name,
					"reason", "context cancelled")
				return
			case <-ticker.C:
				entry := c.Entry(entryId)
				if !entry.Next.IsZero() {
					server.GetRegistry().UpdateNextExecution(wf.Name, entry.Next)
				}
			}
		}
	}()

	// Watch for context cancellation and stop the cron scheduler
	go func() {
		<-ctx.Done()
		logger.L().Infow("Stopping cron trigger for workflow",
			"workflow_name", wf.Name,
			"trigger_schedule", wf.Trigger.Schedule,
			"reason", "context cancelled")

		// Stop the cron scheduler
		cronCtx := c.Stop()
		<-cronCtx.Done()

		// Unregister workflow from registry
		server.GetRegistry().UnregisterWorkflow(wf.Name)

		logger.L().Infow("Cron trigger stopped successfully",
			"workflow_name", wf.Name)
	}()

	return nil
}
