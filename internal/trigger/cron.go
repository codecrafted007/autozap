package trigger

import (
	"fmt"
	"time"

	"github.com/codecrafted007/autozap/internal/action"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/metrics"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/robfig/cron/v3"
)

func StartCronTrigger(wf *workflow.Workflow) error {
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

		for i, act := range wf.Actions {
			switch act.Type {
			case workflow.ActionTypeBash:
				logger.L().Infow("Attempting to execute Bash Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"command", act.Command)
				if err := action.ExecuteBashAction(&act, wf.Name); err != nil {
					logger.L().Errorw("Failed to execute Bash Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", err)
					workflowStatus = "failed"
				}
			case workflow.ActionTypeHTTP:
				logger.L().Infow("Attempting to execute HTTP Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"url", act.URL,
					"method", act.Method)
				if err := action.ExecuteHttpAction(&act, wf.Name); err != nil {
					logger.L().Errorw("Failed to execute Http Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", err)
					workflowStatus = "failed"
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
			}
		}

		// Record workflow execution metrics
		workflowDuration := time.Since(workflowStartTime)
		metrics.RecordWorkflowExecution(wf.Name, workflowStatus, workflowDuration)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job for workflow '%s': %w", wf.Name, err)
	}
	logger.L().Infof("Cron Job %s scheduled for workflow '%s' with entry ID %d",
		wf.Trigger.Schedule, wf.Name, entryId)
	c.Start()
	logger.L().Infow("Cron Trigger started for workflow",
		"workflow_name", wf.Name,
		"trigger_schedule", wf.Trigger.Schedule,
		"entry_id", entryId)

	// Register workflow info metric
	metrics.RegisterWorkflow(wf.Name, string(workflow.TriggerTypeCron), wf.Trigger.Schedule)

	return nil
}
