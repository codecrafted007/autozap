package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger()
}

func TestParseWorkflowFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	t.Run("Valid Cron Workflow", func(t *testing.T) {
		validYAML := `name: test-cron-workflow
description: A test cron workflow
trigger:
  type: cron
  schedule: "*/5 * * * *"
actions:
  - type: bash
    name: test-action
    command: echo "Hello World"
`
		filePath := filepath.Join(tmpDir, "valid-cron.yaml")
		err := os.WriteFile(filePath, []byte(validYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wf, err := ParseWorkflowFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if wf.Name != "test-cron-workflow" {
			t.Errorf("Expected name 'test-cron-workflow', got '%s'", wf.Name)
		}

		if wf.Trigger.Type != workflow.TriggerTypeCron {
			t.Errorf("Expected trigger type 'cron', got '%s'", wf.Trigger.Type)
		}

		if wf.Trigger.Schedule != "*/5 * * * *" {
			t.Errorf("Expected schedule '*/5 * * * *', got '%s'", wf.Trigger.Schedule)
		}

		if len(wf.Actions) != 1 {
			t.Errorf("Expected 1 action, got %d", len(wf.Actions))
		}
	})

	t.Run("Valid FileWatch Workflow", func(t *testing.T) {
		validYAML := `name: test-filewatch-workflow
trigger:
  type: filewatch
  path: /tmp/test
  events:
    - create
    - write
actions:
  - type: bash
    name: test-action
    command: echo "File changed"
`
		filePath := filepath.Join(tmpDir, "valid-filewatch.yaml")
		err := os.WriteFile(filePath, []byte(validYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wf, err := ParseWorkflowFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if wf.Trigger.Type != workflow.TriggerTypeFileWatch {
			t.Errorf("Expected trigger type 'filewatch', got '%s'", wf.Trigger.Type)
		}

		if wf.Trigger.Path != "/tmp/test" {
			t.Errorf("Expected path '/tmp/test', got '%s'", wf.Trigger.Path)
		}

		if len(wf.Trigger.Events) != 2 {
			t.Errorf("Expected 2 events, got %d", len(wf.Trigger.Events))
		}
	})

	t.Run("Valid HTTP Action Workflow", func(t *testing.T) {
		validYAML := `name: test-http-workflow
trigger:
  type: cron
  schedule: "0 * * * *"
actions:
  - type: http
    name: api-call
    url: https://api.example.com/status
    method: GET
    timeout: 10s
    expectStatus: 200
`
		filePath := filepath.Join(tmpDir, "valid-http.yaml")
		err := os.WriteFile(filePath, []byte(validYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wf, err := ParseWorkflowFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(wf.Actions) != 1 {
			t.Fatalf("Expected 1 action, got %d", len(wf.Actions))
		}

		action := wf.Actions[0]
		if action.Type != workflow.ActionTypeHTTP {
			t.Errorf("Expected action type 'http', got '%s'", action.Type)
		}

		if action.URL != "https://api.example.com/status" {
			t.Errorf("Expected URL 'https://api.example.com/status', got '%s'", action.URL)
		}

		if action.Method != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", action.Method)
		}
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, err := ParseWorkflowFile("/nonexistent/path/file.yaml")
		if err == nil {
			t.Fatal("Expected error for nonexistent file, got nil")
		}
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		invalidYAML := `name: test
trigger:
  type: cron
  schedule: "* * * * *"
actions:
  - invalid yaml content [[[
`
		filePath := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(filePath, []byte(invalidYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err = ParseWorkflowFile(filePath)
		if err == nil {
			t.Fatal("Expected error for invalid YAML, got nil")
		}
	})

	t.Run("Multiple Actions Workflow", func(t *testing.T) {
		validYAML := `name: multi-action-workflow
trigger:
  type: cron
  schedule: "0 */2 * * *"
actions:
  - type: bash
    name: first-bash
    command: echo "first"
  - type: http
    name: api-call
    url: https://api.example.com
    method: GET
  - type: bash
    name: second-bash
    command: echo "second"
`
		filePath := filepath.Join(tmpDir, "multi-action.yaml")
		err := os.WriteFile(filePath, []byte(validYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wf, err := ParseWorkflowFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(wf.Actions) != 3 {
			t.Errorf("Expected 3 actions, got %d", len(wf.Actions))
		}
	})

	t.Run("HTTP Action With Headers", func(t *testing.T) {
		validYAML := `name: http-with-headers
trigger:
  type: cron
  schedule: "* * * * *"
actions:
  - type: http
    name: api-call-with-auth
    url: https://api.example.com/data
    method: POST
    headers:
      Authorization: Bearer token123
      Content-Type: application/json
    body: '{"key": "value"}'
    timeout: 30s
    expectStatus: 201
`
		filePath := filepath.Join(tmpDir, "http-headers.yaml")
		err := os.WriteFile(filePath, []byte(validYAML), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		wf, err := ParseWorkflowFile(filePath)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(wf.Actions[0].Headers) != 2 {
			t.Errorf("Expected 2 headers, got %d", len(wf.Actions[0].Headers))
		}
	})
}

func TestValidateWorkflow(t *testing.T) {
	t.Run("Empty Workflow Name", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for empty workflow name, got nil")
		}
	})

	t.Run("No Actions", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for workflow with no actions, got nil")
		}
	})

	t.Run("Cron Trigger Missing Schedule", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for cron trigger without schedule, got nil")
		}
	})

	t.Run("FileWatch Trigger Missing Path", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:   workflow.TriggerTypeFileWatch,
				Path:   "",
				Events: []string{"create"},
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for filewatch trigger without path, got nil")
		}
	})

	t.Run("FileWatch Trigger Missing Events", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:   workflow.TriggerTypeFileWatch,
				Path:   "/tmp/test",
				Events: []string{},
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for filewatch trigger without events, got nil")
		}
	})

	t.Run("Unsupported Trigger Type", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type: "invalid-trigger",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for unsupported trigger type, got nil")
		}
	})

	t.Run("Action Without Name", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for action without name, got nil")
		}
	})

	t.Run("Bash Action Without Command", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: ""},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for bash action without command, got nil")
		}
	})

	t.Run("HTTP Action Without URL", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeHTTP, Name: "test", URL: "", Method: "GET"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for HTTP action without URL, got nil")
		}
	})

	t.Run("HTTP Action Without Method", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeHTTP, Name: "test", URL: "https://example.com", Method: ""},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for HTTP action without method, got nil")
		}
	})

	t.Run("Custom Action Without FunctionName", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeCustom, Name: "test", FunctionName: ""},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for custom action without functionName, got nil")
		}
	})

	t.Run("Unsupported Action Type", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: "invalid-action", Name: "test"},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for unsupported action type, got nil")
		}
	})

	t.Run("Valid Workflow", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "bash-test", Command: "echo test"},
				{Type: workflow.ActionTypeHTTP, Name: "http-test", URL: "https://example.com", Method: "GET"},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error for valid workflow, got: %v", err)
		}
	})

	t.Run("Cron Trigger With Extra FileWatch Fields", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
				Path:     "/tmp/test",
				Events:   []string{"create"},
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error (warnings only), got: %v", err)
		}
	})

	t.Run("FileWatch Trigger With Extra Cron Fields", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeFileWatch,
				Path:     "/tmp/test",
				Events:   []string{"create"},
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error (warnings only), got: %v", err)
		}
	})

	t.Run("Bash Action With Extra HTTP Fields", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{
					Type:    workflow.ActionTypeBash,
					Name:    "test",
					Command: "echo test",
					URL:     "https://example.com",
					Method:  "GET",
				},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error (warnings only), got: %v", err)
		}
	})

	t.Run("HTTP Action With Extra Bash Fields", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{
					Type:    workflow.ActionTypeHTTP,
					Name:    "test",
					URL:     "https://example.com",
					Method:  "GET",
					Command: "echo test",
				},
			},
		}

		err := validateWorkflow(wf)
		if err == nil {
			t.Fatal("Expected error for HTTP action with bash fields, got nil")
		}
	})

	t.Run("Custom Action With Extra Fields", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{
					Type:         workflow.ActionTypeCustom,
					Name:         "test",
					FunctionName: "myFunc",
					Command:      "echo test",
					URL:          "https://example.com",
				},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error (warnings only), got: %v", err)
		}
	})

	t.Run("HTTP Action With Single ExpectStatus", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{
					Type:         workflow.ActionTypeHTTP,
					Name:         "test",
					URL:          "https://example.com",
					Method:       "GET",
					ExpectStatus: 200,
				},
			},
		}

		err := validateWorkflow(wf)
		if err != nil {
			t.Fatalf("Expected no error for HTTP with single expectStatus, got: %v", err)
		}
	})
}
