package trigger

import (
	"testing"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

func init() {
	logger.InitLogger()
}

func TestStartFileWatchTrigger(t *testing.T) {
	t.Run("Invalid Trigger Type", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:   workflow.TriggerTypeCron,
				Path:   "/tmp/test",
				Events: []string{"create"},
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := StartFileWatchTrigger(wf)
		if err == nil {
			t.Fatal("Expected error for invalid trigger type, got nil")
		}
	})

	t.Run("Empty Path", func(t *testing.T) {
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

		err := StartFileWatchTrigger(wf)
		if err == nil {
			t.Fatal("Expected error for empty path, got nil")
		}
	})

	t.Run("Empty Events", func(t *testing.T) {
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

		err := StartFileWatchTrigger(wf)
		if err == nil {
			t.Fatal("Expected error for empty events, got nil")
		}
	})

	t.Run("Invalid Path", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:   workflow.TriggerTypeFileWatch,
				Path:   "/nonexistent/path/that/does/not/exist/12345",
				Events: []string{"create"},
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := StartFileWatchTrigger(wf)
		if err == nil {
			t.Fatal("Expected error for invalid path, got nil")
		}
	})
}
