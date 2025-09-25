package agent

import (
	"fmt"
	"net/http"
	"strings"
)

// ExtractErrorCode extracts a specific error code from various error types
func ExtractErrorCode(err error, operation string) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	// HTTP Status Code patterns (from error strings)
	if strings.Contains(errStr, "status 401") || strings.Contains(errStr, "401") || strings.Contains(errStr, "unauthorized") {
		return "http_401_unauthorized"
	}
	if strings.Contains(errStr, "status 403") || strings.Contains(errStr, "403") || strings.Contains(errStr, "forbidden") {
		return "http_403_forbidden"
	}
	if strings.Contains(errStr, "status 404") || strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
		return "http_404_not_found"
	}
	if strings.Contains(errStr, "status 413") || strings.Contains(errStr, "413") || strings.Contains(errStr, "payload too large") {
		return "http_413_payload_too_large"
	}
	if strings.Contains(errStr, "status 429") || strings.Contains(errStr, "429") || strings.Contains(errStr, "too many requests") {
		return "http_429_rate_limited"
	}
	if strings.Contains(errStr, "status 500") || strings.Contains(errStr, "500") || strings.Contains(errStr, "internal server error") {
		return "http_500_server_error"
	}
	if strings.Contains(errStr, "status 502") || strings.Contains(errStr, "502") || strings.Contains(errStr, "bad gateway") {
		return "http_502_bad_gateway"
	}
	if strings.Contains(errStr, "status 503") || strings.Contains(errStr, "503") || strings.Contains(errStr, "service unavailable") {
		return "http_503_unavailable"
	}
	if strings.Contains(errStr, "status 504") || strings.Contains(errStr, "504") || strings.Contains(errStr, "gateway timeout") {
		return "http_504_timeout"
	}
	if strings.Contains(errStr, "status 507") || strings.Contains(errStr, "507") || strings.Contains(errStr, "insufficient storage") {
		return "http_507_insufficient_storage"
	}

	// Authentication patterns
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "auth") {
		return "auth_failed"
	}
	if strings.Contains(errStr, "token") || strings.Contains(errStr, "oauth") {
		return "token_error"
	}

	// Network patterns
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return "network_timeout"
	}
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connection reset") {
		return "network_connection_error"
	}
	if strings.Contains(errStr, "dns") || strings.Contains(errStr, "no such host") {
		return "network_dns_error"
	}
	if strings.Contains(errStr, "tls") || strings.Contains(errStr, "ssl") || strings.Contains(errStr, "certificate") {
		return "network_tls_error"
	}

	// File operation patterns
	if strings.Contains(errStr, "permission denied") {
		return "permission_denied"
	}
	if strings.Contains(errStr, "quota") || strings.Contains(errStr, "storage") {
		return "quota_exceeded"
	}
	if strings.Contains(errStr, "file too large") || strings.Contains(errStr, "size limit") {
		return "file_too_large"
	}

	// WebDAV specific patterns
	if strings.Contains(errStr, "webdav") || strings.Contains(errStr, "propfind") {
		return "webdav_error"
	}
	if strings.Contains(errStr, "move") || strings.Contains(errStr, "assembly") {
		return "chunk_assembly_failed"
	}

	// Operation-specific fallbacks
	switch operation {
	case "upload":
		return "upload_failed"
	case "download":
		return "download_failed"
	case "directory", "mkdir":
		return "directory_error"
	case "delete", "cleanup":
		return "cleanup_failed"
	case "auth", "token":
		return "auth_failed"
	default:
		return "unknown_error"
	}
}

// ExtractHTTPErrorCode extracts error code from HTTP response
func ExtractHTTPErrorCode(resp *http.Response, err error, operation string) string {
	if err != nil {
		return ExtractErrorCode(err, operation)
	}

	if resp != nil && resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 400:
			return "http_400_bad_request"
		case 401:
			return "http_401_unauthorized"
		case 403:
			return "http_403_forbidden"
		case 404:
			return "http_404_not_found"
		case 413:
			return "http_413_payload_too_large"
		case 429:
			return "http_429_rate_limited"
		case 500:
			return "http_500_server_error"
		case 501:
			return "http_501_not_implemented"
		case 502:
			return "http_502_bad_gateway"
		case 503:
			return "http_503_unavailable"
		case 504:
			return "http_504_timeout"
		case 507:
			return "http_507_insufficient_storage"
		default:
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return fmt.Sprintf("http_%d_client_error", resp.StatusCode)
			} else if resp.StatusCode >= 500 {
				return fmt.Sprintf("http_%d_server_error", resp.StatusCode)
			}
		}
	}

	return "none"
}
