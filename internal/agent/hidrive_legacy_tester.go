package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	hidrive_legacy "github.com/xXRoxXeRXx/cloud-performance-monitor/internal/hidrive_legacy"
)

// RunHiDriveLegacyTest führt einen Upload/Download-Test für HiDrive Legacy API durch
func RunHiDriveLegacyTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "hidrive_legacy"
	uploadErrCode := "none"
	
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "test", "start", 
		"Starting HiDrive Legacy performance test")
	
	// Create OAuth2 client with refresh token
	client, err := hidrive_legacy.NewClientWithOAuth2(cfg.RefreshToken, cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "auth", "error", 
			"OAuth2 client creation failed", 
			WithError(err))
		authErrCode := ExtractErrorCode(err, "oauth2")
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "oauth2_failed").Inc()
		// Set failed test metrics to trigger alerts
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", authErrCode).Set(0)
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", authErrCode).Set(0)
		return err
	}
	Logger.LogOperation(DEBUG, "hidrive_legacy", cfg.InstanceName, "auth", "oauth2_init", 
		"Using OAuth2 client with refresh token")
	
	// Test connection first
	err = client.TestConnection()
	if err != nil {
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "auth", "error", 
			"Connection test failed", 
			WithError(err))
		connErrCode := ExtractErrorCode(err, "connection")
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "auth_failed").Inc()
		// Set failed test metrics to trigger alerts
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", connErrCode).Set(0)
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", connErrCode).Set(0)
		return err
	}
	
	// Ablauf wie andere Tests
	testDir := "performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists
	err = client.EnsureDirectory(testDir)
	if err != nil {
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "directory", "error", 
			"Could not ensure test directory", 
			WithError(err))
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
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "upload", "start", 
		"Starting file upload", 
		WithSize(fileSize))
		
	err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
	uploadDuration := time.Since(startUpload)
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())
	// Always record duration
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
	
	if err != nil {
		uploadErrCode = ExtractErrorCode(err, "upload")
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "upload", "error", 
			"Upload failed", 
			WithError(err),
			WithDuration(uploadDuration),
			WithSize(fileSize))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Inc()
		// Record failed upload in metrics
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(0)
		// Don't record speed for failed uploads
		
		// Circuit breaker: Open on failure
		CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(1)
		return err
	}

	// Record successful upload metrics
	uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
	TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(1)
	// Only record speed for successful uploads
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadSpeed)
	
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "upload", "success", 
		"Upload completed", 
		WithDuration(uploadDuration),
		WithSize(fileSize),
		WithSpeed(uploadSpeed))

	// 3. Download test with metrics
	downloadErrCode := "none"
	startDownload := time.Now()
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "download", "start", 
		"Starting file download")
		
	downloadReader, err := client.DownloadFile(fullPath)
	if err != nil {
		downloadErrCode = ExtractErrorCode(err, "download")
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "download", "error", 
			"Download failed", 
			WithError(err))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
		return err
	}
	defer downloadReader.Close()

	// Read all data to measure download speed including HTTP overhead
	downloadedBytes, err := io.Copy(io.Discard, downloadReader)
	downloadDuration := time.Since(startDownload)
	
	// Record download histogram
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
	// Always record duration
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
	
	if err != nil {
		downloadErrCode := ExtractErrorCode(err, "download")
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "download", "error", 
			"Download read failed", 
			WithError(err),
			WithDuration(downloadDuration))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
		// Don't record speed for failed downloads
		return err
	}

	// Verify file size
	if downloadedBytes != fileSize {
		err = fmt.Errorf("downloaded file size mismatch: expected %d, got %d", fileSize, downloadedBytes)
		Logger.LogOperation(ERROR, "hidrive_legacy", cfg.InstanceName, "download", "error", 
			"File size mismatch", 
			WithError(err),
			WithSize(downloadedBytes))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Set(0)
		return err
	}

	// Record successful download metrics
	downloadSpeed := float64(downloadedBytes) / (1024 * 1024) / downloadDuration.Seconds()
	TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(1)
	// Only record speed for successful downloads
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
	
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "download", "success", 
		"Download completed", 
		WithDuration(downloadDuration),
		WithSize(downloadedBytes),
		WithSpeed(downloadSpeed))

	// 4. Cleanup - delete test file
	Logger.LogOperation(DEBUG, "hidrive_legacy", cfg.InstanceName, "cleanup", "start", 
		"Deleting test file")
	err = client.DeleteFile(fullPath)
	if err != nil {
		Logger.LogOperation(WARN, "hidrive_legacy", cfg.InstanceName, "cleanup", "warning", 
			"Could not delete test file", 
			WithError(err))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "cleanup", "delete_failed").Inc()
		// Don't return error for cleanup failure
	} else {
		Logger.LogOperation(DEBUG, "hidrive_legacy", cfg.InstanceName, "cleanup", "success", 
			"Test file cleanup completed")
	}

	// Circuit breaker: Close on success
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(0)
	
	Logger.LogOperation(INFO, "hidrive_legacy", cfg.InstanceName, "test", "complete", 
		"HiDrive Legacy test completed successfully")
	return nil
}
