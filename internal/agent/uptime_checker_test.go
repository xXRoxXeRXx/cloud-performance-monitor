package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewUptimeChecker(t *testing.T) {
	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: "test-instance",
		URL:          "https://cloud.example.com",
	}

	checker := NewUptimeChecker(config, logger)

	if checker == nil {
		t.Fatal("Expected checker to be created")
	}

	if checker.config != config {
		t.Error("Config not set correctly")
	}

	if checker.logger != logger {
		t.Error("Logger not set correctly")
	}

	if checker.httpClient == nil {
		t.Error("HTTP client not initialized")
	}

	if checker.httpClient.Timeout != UptimeCheckTimeout {
		t.Errorf("Expected timeout %v, got %v", UptimeCheckTimeout, checker.httpClient.Timeout)
	}
}

func TestGetCheckURL(t *testing.T) {
	logger := NewStructuredLogger(INFO, "test", false)

	tests := []struct {
		name        string
		serviceType string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "Nextcloud status.php",
			serviceType: "nextcloud",
			baseURL:     "https://cloud.example.com",
			expectedURL: "https://cloud.example.com/",
		},
		{
			name:        "HiDrive status.php",
			serviceType: "hidrive",
			baseURL:     "https://storage.ionos.fr",
			expectedURL: "https://storage.ionos.fr/",
		},
		{
			name:        "MagentaCLOUD status.php",
			serviceType: "magentacloud",
			baseURL:     "https://magentacloud.de",
			expectedURL: "https://magentacloud.de/",
		},
		{
			name:        "HiDrive Legacy Website",
			serviceType: "hidrive_legacy",
			baseURL:     "https://api.hidrive.strato.com",
			expectedURL: "https://my.hidrive.com/",
		},
		{
			name:        "Dropbox Website",
			serviceType: "dropbox",
			baseURL:     "https://api.dropboxapi.com",
			expectedURL: "https://www.dropbox.com",
		},
		{
			name:        "Unknown Service",
			serviceType: "unknown",
			baseURL:     "https://example.com",
			expectedURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ServiceType:  tt.serviceType,
				InstanceName: "test-instance",
				URL:          tt.baseURL,
			}

			checker := NewUptimeChecker(config, logger)
			actualURL := checker.getCheckURL()

			if actualURL != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, actualURL)
			}
		})
	}
}

func TestUptimeCheck_Success(t *testing.T) {
	// Create a test HTTP server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: server.URL,
		URL:          server.URL,
	}

	checker := NewUptimeChecker(config, logger)
	ctx := context.Background()

	// Perform check
	checker.Check(ctx)

	// Give metrics time to update
	time.Sleep(100 * time.Millisecond)

	// Note: We can't easily test Prometheus metrics in unit tests without more complex setup
	// In a real scenario, you would use prometheus testutil to verify metrics
}

func TestUptimeCheck_Failure(t *testing.T) {
	// Create a test HTTP server that returns 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: server.URL,
		URL:          server.URL,
	}

	checker := NewUptimeChecker(config, logger)
	ctx := context.Background()

	// Perform check
	checker.Check(ctx)

	// Give metrics time to update
	time.Sleep(100 * time.Millisecond)

	// Note: We can't easily test Prometheus metrics in unit tests without more complex setup
}

func TestUptimeCheck_Timeout(t *testing.T) {
	// Create a test HTTP server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: server.URL,
		URL:          server.URL,
	}

	checker := NewUptimeChecker(config, logger)
	
	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Perform check (should timeout)
	checker.Check(ctx)

	// Give metrics time to update
	time.Sleep(100 * time.Millisecond)
}

func TestUptimeCheck_RedirectHandling(t *testing.T) {
	// Create a test HTTP server that redirects
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectServer.Close()

	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectServer.URL, http.StatusMovedPermanently)
	}))
	defer mainServer.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: mainServer.URL,
		URL:          mainServer.URL,
	}

	checker := NewUptimeChecker(config, logger)
	ctx := context.Background()

	// Perform check (should follow redirect)
	checker.Check(ctx)

	// Give metrics time to update
	time.Sleep(100 * time.Millisecond)
}

func TestUptimeCheck_AcceptableStatusCodes(t *testing.T) {
	logger := NewStructuredLogger(INFO, "test", false)

	tests := []struct {
		name       string
		statusCode int
		shouldPass bool
	}{
		{"OK", http.StatusOK, true},
		{"Created", http.StatusCreated, true},
		{"Accepted", http.StatusAccepted, true},
		{"No Content", http.StatusNoContent, true},
		{"Moved Permanently", http.StatusMovedPermanently, false},
		{"Bad Request", http.StatusBadRequest, false},
		{"Unauthorized", http.StatusUnauthorized, false},
		{"Not Found", http.StatusNotFound, false},
		{"Internal Server Error", http.StatusInternalServerError, false},
		{"Service Unavailable", http.StatusServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			config := &Config{
				ServiceType:  "nextcloud",
				InstanceName: server.URL,
				URL:          server.URL,
			}

			checker := NewUptimeChecker(config, logger)
			ctx := context.Background()

			// Perform check
			checker.Check(ctx)

			// Give metrics time to update
			time.Sleep(100 * time.Millisecond)

			// Note: In a real test, we would verify the metrics here
			// For now, we just ensure the check doesn't panic
		})
	}
}

func TestUptimeChecker_Run_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: server.URL,
		URL:          server.URL,
	}

	checker := NewUptimeChecker(config, logger)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Run checker in goroutine
	done := make(chan bool)
	go func() {
		checker.Run(ctx)
		done <- true
	}()

	// Wait a bit to ensure at least one check runs
	time.Sleep(200 * time.Millisecond)

	// Cancel the context
	cancel()

	// Wait for Run to complete
	select {
	case <-done:
		// Success - Run completed after cancellation
	case <-time.After(2 * time.Second):
		t.Error("Run did not complete after context cancellation")
	}
}

func TestUptimeChecker_UserAgent(t *testing.T) {
	receivedUA := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := NewStructuredLogger(INFO, "test", false)
	config := &Config{
		ServiceType:  "nextcloud",
		InstanceName: server.URL,
		URL:          server.URL,
	}

	checker := NewUptimeChecker(config, logger)
	ctx := context.Background()

	checker.Check(ctx)

	// Give server time to process
	time.Sleep(100 * time.Millisecond)

	expectedUA := "Cloud-Performance-Monitor/1.0 (Uptime)"
	if receivedUA != expectedUA {
		t.Errorf("Expected User-Agent %s, got %s", expectedUA, receivedUA)
	}
}
