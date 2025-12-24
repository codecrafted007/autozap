package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codecrafted007/autozap/internal/database"
	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

//go:embed dashboard/*
var dashboardFS embed.FS

// Server holds the HTTP server for metrics and health endpoints
type Server struct {
	httpServer *http.Server
	port       int
	logger     *zap.SugaredLogger
}

// WorkflowStatus represents the status of a single workflow
type WorkflowStatus struct {
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	LastExecution *time.Time `json:"last_execution,omitempty"`
	NextExecution *time.Time `json:"next_execution,omitempty"`
	TriggerType   string     `json:"trigger_type,omitempty"`
}

// HealthResponse represents the response for /health endpoint
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// StatusResponse represents the response for /status endpoint
type StatusResponse struct {
	Status    string           `json:"status"`
	Uptime    string           `json:"uptime"`
	Workflows WorkflowsSummary `json:"workflows"`
	Timestamp time.Time        `json:"timestamp"`
}

// WorkflowsSummary provides a summary of workflow states
type WorkflowsSummary struct {
	Total   int              `json:"total"`
	Running int              `json:"running"`
	Failed  int              `json:"failed"`
	Details []WorkflowStatus `json:"details,omitempty"`
}

var (
	serverStartTime    = time.Now()
	workflowStatuses   = make(map[string]*WorkflowStatus)
	workflowStatusFunc func() []WorkflowStatus
)

// NewServer creates a new HTTP server for metrics and health endpoints
func NewServer(port int) *Server {
	mux := http.NewServeMux()

	// Dashboard UI (embedded files at /dashboard/)
	mux.Handle("/dashboard/", http.FileServer(http.FS(dashboardFS)))

	// Redirect root to dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// API endpoints
	mux.HandleFunc("/api/workflows/active", activeWorkflowsAPIHandler)
	mux.HandleFunc("/api/workflows/history", historyAPIHandler)
	mux.HandleFunc("/api/workflows/stats", statsAPIHandler)
	mux.HandleFunc("/api/workflows/failures", failuresAPIHandler)

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health endpoint (liveness probe)
	mux.HandleFunc("/health", healthHandler)

	// Readiness endpoint
	mux.HandleFunc("/ready", readyHandler)

	// Status endpoint (detailed info)
	mux.HandleFunc("/status", statusHandler)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		port:   port,
		logger: logger.L(),
	}
}

// Start starts the HTTP server in a goroutine
func (s *Server) Start() error {
	s.logger.Infof("Starting HTTP server on port %d", s.port)
	s.logger.Infof("🎨 Dashboard available at: http://localhost:%d/dashboard", s.port)
	s.logger.Infof("📊 Metrics available at: http://localhost:%d/metrics", s.port)
	s.logger.Infof("❤️  Health check at: http://localhost:%d/health", s.port)
	s.logger.Infof("📈 Status at: http://localhost:%d/status", s.port)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// healthHandler handles the /health endpoint (liveness probe)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// readyHandler handles the /ready endpoint (readiness probe)
func readyHandler(w http.ResponseWriter, r *http.Request) {
	// Check if workflows are loaded
	// For now, we'll consider the service ready if it's running
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// statusHandler handles the /status endpoint
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uptime := time.Since(serverStartTime)

	// Get workflow details if available
	var details []WorkflowStatus
	if workflowStatusFunc != nil {
		details = workflowStatusFunc()
	}

	// Calculate summary
	total := len(details)
	running := 0
	failed := 0
	for _, wf := range details {
		if wf.Status == "running" {
			running++
		} else if wf.Status == "failed" {
			failed++
		}
	}

	response := StatusResponse{
		Status: "healthy",
		Uptime: formatDuration(uptime),
		Workflows: WorkflowsSummary{
			Total:   total,
			Running: running,
			Failed:  failed,
			Details: details,
		},
		Timestamp: time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// SetWorkflowStatusFunc sets the function to retrieve workflow statuses
func SetWorkflowStatusFunc(fn func() []WorkflowStatus) {
	workflowStatusFunc = fn
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// API Handlers

// activeWorkflowsAPIHandler handles /api/workflows/active
func activeWorkflowsAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	workflows := GetRegistry().GetAllWorkflows()
	json.NewEncoder(w).Encode(workflows)
}

// historyAPIHandler handles /api/workflows/history
func historyAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	limit := 50
	executions, err := database.GetAllWorkflowHistory(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(executions)
}

// statsAPIHandler handles /api/workflows/stats
func statsAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get recent executions to calculate stats
	since := time.Now().AddDate(0, 0, -7) // Last 7 days
	executions, err := database.GetAllWorkflowHistory(1000)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate aggregate stats
	workflowStats := make(map[string]*database.WorkflowStats)
	for _, exec := range executions {
		if exec.StartedAt.Before(since) {
			continue
		}

		if _, exists := workflowStats[exec.WorkflowName]; !exists {
			workflowStats[exec.WorkflowName] = &database.WorkflowStats{
				WorkflowName: exec.WorkflowName,
			}
		}

		stats := workflowStats[exec.WorkflowName]
		stats.TotalExecutions++

		if exec.Status == "success" {
			stats.SuccessCount++
		} else if exec.Status == "failed" {
			stats.FailedCount++
		}

		if exec.DurationMs != nil {
			stats.AvgDurationMs = (stats.AvgDurationMs*float64(stats.TotalExecutions-1) + float64(*exec.DurationMs)) / float64(stats.TotalExecutions)
		}
	}

	// Calculate success rates
	for _, stats := range workflowStats {
		if stats.TotalExecutions > 0 {
			stats.SuccessRate = (float64(stats.SuccessCount) / float64(stats.TotalExecutions)) * 100
		}
	}

	json.NewEncoder(w).Encode(workflowStats)
}

// failuresAPIHandler handles /api/workflows/failures
func failuresAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	since := time.Now().Add(-24 * time.Hour) // Last 24 hours
	failures, err := database.GetFailedExecutions(since, 50)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get failures: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(failures)
}
