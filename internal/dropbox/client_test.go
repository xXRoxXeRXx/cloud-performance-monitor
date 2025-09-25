package dropbox

import (
	"testing"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_access_token", "test_refresh_token", "test_app_key", "test_app_secret", &utils.DefaultClientLogger{})

	if client.AccessToken != "test_access_token" {
		t.Errorf("Expected AccessToken to be 'test_access_token', got '%s'", client.AccessToken)
	}
	if client.RefreshToken != "test_refresh_token" {
		t.Errorf("Expected RefreshToken to be 'test_refresh_token', got '%s'", client.RefreshToken)
	}
	if client.AppKey != "test_app_key" {
		t.Errorf("Expected AppKey to be 'test_app_key', got '%s'", client.AppKey)
	}
	if client.AppSecret != "test_app_secret" {
		t.Errorf("Expected AppSecret to be 'test_app_secret', got '%s'", client.AppSecret)
	}
	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
}

func TestNewClientWithOAuth2(t *testing.T) {
	client := NewClientWithOAuth2("test_access_token", "test_refresh_token", "test_app_key", "test_app_secret", &utils.DefaultClientLogger{})

	if client.AccessToken != "test_access_token" {
		t.Errorf("Expected AccessToken to be 'test_access_token', got '%s'", client.AccessToken)
	}
	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
}

// Note: Integration tests for EnsureDirectory, DownloadFile, and DeleteFile
// would require valid Dropbox API credentials and are not included here.
// These methods should be tested in integration test suites with proper
// mock servers or real Dropbox API endpoints.
