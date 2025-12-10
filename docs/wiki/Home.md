# Cloud Performance Monitor - Wiki# Cloud Performance Monitor - Wiki# Cloud Performance Monitor - Wiki



Welcome to the Cloud Performance Monitor Wiki!



## 🚨 Alerts (6 Total)Welcome to the Cloud Performance Monitor Wiki! This contains runbooks and troubleshooting guides.Welcome to the Cloud Performance Monitor Wiki! This contains comprehensive runbooks and troubleshooting guides for all monitoring alerts.



| Alert | Severity | Description |

|-------|----------|-------------|

| [ServiceDown](Runbook-ServiceDown) | Critical | Monitor agent not responding |## 📋 Quick Navigation## 📋 Quick Navigation

| CloudServiceUnavailable | Warning | No successful tests in 15 minutes |

| SlowUploadSpeed | Warning | Upload speed below 1 MB/s |

| HighErrorRate | Warning | Error rate above 20% in last hour |

| CircuitBreakerOpen | Warning | Service protection activated |### 🚨 Alerts (6 Total)### 🚨 Critical Alerts

| PrometheusStorageNearFull | Warning | Prometheus storage > 80% full |

- [Service Down](Runbook-ServiceDown) - Monitor agent is not responding

## 📖 Reference

| Alert | Severity | Description |- [Critical Upload Duration](Runbook-CriticalUploadDuration) - Uploads taking longer than 10 minutes

- [Error Code Reference](Error-Code-Reference) - Error codes catalog

|-------|----------|-------------|- [Service Test Failure](Runbook-ServiceTestFailure) - Complete test failures with error codes

## 🛠️ Troubleshooting

| [ServiceDown](Runbook-ServiceDown) | Critical | Monitor agent not responding |- [Critical Error Rate](Runbook-CriticalErrorRate) - Error rate above 50%

### Quick Commands

| CloudServiceUnavailable | Warning | No successful tests in 15 minutes |- [Critical Network Latency](Runbook-CriticalNetworkLatency) - Network latency above 500ms

```bash

# Check service status| SlowUploadSpeed | Warning | Upload speed below 1 MB/s |- [Circuit Breaker Open](Runbook-CircuitBreakerOpen) - Service protection activated

docker compose ps

| HighErrorRate | Warning | Error rate above 20% in last hour |- [Critical SLA Violation](Runbook-CriticalSLAViolation) - Below 95% uptime

# View logs

docker compose logs monitor-agent --tail=100| CircuitBreakerOpen | Warning | Service protection activated |



# Restart services| PrometheusStorageNearFull | Warning | Prometheus storage > 80% full |### ⚠️ Warning Alerts

docker compose restart

- [High Upload Duration](Runbook-HighUploadDuration) - Uploads taking longer than 5 minutes

# Check Prometheus targets

docker exec prometheus wget -qO- http://localhost:9090/api/v1/targets### 📖 Reference Documentation- [Slow Upload Speed](Runbook-SlowUploadSpeed) - Upload speeds below 1 MB/s

```

- [Error Code Reference](Error-Code-Reference) - Error code catalog with descriptions- [High Error Rate](Runbook-HighErrorRate) - Error rate above 10%

### Common Issues

- [High Network Latency](Runbook-HighNetworkLatency) - Network latency above 100ms

**Services not starting?**

```bash## 🛠️ General Troubleshooting- [Connection Timeouts](Runbook-ConnectionTimeouts) - Frequent connection failures

make status

make logs- [Slow Chunk Uploads](Runbook-SlowChunkUploads) - Chunk uploads taking too long

docker system prune

```### Common First Steps- [High Chunk Retry Rate](Runbook-HighChunkRetryRate) - Many chunk upload retries



**No metrics in Grafana?**1. **Check Service Status**: `docker compose ps`- [SLA Violation 99%](Runbook-SLAViolation) - Below 99% uptime

```bash

docker exec prometheus wget -qO- http://monitor-agent:8080/metrics | head -102. **View Logs**: `docker compose logs monitor-agent --tail=100`- [Too Many Alerts](Runbook-TooManyAlerts) - Multiple alerts firing simultaneously

```

3. **Check Metrics**: http://localhost:9090 (Prometheus)

## 📊 Dashboards

4. **View Dashboards**: http://localhost:3000 (Grafana)### 📊 Monitoring Categories

Access Grafana at http://localhost:3003

- **Availability**: Service uptime and responsiveness

- **Daily Performance** - 24-hour metrics

- **Monthly Performance** - 30-day trends### Useful Commands- **Performance**: Upload/download speeds and durations

- **Errors** - Error tracking

```bash- **Reliability**: Error rates and test failures

# Service status

docker compose ps### 📖 Reference Documentation

- [Error Code Reference](Error-Code-Reference) - Complete error code catalog with descriptions

# View agent logs- [Prometheus Metrics](Prometheus-Metrics) - All available metrics and their meanings

docker compose logs monitor-agent- [Configuration Guide](Configuration-Guide) - Environment variables and settings

- **Network**: Latency and connectivity issues

# Restart services- **SLA**: Service level agreement compliance

docker compose restart

## 🛠️ General Troubleshooting

# Rebuild and restart

docker compose up -d --force-recreate### Common First Steps

1. **Check Service Status**: `docker compose ps`

# Check Prometheus targets2. **View Logs**: `docker compose logs [service-name]`

curl http://localhost:9090/api/v1/targets3. **Check Network**: Test connectivity to cloud services

```4. **Verify Configuration**: Ensure .env settings are correct



## 📊 Available Dashboards### Useful Commands

```bash

- **Daily Performance** - Upload/download speeds over 24 hours# Service status

- **Monthly Performance** - 30-day trends and statisticsdocker compose ps

- **Errors** - Error tracking and analysis

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
- [Project README](https://github.com/xXRoxXeRXx/cloud-performance-monitor/blob/main/README.md)
- [Email Configuration Guide](https://github.com/xXRoxXeRXx/cloud-performance-monitor/blob/main/docs/EMAIL_CONFIGURATION.md)
- [Port Security Documentation](https://github.com/xXRoxXeRXx/cloud-performance-monitor/blob/main/docs/PORT_SECURITY.md)