package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// UptimeCheckInterval is the interval for HTTP uptime checks
	UptimeCheckInterval = 60 * time.Second
	// UptimeCheckTimeout is the timeout for a single uptime check
	UptimeCheckTimeout = 30 * time.Second
)

// UptimeChecker performs HTTP health checks on configured instances
type UptimeChecker struct {
	config     *Config
	httpClient *http.Client
	logger     *StructuredLogger
}

// NewUptimeChecker creates a new uptime checker for the given configuration
func NewUptimeChecker(config *Config, logger *StructuredLogger) *UptimeChecker {
	return &UptimeChecker{
		config: config,
		httpClient: &http.Client{
			Timeout: UptimeCheckTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Follow up to 5 redirects
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		logger: logger,
	}
}

// Check performs a single uptime check
func (u *UptimeChecker) Check(ctx context.Context) {
	start := time.Now()
	
	// Determine the URL to check based on service type
	checkURL := u.getCheckURL()
	
	u.logger.LogOperation(INFO, u.config.ServiceType, u.config.InstanceName, 
		"uptime_check", "start", 
		fmt.Sprintf("Starting uptime check for URL: %s", checkURL))

	// Create HTTP request with context
	// Use GET instead of HEAD because status.php endpoints typically require GET
	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		u.logger.LogOperation(ERROR, u.config.ServiceType, u.config.InstanceName,
			"uptime_check", "request_creation", "Failed to create uptime check request", 
			WithError(err))
		u.recordFailure(0, time.Since(start))
		return
	}

	// Set user agent
	req.Header.Set("User-Agent", "Cloud-Performance-Monitor/1.0 (Uptime)")

	// Perform the request
	resp, err := u.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		u.logger.LogOperation(ERROR, u.config.ServiceType, u.config.InstanceName,
			"uptime_check", "http_request", "Uptime check failed",
			WithError(err),
			WithDuration(duration))
		u.recordFailure(0, duration)
		return
	}
	defer resp.Body.Close()

	// Check if status code is in acceptable range (200-299)
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		u.logger.LogOperation(INFO, u.config.ServiceType, u.config.InstanceName,
			"uptime_check", "complete", "Uptime check successful",
			WithStatusCode(resp.StatusCode),
			WithDuration(duration))
		u.recordSuccess(resp.StatusCode, duration)
	} else {
		u.logger.LogOperation(WARN, u.config.ServiceType, u.config.InstanceName,
			"uptime_check", "complete", "Uptime check returned non-success status",
			WithStatusCode(resp.StatusCode),
			WithDuration(duration))
		u.recordFailure(resp.StatusCode, duration)
	}
}

// getCheckURL returns the appropriate URL to check based on service type
func (u *UptimeChecker) getCheckURL() string {
	switch u.config.ServiceType {
	case "nextcloud":
		// Check status.php endpoint (no auth required)
		return fmt.Sprintf("%s/status.php", u.config.URL)
	case "hidrive":
		// Check status.php endpoint (no auth required)
		return fmt.Sprintf("%s/status.php", u.config.URL)
	case "magentacloud":
		// Check status.php endpoint (no auth required)
		return fmt.Sprintf("%s/status.php", u.config.URL)
	case "hidrive_legacy":
		// Check base API URL (no auth required for connectivity test)
		return fmt.Sprintf("%s/2.1/", u.config.URL)
	case "dropbox":
		// Check main Dropbox website (no auth needed for basic connectivity)
		return "https://www.dropbox.com"
	default:
		// Fallback to base URL
		return u.config.URL
	}
}

// recordSuccess records a successful uptime check
func (u *UptimeChecker) recordSuccess(statusCode int, duration time.Duration) {
	labels := prometheus.Labels{
		"service":  u.config.ServiceType,
		"instance": u.config.InstanceName,
	}
	
	cloudUptimeStatus.With(labels).Set(1)
	cloudUptimeResponseTime.With(labels).Set(duration.Seconds())
	cloudUptimeHTTPStatus.With(labels).Set(float64(statusCode))
	cloudUptimeChecksTotal.With(prometheus.Labels{
		"service":  u.config.ServiceType,
		"instance": u.config.InstanceName,
		"result":   "success",
	}).Inc()
}

// recordFailure records a failed uptime check
func (u *UptimeChecker) recordFailure(statusCode int, duration time.Duration) {
	labels := prometheus.Labels{
		"service":  u.config.ServiceType,
		"instance": u.config.InstanceName,
	}
	
	cloudUptimeStatus.With(labels).Set(0)
	cloudUptimeResponseTime.With(labels).Set(duration.Seconds())
	if statusCode > 0 {
		cloudUptimeHTTPStatus.With(labels).Set(float64(statusCode))
	}
	cloudUptimeChecksTotal.With(prometheus.Labels{
		"service":  u.config.ServiceType,
		"instance": u.config.InstanceName,
		"result":   "failure",
	}).Inc()
}

// Run starts the uptime checker loop
func (u *UptimeChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(UptimeCheckInterval)
	defer ticker.Stop()

	u.logger.LogOperation(INFO, u.config.ServiceType, u.config.InstanceName,
		"uptime_checker", "start", 
		fmt.Sprintf("Starting uptime checker with %v interval", UptimeCheckInterval))

	// Perform initial check immediately
	u.Check(ctx)

	for {
		select {
		case <-ctx.Done():
			u.logger.LogOperation(INFO, u.config.ServiceType, u.config.InstanceName,
				"uptime_checker", "stop", "Uptime checker stopped")
			return
		case <-ticker.C:
			u.Check(ctx)
		}
	}
}
