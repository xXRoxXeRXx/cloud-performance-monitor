package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ValidateConfig validates a single configuration
func ValidateConfig(cfg *Config) error {
	// 1. Validate URL
	if err := validateURL(cfg.URL); err != nil {
		return fmt.Errorf("invalid URL for %s: %w", cfg.InstanceName, err)
	}
	
	// 2. Validate service type
	if err := validateServiceType(cfg.ServiceType); err != nil {
		return fmt.Errorf("invalid service type for %s: %w", cfg.InstanceName, err)
	}
	
	// 3. Validate credentials are not empty
	if err := validateCredentials(cfg.Username, cfg.Password); err != nil {
		return fmt.Errorf("invalid credentials for %s: %w", cfg.InstanceName, err)
	}
	
	// 4. Validate test parameters
	if err := validateTestParams(cfg); err != nil {
		return fmt.Errorf("invalid test parameters for %s: %w", cfg.InstanceName, err)
	}
	
	return nil
}

// validateURL validates and parses the URL
func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}
	
	return nil
}

// validateServiceType validates the service type
func validateServiceType(serviceType string) error {
	validTypes := []string{"nextcloud", "hidrive"}
	
	for _, valid := range validTypes {
		if serviceType == valid {
			return nil
		}
	}
	
	return fmt.Errorf("unsupported service type '%s', must be one of: %s", 
		serviceType, strings.Join(validTypes, ", "))
}

// validateCredentials validates username and password
func validateCredentials(username, password string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	
	if len(password) < 8 {
		return fmt.Errorf("password too short (minimum 8 characters)")
	}
	
	return nil
}

// validateTestParams validates test parameters
func validateTestParams(cfg *Config) error {
	if cfg.TestFileSizeMB <= 0 {
		return fmt.Errorf("test file size must be positive, got %d MB", cfg.TestFileSizeMB)
	}
	
	if cfg.TestFileSizeMB > 10240 { // 10GB limit
		return fmt.Errorf("test file size too large, maximum 10240 MB, got %d MB", cfg.TestFileSizeMB)
	}
	
	if cfg.TestIntervalSec < 30 {
		return fmt.Errorf("test interval too short, minimum 30 seconds, got %d", cfg.TestIntervalSec)
	}
	
	if cfg.TestChunkSizeMB <= 0 {
		return fmt.Errorf("chunk size must be positive, got %d MB", cfg.TestChunkSizeMB)
	}
	
	if cfg.TestChunkSizeMB > cfg.TestFileSizeMB {
		return fmt.Errorf("chunk size (%d MB) cannot be larger than file size (%d MB)", 
			cfg.TestChunkSizeMB, cfg.TestFileSizeMB)
	}
	
	return nil
}

// TestCredentials tests if credentials are valid by making a simple request
func TestCredentials(ctx context.Context, cfg *Config) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	var testURL string
	switch cfg.ServiceType {
	case "nextcloud":
		testURL = cfg.URL + "/remote.php/dav/files/" + cfg.Username + "/"
	case "hidrive":
		testURL = cfg.URL + "/remote.php/dav/files/" + cfg.Username + "/"
	default:
		return fmt.Errorf("unsupported service type for credential test: %s", cfg.ServiceType)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}
	
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Depth", "0")
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", cfg.URL, err)
	}
	defer resp.Body.Close()
	
	switch resp.StatusCode {
	case http.StatusOK, http.StatusMultiStatus:
		return nil // Success
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed - invalid username/password")
	case http.StatusForbidden:
		return fmt.Errorf("access forbidden - user may not have WebDAV access")
	case http.StatusNotFound:
		return fmt.Errorf("WebDAV endpoint not found - check URL")
	default:
		return fmt.Errorf("unexpected response status: %d %s", resp.StatusCode, resp.Status)
	}
}
