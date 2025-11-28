# AutoZap Workflow Documentation

## Overview

**AutoZap** is a lightweight, self-hosted, event-driven automation engine written in Go. It serves as a local-first alternative to cloud-based automation platforms like Zapier, designed for DevOps engineers, sysadmins, and homelab users.

### Key Features
- Event-driven workflows (CRON and file watch triggers)
- Multi-action orchestration (bash commands, HTTP requests, custom functions)
- YAML-based workflow configuration
- Structured JSON logging with Uber's Zap
- No cloud dependency - fully self-contained
- Simple and extensible architecture

---

## Application Architecture

### Project Structure

```
autozap/
├── cmd/                          # CLI entry points
│   ├── root.go                   # Root command definition
│   ├── run.go                    # Run subcommand (executes workflows)
│   └── agent.go                  # Agent subcommand (planned feature)
│
├── internal/                     # Core application logic
│   ├── workflow/types.go         # Data structures & types
│   ├── logger/logger.go          # Zap logger configuration
│   ├── parser/parser.go          # YAML parsing & validation
│   ├── trigger/
│   │   ├── cron.go               # CRON-based triggers
│   │   └── filewatch.go          # File system event triggers
│   └── action/
│       ├── bash.go               # Shell command execution
│       └── http.go               # HTTP request actions
│
├── workflows/                    # Example YAML workflows
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
└── go.sum                        # Dependency checksums
```

---

## Core Components

### 1. Entry Point & CLI (main.go, cmd/)

The application uses the **Cobra CLI framework** for command-line interface:

- **main.go**: Initializes the logger and executes CLI commands
- **root.go**: Defines the base `autozap` command
- **run.go**: Implements the `run` subcommand that loads and executes a single workflow file
- **agent.go**: Skeleton for future directory monitoring feature (not yet implemented)

### 2. Workflow Schema (internal/workflow/types.go)

Defines the core data structures:

```go
Workflow
├── Name (string)              # Workflow identifier
├── Description (string)       # Human-readable description
├── Trigger (Trigger)          # What starts the workflow
│   ├── Type (TriggerType)     # "cron" or "filewatch"
│   ├── Schedule (string)      # CRON expression (for cron triggers)
│   ├── Path (string)          # Directory path (for filewatch)
│   └── Events ([]string)      # Events to watch: create, write, remove, rename, chmod
└── Actions ([]Action)         # What to execute when triggered
    ├── Type (ActionType)      # "bash", "http", or "custom"
    ├── Name (string)          # Action identifier
    ├── Command (string)       # For bash actions
    ├── URL, Method, Headers   # For HTTP actions
    └── FunctionName, Arguments # For custom actions
```

### 3. Logging System (internal/logger/logger.go)

- Global Zap logger with production configuration
- JSON output format for structured logging
- ISO8601 timestamps
- Used throughout the application for debug, info, warn, and error logging

### 4. Parser & Validator (internal/parser/parser.go)

Responsibilities:
- Reads workflow YAML files
- Unmarshals into Workflow structs
- Validates workflow structure and required fields
- Validates trigger configurations (schedule for cron, path/events for filewatch)
- Validates actions and their required fields

### 5. Trigger System (internal/trigger/)

#### CRON Trigger (cron.go)
- Uses `robfig/cron/v3` library
- Parses CRON expressions (e.g., `*/5 * * * *` for every 5 minutes)
- Executes all actions sequentially when triggered
- Runs in blocking mode

#### FileWatch Trigger (filewatch.go)
- Uses `fsnotify/fsnotify` library
- Watches specified directory for file system events
- Supports events: create, write, remove, rename, chmod
- Executes all actions when matching event occurs
- Runs in a goroutine with proper lifecycle management

### 6. Action System (internal/action/)

#### Bash Action (bash.go)
- Executes shell commands via `os/exec`
- Captures stdout and stderr
- Logs output and exit codes
- Returns errors on non-zero exit codes

#### HTTP Action (http.go)
- Makes HTTP requests with full method support
- Features:
  - Custom headers
  - Request body support
  - Configurable timeout (context-based)
  - Status code validation (single or list)
  - Body content validation
  - Comprehensive error logging

#### Custom Action (Stub)
- Placeholder for user-defined functions
- Framework exists but not yet implemented

---

## Complete Workflow Execution Flow

### 1. Initialization Phase

```
┌─────────────────────────────────────────────┐
│ User executes:                              │
│ go run main.go run workflows/sample.yaml    │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ main() function                             │
│ - Call logger.InitLogger()                  │
│ - Initialize Zap global logger              │
│ - Set production config & JSON output       │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ cmd.Execute()                               │
│ - Route to 'run' command handler            │
│ - Extract workflow filename from args       │
└────────────────┬────────────────────────────┘
                 │
                 ▼
```

### 2. Workflow Parsing & Validation

