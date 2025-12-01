package action

import (
	"testing"

	"github.com/codecrafted007/autozap/internal/logger"
	"github.com/codecrafted007/autozap/internal/workflow"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger()
}

func TestExecuteBashAction(t *testing.T) {
	t.Run("Successful Command Execution", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "test-echo",
			Command: "echo 'Hello World'",
		}

		err := ExecuteBashAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Command With Exit Code Zero", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "test-true",
			Command: "true",
		}

		err := ExecuteBashAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Command With Non-Zero Exit Code", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "test-false",
			Command: "false",
		}

		err := ExecuteBashAction(action)
		if err == nil {
			t.Fatal("Expected error for command with non-zero exit code, got nil")
		}
	})

	t.Run("Command That Produces Output", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "test-output",
			Command: "echo 'stdout message' && echo 'stderr message' >&2",
		}

		err := ExecuteBashAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid Action Type", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeHTTP,
			Name:    "wrong-type",
			Command: "echo test",
		}

		err := ExecuteBashAction(action)
		if err == nil {
			t.Fatal("Expected error for invalid action type, got nil")
		}
	})

	t.Run("Empty Command", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "empty-command",
			Command: "",
		}

		err := ExecuteBashAction(action)
		if err == nil {
			t.Fatal("Expected error for empty command, got nil")
		}
	})

	t.Run("Command Not Found", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "nonexistent-command",
			Command: "this-command-does-not-exist-12345",
		}

		err := ExecuteBashAction(action)
		if err == nil {
			t.Fatal("Expected error for nonexistent command, got nil")
		}
	})

	t.Run("Multi-line Command", func(t *testing.T) {
		action := &workflow.Action{
			Type: workflow.ActionTypeBash,
			Name: "multiline",
			Command: `
echo "Line 1"
echo "Line 2"
echo "Line 3"
`,
		}

		err := ExecuteBashAction(action)
		if err != nil {
			t.Fatalf("Expected no error for multi-line command, got: %v", err)
		}
	})

	t.Run("Command With Environment Variables", func(t *testing.T) {
		action := &workflow.Action{
			Type:    workflow.ActionTypeBash,
			Name:    "env-test",
			Command: "echo $HOME",
		}

		err := ExecuteBashAction(action)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
}
