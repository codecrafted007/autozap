package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

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
	s.logger.Infof("Metrics available at: http://localhost:%d/metrics", s.port)
	s.logger.Infof("Health check at: http://localhost:%d/health", s.port)
	s.logger.Infof("Status at: http://localhost:%d/status", s.port)

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
