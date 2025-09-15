package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	version := "1.0.0"
	hc := NewHealthChecker(version)
	
	if hc == nil {
		t.Fatal("NewHealthChecker returned nil")
	}
	
	if hc.version != version {
		t.Errorf("Expected version %s, got %s", version, hc.version)
	}
	
	if hc.services == nil {
		t.Error("Services map should be initialized")
	}
}

func TestRegisterService(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	serviceName := "test-service"
	
	hc.RegisterService(serviceName)
	
	if _, exists := hc.services[serviceName]; !exists {
		t.Errorf("Service %s was not registered", serviceName)
	}
	
	service := hc.services[serviceName]
	if service.Name != serviceName {
		t.Errorf("Expected service name %s, got %s", serviceName, service.Name)
	}
	
	if service.Status != "unknown" {
		t.Errorf("Expected initial status 'unknown', got %s", service.Status)
	}
}

func TestUpdateServiceHealth(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	serviceName := "test-service"
	
	// Update health for non-registered service (should auto-register)
	hc.UpdateServiceHealth(serviceName, "healthy", 100*time.Millisecond, nil)
	
	service, exists := hc.services[serviceName]
	if !exists {
		t.Fatalf("Service %s was not auto-registered", serviceName)
	}
	
	if service.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", service.Status)
	}
	
	if service.ResponseTime != 100*time.Millisecond {
		t.Errorf("Expected response time 100ms, got %v", service.ResponseTime)
	}
	
	if service.LastError != "" {
		t.Errorf("Expected empty error, got %s", service.LastError)
	}
}

func TestGetHealthStatus(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	
	// Test with no services
	status := hc.GetHealthStatus()
	if status.Status != "unknown" {
		t.Errorf("Expected status 'unknown' with no services, got %s", status.Status)
	}
	
	// Add healthy service
	hc.RegisterService("healthy-service")
	hc.UpdateServiceHealth("healthy-service", "healthy", 50*time.Millisecond, nil)
	
	status = hc.GetHealthStatus()
	if status.Status != "healthy" {
		t.Errorf("Expected overall status 'healthy', got %s", status.Status)
	}
	
	// Add unhealthy service
	hc.RegisterService("unhealthy-service")
	hc.UpdateServiceHealth("unhealthy-service", "unhealthy", 200*time.Millisecond, nil)
	
	status = hc.GetHealthStatus()
	if status.Status != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy', got %s", status.Status)
	}
	
	if len(status.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(status.Services))
	}
}

func TestHealthHandler(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	hc.RegisterService("test-service")
	hc.UpdateServiceHealth("test-service", "healthy", 50*time.Millisecond, nil)
	
	handler := hc.HealthHandler()
	
	// Test GET request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}
	
	// Test non-GET request
	req = httptest.NewRequest("POST", "/health", nil)
	w = httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for POST, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestLivenessHandler(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	handler := hc.LivenessHandler()
	
	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	
	body := w.Body.String()
	expected := `{"status":"alive"}`
	if body != expected {
		t.Errorf("Expected body %s, got %s", expected, body)
	}
}

func TestReadinessHandler(t *testing.T) {
	hc := NewHealthChecker("1.0.0")
	handler := hc.ReadinessHandler()
	
	// Test with no healthy services
	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d with no healthy services, got %d", 
			http.StatusServiceUnavailable, w.Code)
	}
	
	// Add healthy service
	hc.RegisterService("healthy-service")
	hc.UpdateServiceHealth("healthy-service", "healthy", 50*time.Millisecond, nil)
	
	w = httptest.NewRecorder()
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d with healthy service, got %d", 
			http.StatusOK, w.Code)
	}
}