```
┌─────────────────────────────────────────────┐
│ parser.ParseWorkflowFile(filename)          │
│ 1. Read YAML file from disk                 │
│ 2. Unmarshal into Workflow struct           │
│ 3. Validate workflow structure              │
│    - Check required fields (name, trigger)  │
│    - Validate trigger type                  │
│    - Validate trigger config                │
│    - Validate actions                       │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Return parsed Workflow or error             │
└────────────────┬────────────────────────────┘
                 │
                 ▼
```

### 3. Trigger Type Routing

```
                 ┌──────────────────┐
                 │ Switch on        │
                 │ Trigger Type     │
                 └────────┬─────────┘
                          │
          ┌───────────────┴───────────────┐
          │                               │
          ▼                               ▼
┌──────────────────────┐      ┌──────────────────────┐
│ TriggerTypeCron      │      │ TriggerTypeFileWatch │
└──────────┬───────────┘      └──────────┬───────────┘
           │                              │
           ▼                              ▼
```

### 4. CRON Trigger Workflow

```
┌─────────────────────────────────────────────┐
│ trigger.StartCronTrigger(workflow)          │
│ 1. Create cron scheduler                    │
│ 2. Parse CRON expression                    │
│ 3. Register callback function               │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Start cron scheduler in background          │
│ - cron.Start()                              │
│ - Block forever with select {}              │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Wait for schedule match...                  │
│ ⏰ Schedule fires                           │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Callback function executes                  │
│ 1. Log workflow name & schedule             │
│ 2. Log trigger timestamp                    │
│ 3. Iterate through actions                  │
└────────────────┬────────────────────────────┘
                 │
                 ▼
         [Execute Actions]
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Wait for next schedule...                   │
└─────────────────────────────────────────────┘
```

### 5. FileWatch Trigger Workflow

```
┌─────────────────────────────────────────────┐
│ trigger.StartFileWatchTrigger(workflow)     │
│ 1. Create fsnotify watcher                  │
│ 2. Add directory path to watch list         │
│ 3. Launch event handling goroutine          │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Goroutine: Monitor file system events       │
│ - Listen on watcher.Events channel          │
│ - Listen on watcher.Errors channel          │
│ - Block forever with select {}              │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Wait for file system event...               │
│ 📁 Event detected (create/write/remove/etc) │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Match event against watched event types     │
│ - Is it in workflow.Trigger.Events?         │
└────────────────┬────────────────────────────┘
                 │
            ┌────┴────┐
            │  Match? │
            └────┬────┘
                 │
            Yes  │  No → Continue waiting
                 ▼
┌─────────────────────────────────────────────┐
│ Log event details                           │
│ - Workflow name                             │
│ - Event type                                │
│ - File path                                 │
│ - Timestamp                                 │
└────────────────┬────────────────────────────┘
                 │
                 ▼
         [Execute Actions]
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Continue monitoring for next event...       │
└─────────────────────────────────────────────┘
```

### 6. Action Execution Flow

```
┌─────────────────────────────────────────────┐
│ For Each Action in workflow.Actions         │
│ (Sequential execution)                      │
└────────────────┬────────────────────────────┘
                 │
                 ▼
         ┌───────────────┐
         │ Switch on     │
         │ Action Type   │
         └───────┬───────┘
                 │
    ┌────────────┼────────────┐
    │            │            │
    ▼            ▼            ▼
┌────────┐  ┌────────┐  ┌─────────┐
│  Bash  │  │  HTTP  │  │ Custom  │
└───┬────┘  └───┬────┘  └────┬────┘
    │           │            │
    ▼           ▼            ▼
```

#### Bash Action Execution

```
┌─────────────────────────────────────────────┐
│ action.ExecuteBashAction(action)            │
│ 1. Log action name and command              │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Create exec.Command()                       │
│ - Use /bin/sh -c                            │
│ - Set command string                        │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Execute command                             │
│ - Capture stdout via CombinedOutput()       │
│ - Wait for completion                       │
└────────────────┬────────────────────────────┘
                 │
            ┌────┴────┐
            │ Success?│
            └────┬────┘
                 │
        ┌────────┴────────┐
        │                 │
       Yes               No
        │                 │
        ▼                 ▼
┌──────────────┐   ┌─────────────────┐
│ Log stdout   │   │ Log error       │
│ Log exit 0   │   │ Log stderr      │
│ Return nil   │   │ Log exit code   │
└──────────────┘   │ Return error    │
                   └─────────────────┘
```

#### HTTP Action Execution

