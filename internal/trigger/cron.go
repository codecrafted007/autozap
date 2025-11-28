package trigger

import (
	"fmt"
	"time"

	"github.com/codecrafted007/autozap/internal/action"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/robfig/cron/v3"
)

func StartCronTrigger(wf *workflow.Workflow) error {
	c := cron.New()

	entryId, err := c.AddFunc(wf.Trigger.Schedule, func() {
		logger.L().Infow("Cron Trigger fired for workflow",
			"workflow_name", wf.Name,
			"trigger_schedule", wf.Trigger.Schedule,
			"timestamp", time.Now().Format(time.RFC3339))

		for i, act := range wf.Actions {
			switch act.Type {
			case workflow.ActionTypeBash:
				logger.L().Infow("Attempting to execute Bash Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"command", act.Command)
				if err := action.ExecuteBashAction(&act); err != nil {
					logger.L().Errorw("Failed to execute Bash Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", err)
				}
			case workflow.ActionTypeHTTP:
				logger.L().Infow("Attempting to execute HTTP Action",
					"workflow_name", wf.Name,
					"action_name", act.Name,
					"action_index", i,
					"url", act.URL,
					"method", act.Method)
				if err := action.ExecuteHttpAction(&act); err != nil {
					logger.L().Errorw("Failed to execute Http Action",
						"workflow_name", wf.Name,
						"action_name", act.Name,
						"action_index", i,
						"error", err)
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
			}
		}
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
	return nil
}
