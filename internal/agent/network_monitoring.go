package agent

import (
	"context"
	"net"
	"net/url"
	"time"
)

// MeasureNetworkLatency measures the network latency to a given URL
func MeasureNetworkLatency(targetURL string) (time.Duration, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return 0, err
	}

	// Default to port 443 for HTTPS, 80 for HTTP
	port := parsedURL.Port()
	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Measure TCP connection time
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(parsedURL.Hostname(), port), 5*time.Second)
	latency := time.Since(start)
	
	if err != nil {
		return 0, err
	}
	conn.Close()
	
	return latency, nil
}

// UpdateNetworkLatencyMetrics periodically measures and updates network latency metrics
func UpdateNetworkLatencyMetrics(ctx context.Context, cfg *Config, serviceType string) {
	ticker := time.NewTicker(30 * time.Second) // Measure every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latency, err := MeasureNetworkLatency(cfg.URL)
			if err != nil {
				// Increment connection timeout counter
				ConnectionTimeouts.WithLabelValues(serviceType, cfg.URL).Inc()
			} else {
				// Update latency metric
				NetworkLatency.WithLabelValues(serviceType, cfg.URL).Set(float64(latency.Milliseconds()))
			}
		}
	}
}
