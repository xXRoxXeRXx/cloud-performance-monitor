package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	dropbox "github.com/MarcelWMeyer/cloud-performance-monitor/internal/dropbox"
)

// RunDropboxTest führt einen Upload/Download-Test für Dropbox durch
func RunDropboxTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "dropbox"
	uploadErrCode := "none"
	log.Printf("[Dropbox] >>> RunDropboxTest betreten für %s", cfg.InstanceName)
	log.Printf("Starting Dropbox performance test for instance: %s", cfg.InstanceName)
	
	// Create OAuth2 client - all Dropbox instances use OAuth2 now
	log.Printf("[Dropbox] Using OAuth2 client with refresh token for %s", cfg.InstanceName)
	client := dropbox.NewClientWithOAuth2("", cfg.RefreshToken, cfg.AppKey, cfg.AppSecret)
	
	// Generate initial access token from refresh token
	if err := client.RefreshAccessToken(); err != nil {
		log.Printf("[Dropbox] ERROR: Failed to generate initial access token for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "oauth2_failed").Inc()
		return err
	}
	log.Printf("[Dropbox] OAuth2 access token generated successfully for %s", cfg.InstanceName)
	
	// Ablauf wie Nextcloud-Test
	testDir := "/performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists (Dropbox creates directories automatically)
	err := client.EnsureDirectory(testDir)
	if err != nil {
		log.Printf("[Dropbox] ERROR: Could not validate test directory for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", "directory_validation").Inc()
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
		log.Printf("[Dropbox] ERROR: Upload failed for %s: %v", cfg.InstanceName, err)
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
	
	log.Printf("[Dropbox] Upload completed for %s: %.2f MB/s (%.2f seconds)", 
		cfg.InstanceName, uploadSpeed, uploadDuration.Seconds())

	// 3. Download test with metrics
	startDownload := time.Now()
	downloadReader, err := client.DownloadFile(fullPath)
	if err != nil {
		log.Printf("[Dropbox] ERROR: Download failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "download_failed").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "download_failed").Set(0)
		return err
	}
	defer downloadReader.Close()

	// Read all data to measure actual download speed
	downloadedBytes, err := io.Copy(io.Discard, downloadReader)
	downloadDuration := time.Since(startDownload)
	downloadSpeed := float64(downloadedBytes) / (1024 * 1024) / downloadDuration.Seconds()
	
	// Record download histogram
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
	
	if err != nil {
		log.Printf("[Dropbox] ERROR: Download read failed for %s: %v", cfg.InstanceName, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "read_failed").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "read_failed").Set(0)
		TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
		TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(0)
		return err
	}

	// Verify file size
	if downloadedBytes != fileSize {
		err = fmt.Errorf("downloaded file size mismatch: expected %d, got %d", fileSize, downloadedBytes)
		log.Printf("[Dropbox] ERROR: %v for %s", err, cfg.InstanceName)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Set(0)
		return err
	}

	// Record successful download metrics
	TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "none").Set(1)
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
	
	log.Printf("[Dropbox] Download completed for %s: %.2f MB/s (%.2f seconds)", 
		cfg.InstanceName, downloadSpeed, downloadDuration.Seconds())

	// 4. Cleanup - delete test file
	err = client.DeleteFile(fullPath)
	if err != nil {
		log.Printf("[Dropbox] WARNING: Could not delete test file %s: %v", fullPath, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "cleanup", "delete_failed").Inc()
		// Don't return error for cleanup failure
	} else {
		log.Printf("[Dropbox] Test file cleanup completed for %s", cfg.InstanceName)
	}

	// Record overall test success
	// Total successful tests counter (if needed add this metric to metrics.go)
	
	// Circuit breaker: Close on success
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(0)
	
	log.Printf("[Dropbox] <<< RunDropboxTest completed for %s", cfg.InstanceName)
	return nil
}
