# AutoZap Quick Start Guide

## Run the Demo in 3 Steps

### 1. Start AutoZap
```bash
./demo.sh
```

This will:
- ✅ Build the AutoZap binary
- ✅ Start the agent with all workflows
- ✅ Open the dashboard in your browser

### 2. View the Dashboard

Open: **http://localhost:8080/dashboard**

You'll see:
- 📊 Active workflows with real-time execution counts
- ✅ Success rates (color-coded: green/yellow/red)
- ⏱️ Next execution countdown timers
- ❌ Error messages (if any)
- 🔄 Auto-refresh every 10 seconds

### 3. Use CLI Tools

```bash
# View execution history
./demo-cli.sh history

# View workflow stats
./demo-cli.sh stats demo-quick

# View recent failures
./demo-cli.sh failures
```

### Stop the Demo

```bash
./stop-demo.sh
```

## What's Running?

The demo loads workflows from `./workflows/` including:

- **demo-quick** - Runs every 1 minute (for quick dashboard updates)
- **retry-example** - Demonstrates retry logic
- **system-health-check** - System monitoring examples
- And many more...

## Dashboard Features

| Feature | Description |
|---------|-------------|
| **Active Workflows** | All loaded workflows with status |
| **Execution Stats** | Total runs, successes, failures |
| **Success Rate** | Percentage with color coding |
| **Next Run** | Countdown timer to next execution |
| **Error Tracking** | Last error message displayed |
| **History** | Recent execution timeline |

## Files Created

- `demo.sh` - Main demo startup script
- `stop-demo.sh` - Stop the demo
- `demo-cli.sh` - CLI helper for database queries
- `workflows/demo-quick.yaml` - Fast demo workflow
- `DEMO.md` - Comprehensive demo guide
- `autozap.db` - SQLite database (created on first run)
- `autozap.log` - Application logs

## Platform Support

✅ macOS (Intel & Apple Silicon)
✅ Linux (all distributions)

## Next Steps

1. ✅ Run `./demo.sh` to start
2. ✅ Open http://localhost:8080/dashboard
3. ✅ Watch workflows execute
4. ✅ Use CLI tools to explore data
5. ✅ Create your own workflows in `./workflows/`

## Need Help?

- **Full Guide**: See [DEMO.md](./DEMO.md)
- **Logs**: `tail -f autozap.log`
- **Dashboard**: http://localhost:8080/dashboard
- **API Docs**: See DEMO.md for all endpoints

---

**Happy Automating! 🚀**
