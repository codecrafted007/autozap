package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/parser"
	"github.com/codecrafted007/autozap/internal/trigger"
	"github.com/codecrafted007/autozap/internal/workflow"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent [workflows_directory]",
	Short: "Run AutoZap in agent mode - automatically discover and run all workflows",
	Long: `Agent mode watches a directory for workflow files and runs them all concurrently.

AutoZap will:
- Discover all .yaml and .yml files in the directory
- Parse and validate each workflow
- Start all triggers concurrently
- Hot-reload when new workflows are added
- Gracefully shutdown on SIGTERM/SIGINT

Example:
  autozap agent ./workflows
  autozap agent ./workflows --watch=false  # Disable hot-reload`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default to ./workflows directory
		workflowDir := "./workflows"
		if len(args) > 0 {
			workflowDir = args[0]
		}

		// Get watch flag
		watch, _ := cmd.Flags().GetBool("watch")

		logger.L().Infow("Starting AutoZap Agent",
			"workflow_directory", workflowDir,
			"hot_reload", watch,
		)

		// Check if directory exists
		if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
			logger.L().Errorw("Workflow directory does not exist",
				"directory", workflowDir,
				"error", err,
			)
			return
		}

		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Load and start all workflows
		activeWorkflows := &sync.Map{} // map[string]context.CancelFunc
		if err := loadWorkflows(ctx, workflowDir, activeWorkflows); err != nil {
			logger.L().Errorw("Failed to load workflows",
				"error", err,
			)
			return
		}

		// Setup file watcher for hot-reload
		var watcher *fsnotify.Watcher
		var err error
		if watch {
			watcher, err = setupWorkflowWatcher(ctx, workflowDir, activeWorkflows)
			if err != nil {
				logger.L().Errorw("Failed to setup workflow watcher",
					"error", err,
				)
				return
			}
			defer watcher.Close()
		}

		logger.L().Info("🚀 AutoZap Agent is running. Press Ctrl+C to stop.")

		// Wait for shutdown signal
		<-sigChan
		logger.L().Info("Received shutdown signal. Gracefully stopping all workflows...")

		// Cancel all workflows
		cancel()

		// Give workflows time to cleanup
		time.Sleep(2 * time.Second)

		logger.L().Info("AutoZap Agent stopped successfully")
	},
}

// loadWorkflows discovers and starts all workflow files in a directory
func loadWorkflows(ctx context.Context, workflowDir string, activeWorkflows *sync.Map) error {
	// Find all YAML files
	pattern := filepath.Join(workflowDir, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Also find .yml files
	pattern2 := filepath.Join(workflowDir, "*.yml")
	ymlFiles, err := filepath.Glob(pattern2)
	if err != nil {
		return err
	}
	files = append(files, ymlFiles...)

	if len(files) == 0 {
		logger.L().Warnw("No workflow files found in directory",
			"directory", workflowDir,
		)
		return nil
	}

	logger.L().Infow("Discovered workflow files",
		"count", len(files),
		"directory", workflowDir,
	)

	// Load each workflow
	successCount := 0
	for _, file := range files {
		if err := startWorkflow(ctx, file, activeWorkflows); err != nil {
			logger.L().Errorw("Failed to start workflow",
				"file", file,
				"error", err,
			)
			continue
		}
		successCount++
	}

	logger.L().Infow("Workflows started",
		"total", len(files),
		"successful", successCount,
		"failed", len(files)-successCount,
	)

	return nil
}

// startWorkflow parses and starts a single workflow
func startWorkflow(ctx context.Context, filePath string, activeWorkflows *sync.Map) error {
	// Parse workflow
	wf, err := parser.ParseWorkflowFile(filePath)
	if err != nil {
		return err
	}

	logger.L().Infow("Starting workflow",
		"file", filePath,
		"workflow_name", wf.Name,
		"trigger_type", wf.Trigger.Type,
		"actions_count", len(wf.Actions),
	)

	// Create a context for this workflow
	workflowCtx, workflowCancel := context.WithCancel(ctx)

	// Store the cancel function
	activeWorkflows.Store(filePath, workflowCancel)

	// Start the workflow in a goroutine
	go func() {
		defer workflowCancel()

		switch wf.Trigger.Type {
		case workflow.TriggerTypeCron:
			if err := trigger.StartCronTrigger(wf); err != nil {
				logger.L().Errorw("Failed to start cron trigger",
					"workflow_name", wf.Name,
					"file", filePath,
					"error", err,
				)
				return
			}
		case workflow.TriggerTypeFileWatch:
			if err := trigger.StartFileWatchTrigger(wf); err != nil {
				logger.L().Errorw("Failed to start file watch trigger",
					"workflow_name", wf.Name,
					"file", filePath,
					"error", err,
				)
				return
			}
		default:
			logger.L().Errorw("Unsupported trigger type",
				"workflow_name", wf.Name,
				"trigger_type", wf.Trigger.Type,
			)
			return
		}

		// Wait for context cancellation
		<-workflowCtx.Done()
		logger.L().Infow("Workflow stopped",
			"workflow_name", wf.Name,
			"file", filePath,
		)
	}()

	return nil
}

// setupWorkflowWatcher sets up file system watcher for hot-reload
func setupWorkflowWatcher(ctx context.Context, workflowDir string, activeWorkflows *sync.Map) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Add workflow directory to watcher
	if err := watcher.Add(workflowDir); err != nil {
		watcher.Close()
		return nil, err
	}

	logger.L().Infow("Workflow hot-reload enabled",
		"directory", workflowDir,
	)

	// Watch for file changes in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Only handle YAML files
				if filepath.Ext(event.Name) != ".yaml" && filepath.Ext(event.Name) != ".yml" {
					continue
				}

				// Handle create events
				if event.Op&fsnotify.Create == fsnotify.Create {
					logger.L().Infow("New workflow detected",
						"file", event.Name,
						"operation", "create",
					)

					// Wait a bit for file to be fully written
					time.Sleep(500 * time.Millisecond)

					if err := startWorkflow(ctx, event.Name, activeWorkflows); err != nil {
						logger.L().Errorw("Failed to start new workflow",
							"file", event.Name,
							"error", err,
						)
					}
				}

				// Handle write events (workflow file updated)
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.L().Infow("Workflow file modified",
						"file", event.Name,
						"operation", "write",
					)

					// Stop existing workflow
					if cancel, ok := activeWorkflows.Load(event.Name); ok {
						if cancelFunc, ok := cancel.(context.CancelFunc); ok {
							cancelFunc()
						}
						activeWorkflows.Delete(event.Name)
					}

					// Wait a bit for file to be fully written
					time.Sleep(500 * time.Millisecond)

					// Start updated workflow
					if err := startWorkflow(ctx, event.Name, activeWorkflows); err != nil {
						logger.L().Errorw("Failed to reload workflow",
							"file", event.Name,
							"error", err,
						)
					} else {
						logger.L().Infow("Workflow reloaded successfully",
							"file", event.Name,
						)
					}
				}

				// Handle delete events
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					logger.L().Infow("Workflow file removed",
						"file", event.Name,
						"operation", "remove",
					)

					// Stop workflow
					if cancel, ok := activeWorkflows.Load(event.Name); ok {
						if cancelFunc, ok := cancel.(context.CancelFunc); ok {
							cancelFunc()
						}
						activeWorkflows.Delete(event.Name)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.L().Errorw("Workflow watcher error",
					"error", err,
				)
			}
		}
	}()

	return watcher, nil
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// Add flags
	agentCmd.Flags().Bool("watch", true, "Enable hot-reload for workflow changes")
}
