# Cloud Performance Monitor - Wiki

Welcome to the Cloud Performance Monitor Wiki! This contains comprehensive runbooks and troubleshooting guides for all monitoring alerts.

## 📋 Quick Navigation

### 🚨 Critical Alerts
- [Service Down](Runbook-ServiceDown) - Monitor agent is not responding
- [Critical Upload Duration](Runbook-CriticalUploadDuration) - Uploads taking longer than 10 minutes
- [Service Test Failure](Runbook-ServiceTestFailure) - Complete test failures with error codes
- [Critical Error Rate](Runbook-CriticalErrorRate) - Error rate above 50%
- [Critical Network Latency](Runbook-CriticalNetworkLatency) - Network latency above 500ms
- [Circuit Breaker Open](Runbook-CircuitBreakerOpen) - Service protection activated
- [Critical SLA Violation](Runbook-CriticalSLAViolation) - Below 95% uptime

### ⚠️ Warning Alerts
- [High Upload Duration](Runbook-HighUploadDuration) - Uploads taking longer than 5 minutes
- [Slow Upload Speed](Runbook-SlowUploadSpeed) - Upload speeds below 1 MB/s
- [High Error Rate](Runbook-HighErrorRate) - Error rate above 10%
- [High Network Latency](Runbook-HighNetworkLatency) - Network latency above 100ms
- [Connection Timeouts](Runbook-ConnectionTimeouts) - Frequent connection failures
- [Slow Chunk Uploads](Runbook-SlowChunkUploads) - Chunk uploads taking too long
- [High Chunk Retry Rate](Runbook-HighChunkRetryRate) - Many chunk upload retries
- [SLA Violation 99%](Runbook-SLAViolation) - Below 99% uptime
- [Too Many Alerts](Runbook-TooManyAlerts) - Multiple alerts firing simultaneously

### 📊 Monitoring Categories
- **Availability**: Service uptime and responsiveness
- **Performance**: Upload/download speeds and durations
- **Reliability**: Error rates and test failures

### 📖 Reference Documentation
- [Error Code Reference](Error-Code-Reference) - Complete error code catalog with descriptions
- [Prometheus Metrics](Prometheus-Metrics) - All available metrics and their meanings
- [Configuration Guide](Configuration-Guide) - Environment variables and settings
- **Network**: Latency and connectivity issues
- **SLA**: Service level agreement compliance

## 🛠️ General Troubleshooting

### Common First Steps
1. **Check Service Status**: `docker compose ps`
2. **View Logs**: `docker compose logs [service-name]`
3. **Check Network**: Test connectivity to cloud services
4. **Verify Configuration**: Ensure .env settings are correct

### Useful Commands
```bash
# Service status
docker compose ps

# View all logs
docker compose logs

# Service-specific logs
docker compose logs monitor-agent
docker compose logs alertmanager
docker compose logs prometheus

# Restart services
docker compose restart

# Check metrics
docker exec prometheus wget -qO- http://monitor-agent:8080/metrics | head -20
```

### Emergency Contacts
- **Admin**: As configured in EMAIL_ADMIN
- **DevOps**: As configured in EMAIL_DEVOPS  
- **Network**: As configured in EMAIL_NETWORK
- **Management**: As configured in EMAIL_MANAGEMENT

## 📚 Additional Resources
- [Project README](https://github.com/MarcelWMeyer/cloud-performance-monitor/blob/main/README.md)
- [Email Configuration Guide](https://github.com/MarcelWMeyer/cloud-performance-monitor/blob/main/docs/EMAIL_CONFIGURATION.md)
- [Port Security Documentation](https://github.com/MarcelWMeyer/cloud-performance-monitor/blob/main/docs/PORT_SECURITY.md)