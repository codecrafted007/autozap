package action

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

func ExecuteBashAction(action *workflow.Action) error {
	if action.Type != workflow.ActionTypeBash {
		return fmt.Errorf("Invalid action type for ExecuteBashAction: expected %s, got %s", workflow.ActionTypeBash, action.Type)
	}
	if action.Command == "" {
		return fmt.Errorf("Bash action command cannot be empty")
	}

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
			return fmt.Errorf("bash action %s failed with exit code %d: %w", action.Name, exitError)
		} else {
			logger.L().Errorw("Bash Action failed", logFields...)
			return fmt.Errorf("bash action %s failed to execute:  %v", action.Name, err)
		}
	}
	logger.L().Infow("Bash Action completed successfully", logFields...)
	return nil

}