```
┌─────────────────────────────────────────────┐
│ action.ExecuteHttpAction(action)            │
│ 1. Log action name, method, and URL         │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Create HTTP request                         │
│ - http.NewRequestWithContext()              │
│ - Set method and URL                        │
│ - Set request body (if provided)            │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Add custom headers                          │
│ - Iterate through action.Headers            │
│ - Set each header on request                │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Set timeout context                         │
│ - Parse action.Timeout (default 10s)        │
│ - Create context.WithTimeout()              │
└────────────────┬────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────┐
│ Execute HTTP request                        │
│ - http.Client.Do(request)                   │
│ - Read response body                        │
│ - Close response body                       │
└────────────────┬────────────────────────────┘
                 │
            ┌────┴────┐
            │ Success?│
            └────┬────┘
                 │
        ┌────────┴────────┐
        │                 │
       Yes               No
        │                 │
        ▼                 ▼
┌──────────────────┐   ┌─────────────────────┐
│ Validate status  │   │ Log error details   │
│ - Check against  │   │ - Error message     │
│   expect_status  │   │ - Response body     │
└────────┬─────────┘   │ Return error        │
         │             └─────────────────────┘
         ▼
┌──────────────────┐
│ Status valid?    │
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
   Yes       No
    │         │
    ▼         ▼
┌────────┐  ┌─────────────────┐
│ Validate│  │ Log status error│
│ body    │  │ Return error    │
│ content │  └─────────────────┘
│ (if set)│
└────┬───┘
     │
     ▼
┌────────────────────┐
│ Body contains      │
│ expected string?   │
└────────┬───────────┘
         │
    ┌────┴────┐
    │         │
   Yes       No
    │         │
    ▼         ▼
┌────────┐  ┌─────────────────┐
│ Log     │  │ Log body error  │
│ success │  │ Return error    │
│ Return  │  └─────────────────┘
│ nil     │
└────────┘
```

### 7. Complete End-to-End Flow Diagram

```
START
  │
  ▼
┌─────────────────────────────────┐
│ main.go                         │
│ - Initialize Logger             │
│ - Execute CLI Command           │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ run [workflow_file.yaml]        │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ Parse & Validate YAML           │
│ - Read file                     │
│ - Unmarshal to struct           │
│ - Validate structure            │
│ - Validate trigger              │
│ - Validate actions              │
└────────────┬────────────────────┘
             │
             ▼
      ┌──────────────┐
      │ Trigger Type?│
      └──────┬───────┘
             │
   ┌─────────┴─────────┐
   │                   │
   ▼                   ▼
┌──────────┐      ┌──────────────┐
│   CRON   │      │  FileWatch   │
└────┬─────┘      └──────┬───────┘
     │                   │
     ▼                   ▼
┌──────────────┐    ┌─────────────────┐
│ Create       │    │ Create fsnotify │
│ scheduler    │    │ watcher         │
│ Start cron   │    │ Launch goroutine│
│ Block ∞      │    │ Block ∞         │
└──────┬───────┘    └────────┬────────┘
       │                     │
       ▼                     ▼
┌──────────────┐    ┌─────────────────┐
│ Wait for     │    │ Wait for file   │
│ schedule     │    │ system event    │
└──────┬───────┘    └────────┬────────┘
       │                     │
       │ ⏰ Trigger          │ 📁 Event
       │                     │
       └──────────┬──────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Execute Actions│
         │   (Sequential) │
         └────────┬───────┘
                  │
        ┌─────────┴──────────┐
        │                    │
        ▼                    ▼
┌───────────────┐    ┌──────────────┐
│  Bash Action  │    │ HTTP Action  │
│ - Run command │    │ - Make req   │
│ - Log output  │    │ - Validate   │
│ - Handle error│    │ - Log result │
└───────┬───────┘    └──────┬───────┘
        │                   │
        └─────────┬─────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Log Results    │
         │ (JSON to stdout│
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────┐
         │ Wait for next  │
         │ trigger...     │
         └────────────────┘
```

---

## Workflow Configuration

### YAML Structure

```yaml
name: "workflow-name"
description: "Human-readable description"

trigger:
  # Option 1: CRON-based trigger
  type: "cron"
  schedule: "*/5 * * * *"  # Every 5 minutes

  # Option 2: File watch trigger
  # type: "filewatch"
  # path: "/tmp/watch-dir"
  # events: ["create", "write", "remove"]

actions:
  # Bash action example
  - type: "bash"
    name: "run-script"
    command: "echo 'Hello World'"

  # HTTP action example
  - type: "http"
    name: "api-call"
    url: "https://api.example.com/endpoint"
    method: "POST"
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer token123"
    body: '{"key": "value"}'
    timeout: "10s"
    expect_status: [200, 201]
    expect_body_contains: "success"

  # Custom action example (not yet implemented)
  - type: "custom"
    name: "custom-function"
    functionName: "MyFunction"
    arguments:
      param1: "value1"
      param2: "value2"
```

### CRON Schedule Format

