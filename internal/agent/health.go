package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"` // "healthy", "unhealthy", "unknown"
	LastCheck  time.Time `json:"last_check"`
	LastError  string    `json:"last_error,omitempty"`
	ResponseTime time.Duration `json:"response_time_ms"`
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status    string          `json:"status"`
	Timestamp time.Time       `json:"timestamp"`
	Uptime    time.Duration   `json:"uptime_seconds"`
	Services  []ServiceHealth `json:"services"`
	Version   string          `json:"version"`
}

// HealthChecker manages health checks for all services
type HealthChecker struct {
	mu        sync.RWMutex
	services  map[string]*ServiceHealth
	startTime time.Time
	version   string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		services:  make(map[string]*ServiceHealth),
		startTime: time.Now(),
		version:   version,
	}
}

// RegisterService registers a service for health checking
func (hc *HealthChecker) RegisterService(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	hc.services[name] = &ServiceHealth{
		Name:      name,
		Status:    "unknown",
		LastCheck: time.Now(),
	}
}

// UpdateServiceHealth updates the health status of a service
func (hc *HealthChecker) UpdateServiceHealth(name, status string, responseTime time.Duration, err error) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	service, exists := hc.services[name]
	if !exists {
		service = &ServiceHealth{Name: name}
		hc.services[name] = service
	}
	
	service.Status = status
	service.LastCheck = time.Now()
	service.ResponseTime = responseTime
	
	if err != nil {
		service.LastError = err.Error()
	} else {
		service.LastError = ""
	}
}

// GetHealthStatus returns the current health status
func (hc *HealthChecker) GetHealthStatus() HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	services := make([]ServiceHealth, 0, len(hc.services))
	overallHealthy := true
	
	for _, service := range hc.services {
		services = append(services, *service)
		if service.Status != "healthy" {
			overallHealthy = false
		}
	}
	
	status := "healthy"
	if !overallHealthy {
		status = "unhealthy"
	}
	if len(services) == 0 {
		status = "unknown"
	}
	
	return HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    time.Since(hc.startTime),
		Services:  services,
		Version:   hc.version,
	}
}

// HealthHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		health := hc.GetHealthStatus()
		
		w.Header().Set("Content-Type", "application/json")
		
		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if health.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}
		
		w.WriteHeader(statusCode)
		
		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Failed to encode health status", http.StatusInternalServerError)
			return
		}
	}
}

// LivenessHandler returns a simple liveness probe
func (hc *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"alive"}`)
	}
}

// ReadinessHandler returns a readiness probe
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		health := hc.GetHealthStatus()
		
		w.Header().Set("Content-Type", "application/json")
		
		// Ready if at least one service is healthy
		ready := false
		for _, service := range health.Services {
			if service.Status == "healthy" {
				ready = true
				break
			}
		}
		
		if ready {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"ready"}`)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"status":"not ready"}`)
		}
	}
}
