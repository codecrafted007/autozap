# AutoZap Control Wrapper (autozapctl) - Release Notes

## Version 1.0.0 - Production-Ready Control Tool

### Overview

Replaced demo scripts with a comprehensive, production-ready `autozapctl` wrapper that provides complete lifecycle management for AutoZap.

### What Changed

#### ✅ New Files Created

1. **`autozapctl`** - Main control wrapper script
   - Complete service management (start/stop/restart/status)
   - Configuration file support
   - Data query wrappers
   - Utilities (validate, open dashboard)
   - Systemd service installation
   - Cross-platform (Linux & macOS)

2. **`.autozaprc.example`** - Configuration file template
   - All configurable options documented
   - Can be placed in project dir, home dir, or /etc

3. **`autozap.service`** - Systemd service file
   - Production-ready service definition
   - Security hardening
   - Auto-restart on failure

4. **`AUTOZAPCTL.md`** - Comprehensive documentation
   - Full command reference
   - Configuration guide
   - Examples and use cases
   - Troubleshooting guide

5. **`QUICK_REFERENCE.md`** - Quick command reference
   - Most common commands
   - Quick troubleshooting
   - Migration guide from demo scripts

#### 📝 Files Retained (Still Useful)

- `demo.sh` - Quick demo for first-time users
- `stop-demo.sh` - Demo cleanup
- `demo-cli.sh` - Demo CLI helper
- `DEMO.md` - Demo guide
- `QUICK_START.md` - Quick start guide

These can still be used for quick demos, but `autozapctl` is recommended for regular use.

### New Capabilities

#### Service Management

```bash
autozapctl start [options]    # Start with many options
autozapctl stop               # Graceful shutdown
autozapctl restart [options]  # Restart with new config
autozapctl status             # Detailed status info
```

#### Data Queries (No --db flag needed!)

```bash
autozapctl history [options]
autozapctl stats <workflow>
autozapctl failures [options]
```

#### Utilities

```bash
autozapctl validate <file>    # Validate before running
autozapctl open               # Open dashboard in browser
autozapctl logs [-f]          # View/tail logs
autozapctl install-service    # Install as systemd service
```

#### Configuration File Support

Create `.autozaprc`:
```bash
WORKFLOWS_DIR="./workflows"
DATABASE="./autozap.db"
HTTP_PORT="8080"
LOG_FILE="./autozap.log"
WATCH="true"
LOG_DIR="./logs"  # Optional
```

Loaded automatically from:
1. `./.autozaprc` (project directory)
2. `~/.autozaprc` (user home)
3. `/etc/autozap/config` (system-wide)

### Migration Guide

#### From demo.sh

| Old | New |
|-----|-----|
| `./demo.sh` | `./autozapctl start` |
| `./stop-demo.sh` | `./autozapctl stop` |
| `./demo-cli.sh history` | `./autozapctl history` |
| `./demo-cli.sh stats <w>` | `./autozapctl stats <w>` |
| `./demo-cli.sh failures` | `./autozapctl failures` |

#### From manual commands

| Old | New |
|-----|-----|
| `./autozap agent ./workflows --db ./autozap.db --http-port 8080` | `./autozapctl start` |
| `./autozap history --db ./autozap.db` | `./autozapctl history` |
| `./autozap stats myworkflow --db ./autozap.db` | `./autozapctl stats myworkflow` |
| `kill $(cat autozap.pid)` | `./autozapctl stop` |

### Key Features

✅ **Configuration Management**
- Load settings from config files
- Override with command-line flags
- Multiple config file locations

✅ **Process Management**
- PID tracking
- Health checks
- Graceful shutdown
- Force kill if needed

✅ **Database Path Handling**
- Automatically uses configured database
- No need to specify --db flag for queries
- Consistent across all commands

✅ **Status Monitoring**
- Running status (PID)
- Uptime
- Memory usage
- Active workflow count
- Configuration in use

✅ **Production Ready**
- Systemd service support
- Security hardening
- Resource limits
- Auto-restart on failure

✅ **Cross-Platform**
- Works on Linux
- Works on macOS
- Portable shell commands
- Platform-specific optimizations

### Use Cases

#### Development
```bash
# Quick start
./autozapctl start

# Watch logs
./autozapctl logs -f

# Check status
./autozapctl status
```

#### Production (Linux)
```bash
# Install as service
sudo ./autozapctl install-service
sudo systemctl enable autozap
sudo systemctl start autozap

# Monitor
sudo systemctl status autozap
./autozapctl status
```

#### Testing
```bash
# Validate all workflows
for f in workflows/*.yaml; do
    ./autozapctl validate "$f"
done

# Dry run
./autozapctl start --dry-run
```

#### Debugging
```bash
# Foreground mode
./autozapctl start --foreground

# Or background + tail
./autozapctl start
./autozapctl logs -f
```

### Breaking Changes

None! The `autozapctl` wrapper is purely additive. All existing commands still work:
- `./autozap agent ...`
- `./autozap run ...`
- `./autozap history ...`
- etc.

The wrapper just makes them easier to use.

### Documentation

- **Full Command Reference**: [AUTOZAPCTL.md](./AUTOZAPCTL.md)
- **Quick Reference**: [QUICK_REFERENCE.md](./QUICK_REFERENCE.md)
- **Demo Guide**: [DEMO.md](./DEMO.md)
- **Quick Start**: [QUICK_START.md](./QUICK_START.md)

### Installation

```bash
# Make executable
chmod +x autozapctl

# Optional: Add to PATH
sudo ln -s $(pwd)/autozapctl /usr/local/bin/autozapctl

# Optional: Install as service (Linux)
sudo ./autozapctl install-service
```

### Example Session

```bash
# Start AutoZap
$ ./autozapctl start
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Starting AutoZap
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ℹ Configuration:
  Workflows: ./workflows
  Database:  ./autozap.db
  HTTP Port: 8080
  Log File:  ./autozap.log
  Hot Reload: true

ℹ Starting AutoZap agent...
ℹ Waiting for server to be ready...
✓ AutoZap started successfully (PID: 12345)

ℹ Dashboard:  http://localhost:8080/dashboard
ℹ Metrics:    http://localhost:8080/metrics
ℹ Health:     http://localhost:8080/health

# Check status
$ ./autozapctl status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  AutoZap Status
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ AutoZap is running (PID: 12345)

  Uptime:    00:05
  Memory:    5MB
  Config:    none (using defaults)
  Workflows: ./workflows
  Database:  ./autozap.db
  Log File:  ./autozap.log

  Active Workflows: 13

ℹ Dashboard: http://localhost:8080/dashboard

# View stats
$ ./autozapctl stats my-workflow
📊 Statistics for workflow: my-workflow (Last 7 days)

METRIC            VALUE
------            -----
Total Executions  42
Successful        40 (✓)
Failed            2 (✗)
Success Rate      95.24%
Avg Duration      1.45s

# Stop
$ ./autozapctl stop
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Stopping AutoZap
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ℹ Stopping AutoZap (PID: 12345)...
✓ AutoZap stopped successfully
```

### Future Enhancements

Potential additions for future versions:
- `autozapctl exec <workflow>` - Manually trigger a workflow
- `autozapctl reload` - Reload workflows without restart
- `autozapctl config` - View/edit configuration
- `autozapctl backup` - Backup database
- `autozapctl export` - Export metrics/stats

### Feedback

This is a production-ready tool, but feedback is welcome for improvements and additional features!

---

**Recommendation**: Use `autozapctl` for all AutoZap operations. Keep the demo scripts for quick demonstrations only.
