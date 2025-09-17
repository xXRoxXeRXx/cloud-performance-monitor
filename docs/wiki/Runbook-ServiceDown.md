# Runbook: Service Down

## 🚨 Alert Description
**Monitor agent is not responding**

**Severity**: Critical  
**Category**: Availability

## 📊 Alert Details
- **Expression**: `up{job="monitor-agent"} == 0`
- **Duration**: Alert fires after 1 minute
- **Meaning**: The monitor agent container is not responding to health checks

## 🔍 Investigation Steps

### 1. Check Service Status
```bash
docker compose ps
```
Look for `monitor-agent` status. Should show `Up` and `healthy`.

### 2. Check Container Logs
```bash
docker compose logs monitor-agent --tail=50
```
Look for:
- Startup errors
- Configuration problems
- Authentication failures
- Network connectivity issues

### 3. Verify Health Endpoint
```bash
# From inside Docker network
docker exec monitor-agent curl http://localhost:8080/health
```

## 🛠️ Common Causes & Solutions

### Container Not Running
```bash
# Restart the service
docker compose restart monitor-agent

# If that fails, rebuild
docker compose up -d --force-recreate monitor-agent
```

### Configuration Errors
Check `.env` file:
- Verify cloud service credentials
- Check SMTP settings
- Ensure all required variables are set

### Network Issues
```bash
# Test connectivity to cloud services
docker exec monitor-agent ping cloud.example.com
```

### Resource Issues
```bash
# Check system resources
docker stats monitor-agent

# Check disk space
df -h
```

## 🚀 Resolution Actions

### Immediate Actions
1. **Restart Service**: `docker compose restart monitor-agent`
2. **Check Logs**: Look for obvious errors
3. **Verify Configuration**: Ensure .env is correct

### If Problem Persists
1. **Rebuild Container**: `docker compose build monitor-agent && docker compose up -d`
2. **Check Dependencies**: Ensure other services (Prometheus, Alertmanager) are running
3. **Verify Network**: Test cloud service connectivity

## 📧 Escalation
If service cannot be restored within 15 minutes:
- Contact DevOps team
- Check for infrastructure issues
- Consider switching to backup monitoring

## 🔄 Post-Resolution
1. **Verify Health**: Ensure `/health` endpoint returns healthy status
2. **Check Metrics**: Confirm metrics are being collected
3. **Monitor Alerts**: Ensure alert clears within 2 minutes
4. **Document Issue**: Add to incident log if significant downtime

## 📊 Related Metrics
- `up{job="monitor-agent"}` - Service availability
- `cloud_test_success` - Test execution status
- Container health check status in Docker

## 🔗 Related Runbooks
- [Service Test Failure](Runbook-ServiceTestFailure)
- [Too Many Alerts](Runbook-TooManyAlerts)