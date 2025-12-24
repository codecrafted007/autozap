package action

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/metrics"
	"github.com/codecrafted007/autozap/internal/retry"
	"github.com/codecrafted007/autozap/internal/workflow"
)

func ExecuteBashAction(action *workflow.Action, workflowName ...string) error {
	if action.Type != workflow.ActionTypeBash {
		return fmt.Errorf("invalid action type for ExecuteBashAction: expected %s, got %s", workflow.ActionTypeBash, action.Type)
	}
	if action.Command == "" {
		return fmt.Errorf("bash action command cannot be empty")
	}

	// Track total execution time (including retries)
	totalStartTime := time.Now()

	// Execute with retry logic
	err := retry.ExecuteWithRetry(action.Name, action.Retry, func() error {
		return executeBashActionOnce(action, workflowName...)
	})

	totalDuration := time.Since(totalStartTime)

	// Record metrics if workflow name is provided
	if len(workflowName) > 0 && workflowName[0] != "" {
		status := "success"
		if err != nil {
			status = "failed"
		}
		metrics.RecordActionExecution(workflowName[0], action.Name, string(workflow.ActionTypeBash), status, totalDuration)
	}

	return err
}

// executeBashActionOnce executes a bash action once without retry logic
func executeBashActionOnce(action *workflow.Action, workflowName ...string) error {
	logger.L().Infow("Executing Bash Action",
		"action_name", action.Name,
		"command", action.Command,
	)

	cmd := exec.Command("bash", "-c", action.Command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	logFields := []interface{}{
		"action_name", action.Name,
		"command", action.Command,
		"stdout", stdout.String(),
		"stderr", stderr.String(),
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			logFields = append(logFields, "exit_code", exitError.ExitCode())
			logger.L().Errorw("Bash Action failed", logFields...)
			return fmt.Errorf("bash action %s failed with exit code %d: %w", action.Name, exitError.ExitCode(), exitError)
		} else {
			logger.L().Errorw("Bash Action failed", logFields...)
			return fmt.Errorf("bash action %s failed to execute:  %v", action.Name, err)
		}
	}
	logger.L().Infow("Bash Action completed successfully", logFields...)
	return nil
}
