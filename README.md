⚡ AutoZap — Self-Hosted, Event-Driven Automation Engine

AutoZap is a lightweight, local-first, terminal-friendly automation engine written in Go. It allows users to define powerful workflows in YAML that react to events (like cron schedules or file changes) and perform various actions (like running Bash commands, making HTTP requests, or executing custom logic).

Think of it as “Zapier for infra and Bash scripts” — designed for automation without the cloud, providing full control and observability directly on your servers.

🛍 Project Overview & Vision
AutoZap aims to be a powerful yet minimal self-hosted automation tool that:

* Uses plain YAML for workflow definitions.
* Runs any Bash command.
* Integrates with your infrastructure stack.
* Requires no cloud, no lock-in, and no fluff.

🌟 Why AutoZap?
Traditional automation tools often present limitations for infrastructure and local scripting needs:

| Tool      | Problem                                       |
| --------- | --------------------------------------------- |
| Zapier    | Cloud-based, not infra- or Bash-friendly      |
| Node-RED  | IoT/data-focused, limited CLI/Bash support    |
| Ansible   | Too heavy for simple task automation          |
| CRON only | Lacks chaining, logging, reactive flows       |
| n8n       | Too app-centric, complex for local automation |

AutoZap is built for:

* DevOps engineers who want to automate tasks on servers.
* Homelab users or sysadmins who already write shell scripts.
* Anyone who wants CLI-based task automation with observability.

✅ Current MVP Features
**CLI Tool:** A command-line interface built with Cobra.
**Structured Logging:** Blazing-fast, JSON-formatted logs using Zap.
**YAML Workflow Parsing:** Loads and validates workflows defined in YAML files.
**Trigger Execution:**

* CRON Trigger using `robfig/cron/v3`
* File Watch Trigger using `fsnotify/fsnotify`
  **Action Execution:**
* Bash Action: Run arbitrary Bash commands.
* HTTP Action: Make requests with full config.
* Custom Action: Simulated for future extension.

🛠 Project Structure

```
autozap/
├── cmd/                  # CLI entry points
│   └── root.go
│   └── run.go
├── internal/
│   ├── action/           # bash.go, http.go, custom.go
│   ├── engine.go         # (Planned) Agent runtime loop
│   ├── logger/           # Zap logger setup
│   ├── parser/           # YAML loader/validator
│   ├── trigger/          # cron.go, filewatch.go
│   └── workflow/         # YAML schema structs (types.go)
├── workflows/            # Example YAML definitions
├── go.mod
├── go.sum
└── main.go
```

🚀 Getting Started
**Prerequisites:**

* Go 1.17+
* jq (optional, for pretty JSON logs)

**Install:**

```bash
git clone https://github.com/yourusername/autozap.git
cd autozap
go mod tidy
```

**Run Examples:**

Run a Cron-Triggered Workflow:

```bash
go run main.go run workflows/sample.yaml | jq .
```

Run an HTTP Action Workflow:

```bash
go run main.go run workflows/http-check.yaml | jq .
```

Run a File Watch Triggered Workflow:

```bash
mkdir -p test_watch_dir
go run main.go run workflows/file-monitor.yaml | jq .
```

In another terminal:

```bash
touch test_watch_dir/new_file.txt
echo "hello" > test_watch_dir/existing_file.txt
rm test_watch_dir/new_file.txt
```

Run a Custom Action Workflow:

```bash
go run main.go run workflows/custom-action-example.yaml | jq .
```

📋 Example Workflow YAMLs
**sample.yaml**

```yaml
name: "restart-nginx-example"
description: "Restart Nginx every 5 minutes."
trigger:
  type: "cron"
  schedule: "*/5 * * * *"
actions:
  - type: "bash"
    name: "restart-nginx-service"
    command: "echo 'Simulating Nginx restart...' && echo 'Nginx restart successful' && exit 0"
```

**http-check.yaml**

```yaml
name: "check-google-health"
description: "Ping google.com every minute."
trigger:
  type: "cron"
  schedule: "*/1 * * * *"
actions:
  - type: "http"
    name: "ping-google"
    url: "https://www.google.com"
    method: "GET"
    timeout: "5s"
    expect_status: 200
    expect_body_contains: "<title>Google</title>"
```

**file-monitor.yaml**

```yaml
name: "log-file-changes"
description: "Monitor directory for file changes."
trigger:
  type: "filewatch"
  path: "./test_watch_dir"
  events: ["create", "write", "remove"]
actions:
  - type: "bash"
    name: "log-event-detail"
    command: "echo 'File event detected! AutoZap triggered.'"
```

**custom-action-example.yaml**

```yaml
name: "process-daily-report"
description: "Trigger internal custom report generation."
trigger:
  type: "cron"
  schedule: "0 1 * * *"
actions:
  - type: "custom"
    name: "generate-sales-report"
    function_name: "GenerateSalesReport"
    arguments:
      report_type: "daily"
      email_recipients: ["admin@example.com", "manager@example.com"]
      data_source: "production_db"
```

🚣 Roadmap Ahead
**Phase 1: MVP (Done)**

* Core trigger/action engine
* CLI + logging

**Phase 2: Observability**

* Per-workflow log files
* Prometheus metrics
* Error/status tracking

**Phase 3: Persistence**

* Save last run times
* Avoid duplicate executions
* BoltDB/SQLite history

**Phase 4: Plugin Support**

* HTTP trigger
* Notifications (Slack/Telegram)
* Webhook support

**Phase 5: Dashboard** (Optional)

* Web UI for workflow visualization, status, logs

🤝 Contributing

* Open issues for bugs/ideas
* Submit PRs with fixes/improvements

📄 License
MIT License - see LICENSE file for details.
