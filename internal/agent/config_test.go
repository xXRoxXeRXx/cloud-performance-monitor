package agent

import (
	"os"
	"testing"
)

func TestLoadConfigs(t *testing.T) {
	// Set environment variables
	os.Setenv("NC_INSTANCE_1_URL", "https://test.com")
	os.Setenv("NC_INSTANCE_1_USER", "user")
	os.Setenv("NC_INSTANCE_1_PASS", "validpassword123") // At least 8 characters
	os.Setenv("TEST_FILE_SIZE_MB", "5")
	os.Setenv("TEST_INTERVAL_SECONDS", "600")
	os.Setenv("TEST_CHUNK_SIZE_MB", "5")

	defer func() {
		os.Unsetenv("NC_INSTANCE_1_URL")
		os.Unsetenv("NC_INSTANCE_1_USER")
		os.Unsetenv("NC_INSTANCE_1_PASS")
		os.Unsetenv("TEST_FILE_SIZE_MB")
		os.Unsetenv("TEST_INTERVAL_SECONDS")
		os.Unsetenv("TEST_CHUNK_SIZE_MB")
	}()

	configs, err := LoadConfigs()
	if err != nil {
		t.Errorf("LoadConfigs failed: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}

	cfg := configs[0]
	if cfg.URL != "https://test.com" {
		t.Errorf("Unexpected URL: %s", cfg.URL)
	}
	if cfg.ServiceType != "nextcloud" {
		t.Errorf("Unexpected service type: %s", cfg.ServiceType)
	}
	if cfg.TestFileSizeMB != 5 {
		t.Errorf("Unexpected file size: %d", cfg.TestFileSizeMB)
	}
}

func TestLoadConfigsInvalid(t *testing.T) {
	// No instances
	_, err := LoadConfigs()
	if err == nil {
		t.Error("Expected error for no instances")
	}

	// Invalid file size
	os.Setenv("NC_INSTANCE_1_URL", "https://test.com")
	os.Setenv("NC_INSTANCE_1_USER", "user")
	os.Setenv("NC_INSTANCE_1_PASS", "validpassword123") // Valid password
	os.Setenv("TEST_FILE_SIZE_MB", "-1")

	defer func() {
		os.Unsetenv("NC_INSTANCE_1_URL")
		os.Unsetenv("NC_INSTANCE_1_USER")
		os.Unsetenv("NC_INSTANCE_1_PASS")
		os.Unsetenv("TEST_FILE_SIZE_MB")
	}()

	_, err = LoadConfigs()
	if err == nil {
		t.Error("Expected error for invalid file size")
	}
}
