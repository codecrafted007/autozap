# autozapctl - AutoZap Control Wrapper

Production-ready lifecycle management tool for AutoZap.

## Quick Start

```bash
# Start AutoZap
./autozapctl start

# Check status
./autozapctl status

# View logs
./autozapctl logs -f

# Open dashboard
./autozapctl open

# Stop AutoZap
./autozapctl stop
```

## Installation

### 1. Make executable
```bash
chmod +x autozapctl
```

### 2. Optional: Add to PATH
```bash
# Copy to /usr/local/bin
sudo cp autozapctl /usr/local/bin/

# Or create symlink
sudo ln -s $(pwd)/autozapctl /usr/local/bin/autozapctl
```

### 3. Optional: Install as system service (Linux)
```bash
sudo ./autozapctl install-service
sudo systemctl enable autozap
sudo systemctl start autozap
```

## Commands

### Service Management

#### start
Start the AutoZap agent.

```bash
# Basic start
./autozapctl start

# Custom configuration
./autozapctl start --workflows-dir /opt/workflows --port 9090

# With per-workflow logging
./autozapctl start --log-dir ./logs

# Disable hot-reload
./autozapctl start --no-watch

# Dry run (see what would execute)
./autozapctl start --dry-run

# Foreground mode (for systemd)
./autozapctl start --foreground
```

**Options:**
- `--workflows-dir <dir>` - Workflows directory (default: ./workflows)
- `--db <path>` - Database path (default: ./autozap.db)
- `--port <port>` - HTTP port (default: 8080)
- `--log-dir <dir>` - Per-workflow log directory
- `--no-watch` - Disable hot-reload
- `--dry-run` - Show what would be executed
- `--foreground` - Run in foreground

#### stop
Stop the AutoZap agent gracefully.

```bash
./autozapctl stop
```

#### restart
Restart the AutoZap agent.

```bash
# Restart with same config
./autozapctl restart

# Restart with new config
./autozapctl restart --port 9090
```

#### status
Show running status, uptime, and metrics.

```bash
./autozapctl status
```

Output includes:
- Running status (PID)
- Uptime
- Memory usage
- Configuration file in use
- Active workflow count
- Dashboard URL

### Log Management

#### logs
View application logs.

```bash
# View last 50 lines
./autozapctl logs

# Follow logs (like tail -f)
./autozapctl logs -f
./autozapctl logs --follow
```

### Data Queries

#### history
View workflow execution history.

```bash
# All executions (last 100)
./autozapctl history

# Limit results
./autozapctl history --limit 20

# Specific workflow
./autozapctl history --workflow my-workflow

# Filter by status
./autozapctl history --status failed
```

#### stats
View workflow statistics.

```bash
# Specific workflow
./autozapctl stats my-workflow

# Custom time range
./autozapctl stats my-workflow --days 30
```

#### failures
View recent failures.

```bash
# Last 24 hours
./autozapctl failures

# Last week
./autozapctl failures --hours 168

# Limit results
./autozapctl failures --limit 10
```

### Utilities

#### validate
Validate a workflow file without executing it.

```bash
./autozapctl validate workflows/my-workflow.yaml
```

This checks:
- YAML syntax
- Required fields
- Trigger configuration
- Action types

#### open
Open the dashboard in your default browser.

```bash
./autozapctl open
```

#### install-service
Install AutoZap as a systemd service (Linux only, requires sudo).

```bash
sudo ./autozapctl install-service
```

After installation:
```bash
# Start service
sudo systemctl start autozap

# Enable on boot
sudo systemctl enable autozap

# Check status
sudo systemctl status autozap

# View logs
sudo journalctl -u autozap -f
```

## Configuration

Configuration is loaded from (in priority order):
1. `./.autozaprc` (project directory)
2. `~/.autozaprc` (user home)
3. `/etc/autozap/config` (system-wide)

### Configuration File Format

```bash
# Workflows directory
WORKFLOWS_DIR="./workflows"

# SQLite database path
DATABASE="./autozap.db"

# HTTP server port
HTTP_PORT="8080"

# Log file path
LOG_FILE="./autozap.log"

# PID file path
PID_FILE="./autozap.pid"

# Enable hot-reload
WATCH="true"

# Per-workflow log directory (optional)
LOG_DIR="./logs"
```

### Creating Configuration

```bash
# Copy example config
cp .autozaprc.example .autozaprc

# Edit configuration
vim .autozaprc

# Or create user-specific config
cp .autozaprc.example ~/.autozaprc
```

## Use Cases

### Development

