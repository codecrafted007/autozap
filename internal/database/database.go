package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/codecrafted007/autozap/internal/logger"
)

var db *sql.DB

// WorkflowExecution represents a workflow execution record
type WorkflowExecution struct {
	ID           int64
	WorkflowName string
	StartedAt    time.Time
	CompletedAt  *time.Time
	Status       string // running, success, failed
	Error        *string
	DurationMs   *int64
	TriggerType  string
}

// ActionExecution represents an action execution record
type ActionExecution struct {
	ID                  int64
	WorkflowExecutionID int64
	ActionName          string
	ActionType          string
	StartedAt           time.Time
	CompletedAt         *time.Time
	Status              string // running, success, failed
	Error               *string
	DurationMs          *int64
	Output              *string
}

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	logger.L().Infow("Database initialized successfully", "path", dbPath)
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS workflow_executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workflow_name TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		status TEXT NOT NULL,
		error TEXT,
		duration_ms INTEGER,
		trigger_type TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_workflow_started
	ON workflow_executions(workflow_name, started_at);

	CREATE INDEX IF NOT EXISTS idx_workflow_status
	ON workflow_executions(status);

	CREATE TABLE IF NOT EXISTS action_executions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		workflow_execution_id INTEGER NOT NULL,
		action_name TEXT NOT NULL,
		action_type TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		status TEXT NOT NULL,
		error TEXT,
		duration_ms INTEGER,
		output TEXT,
		FOREIGN KEY (workflow_execution_id) REFERENCES workflow_executions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_action_workflow
	ON action_executions(workflow_execution_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// StartWorkflowExecution creates a new workflow execution record
func StartWorkflowExecution(workflowName, triggerType string) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := db.Exec(`
		INSERT INTO workflow_executions (workflow_name, started_at, status, trigger_type)
		VALUES (?, ?, ?, ?)
	`, workflowName, time.Now(), "running", triggerType)

	if err != nil {
		return 0, fmt.Errorf("failed to insert workflow execution: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// CompleteWorkflowExecution updates a workflow execution as completed
func CompleteWorkflowExecution(id int64, status string, errorMsg *string, duration time.Duration) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	durationMs := duration.Milliseconds()
	completedAt := time.Now()

	_, err := db.Exec(`
		UPDATE workflow_executions
		SET completed_at = ?, status = ?, error = ?, duration_ms = ?
		WHERE id = ?
	`, completedAt, status, errorMsg, durationMs, id)

	if err != nil {
		return fmt.Errorf("failed to update workflow execution: %w", err)
	}

	return nil
}

// StartActionExecution creates a new action execution record
func StartActionExecution(workflowExecID int64, actionName, actionType string) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := db.Exec(`
		INSERT INTO action_executions (workflow_execution_id, action_name, action_type, started_at, status)
		VALUES (?, ?, ?, ?, ?)
	`, workflowExecID, actionName, actionType, time.Now(), "running")

	if err != nil {
		return 0, fmt.Errorf("failed to insert action execution: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// CompleteActionExecution updates an action execution as completed
func CompleteActionExecution(id int64, status string, errorMsg *string, output *string, duration time.Duration) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	durationMs := duration.Milliseconds()
	completedAt := time.Now()

	_, err := db.Exec(`
		UPDATE action_executions
		SET completed_at = ?, status = ?, error = ?, output = ?, duration_ms = ?
		WHERE id = ?
	`, completedAt, status, errorMsg, output, durationMs, id)

	if err != nil {
		return fmt.Errorf("failed to update action execution: %w", err)
	}

	return nil
}

// GetWorkflowHistory returns recent workflow executions
func GetWorkflowHistory(workflowName string, limit int) ([]WorkflowExecution, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, workflow_name, started_at, completed_at, status, error, duration_ms, trigger_type
		FROM workflow_executions
		WHERE workflow_name = ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, workflowName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow history: %w", err)
	}
	defer rows.Close()

	executions := make([]WorkflowExecution, 0)
	for rows.Next() {
		var exec WorkflowExecution
		err := rows.Scan(
			&exec.ID,
			&exec.WorkflowName,
			&exec.StartedAt,
			&exec.CompletedAt,
			&exec.Status,
			&exec.Error,
			&exec.DurationMs,
			&exec.TriggerType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		executions = append(executions, exec)
	}

	return executions, nil
}

// GetAllWorkflowHistory returns recent executions for all workflows
func GetAllWorkflowHistory(limit int) ([]WorkflowExecution, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, workflow_name, started_at, completed_at, status, error, duration_ms, trigger_type
		FROM workflow_executions
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow history: %w", err)
	}
	defer rows.Close()

	executions := make([]WorkflowExecution, 0)
	for rows.Next() {
		var exec WorkflowExecution
		err := rows.Scan(
			&exec.ID,
			&exec.WorkflowName,
			&exec.StartedAt,
			&exec.CompletedAt,
			&exec.Status,
			&exec.Error,
			&exec.DurationMs,
			&exec.TriggerType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		executions = append(executions, exec)
	}

	return executions, nil
}

// GetFailedExecutions returns recent failed workflow executions
func GetFailedExecutions(since time.Time, limit int) ([]WorkflowExecution, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, workflow_name, started_at, completed_at, status, error, duration_ms, trigger_type
		FROM workflow_executions
		WHERE status = 'failed' AND started_at >= ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := db.Query(query, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed executions: %w", err)
	}
	defer rows.Close()

	executions := make([]WorkflowExecution, 0)
	for rows.Next() {
		var exec WorkflowExecution
		err := rows.Scan(
			&exec.ID,
			&exec.WorkflowName,
			&exec.StartedAt,
			&exec.CompletedAt,
			&exec.Status,
			&exec.Error,
			&exec.DurationMs,
			&exec.TriggerType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		executions = append(executions, exec)
	}

	return executions, nil
}

// GetWorkflowStats returns statistics for a workflow
type WorkflowStats struct {
	WorkflowName    string
	TotalExecutions int
	SuccessCount    int
	FailedCount     int
	SuccessRate     float64
	AvgDurationMs   float64
}

func GetWorkflowStats(workflowName string, since time.Time) (*WorkflowStats, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT
			workflow_name,
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
			AVG(duration_ms) as avg_duration
		FROM workflow_executions
		WHERE workflow_name = ? AND started_at >= ?
		GROUP BY workflow_name
	`

	var stats WorkflowStats
	var avgDuration sql.NullFloat64

	err := db.QueryRow(query, workflowName, since).Scan(
		&stats.WorkflowName,
		&stats.TotalExecutions,
		&stats.SuccessCount,
		&stats.FailedCount,
		&avgDuration,
	)

	if err == sql.ErrNoRows {
		// No executions found
		return &WorkflowStats{
			WorkflowName:    workflowName,
			TotalExecutions: 0,
			SuccessCount:    0,
			FailedCount:     0,
			SuccessRate:     0,
			AvgDurationMs:   0,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query workflow stats: %w", err)
	}

	if avgDuration.Valid {
		stats.AvgDurationMs = avgDuration.Float64
	}

	if stats.TotalExecutions > 0 {
		stats.SuccessRate = (float64(stats.SuccessCount) / float64(stats.TotalExecutions)) * 100
	}

	return &stats, nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDB returns the database instance (for testing)
func GetDB() *sql.DB {
	return db
}
