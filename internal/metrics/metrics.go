package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WorkflowExecutions tracks workflow execution counts by status
	WorkflowExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "autozap_workflow_executions_total",
			Help: "Total number of workflow executions by workflow name and status",
		},
		[]string{"workflow", "status"},
	)

	// WorkflowDuration tracks workflow execution duration
	WorkflowDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "autozap_workflow_execution_duration_seconds",
			Help:    "Duration of workflow executions in seconds",
			Buckets: prometheus.DefBuckets, // Default: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"workflow"},
	)

	// ActionExecutions tracks action execution counts
	ActionExecutions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "autozap_action_executions_total",
			Help: "Total number of action executions by workflow, action name, action type, and status",
		},
		[]string{"workflow", "action", "action_type", "status"},
	)

	// ActionDuration tracks action execution duration
	ActionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "autozap_action_execution_duration_seconds",
			Help:    "Duration of action executions in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"workflow", "action", "action_type"},
	)

	// TriggerFires tracks trigger fire counts
	TriggerFires = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "autozap_trigger_fires_total",
			Help: "Total number of trigger fires by workflow and trigger type",
		},
		[]string{"workflow", "trigger_type"},
	)

	// AgentActiveWorkflows tracks the number of active workflows
	AgentActiveWorkflows = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "autozap_agent_active_workflows",
			Help: "Number of currently active workflows in agent mode",
		},
	)

	// AgentUptime tracks agent uptime in seconds
	AgentUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "autozap_agent_uptime_seconds",
			Help: "Agent uptime in seconds",
		},
	)

	// WorkflowLastExecution tracks last execution timestamp
	WorkflowLastExecution = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "autozap_workflow_last_execution_timestamp",
			Help: "Unix timestamp of last workflow execution",
		},
		[]string{"workflow"},
	)

	// WorkflowInfo provides metadata about workflows
	WorkflowInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "autozap_workflow_info",
			Help: "Information about configured workflows (always 1)",
		},
		[]string{"workflow", "trigger_type", "schedule"},
	)
)

// RecordWorkflowExecution records a workflow execution with duration
func RecordWorkflowExecution(workflowName string, status string, duration time.Duration) {
	WorkflowExecutions.WithLabelValues(workflowName, status).Inc()
	WorkflowDuration.WithLabelValues(workflowName).Observe(duration.Seconds())
	WorkflowLastExecution.WithLabelValues(workflowName).SetToCurrentTime()
}

// RecordActionExecution records an action execution with duration
func RecordActionExecution(workflowName, actionName, actionType, status string, duration time.Duration) {
	ActionExecutions.WithLabelValues(workflowName, actionName, actionType, status).Inc()
	ActionDuration.WithLabelValues(workflowName, actionName, actionType).Observe(duration.Seconds())
}

// RecordTriggerFire records a trigger fire event
func RecordTriggerFire(workflowName, triggerType string) {
	TriggerFires.WithLabelValues(workflowName, triggerType).Inc()
}

// RegisterWorkflow registers a workflow in the info metric
func RegisterWorkflow(workflowName, triggerType, schedule string) {
	WorkflowInfo.WithLabelValues(workflowName, triggerType, schedule).Set(1)
}

// UnregisterWorkflow removes a workflow from the info metric
func UnregisterWorkflow(workflowName, triggerType, schedule string) {
	WorkflowInfo.DeleteLabelValues(workflowName, triggerType, schedule)
}

// SetActiveWorkflows sets the number of active workflows
func SetActiveWorkflows(count int) {
	AgentActiveWorkflows.Set(float64(count))
}

// UpdateAgentUptime updates the agent uptime metric
func UpdateAgentUptime(startTime time.Time) {
	AgentUptime.Set(time.Since(startTime).Seconds())
}
