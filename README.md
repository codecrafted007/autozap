# AutoZap

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-61.2%25-brightgreen.svg)](#-testing)
[![Production Workflows](https://img.shields.io/badge/workflows-7%20production--ready-success.svg)](workflows/)
[![Go Report Card](https://goreportcard.com/badge/github.com/codecrafted007/autozap)](https://goreportcard.com/report/github.com/codecrafted007/autozap)
[![CI](https://github.com/codecrafted007/autozap/workflows/CI/badge.svg)](https://github.com/codecrafted007/autozap/actions)

> A lightweight, self-hosted workflow automation engine built in Go. Event-driven infrastructure automation without the cloud dependency.

**Think "Zapier for DevOps" or "Cron on Steroids"** - schedule tasks, watch files, chain actions, and automate your infrastructure with simple YAML configs.

[Features](#-features) • [Production Workflows](#-production-workflows) • [Quick Start](#-quick-start) • [Examples](#-quick-examples) • [Architecture](#%EF%B8%8F-architecture) • [Testing](#-testing) • [Documentation](#-documentation)

---

## 🎯 Why AutoZap?

Modern DevOps teams need automation that's:
- **Self-hosted**: Your infrastructure, your rules, no cloud vendor lock-in
- **Lightweight**: Single Go binary, minimal dependencies, low resource footprint
- **Event-driven**: Respond to cron schedules and file system changes in real-time
- **Observable**: Structured JSON logging with Uber's Zap for production debugging
- **Simple**: YAML-based workflows that your entire team can read and write
- **Extensible**: Plugin architecture for custom triggers and actions

Perfect for: API health monitoring, automated backups, log rotation, deployment automation, infrastructure monitoring, file processing pipelines.

**Includes 7 production-ready workflows** covering Docker cleanup, SSL monitoring, database backups, disk space alerts, system health checks, API monitoring, and log rotation.

---
✅ Current Features
**CLI Tool:** A command-line interface built with Cobra.
**Structured Logging:** Blazing-fast, JSON-formatted logs using Zap.
**YAML Workflow Parsing:** Loads and validates workflows defined in YAML files.
**Trigger Execution:**

## ✨ Features

### Triggers
- **⏰ CRON Scheduling**: Standard cron expressions for time-based automation
- **📁 File System Watching**: React to file create, write, delete, rename, and permission changes
- *(Coming soon)* Webhook triggers, message queue consumers

### Actions
- **💻 Bash Commands**: Execute shell scripts with full stdout/stderr capture
- **🌐 HTTP Requests**: Make API calls with custom headers, body, timeout, and response validation
- **🔌 Custom Functions**: Extensible framework for plugin-based actions
- **⛓️ Sequential Execution**: Reliable, ordered action chains with comprehensive error logging

### Observability
- **📊 Structured Logging**: Production-grade JSON logs with Uber's Zap
- **🚨 Error Handling**: Detailed error messages with exit codes and response bodies
- **📈 Execution Tracking**: Full visibility into workflow triggers and action results

---

## 🎯 Production Workflows

AutoZap includes **7 battle-tested workflows** for real-world DevOps scenarios. These aren't toy examples - they're production-ready automation that solves actual infrastructure problems.

| Workflow | Purpose | Schedule | Impact |
|----------|---------|----------|--------|
| 🐳 **[docker-cleanup.yaml](workflows/docker-cleanup.yaml)** | Clean Docker resources (containers, images, volumes, networks) | Weekly | Prevents disk space issues |
| 🔒 **[ssl-cert-monitor.yaml](workflows/ssl-cert-monitor.yaml)** | Monitor SSL certificate expiry with alerts | Daily | Prevents certificate-related outages |
| 💾 **[postgres-backup.yaml](workflows/postgres-backup.yaml)** | Automated database backups with compression & retention | Daily | Data protection & disaster recovery |
| 💽 **[disk-space-alert.yaml](workflows/disk-space-alert.yaml)** | Monitor disk usage with threshold alerts | Every 15 min | Proactive capacity management |
| 🏥 **[system-health-check.yaml](workflows/system-health-check.yaml)** | CPU, memory, load, swap, and service monitoring | Every 10 min | System reliability & performance |
| 🌐 **[api-health-check.yaml](workflows/api-health-check.yaml)** | API health checks with response time monitoring | Every 5 min | Service availability & SLA compliance |
| 📝 **[log-rotation.yaml](workflows/log-rotation.yaml)** | Automated log rotation, compression, and cleanup | Daily | Storage management & cost optimization |

### What Makes These Production-Ready?

✅ **Error Handling** - Proper exit codes, validation, and fallback logic
✅ **Configurable Thresholds** - Adjust alerts and schedules for your environment
✅ **Integration Ready** - Slack, PagerDuty, OpsGenie webhooks included
✅ **Best Practices** - Follows SRE principles for monitoring and automation
✅ **Well Documented** - Comprehensive README with customization guide

> 📚 **[View All Workflows →](workflows/README.md)** | Includes setup instructions, customization guide, and troubleshooting

---

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/codecrafted007/autozap.git
cd autozap

# Build the binary
go build -o autozap .

# Or install directly
go install github.com/codecrafted007/autozap@latest
```

### Your First Workflow

Create a simple health check monitor:

```yaml
# health-check.yaml
name: "api-health-monitor"
description: "Monitor API health every 5 minutes"

trigger:
  type: "cron"
  schedule: "*/5 * * * *"  # Every 5 minutes

actions:
  - type: "http"
    name: "check-api"
    url: "https://api.example.com/health"
    method: "GET"
    timeout: "10s"
    expect_status: [200]
    expect_body_contains: "healthy"

  - type: "bash"
    name: "log-status"
    command: "echo $(date) - API health check passed >> /var/log/health.log"
```

Run it:

```bash
./autozap run health-check.yaml
```

That's it! AutoZap will check your API every 5 minutes and log the results.

---

## 📚 Quick Examples

Below are simplified examples to get you started. For **production-ready workflows**, see the [Production Workflows](#-production-workflows) section above.

### 🐳 Docker Container Cleanup
```yaml
name: "docker-cleanup"
description: "Remove unused Docker images weekly"

trigger:
  type: "cron"
  schedule: "0 2 * * 0"  # Sunday 2 AM

actions:
  - type: "bash"
    name: "cleanup-images"
    command: "docker image prune -af --filter until=168h"

  - type: "bash"
    name: "cleanup-volumes"
    command: "docker volume prune -f"
```

### 🔒 SSL Certificate Monitoring
```yaml
name: "ssl-cert-check"
description: "Check SSL certificate expiration daily"

trigger:
  type: "cron"
  schedule: "0 9 * * *"  # Daily at 9 AM

actions:
  - type: "bash"
    name: "check-expiry"
    command: |
      expiry=$(echo | openssl s_client -servername example.com -connect example.com:443 2>/dev/null | openssl x509 -noout -enddate | cut -d= -f2)
      echo "Certificate expires: $expiry"

  - type: "http"
    name: "alert-slack"
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    method: "POST"
    body: '{"text": "SSL certificate check completed"}'
```

### 💾 Automated Backups on File Changes
```yaml
name: "auto-backup"
description: "Backup files to S3 when they change"

trigger:
  type: "filewatch"
  path: "/home/user/important-docs"
  events: ["write", "create"]

actions:
  - type: "bash"
    name: "sync-to-s3"
    command: "aws s3 sync /home/user/important-docs s3://backup-bucket/docs/"

  - type: "bash"
    name: "log-backup"
    command: "echo $(date) - Backup completed >> /var/log/backups.log"
```

### 🗄️ Database Backup Automation
```yaml
name: "postgres-backup"
description: "Backup PostgreSQL database nightly"

trigger:
  type: "cron"
  schedule: "0 1 * * *"  # 1 AM daily

actions:
  - type: "bash"
    name: "dump-database"
    command: |
      BACKUP_FILE="/backups/postgres-$(date +%Y%m%d).sql.gz"
      pg_dump -U postgres mydb | gzip > $BACKUP_FILE

  - type: "bash"
    name: "upload-to-s3"
    command: "aws s3 cp /backups/postgres-$(date +%Y%m%d).sql.gz s3://db-backups/"

  - type: "bash"
    name: "cleanup-old-backups"
    command: "find /backups -name 'postgres-*.sql.gz' -mtime +7 -delete"
```

### 📊 API Endpoint Monitoring with Alerting
```yaml
name: "api-monitor"
description: "Monitor critical API endpoints"

trigger:
  type: "cron"
  schedule: "*/2 * * * *"  # Every 2 minutes

actions:
  - type: "http"
    name: "check-endpoint"
    url: "https://api.example.com/v1/status"
    method: "GET"
    timeout: "5s"
    expect_status: [200, 201]

  - type: "http"
    name: "alert-on-failure"
    url: "https://api.pagerduty.com/incidents"
    method: "POST"
    headers:
      Authorization: "Token token=YOUR_TOKEN"
      Content-Type: "application/json"
    body: '{"incident": {"type": "incident", "title": "API endpoint down"}}'
```

### 📝 Log Rotation and Cleanup
```yaml
name: "log-rotation"
description: "Rotate and compress logs daily"

trigger:
  type: "cron"
  schedule: "0 0 * * *"  # Midnight daily

actions:
  - type: "bash"
    name: "rotate-logs"
    command: |
      cd /var/log/myapp
      mv app.log app-$(date +%Y%m%d).log
      gzip app-$(date +%Y%m%d).log
      touch app.log

  - type: "bash"
    name: "cleanup-old-logs"
    command: "find /var/log/myapp -name '*.log.gz' -mtime +30 -delete"
```

### 🚀 Deployment Notification
```yaml
name: "deployment-webhook"
description: "Watch for deployment files and notify team"

trigger:
  type: "filewatch"
  path: "/deployments"
  events: ["create"]

actions:
  - type: "bash"
    name: "read-deployment-info"
    command: "cat /deployments/*.json"

  - type: "http"
    name: "notify-team"
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK"
    method: "POST"
    body: '{"text": "🚀 New deployment detected!"}'
```

### 💽 Disk Space Monitoring
```yaml
name: "disk-space-alert"
description: "Alert when disk space is low"

trigger:
  type: "cron"
  schedule: "0 */4 * * *"  # Every 4 hours

actions:
  - type: "bash"
    name: "check-disk-space"
    command: |
      usage=$(df -h / | tail -1 | awk '{print $5}' | sed 's/%//')
      if [ $usage -gt 80 ]; then
        echo "WARNING: Disk usage at ${usage}%"
        exit 1
      fi
      echo "OK: Disk usage at ${usage}%"

  - type: "http"
    name: "send-alert"
    url: "https://api.opsgenie.com/v2/alerts"
    method: "POST"
    headers:
      Authorization: "GenieKey YOUR_KEY"
    body: '{"message": "Disk space critical", "priority": "P1"}'
```

> 🚀 **Want production-ready workflows?** Check out our [7 battle-tested examples](workflows/README.md) with full documentation, error handling, and real-world use cases.

---

## 🏗️ Architecture

AutoZap follows a clean, modular architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────┐
│                  CLI (Cobra)                    │
│              main.go, cmd/run.go                │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│            Parser & Validator                   │
│    YAML → Workflow Struct → Validation          │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
        ┌────────────────────────┐
        │   Trigger Dispatcher   │
        └────────┬───────────────┘
                 │
        ┌────────┴─────────┐
        │                  │
        ▼                  ▼
┌──────────────┐   ┌──────────────┐
│ CRON Trigger │   │ File Watcher │
│  (robfig)    │   │  (fsnotify)  │
└──────┬───────┘   └──────┬───────┘
       │                  │
       └────────┬─────────┘
                │
                ▼
      ┌─────────────────┐
      │ Action Executor │
      │  (Sequential)   │
      └────────┬────────┘
               │
        ┌──────┴──────┐
        │             │
        ▼             ▼
┌─────────────┐  ┌──────────┐
│ Bash Action │  │ HTTP     │
│  (os/exec)  │  │ Action   │
└─────────────┘  └──────────┘
        │             │
        └──────┬──────┘
               │
               ▼
      ┌────────────────┐
      │ Zap Logger     │
      │ (JSON Output)  │
      └────────────────┘
```

### Key Design Decisions

**Why Go?**
- Native concurrency with goroutines for parallel workflow management
- Single binary deployment - no runtime dependencies
- Strong standard library (os/exec, net/http, context)
- Fast compilation and excellent performance
- Type safety for workflow validation

**Event-Driven Architecture**
- More powerful than simple cron jobs
- Real-time response to file system changes
- Extensible trigger system for future webhook/queue support
- Non-blocking execution with goroutines

**Modular Design**
- Clear separation: Parser → Triggers → Actions → Logger
- Easy to extend with new trigger types (webhooks, queues)
- Easy to extend with new action types (database, SSH, etc.)
- Highly testable components

> 📖 For detailed workflow documentation, see [autozap_workflow.md](autozap_workflow.md)

---

## 🛠️ Development

### Prerequisites
- Go 1.21 or higher
- Make (optional)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/codecrafted007/autozap.git
cd autozap

# Install dependencies
go mod download

# Build
go build -o autozap .

# Run tests
go test -v ./...

# Run with race detection
go test -race ./...

# Lint (requires golangci-lint)
golangci-lint run
```

### Project Structure

```
autozap/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── run.go             # Run workflow command
│   └── agent.go           # Agent mode (planned)
├── internal/
│   ├── workflow/          # Workflow types and structures
│   ├── parser/            # YAML parser and validator
│   ├── trigger/           # Trigger implementations
│   │   ├── cron.go       # CRON trigger
│   │   └── filewatch.go  # File watcher trigger
│   ├── action/            # Action implementations
│   │   ├── bash.go       # Bash command action
│   │   └── http.go       # HTTP request action
│   └── logger/            # Zap logger setup
├── workflows/             # Example workflows
├── main.go               # Application entry point
└── go.mod                # Go module definition
```

---

## 🧪 Testing

AutoZap maintains **61.2% test coverage** with comprehensive unit tests across all core packages.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage by Package

| Package | Coverage | Test Files |
|---------|----------|------------|
| `internal/parser` | 89.3% | `parser_test.go` |
| `internal/logger` | 90.0% | `logger_test.go` |
| `internal/action` | 84.6% | `bash_test.go`, `http_test.go` |
| `internal/trigger` | 31.7% | `cron_test.go`, `filewatch_test.go` |
| `internal/workflow` | 10.5% | `types_test.go` |
| **Overall** | **61.2%** | **7 test files** |

### What's Tested

✅ **Parser Package** (89.3%)
- YAML workflow file parsing
- Workflow validation (triggers, actions, fields)
- Error handling for invalid configurations
- Support for all trigger and action types

✅ **Action Package** (84.6%)
- Bash command execution (success, failure, exit codes)
- HTTP requests (GET, POST, PUT, DELETE)
- HTTP response validation (status codes, body content)
- Timeout handling and custom headers
- Error scenarios and edge cases

✅ **Logger Package** (90.0%)
- Logger initialization
- Structured logging functionality
- Error handling for uninitialized logger

✅ **Trigger Package** (31.7%)
- CRON trigger validation
- File watch trigger validation
- Invalid configuration handling

### Test Structure

Tests follow Go conventions with test files located alongside source files:

```
internal/
├── action/
│   ├── bash.go
│   ├── bash_test.go      # 9 test cases
│   ├── http.go
│   └── http_test.go      # 16 test cases
├── parser/
│   ├── parser.go
│   └── parser_test.go    # 35+ test cases
└── ...
```

### Writing New Tests

When contributing, ensure:
1. Test files are named `*_test.go`
2. Test functions start with `Test`
3. Use table-driven tests for multiple scenarios
4. Include both positive and negative test cases
5. Run `go test ./...` before submitting PRs

Example test structure:
```go
func TestMyFunction(t *testing.T) {
    t.Run("Success Case", func(t *testing.T) {
        // Test implementation
    })

    t.Run("Error Case", func(t *testing.T) {
        // Test error handling
    })
}
```

---

## 🚦 Project Status

**Alpha Release** - Core functionality is working and stable for personal use. Not yet recommended for mission-critical production workloads.

### Implemented ✅
- CRON-based scheduling with robfig/cron
- File system watching with fsnotify
- Bash command execution with full output capture
- HTTP requests with validation (status codes, body matching)
- Structured JSON logging with Uber's Zap
- YAML workflow parsing and validation
- Sequential action execution with error handling

```bash
git clone https://github.com/codecrafted007/autozap.git
cd autozap
go mod tidy
```

### Roadmap 🗓️
- [ ] **Agent Mode**: Monitor directory for multiple workflows
- [ ] **Workflow State**: Track execution history in SQLite
- [ ] **Templating**: Variable substitution and dynamic values
- [ ] **Retry Logic**: Automatic retries with exponential backoff
- [ ] **Conditionals**: Skip actions based on previous results
- [ ] **Webhook Trigger**: HTTP endpoint to trigger workflows
- [ ] **Prometheus Metrics**: Expose workflow metrics
- [ ] **Web UI**: Dashboard for workflow management
- [ ] **Plugin System**: External action/trigger plugins
- [ ] **Secrets Management**: Encrypted credential storage

---

## 📋 Documentation

- **[Production Workflows](workflows/README.md)** - 7 battle-tested workflows with setup guide
- **[Workflow Documentation](autozap_workflow.md)** - Complete workflow execution guide
- **[Examples Directory](workflows/)** - All workflow YAML files
- **[Contributing](CONTRIBUTING.md)** - How to contribute to AutoZap

---

## 🤝 Contributing

Contributions are welcome! Whether it's:
- 🐛 Bug reports and fixes
- ✨ New features or triggers/actions
- 📝 Documentation improvements
- 💡 Architecture suggestions

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with these excellent libraries:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications
- [cron](https://github.com/robfig/cron) - CRON scheduling

---

## 💬 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/codecrafted007/autozap/issues)
- **Discussions**: [GitHub Discussions](https://github.com/codecrafted007/autozap/discussions)

---

<p align="center">
  <sub>Built with ❤️ using Go</sub>
</p>
