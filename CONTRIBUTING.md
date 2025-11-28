# Contributing to AutoZap

First off, thank you for considering contributing to AutoZap! It's people like you that make AutoZap such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our commitment to providing a welcoming and inspiring community for all. Please be respectful and constructive in your interactions.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (workflow YAML files, commands run, etc.)
- **Describe the behavior you observed** and what you expected to see
- **Include logs** (AutoZap outputs JSON logs - please include relevant log entries)
- **Specify your environment**: Go version, OS, AutoZap version

**Example bug report:**

```markdown
## Bug: File watcher doesn't detect rename events

**Steps to reproduce:**
1. Start AutoZap with `workflows/file-monitor.yaml`
2. Rename a file in the watched directory
3. No workflow triggers

**Expected:** Workflow should trigger on rename event

**Actual:** No trigger occurs

**Environment:**
- AutoZap version: main branch (commit abc123)
- Go version: 1.21.0
- OS: macOS 14.0
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful** to most AutoZap users
- **List examples** of how it would work

**Example enhancement:**

```markdown
## Enhancement: Add Slack notification action

**Description:**
Add a native Slack action type that handles Slack webhook formatting automatically.

**Why it's useful:**
Currently users need to use HTTP actions with manual JSON formatting. A dedicated Slack action would simplify workflow definitions.

**Example usage:**
```yaml
actions:
  - type: "slack"
    name: "notify-team"
    webhook_url: "${SECRET.slack_webhook}"
    message: "Deployment completed successfully"
    channel: "#devops"
\`\`\`
```

### Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Follow the code style** (run `go fmt` and `golangci-lint`)
3. **Add tests** if you're adding functionality
4. **Update documentation** if you're changing behavior
5. **Ensure tests pass** (`go test ./...`)
6. **Write a clear commit message**

**Pull Request Process:**

1. Update the README.md with details of changes if applicable
2. Update the CHANGELOG.md (if we have one) with your changes
3. The PR will be merged once you have the sign-off of a maintainer

**Good PR titles:**
- `feat: add retry logic to HTTP actions`
- `fix: file watcher not detecting chmod events`
- `docs: improve quick start guide`
- `test: add unit tests for parser validation`

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/autozap.git
cd autozap

# Install dependencies
go mod download

# Build
go build -o autozap .

# Run tests
go test -v ./...

# Run linter
golangci-lint run
```

## Code Style

- Follow standard Go conventions (use `go fmt`)
- Write meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused
- Use structured logging with Zap logger

**Example:**

```go
// ExecuteBashAction runs a bash command and captures output
func ExecuteBashAction(action workflow.Action) error {
    logger.L.Info("executing bash action",
        zap.String("name", action.Name),
        zap.String("command", action.Command),
    )

    // Implementation...
}
```

## Project Structure

Understanding the codebase:

```
autozap/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Base command
│   ├── run.go             # Run workflow command
│   └── agent.go           # Agent mode (future)
├── internal/
│   ├── workflow/          # Data structures
│   ├── parser/            # YAML parsing & validation
│   ├── trigger/           # Trigger implementations
│   │   ├── cron.go
│   │   └── filewatch.go
│   ├── action/            # Action implementations
│   │   ├── bash.go
│   │   └── http.go
│   └── logger/            # Zap logger setup
├── workflows/             # Example workflows
└── main.go               # Entry point
```

## Adding New Features

### Adding a New Trigger Type

1. Define the trigger type in `internal/workflow/types.go`:
```go
const (
    TriggerTypeWebhook TriggerType = "webhook"
)
```

2. Add validation in `internal/parser/parser.go`:
```go
case workflow.TriggerTypeWebhook:
    if wf.Trigger.Port == 0 {
        return fmt.Errorf("webhook trigger requires 'port' field")
    }
```

3. Implement trigger in `internal/trigger/webhook.go`:
```go
package trigger

func StartWebhookTrigger(wf *workflow.Workflow) error {
    // Implementation
}
```

4. Register in `cmd/run.go`:
```go
case workflow.TriggerTypeWebhook:
    err = trigger.StartWebhookTrigger(workflow)
```

### Adding a New Action Type

Similar process - modify `types.go`, `parser.go`, create `internal/action/newaction.go`, and register in trigger handlers.

## Testing

- Write unit tests for new functionality
- Use table-driven tests for validation logic
- Mock external dependencies (HTTP calls, file system)

**Example test:**

```go
func TestParseCronTrigger(t *testing.T) {
    tests := []struct {
        name    string
        yaml    string
        wantErr bool
    }{
        {
            name: "valid cron",
            yaml: `
name: "test"
trigger:
  type: "cron"
  schedule: "*/5 * * * *"
`,
            wantErr: false,
        },
        {
            name: "missing schedule",
            yaml: `
name: "test"
trigger:
  type: "cron"
`,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := parser.ParseWorkflow([]byte(tt.yaml))
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseWorkflow() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Documentation

- Update README.md for user-facing changes
- Add examples in `workflows/` directory
- Update `autozap_workflow.md` for architectural changes
- Include code comments for complex logic

## Community

- Be respectful and constructive
- Help others in issues and discussions
- Share your use cases and workflows
- Provide feedback on PRs

## Questions?

Feel free to open an issue with the `question` label or start a discussion on GitHub Discussions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to AutoZap! 🚀
