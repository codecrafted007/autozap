# AutoZap Quick Reference

## autozapctl - The Control Wrapper

### Most Common Commands

```bash
# Start AutoZap
./autozapctl start

# Stop AutoZap
./autozapctl stop

# Check if running
./autozapctl status

# View logs
./autozapctl logs -f

# Restart
./autozapctl restart
```

### Data Queries

```bash
# Recent executions
./autozapctl history

# Workflow stats
./autozapctl stats <workflow-name>

# Recent failures
./autozapctl failures
```

### Utilities

```bash
# Validate workflow file
./autozapctl validate <file.yaml>

# Open dashboard in browser
./autozapctl open

# Show help
./autozapctl help
```

## Configuration

Create `.autozaprc` in project directory:

```bash
WORKFLOWS_DIR="./workflows"
DATABASE="./autozap.db"
HTTP_PORT="8080"
LOG_FILE="./autozap.log"
WATCH="true"
```

## Start Options

```bash
# Custom port
./autozapctl start --port 9090

# Custom workflows directory
./autozapctl start --workflows-dir /opt/workflows

# Per-workflow logging
./autozapctl start --log-dir ./logs

# Disable hot-reload
./autozapctl start --no-watch

# Dry run
./autozapctl start --dry-run
```

## Dashboard URLs

When running on default port 8080:

- **Dashboard**: http://localhost:8080/dashboard
- **Metrics**: http://localhost:8080/metrics
- **Health**: http://localhost:8080/health
- **API**: http://localhost:8080/api/workflows/active

## Production Deployment (Linux)

```bash
# Install as systemd service
sudo ./autozapctl install-service

# Enable on boot
sudo systemctl enable autozap

# Start service
sudo systemctl start autozap

# Check status
sudo systemctl status autozap

# View logs
sudo journalctl -u autozap -f
```

## Workflow File Example

```yaml
name: my-workflow
description: Example workflow
trigger:
  type: cron
  schedule: "*/5 * * * *"
actions:
  - type: bash
    name: my-action
    command: echo "Hello"
    retry:
      maxAttempts: 3
      initialDelay: "1s"
```

## Troubleshooting

```bash
# View logs
./autozapctl logs -f

# Check status
./autozapctl status

# Validate workflow
./autozapctl validate workflows/my-workflow.yaml

# Check database
sqlite3 autozap.db "SELECT * FROM workflow_executions LIMIT 5;"

# Kill stale process
rm -f autozap.pid
./autozapctl start
```

## Migration from demo.sh

| Old Command | New Command |
|-------------|-------------|
| `./demo.sh` | `./autozapctl start` |
| `./stop-demo.sh` | `./autozapctl stop` |
| `./demo-cli.sh history` | `./autozapctl history` |
| `./demo-cli.sh stats <workflow>` | `./autozapctl stats <workflow>` |
| `./demo-cli.sh failures` | `./autozapctl failures` |

## Key Features

✅ **Service Management** - start, stop, restart, status
✅ **Configuration Files** - .autozaprc support
✅ **Auto Database Path** - no need to specify --db
✅ **Health Checks** - built-in status monitoring
✅ **Log Management** - easy log viewing
✅ **Systemd Integration** - production-ready
✅ **Cross-Platform** - Linux & macOS

## Documentation

- **Full Guide**: [AUTOZAPCTL.md](./AUTOZAPCTL.md)
- **Demo Guide**: [DEMO.md](./DEMO.md)
- **Quick Start**: [QUICK_START.md](./QUICK_START.md)
