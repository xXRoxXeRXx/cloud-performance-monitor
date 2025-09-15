package hidrive_legacy

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_access_token")

	if client.AccessToken != "test_access_token" {
		t.Errorf("Expected AccessToken to be 'test_access_token', got '%s'", client.AccessToken)
	}
	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
}

func TestNewClientWithOAuth2(t *testing.T) {
	// Note: This test would require valid OAuth2 credentials to work properly.
	// For unit testing, we skip the actual OAuth2 flow and just test that the
	// function doesn't panic with invalid credentials.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewClientWithOAuth2 panicked: %v", r)
		}
	}()

	// This will fail due to invalid credentials, but should not panic
	_, err := NewClientWithOAuth2("invalid_refresh_token", "invalid_client_id", "invalid_client_secret")
	if err == nil {
		t.Error("Expected NewClientWithOAuth2 to fail with invalid credentials")
	}
}

// Note: Integration tests for EnsureDirectory, DownloadFile, and DeleteFile
// would require valid HiDrive API credentials and are not included here.
// These methods should be tested in integration test suites with proper
// mock servers or real HiDrive API endpoints.
