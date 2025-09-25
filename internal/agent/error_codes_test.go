package agent

import (
	"errors"
	"net/http"
	"testing"
)

func TestExtractErrorCode(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		operation string
		expected  string
	}{
		{
			name:      "no error",
			err:       nil,
			operation: "upload",
			expected:  "none",
		},
		{
			name:      "HTTP 504 gateway timeout",
			err:       errors.New("HTTP 504 Gateway Timeout"),
			operation: "upload",
			expected:  "http_504_timeout",
		},
		{
			name:      "HTTP 401 unauthorized",
			err:       errors.New("HTTP status 401 Unauthorized"),
			operation: "upload",
			expected:  "http_401_unauthorized",
		},
		{
			name:      "HTTP 503 service unavailable",
			err:       errors.New("Service unavailable"),
			operation: "download",
			expected:  "http_503_unavailable",
		},
		{
			name:      "Network timeout",
			err:       errors.New("context deadline exceeded"),
			operation: "upload",
			expected:  "network_timeout",
		},
		{
			name:      "Connection refused",
			err:       errors.New("connection refused"),
			operation: "download",
			expected:  "network_connection_error",
		},
		{
			name:      "Generic upload error",
			err:       errors.New("some generic error"),
			operation: "upload",
			expected:  "upload_failed",
		},
		{
			name:      "Generic download error",
			err:       errors.New("some generic error"),
			operation: "download",
			expected:  "download_failed",
		},
		{
			name:      "Webdav error",
			err:       errors.New("WebDAV PROPFIND failed"),
			operation: "upload",
			expected:  "webdav_error",
		},
		{
			name:      "Chunk assembly failed",
			err:       errors.New("MOVE assembly failed"),
			operation: "upload",
			expected:  "chunk_assembly_failed",
		},
		{
			name:      "HTTP 409 conflict",
			err:       errors.New("upload of chunk 1 failed with status 409 Conflict"),
			operation: "upload",
			expected:  "http_409_conflict",
		},
		{
			name:      "HTTP 412 precondition failed",
			err:       errors.New("Chunk 19 upload failed with status 412 Precondition Failed"),
			operation: "upload",
			expected:  "http_412_precondition_failed",
		},
		{
			name:      "Client timeout exceeded",
			err:       errors.New("Client.Timeout exceeded while awaiting headers"),
			operation: "upload",
			expected:  "network_timeout",
		},
		{
			name:      "HTTP 429 too many requests",
			err:       errors.New("finish session failed with status 429: too_many_write_operations"),
			operation: "upload",
			expected:  "http_429_rate_limited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractErrorCode(tt.err, tt.operation)
			if result != tt.expected {
				t.Errorf("ExtractErrorCode(%v, %s) = %s, want %s", tt.err, tt.operation, result, tt.expected)
			}
		})
	}
}

func TestExtractHTTPErrorCode(t *testing.T) {
	tests := []struct {
		name       string
		resp       *http.Response
		err        error
		operation  string
		expected   string
	}{
		{
			name:      "no error, status 200",
			resp:      &http.Response{StatusCode: 200},
			err:       nil,
			operation: "upload",
			expected:  "none",
		},
		{
			name:      "HTTP 404 Not Found",
			resp:      &http.Response{StatusCode: 404},
			err:       nil,
			operation: "download",
			expected:  "http_404_not_found",
		},
		{
			name:      "HTTP 504 Gateway Timeout",
			resp:      &http.Response{StatusCode: 504},
			err:       nil,
			operation: "upload",
			expected:  "http_504_timeout",
		},
		{
			name:      "HTTP 507 Insufficient Storage",
			resp:      &http.Response{StatusCode: 507},
			err:       nil,
			operation: "upload",
			expected:  "http_507_insufficient_storage",
		},
		{
			name:      "Network error takes precedence",
			resp:      nil,
			err:       errors.New("connection timeout"),
			operation: "upload",
			expected:  "network_timeout",
		},
		{
			name:      "Unknown 4xx error",
			resp:      &http.Response{StatusCode: 418},
			err:       nil,
			operation: "upload",
			expected:  "http_418_client_error",
		},
		{
			name:      "Unknown 5xx error",
			resp:      &http.Response{StatusCode: 599},
			err:       nil,
			operation: "upload",
			expected:  "http_599_server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHTTPErrorCode(tt.resp, tt.err, tt.operation)
			if result != tt.expected {
				t.Errorf("ExtractHTTPErrorCode(%v, %v, %s) = %s, want %s", tt.resp, tt.err, tt.operation, result, tt.expected)
			}
		})
	}
}
