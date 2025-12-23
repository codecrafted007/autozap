# AutoZap Workflow Examples

This directory contains production-ready workflow examples demonstrating real-world DevOps and infrastructure automation use cases.

## 📁 Workflow Categories

### 🐳 Container Management
- **`docker-cleanup.yaml`** - Automated Docker cleanup (containers, images, volumes, networks)
  - Runs weekly on Sundays at 2 AM
  - Removes stopped containers, dangling images, unused volumes
  - Cleans images older than 7 days
  - Shows disk space saved after cleanup

### 🔒 Security & Monitoring
- **`ssl-cert-monitor.yaml`** - SSL certificate expiry monitoring
  - Daily checks at 9 AM
  - Monitors multiple domains
  - Alerts when certificates expire within 30 days
  - Slack notifications for expiring certificates

### 💾 Backup & Recovery
- **`postgres-backup.yaml`** - PostgreSQL database backup automation
  - Daily backups at 1 AM
  - Automatic compression (gzip)
  - Keeps last 7 days of backups
  - Verification and Slack notifications

### 📊 System Monitoring
- **`disk-space-alert.yaml`** - Disk space monitoring and alerting
  - Checks every 15 minutes
  - Monitors root and /var partitions
  - Alerts when usage exceeds 80%
  - Tracks inode usage
  - Identifies largest directories

- **`system-health-check.yaml`** - Comprehensive system health monitoring
  - Runs every 10 minutes
  - Monitors CPU, memory, load average
  - Detects zombie processes
  - Checks swap usage
  - Verifies critical services (SSH, Nginx, Docker)
  - Slack notifications

### 🌐 API & Service Monitoring
- **`api-health-check.yaml`** - API health and performance monitoring
  - Checks every 5 minutes
  - Tests health endpoints (API, database, Redis)
  - Measures API response time
  - Tests authentication endpoints
  - Monitors error rates
  - Slack notifications

### 📝 Log Management
- **`log-rotation.yaml`** - Automated log rotation and cleanup
  - Daily execution at midnight
  - Rotates and compresses application logs
  - Cleans Nginx logs
  - Deletes logs older than 90 days
  - Truncates logs larger than 1GB
  - Cleans systemd journal (keeps 7 days)

### 📚 Example Workflows
- **`sample.yaml`** - Basic cron workflow example
- **`file-monitor.yaml`** - File system monitoring example
- **`http-check.yaml`** - HTTP health check example
- **`custom-action-example.yaml`** - Custom action example
- **`sequence-action-example.yaml`** - Action sequencing example

## 🚀 Usage

### Run a Workflow
```bash
./autozap run workflows/docker-cleanup.yaml
```

### List All Workflows
```bash
./autozap list workflows/
```

### Validate a Workflow
```bash
./autozap validate workflows/postgres-backup.yaml
```

## ⚙️ Customization Guide

### 1. Update Slack Webhook URLs
Replace `https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK` with your actual Slack webhook URL in:
- `ssl-cert-monitor.yaml`
- `postgres-backup.yaml`
- `disk-space-alert.yaml`
- `api-health-check.yaml`
- `system-health-check.yaml`

### 2. Adjust Schedules
Modify the `schedule` field in the `trigger` section:
- `*/5 * * * *` - Every 5 minutes
- `0 * * * *` - Every hour
- `0 0 * * *` - Daily at midnight
- `0 0 * * 0` - Weekly on Sundays
- `@hourly`, `@daily`, `@weekly` - Common shortcuts

### 3. Update Paths and URLs
- **SSL monitoring**: Change domain names in `ssl-cert-monitor.yaml`
- **Database backups**: Update backup paths in `postgres-backup.yaml`
- **API monitoring**: Replace API URLs in `api-health-check.yaml`
- **Log rotation**: Adjust log directories in `log-rotation.yaml`

### 4. Adjust Thresholds
- **Disk space**: Default 80% (modify in `disk-space-alert.yaml`)
- **CPU usage**: Default 80% (modify in `system-health-check.yaml`)
- **Memory usage**: Default 85% (modify in `system-health-check.yaml`)
- **API response time**: Default 2000ms (modify in `api-health-check.yaml`)

## 🎯 Production Deployment

### Docker Deployment
```bash
docker run -d \
  -v $(pwd)/workflows:/workflows:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  autozap:latest run /workflows/docker-cleanup.yaml
```

### Kubernetes CronJob
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: docker-cleanup
spec:
  schedule: "0 2 * * 0"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: autozap
            image: autozap:latest
            args: ["run", "/workflows/docker-cleanup.yaml"]
            volumeMounts:
            - name: workflows
              mountPath: /workflows
          volumes:
          - name: workflows
            configMap:
              name: autozap-workflows
```

## 🔧 Troubleshooting

### Workflow Not Running
1. Check cron schedule syntax: `*/5 * * * *`
2. Verify file permissions: `chmod +x workflows/*.yaml`
3. Check logs for errors: `./autozap run workflow.yaml --verbose`

### Permission Errors
- Docker commands may require sudo or docker group membership
- Log rotation may require root permissions for system logs
- Database backups need appropriate database credentials

### Slack Notifications Not Working
- Verify webhook URL is correct
- Check network connectivity
- Ensure JSON body is properly formatted
- Test webhook manually: `curl -X POST -H 'Content-Type: application/json' -d '{"text":"test"}' YOUR_WEBHOOK_URL`

## 📖 Learn More

- [AutoZap Documentation](../README.md)
- [Writing Custom Workflows](../docs/workflows.md)
- [Trigger Types](../docs/triggers.md)
- [Action Types](../docs/actions.md)

## 💡 Best Practices

1. **Test locally first** - Always test workflows in development before production
2. **Use descriptive names** - Make action names clear and meaningful
3. **Add proper error handling** - Use exit codes and validation
4. **Monitor execution** - Check logs regularly for failures
5. **Start conservative** - Begin with longer intervals, then optimize
6. **Document customizations** - Keep notes on what you've changed
7. **Version control** - Store workflows in Git
8. **Use environment variables** - For sensitive data like API keys

## 🤝 Contributing

Have a useful workflow to share? We'd love to see it! Please submit a PR with:
- The workflow YAML file
- Documentation in this README
- Comments explaining the workflow purpose
- Any prerequisites or setup instructions

---

**Need help?** Open an issue at [github.com/codecrafted007/autozap/issues](https://github.com/codecrafted007/autozap/issues)
