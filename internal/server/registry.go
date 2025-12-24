package server

import (
	"sync"
	"time"

	"github.com/codecrafted007/autozap/internal/workflow"
)

// WorkflowRegistry tracks active workflows and their status
type WorkflowRegistry struct {
	workflows map[string]*WorkflowInfo
	mu        sync.RWMutex
}

// WorkflowInfo contains runtime information about a workflow
type WorkflowInfo struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	TriggerType   string                 `json:"trigger_type"`
	Schedule      string                 `json:"schedule,omitempty"`
	Status        string                 `json:"status"` // active, stopped, error
	RegisteredAt  time.Time              `json:"registered_at"`
	LastExecution *time.Time             `json:"last_execution,omitempty"`
	NextExecution *time.Time             `json:"next_execution,omitempty"`
	TotalRuns     int                    `json:"total_runs"`
	SuccessCount  int                    `json:"success_count"`
	FailureCount  int                    `json:"failure_count"`
	LastError     string                 `json:"last_error,omitempty"`
	Actions       []WorkflowActionInfo   `json:"actions"`
}

// WorkflowActionInfo contains information about an action
type WorkflowActionInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var registry *WorkflowRegistry

func init() {
	registry = &WorkflowRegistry{
		workflows: make(map[string]*WorkflowInfo),
	}
}

// GetRegistry returns the global workflow registry
func GetRegistry() *WorkflowRegistry {
	return registry
}

// RegisterWorkflow registers a new workflow in the registry
func (r *WorkflowRegistry) RegisterWorkflow(wf *workflow.Workflow) {
	r.mu.Lock()
	defer r.mu.Unlock()

	actions := make([]WorkflowActionInfo, len(wf.Actions))
	for i, action := range wf.Actions {
		actions[i] = WorkflowActionInfo{
			Name: action.Name,
			Type: string(action.Type),
		}
	}

	info := &WorkflowInfo{
		Name:         wf.Name,
		Description:  wf.Description,
		TriggerType:  string(wf.Trigger.Type),
		Schedule:     wf.Trigger.Schedule,
		Status:       "active",
		RegisteredAt: time.Now(),
		Actions:      actions,
	}

	r.workflows[wf.Name] = info
}

// UnregisterWorkflow removes a workflow from the registry
func (r *WorkflowRegistry) UnregisterWorkflow(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if info, exists := r.workflows[name]; exists {
		info.Status = "stopped"
	}
}

// UpdateExecutionStats updates execution statistics for a workflow
func (r *WorkflowRegistry) UpdateExecutionStats(name string, success bool, errorMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.workflows[name]
	if !exists {
		return
	}

	now := time.Now()
	info.LastExecution = &now
	info.TotalRuns++

	if success {
		info.SuccessCount++
		info.LastError = ""
	} else {
		info.FailureCount++
		if errorMsg != "" {
			info.LastError = errorMsg
		}
	}
}

// UpdateNextExecution updates the next scheduled execution time
func (r *WorkflowRegistry) UpdateNextExecution(name string, nextTime time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.workflows[name]
	if !exists {
		return
	}

	info.NextExecution = &nextTime
}

// GetWorkflow returns information about a specific workflow
func (r *WorkflowRegistry) GetWorkflow(name string) (*WorkflowInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.workflows[name]
	return info, exists
}

// GetAllWorkflows returns all registered workflows
func (r *WorkflowRegistry) GetAllWorkflows() []*WorkflowInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workflows := make([]*WorkflowInfo, 0, len(r.workflows))
	for _, info := range r.workflows {
		workflows = append(workflows, info)
	}

	return workflows
}

// GetActiveWorkflows returns only active workflows
func (r *WorkflowRegistry) GetActiveWorkflows() []*WorkflowInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	workflows := make([]*WorkflowInfo, 0)
	for _, info := range r.workflows {
		if info.Status == "active" {
			workflows = append(workflows, info)
		}
	}

	return workflows
}

// GetWorkflowCount returns the total number of registered workflows
func (r *WorkflowRegistry) GetWorkflowCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.workflows)
}
