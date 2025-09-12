package main

import (
	"context"
	"log"
	"net/http"
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

	// Start the Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Metrics server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("ERROR: Metrics server failed: %v", err)
		}
	}()

	// Start sequential monitoring
	startSequentialMonitoring(allConfigs)
}

// startSequentialMonitoring runs tests for all instances sequentially - one after another
func startSequentialMonitoring(configs []*agent.Config) {
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
	runTestCycle(configs, clients)
	
	// Start the periodic testing loop
	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		log.Println("=== Starting new test cycle ===")
		runTestCycle(configs, clients)
	}
}

// runTestCycle runs tests for all instances sequentially - one after another
func runTestCycle(configs []*agent.Config, clients map[*agent.Config]interface{}) {
	cycleStart := time.Now()
	log.Printf("Starting test cycle with %d instances", len(configs))
	
	for i, cfg := range configs {
		log.Printf("[%d/%d] Starting test for instance %s (%s)", 
			i+1, len(configs), cfg.InstanceName, cfg.ServiceType)
		
		testStart := time.Now()
		runTestForInstance(cfg, clients[cfg])
		testDuration := time.Since(testStart)
		
		log.Printf("[%d/%d] Completed test for instance %s (%s) in %v", 
			i+1, len(configs), cfg.InstanceName, cfg.ServiceType, testDuration)
	}
	
	cycleDuration := time.Since(cycleStart)
	log.Printf("=== Test cycle completed in %v ===", cycleDuration)
}

// runTestForInstance runs a single test for the given instance
func runTestForInstance(cfg *agent.Config, client interface{}) {
	startTime := time.Now()
	
	switch cfg.ServiceType {
	case "nextcloud":
		if ncClient, ok := client.(*nextcloud.Client); ok {
			agent.RunTest(cfg, ncClient)
		} else {
			log.Printf("ERROR: Invalid client type for Nextcloud instance %s", cfg.InstanceName)
		}
	case "hidrive":
		ctx := context.Background()
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