```bash
# Start with hot-reload
./autozapctl start

# Watch logs
./autozapctl logs -f

# Check status frequently
watch -n 5 ./autozapctl status
```

### Production

```bash
# Install as service
sudo ./autozapctl install-service
sudo systemctl enable autozap
sudo systemctl start autozap

# Monitor
sudo systemctl status autozap
sudo journalctl -u autozap -f

# Check application status
./autozapctl status
```

### Testing

```bash
# Validate workflows
./autozapctl validate workflows/*.yaml

# Dry run
./autozapctl start --dry-run

# Check history after tests
./autozapctl history --limit 10
```

### Debugging

```bash
# Start in foreground (see output directly)
./autozapctl start --foreground

# Or start normally and tail logs
./autozapctl start
./autozapctl logs -f

# Check failures
./autozapctl failures

# Check specific workflow stats
./autozapctl stats problematic-workflow
```

## Examples

### Basic Workflow

```bash
# 1. Create workflow
cat > workflows/hello.yaml <<EOF
name: hello-world
trigger:
  type: cron
  schedule: "*/5 * * * *"
actions:
  - type: bash
    name: say-hello
    command: echo "Hello from AutoZap!"
EOF

# 2. Start AutoZap
./autozapctl start

# 3. Watch it run
./autozapctl logs -f

# 4. Check stats
./autozapctl stats hello-world
```

### Multiple Environments

```bash
# Development
./autozapctl start --port 8080 --workflows-dir ./workflows-dev

# Staging
./autozapctl start --port 8081 --workflows-dir ./workflows-staging --db staging.db

# Production (via systemd)
sudo systemctl start autozap
```

### Monitoring Setup

```bash
# Terminal 1: Status dashboard
watch -n 5 './autozapctl status'

# Terminal 2: Live logs
./autozapctl logs -f

# Terminal 3: Web dashboard
./autozapctl open
```

## Troubleshooting

### AutoZap won't start

```bash
# Check if already running
./autozapctl status

# Check logs
./autozapctl logs

# Try starting in foreground to see errors
./autozapctl start --foreground
```

### Port already in use

```bash
# Use different port
./autozapctl start --port 9090

# Or find what's using the port
lsof -i :8080
```

### Stale PID file

```bash
# Remove manually if AutoZap crashed
rm -f autozap.pid

# Start again
./autozapctl start
```

### Database locked

```bash
# Stop AutoZap
./autozapctl stop

# Check for other processes using the database
lsof autozap.db

# Start again
./autozapctl start
```

### Configuration not loading

```bash
# Check which config is being used
./autozapctl status | grep "Config:"

# Verify file exists
ls -la .autozaprc

# Check syntax
bash -n .autozaprc
```

## Best Practices

### 1. Use Configuration Files

Instead of passing options every time:
```bash
# Bad
./autozapctl start --workflows-dir /opt/workflows --port 9090

# Good
echo 'WORKFLOWS_DIR="/opt/workflows"' > .autozaprc
echo 'HTTP_PORT="9090"' >> .autozaprc
./autozapctl start
```

### 2. Monitor Regularly

```bash
# Add to cron for daily reports
0 9 * * * /opt/autozap/autozapctl failures --hours 24 | mail -s "AutoZap Daily Report" admin@example.com
```

### 3. Validate Before Deploying

```bash
# Validate all workflows
for workflow in workflows/*.yaml; do
    ./autozapctl validate "$workflow" || exit 1
done
```

### 4. Use Systemd in Production

```bash
# More reliable than manual start/stop
sudo ./autozapctl install-service
sudo systemctl enable autozap
```

### 5. Backup Database Regularly

```bash
# Daily backup
0 2 * * * cp /opt/autozap/autozap.db /backups/autozap-$(date +\%Y\%m\%d).db
```

## Integration

### Prometheus Monitoring

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'autozap'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboard

Import metrics from `http://localhost:8080/metrics`:
- `autozap_workflow_executions_total`
- `autozap_workflow_execution_duration_seconds`
- `autozap_trigger_fires_total`

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name autozap.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Migration from Demo Scripts

If you were using `demo.sh`:

```bash
# Old way
./demo.sh
./stop-demo.sh
./demo-cli.sh history

# New way
./autozapctl start
./autozapctl stop
./autozapctl history
```

The `autozapctl` wrapper is more powerful and production-ready!

## Support

For issues or questions:
- Check logs: `./autozapctl logs -f`
- View status: `./autozapctl status`
- Validate workflows: `./autozapctl validate <file>`
- Check database: `sqlite3 autozap.db "SELECT * FROM workflow_executions LIMIT 10;"`
