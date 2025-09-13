package agent

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the configuration for a single storage instance (Nextcloud or HiDrive)
type Config struct {
	InstanceName      string
	ServiceType       string // "nextcloud" oder "hidrive"
	URL               string
	Username          string
	Password          string
	TestFileSizeMB    int
	TestIntervalSec   int
	TestChunkSizeMB   int
}

const (
	DefaultFileSizeMB  = 10
	DefaultIntervalSec = 300
	DefaultChunkSizeMB = 5  // Kleinere Chunks für HiDrive (5MB statt 10MB)
)

// LoadConfigs loads configurations for all specified Nextcloud and HiDrive Next instances
func LoadConfigs() ([]*Config, error) {
       var configs []*Config

       // Nextcloud Instanzen
       i := 1
       for {
	       urlKey := fmt.Sprintf("NC_INSTANCE_%d_URL", i)
	       userKey := fmt.Sprintf("NC_INSTANCE_%d_USER", i)
	       passKey := fmt.Sprintf("NC_INSTANCE_%d_PASS", i)
	       url := os.Getenv(urlKey)
	       if i == 1 && url == "" {
		       // Noch nicht zwingend Fehler, HiDrive könnte konfiguriert sein
		       break
	       }
	       if url == "" {
		       break
	       }
	       user := os.Getenv(userKey)
	       pass := os.Getenv(passKey)
	       if user == "" || pass == "" {
		       return nil, fmt.Errorf("error: %s and %s must be set for instance %d", userKey, passKey, i)
	       }
	       fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
	       if fileSize == 0 {
		       fileSize = DefaultFileSizeMB
	       }
	       if fileSize <= 0 {
		       return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
	       }
	       interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
	       if interval == 0 {
		       interval = DefaultIntervalSec
	       }
	       if interval <= 0 {
		       return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
	       }
	       chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
	       if chunkSize == 0 {
		       chunkSize = DefaultChunkSizeMB
	       }
	       if chunkSize <= 0 {
		       return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
	       }
	       config := &Config{
		       InstanceName: url,
		       ServiceType:  "nextcloud",
		       URL:          url,
		       Username:     user,
		       Password:     pass,
		       TestFileSizeMB:  fileSize,
		       TestIntervalSec: interval,
		       TestChunkSizeMB: chunkSize,
	       }
	       configs = append(configs, config)
	       i++
       }

       // HiDrive Instanzen
       j := 1
       for {
	       urlKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_URL", j)
	       userKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_USER", j)
	       passKey := fmt.Sprintf("HIDRIVE_INSTANCE_%d_PASS", j)
	       url := os.Getenv(urlKey)
	       if j == 1 && url == "" && len(configs) == 0 {
		       return nil, fmt.Errorf("error: at least NC_INSTANCE_1_URL or HIDRIVE_INSTANCE_1_URL must be set")
	       }
	       if url == "" {
		       break
	       }
	       user := os.Getenv(userKey)
	       pass := os.Getenv(passKey)
	       if user == "" || pass == "" {
		       return nil, fmt.Errorf("error: %s and %s must be set for HiDrive instance %d", userKey, passKey, j)
	       }
	       fileSize, _ := strconv.Atoi(os.Getenv("TEST_FILE_SIZE_MB"))
	       if fileSize == 0 {
		       fileSize = DefaultFileSizeMB
	       }
	       if fileSize <= 0 {
		       return nil, fmt.Errorf("error: TEST_FILE_SIZE_MB must be positive, got %d", fileSize)
	       }
	       interval, _ := strconv.Atoi(os.Getenv("TEST_INTERVAL_SECONDS"))
	       if interval == 0 {
		       interval = DefaultIntervalSec
	       }
	       if interval <= 0 {
		       return nil, fmt.Errorf("error: TEST_INTERVAL_SECONDS must be positive, got %d", interval)
	       }
	       chunkSize, _ := strconv.Atoi(os.Getenv("TEST_CHUNK_SIZE_MB"))
	       if chunkSize == 0 {
		       chunkSize = DefaultChunkSizeMB
	       }
	       if chunkSize <= 0 {
		       return nil, fmt.Errorf("error: TEST_CHUNK_SIZE_MB must be positive, got %d", chunkSize)
	       }
	       config := &Config{
		       InstanceName: url,
		       ServiceType:  "hidrive",
		       URL:          url,
		       Username:     user,
		       Password:     pass,
		       TestFileSizeMB:  fileSize,
		       TestIntervalSec: interval,
		       TestChunkSizeMB: chunkSize,
	       }
	       configs = append(configs, config)
	       j++
       }

       if len(configs) == 0 {
	       return nil, fmt.Errorf("error: no instances configured. Please set NC_INSTANCE_1_... or HIDRIVE_INSTANCE_1_... variables")
       }

       // Validate all configurations
       for i, cfg := range configs {
	       if err := ValidateConfig(cfg); err != nil {
		       return nil, fmt.Errorf("configuration validation failed for instance %d: %w", i+1, err)
	       }
	       fmt.Printf("[Config] Validated instance: %s, ServiceType: %s, URL: %s\n", cfg.InstanceName, cfg.ServiceType, cfg.URL)
       }

       // Debug: Log all loaded configs
       for _, c := range configs {
	       fmt.Printf("[Config] Loaded instance: %s, ServiceType: %s, URL: %s\n", c.InstanceName, c.ServiceType, c.URL)
       }
       return configs, nil
}