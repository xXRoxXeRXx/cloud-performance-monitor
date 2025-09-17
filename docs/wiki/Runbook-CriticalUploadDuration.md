# Runbook: Critical Upload Duration

## 🚨 Alert Description
**Upload took longer than 10 minutes - critical performance issue**

**Severity**: Critical  
**Category**: Performance

## 📊 Alert Details
- **Expression**: `cloud_test_duration_seconds{type="upload"} > 600`
- **Duration**: Alert fires after 1 minute
- **Threshold**: 600 seconds (10 minutes)
- **Impact**: Severe performance degradation affecting user experience

## 🔍 Investigation Steps

### 1. Identify Affected Services
```bash
# Check which services are affected
docker exec prometheus wget -qO- 'http://localhost:9090/api/v1/query?query=cloud_test_duration_seconds{type="upload"} > 600'
```

### 2. Check Upload Performance Metrics
```bash
# Current upload speeds
docker exec prometheus wget -qO- 'http://localhost:9090/api/v1/query?query=cloud_test_speed_mbytes_per_sec{type="upload"}'

# Recent upload durations
docker exec prometheus wget -qO- 'http://localhost:9090/api/v1/query?query=cloud_test_duration_seconds{type="upload"}'
```

### 3. Review Service Logs
```bash
# Check for upload errors
docker compose logs monitor-agent | grep -i "upload\|chunk\|error"

# Look for network issues
docker compose logs monitor-agent | grep -E "(timeout|connection|504|503)"
```

## 🛠️ Common Causes & Solutions

### 1. Network Connectivity Issues
**Symptoms**: High latency, timeouts, intermittent failures
```bash
# Check network latency
docker exec monitor-agent ping -c 4 [cloud-service-host]

# Test connectivity
docker exec monitor-agent curl -I [cloud-service-url]
```

**Solution**: 
- Contact network team if latency > 500ms
- Check for ISP issues
- Verify firewall rules

### 2. Cloud Service Performance
**Symptoms**: Consistent slow uploads across all instances
```bash
# Check service status pages
# - Nextcloud: Instance-specific status
# - Dropbox: status.dropbox.com
# - IONOS: status.ionos.com
```

**Solution**:
- Check vendor status pages
- Contact cloud service provider
- Consider temporary rate limiting

### 3. Large File Size vs. Connection Speed
**Symptoms**: Duration correlates with file size
```bash
# Check current test file size
echo $TEST_FILE_SIZE_MB

# Check chunk size configuration
echo $TEST_CHUNK_SIZE_MB
```

**Solution**:
```bash
# Reduce test file size temporarily
# In .env file:
TEST_FILE_SIZE_MB=50  # Reduce from 100MB

# Restart services
docker compose restart monitor-agent
```

### 4. Authentication/API Rate Limiting
**Symptoms**: 429 Too Many Requests, auth errors
```bash
# Check for authentication errors
docker compose logs monitor-agent | grep -E "(auth|token|401|403|429)"
```

**Solution**:
- Check API quotas
- Verify credentials
- Implement backoff strategy

## 🚀 Immediate Actions

### 1. Reduce Test Frequency (Emergency)
```bash
# Edit .env file
TEST_INTERVAL_SECONDS=1800  # Increase from 900 to 30 minutes

# Restart service
docker compose restart monitor-agent
```

### 2. Check System Resources
```bash
# Monitor system performance
docker stats monitor-agent

# Check disk I/O
iostat -x 1 5
```

### 3. Verify Network Path
```bash
# Traceroute to cloud service
docker exec monitor-agent traceroute [cloud-service-host]

# Check for packet loss
docker exec monitor-agent ping -c 20 [cloud-service-host]
```

## 📈 Performance Thresholds

| Duration | Action Required |
|----------|----------------|
| 300-600s | Warning - Monitor closely |
| 600-1200s | **Critical** - Immediate investigation |
| >1200s | Emergency - Consider service degradation |

## 🔄 Resolution Steps

### Short-term (< 15 minutes)
1. Reduce test file size
2. Increase test interval
3. Check for obvious network issues

### Medium-term (< 1 hour)
1. Contact cloud service provider
2. Investigate network path issues
3. Review historical performance trends

### Long-term
1. Implement adaptive file sizing
2. Add multiple upload location options
3. Enhanced network monitoring

## 📧 Escalation Path
1. **Immediate**: DevOps team
2. **15 minutes**: Network team + Management
3. **30 minutes**: Cloud service provider
4. **1 hour**: Consider service degradation notice

## 📊 Related Metrics
- `cloud_test_speed_mbytes_per_sec{type="upload"}` - Upload speed
- `cloud_network_latency_ms` - Network latency
- `cloud_chunk_retries_total` - Chunk upload failures
- `cloud_test_errors_total{type="upload"}` - Upload errors

## 🔗 Related Runbooks
- [High Upload Duration](Runbook-HighUploadDuration)
- [Slow Upload Speed](Runbook-SlowUploadSpeed)
- [High Network Latency](Runbook-HighNetworkLatency)
- [Slow Chunk Uploads](Runbook-SlowChunkUploads)