# Error Code Reference

## Overview
This document provides a comprehensive reference for all error codes used in the Prometheus metrics of the Cloud Performance Monitor.

## Error Code Format
Error codes appear in the `error_code` label of Prometheus metrics:
```
cloud_test_success{error_code="http_504_timeout",instance="...",service="...",type="..."} 0
cloud_test_errors_total{error_code="http_504_timeout",instance="...",service="...",type="..."} 1
```

## HTTP Status Code Errors

| Error Code | Description | HTTP Status | Common Causes |
|------------|-------------|-------------|---------------|
| `http_400_bad_request` | Bad Request | 400 | Invalid request format |
| `http_401_unauthorized` | Unauthorized | 401 | Authentication failed |
| `http_403_forbidden` | Forbidden | 403 | Permission denied |
| `http_404_not_found` | Not Found | 404 | Service/resource not found |
| `http_413_payload_too_large` | Payload Too Large | 413 | File size exceeds limit |
| `http_429_rate_limited` | Too Many Requests | 429 | Rate limiting active |
| `http_500_server_error` | Internal Server Error | 500 | Server-side issues |
| `http_501_not_implemented` | Not Implemented | 501 | Feature not supported |
| `http_502_bad_gateway` | Bad Gateway | 502 | Proxy/gateway error |
| `http_503_unavailable` | Service Unavailable | 503 | Maintenance/overload |
| `http_504_timeout` | Gateway Timeout | 504 | Request timeout |
| `http_507_insufficient_storage` | Insufficient Storage | 507 | Storage quota exceeded |
| `http_XXX_client_error` | Unknown Client Error | 4XX | Unspecified client error |
| `http_XXX_server_error` | Unknown Server Error | 5XX | Unspecified server error |

## Authentication Errors

| Error Code | Description | Common Causes |
|------------|-------------|---------------|
| `auth_failed` | Generic Authentication Failed | Invalid credentials, auth method issues |
| `token_error` | OAuth2 Token Issues | Token refresh failed, invalid token |

## Network Errors

| Error Code | Description | Common Causes |
|------------|-------------|---------------|
| `network_timeout` | Network Timeout | Connection timeout, deadline exceeded |
| `network_connection_error` | Connection Error | Connection refused/reset |
| `network_dns_error` | DNS Resolution Failed | DNS lookup failed |
| `network_tls_error` | TLS/SSL Issues | Certificate problems, TLS handshake failed |

## File Operation Errors

| Error Code | Description | Common Causes |
|------------|-------------|---------------|
| `upload_failed` | Generic Upload Error | Upload operation failed |
| `download_failed` | Generic Download Error | Download operation failed |
| `directory_error` | Directory Operation Failed | Cannot create/access directory |
| `cleanup_failed` | File Cleanup Failed | Cannot delete test files |
| `permission_denied` | Permission Denied | Insufficient permissions |
| `quota_exceeded` | Storage Quota Exceeded | Not enough storage space |
| `file_too_large` | File Size Limit Exceeded | File exceeds service limits |
| `size_mismatch` | File Size Mismatch | Downloaded size != uploaded size |
| `read_error` | File Read Error | Cannot read downloaded file |

## WebDAV Specific Errors

| Error Code | Description | Common Causes |
|------------|-------------|---------------|
| `webdav_error` | WebDAV Protocol Error | PROPFIND, MKCOL failures |
| `chunk_assembly_failed` | Chunk Assembly Failed | MOVE operation failed |

## Generic Errors

| Error Code | Description | Common Causes |
|------------|-------------|---------------|
| `unknown_error` | Unknown Error | Unclassified errors |
| `none` | No Error | Successful operation |

## Error Code Priority
Error codes are assigned based on priority:

1. **HTTP Status Codes** - If HTTP response code is available
2. **Authentication Patterns** - Token, OAuth, auth-related errors
3. **Network Patterns** - Timeout, connection, DNS, TLS issues
4. **File Operation Patterns** - Permission, quota, size issues
5. **WebDAV Patterns** - Protocol-specific issues
6. **Operation Fallbacks** - Generic operation-based codes

## Monitoring Error Codes

### Prometheus Queries

```promql
# Count errors by type
sum by (error_code) (cloud_test_errors_total)

# Success rate by error code
(
  sum by (service, instance) (cloud_test_success{error_code="none"})
  /
  sum by (service, instance) (cloud_test_success)
) * 100

# Most common errors
topk(10, sum by (error_code) (cloud_test_errors_total{error_code!="none"}))

# HTTP error distribution
sum by (error_code) (cloud_test_errors_total{error_code=~"http_.*"})
```

### Grafana Dashboards

Error codes are displayed in:
- Service overview panels
- Error distribution charts
- Alert condition tables
- Historical trending graphs

## Integration with Alerts

All alerts include `error_code` labels for precise identification:

```yaml
groups:
  - name: cloud_storage_alerts
    rules:
      - alert: ServiceTestFailure
        expr: cloud_test_success{error_code!="none"} == 0
        labels:
          error_code: "{{ $labels.error_code }}"
```

## Troubleshooting by Error Code

For detailed troubleshooting steps for each error code, see:
- [Runbook: Service Test Failure](Runbook-ServiceTestFailure.md)
- [Runbook: Network Issues](Runbook-NetworkIssues.md)
- [Runbook: Authentication Problems](Runbook-AuthenticationFailure.md)

## Adding New Error Codes

To add new error codes:

1. Update `internal/agent/error_codes.go`
2. Add test cases in `internal/agent/error_codes_test.go`
3. Update this reference document
4. Update relevant runbooks
5. Test with actual error conditions