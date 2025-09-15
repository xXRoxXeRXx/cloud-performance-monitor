package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MarcelWMeyer/cloud-performance-monitor/internal/agent"
	"github.com/MarcelWMeyer/cloud-performance-monitor/internal/nextcloud"
	"github.com/MarcelWMeyer/cloud-performance-monitor/internal/magentacloud"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Version = "1.0.0"
	DefaultShutdownTimeout = 30 * time.Second
)

func main() {
	// Initialize structured logging
	logLevel := agent.GetLogLevel()
	logFormat := agent.GetLogFormat()
	agent.InitLogger(logLevel, "monitor-agent", logFormat)
	
	agent.Logger.InfoWithFields("monitor-agent", Version, 
		"Starting Cloud Performance Monitor", "", "")
	
	// Create shutdown manager
	shutdownManager := agent.NewShutdownManager(DefaultShutdownTimeout)
	
	// Load all configurations
	allConfigs, err := agent.LoadConfigs()
	if err != nil {
		agent.Logger.Error("Could not load configuration", err)
		os.Exit(1)
	}
	
	agent.Logger.InfoWithFields("monitor-agent", "", 
		fmt.Sprintf("Loaded %d service configurations", len(allConfigs)), "", "")
	
	// Create health checker
	healthChecker := agent.NewHealthChecker(Version)
	
	// Register all services with health checker
	for _, cfg := range allConfigs {
		healthChecker.RegisterService(cfg.InstanceName)
	}
	
	// Create test manager
	testManager := agent.NewTestManager(shutdownManager)
	
	// Setup HTTP server with health endpoints
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthChecker.HealthHandler())
	mux.HandleFunc("/health/live", healthChecker.LivenessHandler())
	mux.HandleFunc("/health/ready", healthChecker.ReadinessHandler())
	
	// Create HTTP server manager
	httpManager := agent.NewHTTPServerManager(":8080", mux)
	
	// Register HTTP server shutdown hook
	shutdownManager.AddHook(httpManager.ShutdownHook())
	
	// Start HTTP server
	go func() {
		if err := httpManager.Start(); err != nil {
			agent.Logger.Error("HTTP server failed to start", err)
			os.Exit(1)
		}
	}()
	
	agent.Logger.InfoWithFields("http-server", ":8080", 
		"HTTP server started with endpoints: /metrics, /health, /health/live, /health/ready", "", "")
	
	// Start sequential monitoring
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startSequentialMonitoring(shutdownManager.Context(), allConfigs, healthChecker, testManager)
	}()
	
	// Start network latency monitoring for all instances
	for _, cfg := range allConfigs {
		wg.Add(1)
		go func(config *agent.Config) {
			defer wg.Done()
			agent.UpdateNetworkLatencyMetrics(shutdownManager.Context(), config, config.ServiceType)
		}(cfg)
	}
	
	// Wait for shutdown signal and perform graceful shutdown
	if err := shutdownManager.WaitForShutdown(); err != nil {
		agent.Logger.Error("Shutdown completed with errors", err)
		os.Exit(1)
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	agent.Logger.Info("Application shutdown completed successfully")
}

// startSequentialMonitoring runs tests for all instances sequentially - one after another
func startSequentialMonitoring(ctx context.Context, configs []*agent.Config, healthChecker *agent.HealthChecker, testManager *agent.TestManager) {
	agent.Logger.InfoWithFields("monitor-agent", "", 
		fmt.Sprintf("Starting sequential monitoring for %d instances", len(configs)), "", "")
	
	if len(configs) == 0 {
		agent.Logger.Warn("No instances configured, exiting")
		return
	}
	
	// Use the first config's interval as base interval
	baseInterval := time.Duration(configs[0].TestIntervalSec) * time.Second
	agent.Logger.InfoWithFields("monitor-agent", "", 
		fmt.Sprintf("Test cycle interval: %v", baseInterval), "", "")
	
	// Create clients for all instances
	clients := make(map[*agent.Config]interface{})
	for _, cfg := range configs {
		switch cfg.ServiceType {
		case "nextcloud":
			clients[cfg] = nextcloud.NewClient(cfg.URL, cfg.Username, cfg.Password)
		case "hidrive":
			clients[cfg] = cfg // HiDrive client is created in the test function
		case "hidrive_legacy":
			clients[cfg] = cfg // HiDrive Legacy client is created in the test function
		case "dropbox":
			clients[cfg] = cfg // Dropbox client is created in the test function
		case "magentacloud":
			clients[cfg] = magentacloud.NewClient(cfg.URL, cfg.Username, cfg.Password, cfg.ANID)
		}
	}
	
	// Run initial test cycle immediately
	agent.Logger.Info("Starting initial test cycle")
	if !runTestCycle(ctx, configs, clients, healthChecker, testManager) {
		return // Shutdown signal received
	}
	
	// Start the periodic testing loop
	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			agent.Logger.Info("Shutdown signal received, stopping monitoring...")
			return
		case <-ticker.C:
			agent.Logger.Info("Starting new test cycle")
			if !runTestCycle(ctx, configs, clients, healthChecker, testManager) {
				return // Shutdown signal received during test
			}
		}
	}
}

