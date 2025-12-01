package workflow

import (
	"testing"
)

func TestTriggerTypeString(t *testing.T) {
	tests := []struct {
		name     string
		trigger  TriggerType
		expected string
	}{
		{"Cron", TriggerTypeCron, "cron"},
		{"FileWatch", TriggerTypeFileWatch, "filewatch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.trigger.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestActionTypeString(t *testing.T) {
	tests := []struct {
		name     string
		action   ActionType
		expected string
	}{
		{"Bash", ActionTypeBash, "bash"},
		{"HTTP", ActionTypeHTTP, "http"},
		{"Custom", ActionTypeCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestWorkflowStructure(t *testing.T) {
	t.Run("Valid Workflow Structure", func(t *testing.T) {
		wf := Workflow{
			Name:        "test-workflow",
			Description: "A test workflow",
			Trigger: Trigger{
				Type:     TriggerTypeCron,
				Schedule: "*/5 * * * *",
			},
			Actions: []Action{
				{
					Type:    ActionTypeBash,
					Name:    "test-action",
					Command: "echo test",
				},
			},
		}

		if wf.Name != "test-workflow" {
			t.Errorf("Expected name 'test-workflow', got '%s'", wf.Name)
		}

		if wf.Trigger.Type != TriggerTypeCron {
			t.Errorf("Expected trigger type 'cron', got '%s'", wf.Trigger.Type)
		}

		if len(wf.Actions) != 1 {
			t.Errorf("Expected 1 action, got %d", len(wf.Actions))
		}
	})

	t.Run("Action With All HTTP Fields", func(t *testing.T) {
		action := Action{
			Type:   ActionTypeHTTP,
			Name:   "api-call",
			URL:    "https://api.example.com",
			Method: "POST",
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
			Body:               `{"key": "value"}`,
			Timeout:            "10s",
			ExpectStatus:       200,
			ExpectBodyContains: "success",
		}

		if action.Type != ActionTypeHTTP {
			t.Errorf("Expected action type 'http', got '%s'", action.Type)
		}

		if action.URL != "https://api.example.com" {
			t.Errorf("Expected URL 'https://api.example.com', got '%s'", action.URL)
		}

		if len(action.Headers) != 2 {
			t.Errorf("Expected 2 headers, got %d", len(action.Headers))
		}
	})

	t.Run("FileWatch Trigger Structure", func(t *testing.T) {
		trigger := Trigger{
			Type:   TriggerTypeFileWatch,
			Path:   "/tmp/watch",
			Events: []string{"create", "write", "remove"},
		}

		if trigger.Type != TriggerTypeFileWatch {
			t.Errorf("Expected trigger type 'filewatch', got '%s'", trigger.Type)
		}

		if trigger.Path != "/tmp/watch" {
			t.Errorf("Expected path '/tmp/watch', got '%s'", trigger.Path)
		}

		if len(trigger.Events) != 3 {
			t.Errorf("Expected 3 events, got %d", len(trigger.Events))
		}
	})

	t.Run("Custom Action Structure", func(t *testing.T) {
		action := Action{
			Type:         ActionTypeCustom,
			Name:         "custom-function",
			FunctionName: "myCustomFunction",
			Arguments: map[string]interface{}{
				"arg1": "value1",
				"arg2": 123,
				"arg3": true,
			},
		}

		if action.Type != ActionTypeCustom {
			t.Errorf("Expected action type 'custom', got '%s'", action.Type)
		}

		if action.FunctionName != "myCustomFunction" {
			t.Errorf("Expected functionName 'myCustomFunction', got '%s'", action.FunctionName)
		}

		if len(action.Arguments) != 3 {
			t.Errorf("Expected 3 arguments, got %d", len(action.Arguments))
		}
	})
}

func TestActionTypeConstants(t *testing.T) {
	if ActionTypeBash != "bash" {
		t.Errorf("Expected ActionTypeBash to be 'bash', got '%s'", ActionTypeBash)
	}

	if ActionTypeHTTP != "http" {
		t.Errorf("Expected ActionTypeHTTP to be 'http', got '%s'", ActionTypeHTTP)
	}

	if ActionTypeCustom != "custom" {
		t.Errorf("Expected ActionTypeCustom to be 'custom', got '%s'", ActionTypeCustom)
	}
}

func TestTriggerTypeConstants(t *testing.T) {
	if TriggerTypeCron != "cron" {
		t.Errorf("Expected TriggerTypeCron to be 'cron', got '%s'", TriggerTypeCron)
	}

	if TriggerTypeFileWatch != "filewatch" {
		t.Errorf("Expected TriggerTypeFileWatch to be 'filewatch', got '%s'", TriggerTypeFileWatch)
	}
}

func TestMultipleActionsWorkflow(t *testing.T) {
	wf := Workflow{
		Name:        "multi-action-workflow",
		Description: "Workflow with multiple actions",
		Trigger: Trigger{
			Type:     TriggerTypeCron,
			Schedule: "0 0 * * *",
		},
		Actions: []Action{
			{
				Type:    ActionTypeBash,
				Name:    "first-action",
				Command: "echo first",
			},
			{
				Type:   ActionTypeHTTP,
				Name:   "second-action",
				URL:    "https://api.example.com",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body:         `{"data": "test"}`,
				Timeout:      "30s",
				ExpectStatus: 201,
			},
			{
				Type:         ActionTypeCustom,
				Name:         "third-action",
				FunctionName: "customFunc",
				Arguments: map[string]interface{}{
					"param1": "value1",
					"param2": 42,
				},
			},
		},
	}

	if len(wf.Actions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(wf.Actions))
	}

	if wf.Actions[0].Type != ActionTypeBash {
		t.Errorf("Expected first action type to be 'bash', got '%s'", wf.Actions[0].Type)
	}

	if wf.Actions[1].Type != ActionTypeHTTP {
		t.Errorf("Expected second action type to be 'http', got '%s'", wf.Actions[1].Type)
	}

	if wf.Actions[2].Type != ActionTypeCustom {
		t.Errorf("Expected third action type to be 'custom', got '%s'", wf.Actions[2].Type)
	}
}

func TestEmptyOptionalFields(t *testing.T) {
	wf := Workflow{
		Name:        "minimal-workflow",
		Description: "",
		Trigger: Trigger{
			Type:     TriggerTypeCron,
			Schedule: "* * * * *",
		},
		Actions: []Action{
			{
				Type:    ActionTypeBash,
				Name:    "minimal-bash",
				Command: "echo test",
			},
		},
	}

	if wf.Description != "" {
		t.Errorf("Expected empty description, got '%s'", wf.Description)
	}

	if wf.Actions[0].Command != "echo test" {
		t.Errorf("Expected command 'echo test', got '%s'", wf.Actions[0].Command)
	}
}
