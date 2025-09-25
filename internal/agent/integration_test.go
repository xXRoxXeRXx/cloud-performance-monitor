//go:build integration

package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/nextcloud"
)

// mockNextcloudServer creates a mock Nextcloud WebDAV server for testing
func mockNextcloudServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock MKCOL (create directory)
	mux.HandleFunc("/remote.php/dav/files/testuser/performance_tests/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "MKCOL" {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock file operations (PUT, GET, DELETE)
	mux.HandleFunc("/remote.php/dav/files/testuser/performance_tests/testfile_", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			w.WriteHeader(http.StatusCreated)
		case "GET":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock file content"))
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}
func mockDropboxServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock OAuth2 token endpoint
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"access_token": "mock_access_token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock create folder endpoint
	mux.HandleFunc("/2/files/create_folder_v2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"metadata": {
					"name": "performance_tests",
					"path_lower": "/performance_tests",
					"path_display": "/performance_tests"
				}
			}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock upload session start
	mux.HandleFunc("/2/files/upload_session/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"session_id": "mock_session_id"
			}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock upload session append
	mux.HandleFunc("/2/files/upload_session/append_v2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock upload session finish
	mux.HandleFunc("/2/files/upload_session/finish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"name": "testfile_123.tmp",
				"path_lower": "/performance_tests/testfile_123.tmp",
				"path_display": "/performance_tests/testfile_123.tmp",
				"size": 1048576
			}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock download
	mux.HandleFunc("/2/files/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("mock file content"))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Mock delete
	mux.HandleFunc("/2/files/delete_v2", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"metadata": {
					"name": "testfile_123.tmp",
					"path_lower": "/performance_tests/testfile_123.tmp",
					"path_display": "/performance_tests/testfile_123.tmp"
				}
			}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}
func TestNextcloudIntegration(t *testing.T) {
	server := mockNextcloudServer()
	defer server.Close()

	// Create test config
	cfg := &Config{
		URL:             server.URL + "/remote.php/dav/files/testuser",
		Username:        "testuser",
		Password:        "testpass",
		ServiceType:     "nextcloud",
		InstanceName:    "test-instance",
		TestFileSizeMB:  1,
		TestChunkSizeMB: 1,
	}

	// Create client
	client := nextcloud.NewClient(cfg.URL, cfg.Username, cfg.Password)

	// Run the test
	RunTest(cfg, client)

	// Verify metrics were recorded
	// Note: In a real test, we'd check Prometheus metrics, but for now we just ensure no panics
	time.Sleep(100 * time.Millisecond) // Allow async operations to complete
}
