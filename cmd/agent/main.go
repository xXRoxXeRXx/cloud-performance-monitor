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

	// Start a monitoring goroutine for each configured instance
	for _, cfg := range allConfigs {
		go startMonitoringInstance(cfg)
	}

	// Keep the main process alive
	select {}
}

func startMonitoringInstance(cfg *agent.Config) {
	log.Printf("Agent starting for instance %s (%s). Tests will run every %d seconds.", cfg.InstanceName, cfg.ServiceType, cfg.TestIntervalSec)
	ticker := time.NewTicker(time.Duration(cfg.TestIntervalSec) * time.Second)
	defer ticker.Stop()

	switch cfg.ServiceType {
	case "nextcloud":
		ncClient := nextcloud.NewClient(cfg.URL, cfg.Username, cfg.Password)
		agent.RunTest(cfg, ncClient)
		for range ticker.C {
			agent.RunTest(cfg, ncClient)
		}
       case "hidrive":
	       defer func() {
		       if r := recover(); r != nil {
			       log.Printf("[HiDrive] PANIC in Goroutine für %s: %v", cfg.URL, r)
		       }
	       }()
	       log.Printf("[HiDrive] Goroutine gestartet für %s", cfg.URL)
	       log.Printf("[HiDrive] Vor erstem Test für %s", cfg.URL)
			   ctx := context.Background()
			   if err := agent.RunHiDriveTest(ctx, cfg); err != nil {
		       log.Printf("[HiDrive] Test error: %v", err)
	       }
	       log.Printf("[HiDrive] Nach erstem Test für %s", cfg.URL)
	       for range ticker.C {
		       log.Printf("[HiDrive] Vor periodischem Test für %s", cfg.URL)
		       if err := agent.RunHiDriveTest(ctx, cfg); err != nil {
			       log.Printf("[HiDrive] Test error: %v", err)
		       }
		       log.Printf("[HiDrive] Nach periodischem Test für %s", cfg.URL)
	       }
	default:
		log.Printf("Unknown service type: %s", cfg.ServiceType)
	}
}