// runTestCycle runs tests for all instances sequentially - one after another
// Returns false if shutdown signal received, true if completed normally
func runTestCycle(ctx context.Context, configs []*agent.Config, clients map[*agent.Config]interface{}, healthChecker *agent.HealthChecker, testManager *agent.TestManager) bool {
	cycleStart := time.Now()
	agent.Logger.InfoWithFields("monitor-agent", "", 
		fmt.Sprintf("Starting test cycle with %d instances", len(configs)), "", "")
	
	for i, cfg := range configs {
		// Check for shutdown signal before each test
		select {
		case <-ctx.Done():
			agent.Logger.InfoWithFields("monitor-agent", "", 
				fmt.Sprintf("Shutdown signal received during test cycle, stopping after %d/%d tests", i, len(configs)), "", "")
			return false
		default:
		}
		
		agent.Logger.InfoWithFields(cfg.ServiceType, cfg.InstanceName, 
			fmt.Sprintf("Starting test [%d/%d]", i+1, len(configs)), "", "")
		
		testStart := time.Now()
		
		// Run test directly (synchronously) for sequential execution
		err := runTestForInstance(ctx, cfg, clients[cfg], healthChecker)
		if err != nil {
			agent.Logger.ErrorWithFields(cfg.ServiceType, cfg.InstanceName, 
				"Test failed", err)
		}
		
		testDuration := time.Since(testStart)
		
		agent.Logger.InfoWithFields(cfg.ServiceType, cfg.InstanceName, 
			fmt.Sprintf("Completed test [%d/%d]", i+1, len(configs)), 
			testDuration.String(), "")
	}
	
	cycleDuration := time.Since(cycleStart)
	agent.Logger.InfoWithFields("monitor-agent", "", 
		"Test cycle completed", cycleDuration.String(), "")
	return true
}

// runTestForInstance runs a single test for the given instance
func runTestForInstance(ctx context.Context, cfg *agent.Config, client interface{}, healthChecker *agent.HealthChecker) error {
	startTime := time.Now()
	var err error
	
	switch cfg.ServiceType {
	case "nextcloud":
		if ncClient, ok := client.(*nextcloud.Client); ok {
			agent.RunTest(cfg, ncClient)
		} else {
			err = fmt.Errorf("invalid client type for Nextcloud instance %s", cfg.InstanceName)
		}
	case "hidrive":
		err = agent.RunHiDriveTest(ctx, cfg)
	case "hidrive_legacy":
		err = agent.RunHiDriveLegacyTest(ctx, cfg)
	case "dropbox":
		err = agent.RunDropboxTest(ctx, cfg)
	case "magentacloud":
		err = agent.RunMagentaCloudTest(ctx, cfg)
	default:
		err = fmt.Errorf("unknown service type: %s", cfg.ServiceType)
	}
	
	duration := time.Since(startTime)
	
	// Update health status
	status := "healthy"
	if err != nil {
		status = "unhealthy"
		agent.Logger.ErrorWithFields(cfg.ServiceType, cfg.InstanceName, 
			"Test failed", err)
	} else {
		agent.Logger.InfoWithFields(cfg.ServiceType, cfg.InstanceName, 
			"Test completed successfully", duration.String(), "")
	}
	
	healthChecker.UpdateServiceHealth(cfg.InstanceName, status, duration, err)
	
	return err
}
