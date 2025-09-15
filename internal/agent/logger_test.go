package agent

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNewStructuredLogger(t *testing.T) {
	logger := NewStructuredLogger(INFO, "test-service", true)
	
	if logger == nil {
		t.Fatal("NewStructuredLogger returned nil")
	}
	
	if logger.level != INFO {
		t.Errorf("Expected level INFO, got %v", logger.level)
	}
	
	if logger.service != "test-service" {
		t.Errorf("Expected service 'test-service', got %s", logger.service)
	}
	
	if !logger.jsonFormat {
		t.Error("Expected JSON format to be true")
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}
	
	for _, test := range tests {
		if got := test.level.String(); got != test.expected {
			t.Errorf("LogLevel(%d).String() = %s, want %s", 
				test.level, got, test.expected)
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"debug", DEBUG},
		{"INFO", INFO},
		{"info", INFO},
		{"WARN", WARN},
		{"WARNING", WARN},
		{"warn", WARN},
		{"ERROR", ERROR},
		{"error", ERROR},
		{"invalid", INFO}, // defaults to INFO
		{"", INFO},        // defaults to INFO
	}
	
	for _, test := range tests {
		if got := ParseLogLevel(test.input); got != test.expected {
			t.Errorf("ParseLogLevel(%s) = %v, want %v", 
				test.input, got, test.expected)
		}
	}
}

func TestStructuredLogger_LogFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(WARN, "test-service", false)
	logger.SetOutput(&buf)
	
	// These should not be logged (below WARN level)
	logger.Debug("debug message")
	logger.Info("info message")
	
	// These should be logged (WARN level and above)
	logger.Warn("warn message")
	logger.Error("error message", nil)
	
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	
	// Should only have 2 lines (WARN and ERROR)
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d: %v", len(lines), lines)
	}
	
	// Check that debug and info messages are not present
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at WARN level")
	}
	
	if strings.Contains(output, "info message") {
		t.Error("Info message should not be logged at WARN level")
	}
	
	// Check that warn and error messages are present
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be logged at WARN level")
	}
	
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be logged at WARN level")
	}
}

func TestStructuredLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(INFO, "test-service", false)
	logger.SetOutput(&buf)
	
	logger.InfoWithFields("nextcloud", "test-instance", "test message", "1.5s", "success")
	
	output := buf.String()
	
	// Check that all expected components are present in text format
	expectedComponents := []string{
		"INFO",
		"[nextcloud]",
		"[test-instance]",
		"test message",
		"(duration: 1.5s)",
		"(status: success)",
	}
	
	for _, component := range expectedComponents {
		if !strings.Contains(output, component) {
			t.Errorf("Expected component '%s' not found in output: %s", 
				component, output)
		}
	}
}

func TestStructuredLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(INFO, "test-service", true)
	logger.SetOutput(&buf)
	
	logger.InfoWithFields("nextcloud", "test-instance", "test message", "1.5s", "success")
	
	output := buf.String()
	
	// Check that JSON components are present
	expectedJSONComponents := []string{
		`"level":"INFO"`,
		`"service":"nextcloud"`,
		`"instance":"test-instance"`,
		`"message":"test message"`,
		`"duration":"1.5s"`,
		`"status":"success"`,
		`"timestamp"`,
	}
	
	for _, component := range expectedJSONComponents {
		if !strings.Contains(output, component) {
			t.Errorf("Expected JSON component '%s' not found in output: %s", 
				component, output)
		}
	}
}

func TestStructuredLogger_TestResult(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructuredLogger(INFO, "test-service", false)
	logger.SetOutput(&buf)
	
	// Test successful result
	logger.TestResult("nextcloud", "instance1", "UPLOAD", "completed", "2.3s", "SUCCESS", nil)
	
	output := buf.String()
	
	if !strings.Contains(output, "INFO") {
		t.Error("Successful test should log at INFO level")
	}
	
	if !strings.Contains(output, "UPLOAD completed") {
		t.Error("Test result should include test type and message")
	}
	
	// Clear buffer and test failed result
	buf.Reset()
	
	// Test failed result
	logger.TestResult("nextcloud", "instance1", "DOWNLOAD", "failed", "0.5s", "FAILED", 
		errors.New("test error"))
	
	output = buf.String()
	
	if !strings.Contains(output, "ERROR") {
		t.Error("Failed test should log at ERROR level")
	}
}
