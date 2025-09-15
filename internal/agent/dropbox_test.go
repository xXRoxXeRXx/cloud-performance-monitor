package agent

import (
	"os"
	"testing"
)

func TestDropboxConfig(t *testing.T) {
	// Set environment variables for Dropbox with proper refresh token setup
	os.Setenv("DROPBOX_INSTANCE_1_REFRESH_TOKEN", "test-refresh-token-123")
	os.Setenv("DROPBOX_INSTANCE_1_APP_KEY", "test-app-key")
	os.Setenv("DROPBOX_INSTANCE_1_APP_SECRET", "test-app-secret")
	os.Setenv("DROPBOX_INSTANCE_1_NAME", "test-dropbox")
	os.Setenv("TEST_FILE_SIZE_MB", "5")
	os.Setenv("TEST_INTERVAL_SECONDS", "600")
	os.Setenv("TEST_CHUNK_SIZE_MB", "2")

	defer func() {
		os.Unsetenv("DROPBOX_INSTANCE_1_REFRESH_TOKEN")
		os.Unsetenv("DROPBOX_INSTANCE_1_APP_KEY")
		os.Unsetenv("DROPBOX_INSTANCE_1_APP_SECRET")
		os.Unsetenv("DROPBOX_INSTANCE_1_NAME")
		os.Unsetenv("TEST_FILE_SIZE_MB")
		os.Unsetenv("TEST_INTERVAL_SECONDS")
		os.Unsetenv("TEST_CHUNK_SIZE_MB")
	}()

	configs, err := LoadConfigs()
	if err != nil {
		t.Errorf("LoadConfigs failed: %v", err)
		return
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
		return
	}

	cfg := configs[0]
	if cfg.ServiceType != "dropbox" {
		t.Errorf("Expected service type 'dropbox', got %s", cfg.ServiceType)
	}
	if cfg.RefreshToken != "test-refresh-token-123" {
		t.Errorf("Expected refresh token 'test-refresh-token-123', got %s", cfg.RefreshToken)
	}
	if cfg.InstanceName != "test-dropbox" {
		t.Errorf("Expected instance name 'test-dropbox', got %s", cfg.InstanceName)
	}
	if cfg.TestFileSizeMB != 5 {
		t.Errorf("Expected file size 5, got %d", cfg.TestFileSizeMB)
	}
}

func TestDropboxConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Valid Dropbox config",
			config: &Config{
				InstanceName:    "test-dropbox",
				ServiceType:     "dropbox",
				RefreshToken:    "test-refresh-token",
				AppKey:          "test-app-key",
				AppSecret:       "test-app-secret",
				TestFileSizeMB:  10,
				TestIntervalSec: 300,
				TestChunkSizeMB: 5,
			},
			expectError: false,
		},
		{
			name: "Missing refresh token",
			config: &Config{
				InstanceName:    "test-dropbox",
				ServiceType:     "dropbox",
				RefreshToken:    "",
				AppKey:          "test-app-key",
				AppSecret:       "test-app-secret",
				TestFileSizeMB:  10,
				TestIntervalSec: 300,
				TestChunkSizeMB: 5,
			},
			expectError: true,
		},
		{
			name: "Invalid file size",
			config: &Config{
				InstanceName:    "test-dropbox",
				ServiceType:     "dropbox",
				AccessToken:     "test-token",
				TestFileSizeMB:  0,
				TestIntervalSec: 300,
				TestChunkSizeMB: 5,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateConfig(tc.config)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDropboxConfigDefault(t *testing.T) {
	// Test with minimal configuration (refresh token + app credentials)
	os.Setenv("DROPBOX_INSTANCE_1_REFRESH_TOKEN", "test-refresh-token-minimal")
	os.Setenv("DROPBOX_INSTANCE_1_APP_KEY", "test-app-key")
	os.Setenv("DROPBOX_INSTANCE_1_APP_SECRET", "test-app-secret")

	defer func() {
		os.Unsetenv("DROPBOX_INSTANCE_1_REFRESH_TOKEN")
		os.Unsetenv("DROPBOX_INSTANCE_1_APP_KEY")
		os.Unsetenv("DROPBOX_INSTANCE_1_APP_SECRET")
	}()

	configs, err := LoadConfigs()
	if err != nil {
		t.Errorf("LoadConfigs failed: %v", err)
		return
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
		return
	}

	cfg := configs[0]
	
	// Check defaults
	if cfg.InstanceName != "dropbox-instance-1" {
		t.Errorf("Expected default instance name 'dropbox-instance-1', got %s", cfg.InstanceName)
	}
	if cfg.TestFileSizeMB != DefaultFileSizeMB {
		t.Errorf("Expected default file size %d, got %d", DefaultFileSizeMB, cfg.TestFileSizeMB)
	}
	if cfg.TestIntervalSec != DefaultIntervalSec {
		t.Errorf("Expected default interval %d, got %d", DefaultIntervalSec, cfg.TestIntervalSec)
	}
	if cfg.TestChunkSizeMB != DefaultChunkSizeMB {
		t.Errorf("Expected default chunk size %d, got %d", DefaultChunkSizeMB, cfg.TestChunkSizeMB)
	}
}