Standard 5-field CRON expression:
```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, both 0 and 7 are Sunday)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

Examples:
- `*/1 * * * *` - Every minute
- `*/5 * * * *` - Every 5 minutes
- `0 * * * *` - Every hour at minute 0
- `0 9 * * *` - Every day at 9:00 AM
- `0 */6 * * *` - Every 6 hours
- `0 0 * * 0` - Every Sunday at midnight

### File Watch Events

Supported event types:
- `create` - New file created in watched directory
- `write` - File modified/written
- `remove` - File deleted
- `rename` - File renamed
- `chmod` - File permissions changed

---

## Usage Examples

### Running a Workflow

```bash
# Execute a single workflow file
go run main.go run workflows/sample.yaml

# Or with compiled binary
./autozap run workflows/http-check.yaml
```

### Example: Health Check Monitor

```yaml
name: "health-check-monitor"
description: "Check API health every minute"

trigger:
  type: "cron"
  schedule: "*/1 * * * *"

actions:
  - type: "http"
    name: "check-api"
    url: "https://api.myapp.com/health"
    method: "GET"
    timeout: "5s"
    expect_status: [200]
    expect_body_contains: "healthy"

  - type: "bash"
    name: "log-check"
    command: "echo $(date) - Health check passed >> /var/log/health.log"
```

### Example: File Backup on Change

```yaml
name: "auto-backup"
description: "Backup files when they change"

trigger:
  type: "filewatch"
  path: "/home/user/documents"
  events: ["write", "create"]

actions:
  - type: "bash"
    name: "backup-to-s3"
    command: "aws s3 sync /home/user/documents s3://my-backup-bucket/"

  - type: "http"
    name: "notify-webhook"
    url: "https://hooks.slack.com/services/XXX/YYY/ZZZ"
    method: "POST"
    body: '{"text": "Backup completed"}'
```

---

## Data Flow Summary

```
┌──────────────┐
│  YAML File   │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│  Parser/Validator    │
│  - Read YAML         │
│  - Validate schema   │
│  - Check constraints │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Workflow Struct     │
│  (in memory)         │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Trigger Handler     │
│  - CRON scheduler    │
│  - File watcher      │
└──────┬───────────────┘
       │
       │ Event fired
       ▼
┌──────────────────────┐
│  Action Executor     │
│  - Sequential exec   │
│  - Error handling    │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Bash/HTTP Handler   │
│  - Execute           │
│  - Validate          │
│  - Capture output    │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Logger (Zap)        │
│  - JSON format       │
│  - stdout output     │
└──────────────────────┘
```

---

## Current Status & Limitations

### Implemented Features ✓
- CLI with Cobra framework
- YAML workflow parsing and validation
- CRON trigger scheduling
- File watch trigger with event filtering
- Bash action execution
- HTTP action with full validation
- Structured JSON logging
- Sequential action execution
- Error handling and logging

### Planned Features (Not Yet Implemented)
- Agent command for directory monitoring
- Custom action execution
- Workflow persistence (execution history)
- Conditional action execution
- Template variables in commands
- Parallel action execution
- Retry logic for failed actions
- Prometheus metrics
- Web UI dashboard
- Webhook triggers

### Known Limitations
1. No template/variable substitution in workflow definitions
2. Actions execute sequentially only (no parallelism)
3. No execution history or persistence
4. No retry mechanism for failed actions
5. No conditional logic (if/else)
6. Application blocks forever when running (no graceful shutdown)
7. Single workflow per execution (agent mode not implemented)

---

## Dependencies

### Core Go Libraries
- `github.com/fsnotify/fsnotify` - File system notifications
- `github.com/robfig/cron/v3` - CRON scheduling
- `github.com/spf13/cobra` - CLI framework
- `go.uber.org/zap` - Structured logging
- `gopkg.in/yaml.v3` - YAML parsing

### Standard Library Usage
- `os/exec` - Command execution
- `net/http` - HTTP client
- `context` - Timeout management
- `io` - I/O operations
- `encoding/json` - JSON handling

---

## Extending AutoZap

### Adding a New Trigger Type

1. Define the trigger type in `internal/workflow/types.go`
2. Add validation logic in `internal/parser/parser.go`
3. Create implementation in `internal/trigger/new_trigger.go`
4. Register trigger in `cmd/run.go` switch statement

### Adding a New Action Type

1. Define the action type in `internal/workflow/types.go`
2. Add validation logic in `internal/parser/parser.go`
3. Create implementation in `internal/action/new_action.go`
4. Register action in trigger handlers (cron.go, filewatch.go)

---

## Conclusion

AutoZap provides a clean, extensible foundation for event-driven automation with clear separation of concerns:

- **Triggers** detect when workflows should run
- **Actions** define what should happen
- **Parser** ensures workflows are valid
- **Logger** provides observability
- **CLI** provides user interface

The architecture supports easy extension with new trigger types, action types, and features while maintaining simplicity and reliability.
