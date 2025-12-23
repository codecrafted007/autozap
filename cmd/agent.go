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
	"github.com/codecrafted007/autozap/internal/metrics"
	"github.com/codecrafted007/autozap/internal/parser"
	"github.com/codecrafted007/autozap/internal/server"
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

		// Get flags
		watch, _ := cmd.Flags().GetBool("watch")
		logDir, _ := cmd.Flags().GetString("log-dir")
		httpPort, _ := cmd.Flags().GetInt("http-port")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			logger.L().Info("[DRY RUN MODE] No workflows will be executed")
		}

		logger.L().Infow("Starting AutoZap Agent",
			"workflow_directory", workflowDir,
			"hot_reload", watch,
			"log_directory", logDir,
			"http_port", httpPort,
			"dry_run", dryRun,
		)

		// Start HTTP server for metrics and health endpoints
		srv := server.NewServer(httpPort)
		if err := srv.Start(); err != nil {
			logger.L().Errorw("Failed to start HTTP server",
				"error", err,
			)
			return
		}

		// Track agent start time for uptime metric
		agentStartTime := time.Now()

		// Validate log directory if specified
		if logDir != "" {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				logger.L().Errorw("Failed to create log directory",
					"directory", logDir,
					"error", err,
				)
				return
			}
			logger.L().Infow("Per-workflow logging enabled",
				"log_directory", logDir,
			)
		}

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
		if err := loadWorkflows(ctx, workflowDir, logDir, activeWorkflows, dryRun); err != nil {
			logger.L().Errorw("Failed to load workflows",
				"error", err,
			)
			return
		}

		// In dry-run mode, exit after showing what would be done
		if dryRun {
			logger.L().Info("[DRY RUN] Dry run complete. No workflows were started.")
			return
		}

		// Update active workflows metric
		count := 0
		activeWorkflows.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		metrics.SetActiveWorkflows(count)

		// Start goroutine to periodically update agent uptime
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					metrics.UpdateAgentUptime(agentStartTime)
				}
			}
		}()

		// Setup file watcher for hot-reload
		var watcher *fsnotify.Watcher
		var err error
		if watch {
			watcher, err = setupWorkflowWatcher(ctx, workflowDir, logDir, activeWorkflows)
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

		// Shutdown HTTP server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Stop(shutdownCtx); err != nil {
			logger.L().Errorw("Error shutting down HTTP server", "error", err)
		}

		logger.L().Info("AutoZap Agent stopped successfully")
	},
}

// loadWorkflows discovers and starts all workflow files in a directory
func loadWorkflows(ctx context.Context, workflowDir, logDir string, activeWorkflows *sync.Map, dryRun bool) error {
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

	// In dry-run mode, show what would be started
	if dryRun {
		logger.L().Infof("[DRY RUN] Would start %d workflows:", len(files))
		for i, file := range files {
			wf, err := parser.ParseWorkflowFile(file)
			if err != nil {
				logger.L().Errorf("[DRY RUN] Would fail to load: %s (error: %v)", file, err)
				continue
			}
			logger.L().Infof("[DRY RUN]   %d. %s", i+1, wf.Name)
			logger.L().Infof("[DRY RUN]      File: %s", file)
			logger.L().Infof("[DRY RUN]      Trigger: %s", wf.Trigger.Type)

			switch wf.Trigger.Type {
			case workflow.TriggerTypeCron:
				logger.L().Infof("[DRY RUN]      Schedule: %s", wf.Trigger.Schedule)
			case workflow.TriggerTypeFileWatch:
				logger.L().Infof("[DRY RUN]      Watch: %s", wf.Trigger.Path)
			}

			logger.L().Infof("[DRY RUN]      Actions: %d", len(wf.Actions))
		}
		return nil
	}

	// Load each workflow
	successCount := 0
	for _, file := range files {
		if err := startWorkflow(ctx, file, logDir, activeWorkflows); err != nil {
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
func startWorkflow(ctx context.Context, filePath, logDir string, activeWorkflows *sync.Map) error {
	// Parse workflow
	wf, err := parser.ParseWorkflowFile(filePath)
	if err != nil {
		return err
	}

	// Create workflow-specific logger
	workflowLogger, err := logger.NewWorkflowLogger(wf.Name, logDir)
	if err != nil {
		logger.L().Errorw("Failed to create workflow logger",
			"workflow_name", wf.Name,
			"error", err,
		)
		// Fallback to global logger
		workflowLogger = logger.L().With("workflow_name", wf.Name)
	}

	workflowLogger.Infow("Starting workflow",
		"file", filePath,
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
				workflowLogger.Errorw("Failed to start cron trigger",
					"file", filePath,
					"error", err,
				)
				return
			}
		case workflow.TriggerTypeFileWatch:
			if err := trigger.StartFileWatchTrigger(wf); err != nil {
				workflowLogger.Errorw("Failed to start file watch trigger",
					"file", filePath,
					"error", err,
				)
				return
			}
		default:
			workflowLogger.Errorw("Unsupported trigger type",
				"trigger_type", wf.Trigger.Type,
			)
			return
		}

		// Wait for context cancellation
		<-workflowCtx.Done()
		workflowLogger.Infow("Workflow stopped",
			"file", filePath,
		)
	}()

	return nil
}

// setupWorkflowWatcher sets up file system watcher for hot-reload
func setupWorkflowWatcher(ctx context.Context, workflowDir, logDir string, activeWorkflows *sync.Map) (*fsnotify.Watcher, error) {
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

					if err := startWorkflow(ctx, event.Name, logDir, activeWorkflows); err != nil {
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
					if err := startWorkflow(ctx, event.Name, logDir, activeWorkflows); err != nil {
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
	agentCmd.Flags().String("log-dir", "", "Directory for per-workflow log files (default: stdout)")
	agentCmd.Flags().Int("http-port", 8080, "HTTP port for metrics and health endpoints")
	agentCmd.Flags().Bool("dry-run", false, "Show what would be executed without starting workflows")
}
