package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	hidrive_legacy "github.com/MarcelWMeyer/cloud-performance-monitor/internal/hidrive_legacy"
)

// RunHiDriveLegacyTest führt einen Upload/Download-Test für HiDrive Legacy API durch
func RunHiDriveLegacyTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "hidrive_legacy"
	uploadErrCode := "none"
	log.Printf("[HiDrive Legacy] >>> RunHiDriveLegacyTest betreten für %s", cfg.InstanceName)
	log.Printf("Starting HiDrive Legacy performance test for instance: %s", cfg.InstanceName)
	
	// Create OAuth2 client with refresh token
	client, err := hidrive_legacy.NewClientWithOAuth2(cfg.RefreshToken, cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		log.Printf("[HiDrive Legacy] ERROR: OAuth2 client creation failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "oauth2_failed").Inc()
		return err
	}
	log.Printf("[HiDrive Legacy] Using OAuth2 client with refresh token for %s", cfg.InstanceName)
	
	// Test connection first
	err = client.TestConnection()
	if err != nil {
		log.Printf("[HiDrive Legacy] ERROR: Connection test failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "auth_failed").Inc()
		return err
	}
	
	// Ablauf wie andere Tests
	testDir := "performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists
	err = client.EnsureDirectory(testDir)
	if err != nil {
		log.Printf("[HiDrive Legacy] ERROR: Could not ensure test directory for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", "directory_creation").Inc()
		return err
	}

	// 1. Generate temp file using streaming reader
	fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
	reader := io.LimitReader(&randomReader{}, fileSize)
	chunkSize := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	
	// Record chunk size for monitoring
	ChunkSize.WithLabelValues(serviceLabel, cfg.InstanceName).Set(float64(chunkSize))
	// Initialize circuit breaker state (0 = closed)
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(0)

	// 2. Upload test with enhanced metrics
	startUpload := time.Now()
	err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
	uploadDuration := time.Since(startUpload)
	uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())
	
	if err != nil {
		uploadErrCode = "upload_failed"
		log.Printf("[HiDrive Legacy] ERROR: Upload failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Inc()
		// Record failed upload in metrics
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(0)
		TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
		TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(0)
		
		// Circuit breaker: Open on failure
		CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(1)
		return err
	}

	// Record successful upload metrics
	TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(1)
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadSpeed)
	
	log.Printf("[HiDrive Legacy] Upload completed for %s: %.2f MB/s (%.2f seconds)", 
		cfg.InstanceName, uploadSpeed, uploadDuration.Seconds())

	// 3. Download test with metrics
	startDownload := time.Now()
	downloadReader, err := client.DownloadFile(fullPath)
	if err != nil {
		log.Printf("[HiDrive Legacy] ERROR: Download failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "download_failed").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "download_failed").Set(0)
		return err
	}
	defer downloadReader.Close()

	// Read all data to measure actual download speed - measure only transfer time
	transferStart := time.Now()
	downloadedBytes, err := io.Copy(io.Discard, downloadReader)
	transferDuration := time.Since(transferStart)
	downloadDuration := time.Since(startDownload) // Total duration for metrics
	downloadSpeed := float64(downloadedBytes) / (1024 * 1024) / transferDuration.Seconds()
	
	// Record download histogram
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
	
	if err != nil {
		log.Printf("[HiDrive Legacy] ERROR: Download read failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "read_failed").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "read_failed").Set(0)
		TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
		TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(0)
		return err
	}

	// Verify file size
	if downloadedBytes != fileSize {
		err = fmt.Errorf("downloaded file size mismatch: expected %d, got %d", fileSize, downloadedBytes)
		log.Printf("[HiDrive Legacy] ERROR: %v for %s", err, cfg.InstanceName)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Set(0)
		return err
	}

	// Record successful download metrics
	TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "none").Set(1)
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
	
	log.Printf("[HiDrive Legacy] Download completed for %s: %.2f MB/s (%.2f seconds)", 
		cfg.InstanceName, downloadSpeed, downloadDuration.Seconds())

	// 4. Cleanup - delete test file
	err = client.DeleteFile(fullPath)
	if err != nil {
		log.Printf("[HiDrive Legacy] WARNING: Could not delete test file %s: %v", fullPath, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "cleanup", "delete_failed").Inc()
		// Don't return error for cleanup failure
	} else {
		log.Printf("[HiDrive Legacy] Test file cleanup completed for %s", cfg.InstanceName)
	}

	// Circuit breaker: Close on success
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(0)
	
	log.Printf("[HiDrive Legacy] <<< RunHiDriveLegacyTest completed for %s", cfg.InstanceName)
	return nil
}
