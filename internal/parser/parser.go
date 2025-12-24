package parser

import (
	"fmt"
	"os"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
	"gopkg.in/yaml.v3"
)

func ParseWorkflowFile(filePath string) (*workflow.Workflow, error) {
	// This function will read the YAML file at filePath,
	// parse it into a workflow.Workflow struct, and return it.
	// For now, we will just return nil and nil to avoid compilation errors.

	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("workflow file not found: %s", filePath)
	}

	yamFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %s %w", filePath, err)
	}

	var wf workflow.Workflow

	if err := yaml.Unmarshal(yamFile, &wf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow YAML file: %s %w", filePath, err)
	}

	if err := validateWorkflow(&wf); err != nil {
		return nil, fmt.Errorf("workflow validation failed for file %s: %w", filePath, err)
	}
	logger.L().Infof("Successfully parsed workflow file: %s", filePath)
	return &wf, nil
}

func validateWorkflow(wf *workflow.Workflow) error {
	if wf.Name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}

	if len(wf.Actions) == 0 {
		return fmt.Errorf("workflow must define at least one action")
	}

	switch wf.Trigger.Type {
	case workflow.TriggerTypeCron:
		if wf.Trigger.Schedule == "" {
			return fmt.Errorf("cron trigger requires a 'schedule'")
		}

		if wf.Trigger.Path != "" || len(wf.Trigger.Events) > 0 {
			logger.L().Warnf("cron trigger has unexpected 'path' or 'event' these will be ignored.")
		}
	case workflow.TriggerTypeFileWatch:
		if wf.Trigger.Path == "" {
			return fmt.Errorf("filewatch trigger requires a 'path'")
		}

		if len(wf.Trigger.Events) == 0 {
			return fmt.Errorf("filewatch trigger requires at least one 'event'")
		}

		// Validate event names at parse time
		if err := validateFileWatchEvents(wf.Trigger.Events); err != nil {
			return fmt.Errorf("filewatch trigger validation failed: %w", err)
		}

		if wf.Trigger.Schedule != "" {
			logger.L().Warnf("Filewatch trigger has unexpected 'schedule' field; it will be ignored.")
		}
	default:
		return fmt.Errorf("unsupported trigger type: %s", wf.Trigger.Type)

	}

	// Validate Actions
	for i, action := range wf.Actions {
		if action.Name == "" {
			return fmt.Errorf("action at index %d must have a 'name' ", i)
		}

		switch action.Type {
		case workflow.ActionTypeBash:
			if action.Command == "" {
				return fmt.Errorf("bash action %s at index %d must have a 'command'", action.Name, i)
			}
			//Warn if HTTP/Custom fields are present
			if action.URL != "" || action.Method != "" || len(action.Headers) > 0 || action.Body != "" {
				logger.L().Warnf("Bash action %s at index %d has unexpected HTTP fields; they will be ignored.", action.Name, i)
			}
		case workflow.ActionTypeHTTP:
			if action.URL == "" {
				return fmt.Errorf("HTTP action %s at index %d must have a 'url'", action.Name, i)
			}
			if action.Method == "" {
				return fmt.Errorf("HTTP action %s at index %d must have a 'method'", action.Name, i)
			}

			// ExpectStatus validation is handled at runtime with proper type conversion
			// We allow int, float64, or []interface{} from YAML unmarshaling

			// Warn if Bash/Custom fields are present
			if action.Command != "" || action.FunctionName != "" || action.Arguments != nil {
				return fmt.Errorf("HTTP action %s at index %d has unexpected Bash or Custom fields; they will be ignored", action.Name, i)
			}
		case workflow.ActionTypeCustom:
			if action.FunctionName == "" {
				return fmt.Errorf("custom action %s at index %d must have a 'functionName'", action.Name, i)
			}
			if action.Command != "" || action.URL != "" || action.Method != "" || len(action.Headers) > 0 || action.Body != "" {
				logger.L().Warnf("Custom action %s at index %d has unexpected Bash or HTTP fields; they will be ignored.", action.Name, i)
			}
		default:
			return fmt.Errorf("action %s at index %d has unsupported type: %s", action.Name, i, action.Type)
		}
	}

	return nil
}

// validateFileWatchEvents checks if all event names are valid
func validateFileWatchEvents(events []string) error {
	validEvents := map[string]bool{
		"create": true,
		"write":  true,
		"remove": true,
		"rename": true,
		"chmod":  true,
	}

	for _, event := range events {
		if !validEvents[event] {
			return fmt.Errorf("invalid filewatch event: '%s'. Valid events are: create, write, remove, rename, chmod", event)
		}
	}

	return nil
}
