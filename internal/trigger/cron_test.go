package trigger

import (
	"testing"

	"github.com/codecrafted007/autozap/internal/workflow"
)

func TestStartCronTrigger(t *testing.T) {
	t.Run("Invalid Cron Schedule", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "invalid cron schedule",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := StartCronTrigger(wf)
		if err == nil {
			t.Fatal("Expected error for invalid cron schedule, got nil")
		}
	})

	t.Run("Valid Cron Schedule", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "test-workflow",
			Trigger: workflow.Trigger{
				Type:     workflow.TriggerTypeCron,
				Schedule: "* * * * *",
			},
			Actions: []workflow.Action{
				{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
			},
		}

		err := StartCronTrigger(wf)
		if err != nil {
			t.Fatalf("Expected no error for valid cron schedule, got: %v", err)
		}
	})

	t.Run("Standard Cron Expressions", func(t *testing.T) {
		schedules := []string{
			"*/5 * * * *",     // Every 5 minutes
			"0 * * * *",       // Every hour
			"0 0 * * *",       // Every day at midnight
			"0 0 * * 0",       // Every Sunday at midnight
			"0 0 1 * *",       // First day of every month
			"@hourly",         // Predefined schedule
			"@daily",          // Predefined schedule
			"@weekly",         // Predefined schedule
		}

		for _, schedule := range schedules {
			wf := &workflow.Workflow{
				Name: "test-workflow-" + schedule,
				Trigger: workflow.Trigger{
					Type:     workflow.TriggerTypeCron,
					Schedule: schedule,
				},
				Actions: []workflow.Action{
					{Type: workflow.ActionTypeBash, Name: "test", Command: "echo test"},
				},
			}

			err := StartCronTrigger(wf)
			if err != nil {
				t.Errorf("Expected no error for schedule '%s', got: %v", schedule, err)
			}
		}
	})
}
