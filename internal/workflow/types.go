package workflow

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Workflow struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Trigger     Trigger  `yaml:"trigger"`
	Actions     []Action `yaml:"actions"`
}

type TriggerType string

const (
	TriggerTypeCron      TriggerType = "cron"
	TriggerTypeFileWatch TriggerType = "filewatch"
)

func (tt *TriggerType) UnmarshalYaml(value *yaml.Node) error {
	var s string

	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("failed to decode TriggerType: %w", err)
	}

	// Validate the string against our predefined constants
	switch strings.ToLower(s) {
	case string(TriggerTypeCron):
		*tt = TriggerTypeCron
	case string(TriggerTypeFileWatch):
		*tt = TriggerTypeFileWatch
	default:
		return fmt.Errorf("invalid trigger type '%s'. Must be one of: %s, %s", s, TriggerTypeCron, TriggerTypeFileWatch)
	}
	return nil
}

type Trigger struct {
	Type     TriggerType `yaml:"type"`               //custom TriggerType enum
	Schedule string      `yaml:"schedule,omitempty"` // Mandatory for cron, omitted otherwise
	Path     string      `yaml:"path,omitempty"`     // Will be used for filewatch trigger later
	Events   []string    `yaml:"events,omitempty"`   // for filewatch, omitted otherwise
}

// ActionType defines the type of action to be performed (e.g., "bash", "http", etc.)
// This is a placeholder for future action types, such as HTTP requests, file operations, etc
// This acts as a enum
type ActionType string

const (
	ActionTypeBash   ActionType = "bash"
	ActionTypeHTTP   ActionType = "http"
	ActionTypeCustom ActionType = "custom" // For user-defined actions
)

// This allows yaml parser to convert string from yaml file directly to ActionType
func (at *ActionType) UnmarshalYaml(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("failed to decode ActionType: %w", err)
	}

	switch strings.ToLower(s) {
	case string(ActionTypeBash):
		*at = ActionTypeBash
	case string(ActionTypeHTTP):
		*at = ActionTypeHTTP
	case string(ActionTypeCustom):
		*at = ActionTypeCustom
	default:
		return fmt.Errorf("invalid action type '%s'. Must be one of: %s, %s, %s", s, ActionTypeBash, ActionTypeHTTP, ActionTypeCustom)
	}
	return nil
}

func (at ActionType) String() string {
	return string(at)
}

func (tt TriggerType) String() string {
	return string(tt)
}

type Action struct {
	Type ActionType `yaml:"type"`
	Name string     `yaml:"name"`
	// Field for ActionType bash
	Command string `yaml:"command,omitempty"` // For bash actions

	//Field for ActionType Http

	URL                string            `yaml:"url,omitempty"`
	Method             string            `yaml:"method,omitempty"`
	Headers            map[string]string `yaml:"headers,omitempty"`            // e.g., {"Content-Type": "application/json"}
	Body               string            `yaml:"body,omitempty"`               // For HTTP actions
	Timeout            string            `yaml:"timeout,omitempty"`            // e.g., "10s", will be parsed to time.Duration
	ExpectStatus       interface{}       `yaml:"expectStatus,omitempty"`       // Can be int or []int for multiple valid codes
	ExpectBodyContains string            `yaml:"expectBodyContains,omitempty"` // For HTTP actions

	// Fields for ActionTypeCustom

	FunctionName string                 `yaml:"functionName,omitempty"`
	Arguments    map[string]interface{} `yaml:"arguments,omitempty"` // using interface for flexibility
}
