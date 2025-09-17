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

### HTTP Status Code Errors

#### `http_401_unauthorized` - Authentication Failed
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
- Check if password contains special characters requiring escaping ($$ for $)
- Confirm app passwords for MagentaCLOUD
- Refresh OAuth2 tokens for Dropbox/HiDrive Legacy

#### `http_403_forbidden` - Access Forbidden
**Symptoms**: 403 Forbidden, Permission denied
**Solutions**:
- Check user permissions on cloud service
- Verify user has write access to test directory
- For MagentaCLOUD: Confirm ANID is correct

#### `http_404_not_found` - Resource Not Found
**Symptoms**: 404 Not Found, Service endpoint not available
**Solutions**:
- Verify service URL in .env
- Check if cloud service is operational
- Confirm WebDAV endpoint path

#### `http_413_payload_too_large` - File Too Large
**Symptoms**: 413 Payload Too Large, File size exceeds limit
**Solutions**:
- Reduce TEST_FILE_SIZE_MB in .env
- Check cloud service upload limits
- Verify TEST_CHUNK_SIZE_MB is appropriate

#### `http_429_rate_limited` - Rate Limited
**Symptoms**: 429 Too Many Requests, Rate limiting active
**Solutions**:
- Increase TEST_INTERVAL_SECONDS in .env
- Check cloud service rate limits
- Implement exponential backoff

#### `http_500_server_error` - Internal Server Error
**Symptoms**: 500 Internal Server Error, Server-side issue
**Solutions**:
- Check cloud service status pages
- Retry tests in a few minutes
- Contact cloud service support if persistent

#### `http_502_bad_gateway` - Bad Gateway
**Symptoms**: 502 Bad Gateway, Proxy/gateway error
**Solutions**:
- Cloud service proxy issues
- Check service status
- Wait and retry

#### `http_503_unavailable` - Service Unavailable
**Symptoms**: 503 Service Unavailable, Maintenance mode
**Solutions**:
- Cloud service in maintenance
- Check service announcements
- Wait for maintenance completion

#### `http_504_timeout` - Gateway Timeout
**Symptoms**: 504 Gateway Timeout, Request timeout
**Solutions**:
- Increase timeout values
- Check network connectivity
- Reduce file size or chunk size

#### `http_507_insufficient_storage` - Insufficient Storage
**Symptoms**: 507 Insufficient Storage, Quota exceeded
**Solutions**:
- Check cloud service quota
- Clean up old files
- Reduce test file size

### Authentication Errors

#### `auth_failed` - Generic Authentication Failed
**Symptoms**: Authentication errors not covered by HTTP codes
**Solutions**:
- Check all authentication parameters
- Verify service-specific auth requirements

#### `token_error` - OAuth2 Token Issues
**Symptoms**: Token refresh or validation failed
**Solutions**:
- Regenerate OAuth2 refresh tokens
- Check OAuth2 client credentials
- Verify token expiration

### Network Errors

#### `network_timeout` - Network Timeout
**Symptoms**: Connection timeouts, deadline exceeded
```bash
# Test connectivity
docker exec monitor-agent ping -c 4 [service-host]
docker exec monitor-agent nslookup [service-host]
```

**Solutions**:
- Check network connectivity
- Verify DNS resolution
- Increase timeout values in config
- Check firewall settings

#### `network_connection_error` - Connection Refused/Reset
**Symptoms**: Connection refused, connection reset
**Solutions**:
- Service might be down
- Check service status pages
- Verify service ports are open
- Check firewall/proxy settings

#### `network_dns_error` - DNS Resolution Failed
**Symptoms**: No such host, DNS resolution failed
**Solutions**:
- Check DNS settings
- Verify domain names in .env
- Test with different DNS servers
- Check /etc/hosts if needed

#### `network_tls_error` - TLS/SSL Certificate Issues
**Symptoms**: Certificate validation failed, SSL errors
**Solutions**:
- Update CA certificates
- Check certificate expiration
- Verify TLS configuration
- Use openssl to test: `openssl s_client -connect [host]:443`

### File Operation Errors

#### `upload_failed` - Generic Upload Error
**Symptoms**: Upload operation failed without specific HTTP code
**Solutions**:
- Check upload logs for details
- Verify file permissions
- Check available storage
- Try smaller file size

#### `download_failed` - Generic Download Error
**Symptoms**: Download operation failed without specific HTTP code
**Solutions**:
- Verify file exists on service
- Check download permissions
- Test with different file
- Check network stability

#### `directory_error` - Directory Operation Failed
**Symptoms**: Cannot create or access test directory
**Solutions**:
- Check user permissions
- Verify path structure
- Test manual directory creation
- Check service-specific path requirements

#### `cleanup_failed` - File Cleanup Failed
**Symptoms**: Cannot delete test files
**Solutions**:
- Check delete permissions
- Manual cleanup may be needed
- Verify file is not locked
- Check service quotas

#### `permission_denied` - Permission Denied
**Symptoms**: Insufficient permissions for operation
**Solutions**:
- Verify user permissions on cloud service
- Check directory permissions
- Ensure write access to test paths
- Verify app password permissions

#### `quota_exceeded` - Storage Quota Exceeded
**Symptoms**: Not enough storage space
**Solutions**:
- Check cloud service storage quota
- Clean up old files
- Reduce TEST_FILE_SIZE_MB
- Upgrade storage plan if needed

#### `file_too_large` - File Size Limit Exceeded
**Symptoms**: File exceeds service limits
**Solutions**:
- Reduce TEST_FILE_SIZE_MB in .env
- Check service file size limits
- Adjust TEST_CHUNK_SIZE_MB
- Split into smaller chunks

### WebDAV Specific Errors

#### `webdav_error` - WebDAV Protocol Error
**Symptoms**: WebDAV PROPFIND, MKCOL, or other WebDAV errors
**Solutions**:
- Check WebDAV endpoint URL
- Verify WebDAV is enabled on service
- Test with WebDAV client tools
- Check authentication method

#### `chunk_assembly_failed` - Chunk Assembly Failed
**Symptoms**: MOVE operation failed during chunk assembly
**Solutions**:
- Check chunk upload logs
- Verify all chunks uploaded successfully
- Check destination path permissions
- Retry with smaller chunks

### Generic Errors

#### `size_mismatch` - File Size Mismatch
**Symptoms**: Downloaded file size differs from uploaded
**Solutions**:
- Check for data corruption
- Verify network stability
- Check service integrity
- Test with different file size

#### `read_error` - File Read Error
**Symptoms**: Cannot read downloaded file
**Solutions**:
- Check file corruption
- Verify download completed
- Test network stability
- Check available disk space

#### `unknown_error` - Unknown Error
**Symptoms**: Unclassified error
**Solutions**:
- Check detailed logs for context
- Review error messages
- Contact support with logs
- Try alternative test parameters
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