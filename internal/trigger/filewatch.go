package trigger

import (
	"fmt"
	"time"

	"github.com/codecrafted007/autozap/internal/action"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/fsnotify/fsnotify"
)

func StartFileWatchTrigger(wf *workflow.Workflow) error {

	if wf.Trigger.Type != workflow.TriggerTypeFileWatch {
		err := fmt.Errorf("invalid trigger type for StartFileWatchTrigger: expected '%s', got '%s'", workflow.TriggerTypeFileWatch.String(), wf.Trigger.Type.String())
		logger.L().Errorw("Failed to start file watch trigger due to incorrect type",
			"workflow_name", wf.Name,
			"expected_type", workflow.TriggerTypeFileWatch.String(),
			"received_type", wf.Trigger.Type.String(),
			"error", err, // Log the error object
		)
		return err
	}

	if wf.Trigger.Path == "" {
		logger.L().Errorf("Filewatch trigger requires a file path to watch")
		return fmt.Errorf("file path cannot be empty for filewatch trigger")
	}

	if len(wf.Trigger.Events) == 0 {
		logger.L().Errorf("Filewatch trigger requires at least one event type to watch")
		return fmt.Errorf("at least one event type must be specified for filewatch trigger")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		watcher.Close() // Ensure we close the watcher if it was created
		logger.L().Errorw("Failed to create file watcher",
			"workflow_name", wf.Name,
			"error", err,
		)
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add the path to watch
	err = watcher.Add(wf.Trigger.Path)
	if err != nil {
		watcher.Close() // Ensure watcher is closed on error
		err = fmt.Errorf("failed to add path '%s' to watcher for workflow '%s': %w", wf.Trigger.Path, wf.Name, err)
		logger.L().Errorw("File watch trigger setup error",
			"workflow_name", wf.Name,
			"path", wf.Trigger.Path,
			"error", err,
		)
		return err
	}

	logger.L().Infow("File watch trigger started",
		"workflow_name", wf.Name,
		"watching_path", wf.Trigger.Path,
		"events_to_watch", wf.Trigger.Events,
	)

	// Start go routine to handle file events
	go func() {
		defer watcher.Close() // Ensure the watcher is closed when done
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.L().Errorw("File watcher events channel closed", "workflow_name", wf.Name)
					return
				}
				// Log for debugging what type of event is received
				logger.L().Debugw("Raw fsnotify event received",
					"workflow_name", wf.Name,
					"event_name", event.Name,
					"event_op", event.Op.String(),
				)
				shouldTrigger := false
				for _, ev := range wf.Trigger.Events {
					switch ev {
					case "create":
						if event.Op&fsnotify.Create == fsnotify.Create {
							shouldTrigger = true
						}
					case "write":
						if event.Op&fsnotify.Write == fsnotify.Write {
							shouldTrigger = true
						}
					case "remove":
						if event.Op&fsnotify.Remove == fsnotify.Remove {
							shouldTrigger = true
						}
					case "rename":
						if event.Op&fsnotify.Rename == fsnotify.Rename {
							shouldTrigger = true
						}
					case "chmod":
						if event.Op&fsnotify.Chmod == fsnotify.Chmod {
							shouldTrigger = true
						}
					default:
						logger.L().Errorw("Unsupported file event type",
							"workflow_name", wf.Name,
							"event_type", ev,
						)
					}

					if shouldTrigger {
						break // Found a matching event, no need to check further
					}
				}

				if shouldTrigger {
					logger.L().Infow("File watch trigger fired for worflow",
						"workflow_name", wf.Name,
						"event_type", event.Op.String(),
						"file_path", event.Name,
						"timestamp", time.Now().Format(time.RFC3339),
					)

					// Exceute actions
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
							// TODO: Implement HTTP action execution
						case workflow.ActionTypeCustom:
							logger.L().Infow("Custom action type detected, but execution not yet implemented (triggered by filewatch).",
								"workflow_name", wf.Name,
								"action_index", i,
								"action_name", act.Name,
								"action_type", act.Type.String(),
							)
							// TODO: Implement HTTP action execution
						default:
							logger.L().Warnw("Unsupported action type encountered for execution (triggered by filewatch)",
								"workflow_name", wf.Name,
								"action_index", i,
								"action_name", act.Name,
								"action_type", act.Type.String(),
							)
						}
					} // End of execte actions
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					logger.L().Errorw("File watcher errors channel closed", "workflow_name", wf.Name)
					return
				}
				logger.L().Errorw("File watcher error",
					"workflow_name", wf.Name,
					"error", err,
				)
			}
		}

	}()

	return nil
}
