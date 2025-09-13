package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MarcelWMeyer/nextcloud-performance-monitor/internal/agent"
	"github.com/MarcelWMeyer/nextcloud-performance-monitor/internal/nextcloud"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load all configurations
	allConfigs, err := agent.LoadConfigs()
	if err != nil {
		log.Fatalf("ERROR: Could not load configuration: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the Prometheus metrics server
	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}
	
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Metrics server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ERROR: Metrics server failed: %v", err)
		}
	}()

	// Start sequential monitoring in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startSequentialMonitoring(ctx, allConfigs)
	}()
	
	// Start network latency monitoring for all instances
	for _, cfg := range allConfigs {
		wg.Add(1)
		go func(config *agent.Config) {
			defer wg.Done()
			agent.UpdateNetworkLatencyMetrics(ctx, config, config.ServiceType)
		}(cfg)
	}

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %s, initiating graceful shutdown...", sig)

	// Cancel context to signal shutdown
	cancel()

	// Shutdown metrics server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown error: %v", err)
	} else {
		log.Println("Metrics server shut down gracefully")
	}

	// Wait for monitoring to finish current tests
	log.Println("Waiting for running tests to complete...")
	wg.Wait()
	log.Println("Graceful shutdown completed")
}

// startSequentialMonitoring runs tests for all instances sequentially - one after another
func startSequentialMonitoring(ctx context.Context, configs []*agent.Config) {
	log.Printf("Starting sequential monitoring for %d instances", len(configs))
	
	if len(configs) == 0 {
		log.Println("No instances configured, exiting")
		return
	}
	
	// Use the first config's interval as base interval
	baseInterval := time.Duration(configs[0].TestIntervalSec) * time.Second
	log.Printf("Test cycle interval: %v", baseInterval)
	
	// Create clients for all instances
	clients := make(map[*agent.Config]interface{})
	for _, cfg := range configs {
		switch cfg.ServiceType {
		case "nextcloud":
			clients[cfg] = nextcloud.NewClient(cfg.URL, cfg.Username, cfg.Password)
		case "hidrive":
			clients[cfg] = cfg // HiDrive client is created in the test function
		}
	}
	
	// Run initial test cycle immediately
	log.Println("=== Starting initial test cycle ===")
	if !runTestCycle(ctx, configs, clients) {
		return // Shutdown signal received
	}
	
	// Start the periodic testing loop
	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, stopping monitoring...")
			return
		case <-ticker.C:
			log.Println("=== Starting new test cycle ===")
			if !runTestCycle(ctx, configs, clients) {
				return // Shutdown signal received during test
			}
		}
	}
}

// runTestCycle runs tests for all instances sequentially - one after another
// Returns false if shutdown signal received, true if completed normally
func runTestCycle(ctx context.Context, configs []*agent.Config, clients map[*agent.Config]interface{}) bool {
	cycleStart := time.Now()
	log.Printf("Starting test cycle with %d instances", len(configs))
	
	for i, cfg := range configs {
		// Check for shutdown signal before each test
		select {
		case <-ctx.Done():
			log.Printf("Shutdown signal received during test cycle, stopping after %d/%d tests", i, len(configs))
			return false
		default:
		}
		
		log.Printf("[%d/%d] Starting test for instance %s (%s)", 
			i+1, len(configs), cfg.InstanceName, cfg.ServiceType)
		
		testStart := time.Now()
		runTestForInstance(ctx, cfg, clients[cfg])
		testDuration := time.Since(testStart)
		
		log.Printf("[%d/%d] Completed test for instance %s (%s) in %v", 
			i+1, len(configs), cfg.InstanceName, cfg.ServiceType, testDuration)
	}
	
	cycleDuration := time.Since(cycleStart)
	log.Printf("=== Test cycle completed in %v ===", cycleDuration)
	return true
}

// runTestForInstance runs a single test for the given instance
func runTestForInstance(ctx context.Context, cfg *agent.Config, client interface{}) {
	startTime := time.Now()
	
	switch cfg.ServiceType {
	case "nextcloud":
		if ncClient, ok := client.(*nextcloud.Client); ok {
			agent.RunTest(cfg, ncClient)
		} else {
			log.Printf("ERROR: Invalid client type for Nextcloud instance %s", cfg.InstanceName)
		}
	case "hidrive":
		if err := agent.RunHiDriveTest(ctx, cfg); err != nil {
			log.Printf("[HiDrive] Test error for %s: %v", cfg.URL, err)
		}
	default:
		log.Printf("Unknown service type: %s", cfg.ServiceType)
	}
	
	duration := time.Since(startTime)
	log.Printf("Test completed for %s (%s) in %v", cfg.InstanceName, cfg.ServiceType, duration)
}

// Legacy function - removed parallel execution
func startMonitoringInstance(cfg *agent.Config) {
	// This function is no longer used but kept for backward compatibility
	log.Printf("WARNING: startMonitoringInstance called for %s - this should not happen in sequential mode", cfg.InstanceName)
}
