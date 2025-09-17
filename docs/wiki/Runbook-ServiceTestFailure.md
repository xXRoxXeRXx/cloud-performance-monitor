# Runbook: Service Test Failure

## 🚨 Alert Description
**Complete test failure with error code**

**Severity**: Critical  
**Category**: Reliability

## 📊 Alert Details
- **Expression**: `cloud_test_success{error_code!="none"} == 0`
- **Duration**: Alert fires after 1 minute
- **Meaning**: Test completely failed to execute with specific error

## 🔍 Investigation Steps

### 1. Identify Error Code
```bash
# Check current test status
docker exec prometheus wget -qO- 'http://localhost:9090/api/v1/query?query=cloud_test_success{error_code!="none"}'
```

### 2. Review Recent Logs
```bash
# Check for ERROR level entries
docker compose logs monitor-agent | grep -E "\[ERROR\].*$(date +%Y-%m-%d)"

# Look for specific error patterns
docker compose logs monitor-agent | grep -E "(authentication|network|timeout|chunk|upload)"
```

## 🛠️ Common Error Codes & Solutions

### Authentication Errors

#### `auth_failed` - Authentication Failed
**Symptoms**: 401 Unauthorized, Invalid credentials
```bash
# Check credentials in .env
grep -E "(USER|PASS|TOKEN)" .env

# Test authentication manually
# For Nextcloud/HiDrive:
curl -u "username:password" [service-url]/status.php

# For Dropbox:
curl -H "Authorization: Bearer [token]" https://api.dropboxapi.com/2/users/get_current_account
```

**Solutions**:
- Verify username/password in .env
- Check if app password is required
- Verify OAuth tokens are not expired
- Check for account lockouts

#### `token_expired` - OAuth Token Expired
**Symptoms**: 401 with token-related error messages
```bash
# Check token refresh logs
docker compose logs monitor-agent | grep -i "token\|oauth\|refresh"
```

**Solutions**:
- Regenerate OAuth tokens (Dropbox, HiDrive Legacy)
- Update refresh tokens in .env
- Check token expiration dates

### Network Errors

#### `connection_timeout` - Network Timeout
**Symptoms**: Connection timeouts, DNS failures
```bash
# Test connectivity
docker exec monitor-agent ping -c 4 [service-host]
docker exec monitor-agent nslookup [service-host]
```

**Solutions**:
- Check network connectivity
- Verify DNS resolution
- Check firewall rules
- Contact network team

#### `ssl_error` - SSL/TLS Issues
**Symptoms**: SSL handshake failures, certificate errors
```bash
# Test SSL connection
docker exec monitor-agent openssl s_client -connect [service-host]:443 -servername [service-host]
```

**Solutions**:
- Check certificate validity
- Verify SSL/TLS configuration
- Update CA certificates if needed

### Service-Specific Errors

#### `directory_creation_failed` - Cannot Create Upload Directory
**Symptoms**: 403 Forbidden, insufficient permissions
```bash
# Check account permissions
# Verify WebDAV access is enabled
# Check for quota limits
```

**Solutions**:
- Verify account has write permissions
- Check storage quota
- Ensure WebDAV is enabled
- Check directory permissions

#### `chunk_assembly_failed` - Chunked Upload Assembly Failed
**Symptoms**: Individual chunks upload but final assembly fails
```bash
# Check chunk upload logs
docker compose logs monitor-agent | grep -i "chunk.*assembly\|move\|final"
```

**Solutions**:
- Check storage space
- Verify file size limits
- Check for concurrent upload limits
- Review chunk size configuration

## 🚀 Immediate Actions

### 1. Check Service Dependencies
```bash
# Verify all services are running
docker compose ps

# Check container health
docker compose ps --format "table {{.Names}}\t{{.Status}}"
```

### 2. Test Service Connectivity
```bash
# Quick connectivity test
docker exec monitor-agent curl -I [service-url] --max-time 10
```

### 3. Review Configuration
```bash
# Verify .env configuration
# Check for missing variables
# Verify credentials format
```

## 📋 Error Code Reference

| Error Code | Meaning | Severity | Common Cause |
|------------|---------|----------|--------------|
| `auth_failed` | Authentication failure | Critical | Wrong credentials |
| `token_expired` | OAuth token expired | Critical | Token needs refresh |
| `connection_timeout` | Network timeout | Critical | Network issues |
| `ssl_error` | SSL/TLS problem | Critical | Certificate issues |
| `directory_creation_failed` | Cannot create directory | Critical | Permission issues |
| `chunk_assembly_failed` | Chunked upload failed | Critical | Storage/quota issues |
| `file_too_large` | File exceeds limits | Warning | File size limits |
| `quota_exceeded` | Storage quota full | Critical | Storage full |
| `rate_limited` | API rate limiting | Warning | Too many requests |

## 🔄 Resolution Process

### Step 1: Quick Fixes (< 5 minutes)
1. Check service status
2. Verify basic connectivity
3. Review recent configuration changes

### Step 2: Detailed Investigation (< 15 minutes)
1. Analyze error codes
2. Check service-specific issues
3. Review authentication status

### Step 3: Corrective Actions (< 30 minutes)
1. Fix identified issues
2. Update configuration if needed
3. Restart services
4. Verify resolution

## 📧 Escalation

### Immediate (< 5 minutes)
- Check for simple configuration issues
- Verify service availability

### After 15 minutes
- Contact cloud service provider
- Escalate to network team if connectivity issues

### After 30 minutes
- Consider temporary service degradation
- Notify management
- Implement backup monitoring

## 📊 Related Metrics
- `cloud_test_success` - Test success rate
- `cloud_test_errors_total` - Error counts by type
- `cloud_test_duration_seconds` - Test execution time
- `up{job="monitor-agent"}` - Service availability

## 🔗 Related Runbooks
- [Service Down](Runbook-ServiceDown)
- [High Error Rate](Runbook-HighErrorRate)
- [Critical Error Rate](Runbook-CriticalErrorRate)