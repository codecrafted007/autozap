# AutoZap Demo Guide

This guide helps you quickly set up and run AutoZap locally to see the dashboard and workflow execution in action.

## Quick Start

### Prerequisites

- Go 1.21 or higher installed
- Linux or macOS operating system
- Terminal access

### Running the Demo

1. **Start the demo:**
   ```bash
   ./demo.sh
   ```

   This script will:
   - Build the AutoZap binary
   - Clean up any old demo data
   - Start AutoZap in agent mode
   - Load all workflows from `./workflows/` directory
   - Open the dashboard in your browser (if possible)

2. **Access the Dashboard:**

   Open your browser to: **http://localhost:8080/dashboard**

   The dashboard shows:
   - Active workflows with real-time status
   - Execution counts (total runs, successes, failures)
   - Success rate percentages with color coding
   - Last error messages (if any)
   - Next scheduled execution times with countdown timers
   - Auto-refresh every 10 seconds

3. **Stop the demo:**
   ```bash
   ./stop-demo.sh
   ```

## What You'll See

### Workflows Loaded

The demo loads all workflows from the `workflows/` directory, including:

- **demo-quick** - Runs every minute for quick dashboard population
- **retry-example** - Demonstrates retry logic with exponential backoff
- **docker-cleanup** - Sample Docker cleanup workflow
- **system-health-check** - System monitoring workflow
- **api-health-check** - HTTP endpoint monitoring
- And more...

### Dashboard Features

1. **Active Workflow Cards**
   - Workflow name and description
   - Trigger type and schedule
   - Total runs with success/failure breakdown
   - Color-coded success rate:
     - 🟢 Green: > 90% success rate
     - 🟡 Yellow: 70-90% success rate
     - 🔴 Red: < 70% success rate
   - Last error message (if any)
   - Next execution countdown timer

2. **Real-time Updates**
   - Dashboard auto-refreshes every 10 seconds
   - Countdown timers update every second
   - See new executions as they happen

3. **Workflow History**
   - Recent execution history
   - Execution duration
   - Status indicators
   - Detailed error messages

## Available Endpoints

### Web UI
- **Dashboard**: http://localhost:8080/dashboard
- **Prometheus Metrics**: http://localhost:8080/metrics
- **Health Check**: http://localhost:8080/health
- **Status**: http://localhost:8080/status

### REST APIs
- **Active Workflows**: http://localhost:8080/api/workflows/active
- **Execution History**: http://localhost:8080/api/workflows/history
- **Workflow Stats**: http://localhost:8080/api/workflows/stats
- **Recent Failures**: http://localhost:8080/api/workflows/failures

## Command Line Interface

While the demo is running, you can use these CLI commands:

**Using the demo CLI helper** (recommended - automatically uses correct database path):

```bash
# View execution history
./demo-cli.sh history

# View history with limit
./demo-cli.sh history --limit 20

# View workflow statistics
./demo-cli.sh stats demo-quick

# View recent failures
./demo-cli.sh failures

# View failures for last 24 hours
./demo-cli.sh failures --hours 24
```

**Using autozap directly** (requires --db flag):

```bash
# View execution history
./autozap history --db ./autozap.db

# View history for specific workflow
./autozap history --db ./autozap.db --workflow demo-quick --limit 20

# View workflow statistics
./autozap stats demo-quick --db ./autozap.db

# View recent failures
./autozap failures --db ./autozap.db --hours 24
```

## Logs

View live logs:
```bash
tail -f autozap.log
```

The logs show:
- Workflow registrations
- Trigger fires
- Action executions
- Success/failure outcomes
- Retry attempts
- Error details

## Demo Workflow

The `demo-quick` workflow is specifically designed for quick demonstration:

- **Schedule**: Every 1 minute (*/1 * * * *)
- **Actions**:
  1. Print timestamp (bash action)
  2. Check HTTP endpoint (http action with retry logic)

This workflow executes frequently so you can see the dashboard populate with data quickly.

## Testing Scenarios

### 1. Successful Executions
Watch the dashboard as workflows execute successfully. You'll see:
- Total runs increment
- Success count increase
- Success rate remain high (green)
- Next execution timer reset

### 2. Retry Logic
The `retry-example` workflow demonstrates:
- HTTP requests with retry on 503 errors
- Bash commands with retry logic
- Exponential backoff delays
- Check logs to see retry attempts

### 3. Workflow Monitoring
- Multiple workflows running simultaneously
- Different trigger types (cron, filewatch)
- Execution history tracking
- Error tracking and reporting

## Troubleshooting

### Server Doesn't Start
```bash
# Check if port 8080 is already in use
lsof -i :8080

# Check logs for errors
cat autozap.log
```

### Workflows Not Loading
```bash
# Verify workflow files exist
ls -la workflows/

# Check workflow validation in logs
grep -i error autozap.log
```

### Database Issues
```bash
# Remove database and restart
./stop-demo.sh
rm -f autozap.db
./demo.sh
```

## Database Location

The demo creates an SQLite database at:
```
./autozap.db
```

This stores:
- Workflow execution history
- Action execution details
- Timestamps and durations
- Error messages

## Platform Support

The demo script is tested and works on:
- ✅ macOS (Intel and Apple Silicon)
- ✅ Linux (Ubuntu, Debian, CentOS, Fedora, etc.)

The script uses portable shell commands and automatically detects the OS for browser opening.

## Next Steps

After running the demo:

1. **Explore the Dashboard**
   - Click through different sections
   - Watch real-time updates
   - Review execution history

2. **Create Your Own Workflows**
   - Copy examples from `workflows/` directory
   - Modify triggers and actions
   - Test retry logic
   - Monitor results on dashboard

3. **Integrate with Your Systems**
   - Add production workflows
   - Configure monitoring endpoints
   - Set up alerting
   - Export Prometheus metrics

4. **Production Deployment**
   - Use systemd or supervisord
   - Configure persistent storage
   - Set up reverse proxy
   - Enable authentication

## Cleaning Up

To completely clean up the demo:

```bash
# Stop AutoZap
./stop-demo.sh

# Remove demo files
rm -f autozap.db autozap.log autozap.pid

# Remove binary (optional)
rm -f autozap
```

## Support

For issues or questions:
- Check logs: `autozap.log`
- Review workflow files in `workflows/`
- Check database: `sqlite3 autozap.db "SELECT * FROM workflow_executions;"`
