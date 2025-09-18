package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	magentacloud "github.com/MarcelWMeyer/cloud-performance-monitor/internal/magentacloud"
)

// RunMagentaCloudTest führt einen Upload/Download-Test für MagentaCLOUD durch
func RunMagentaCloudTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "magentacloud"
	uploadErrCode := "none"
	
	Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "test", "start", 
		"Starting MagentaCLOUD performance test")
	
	client := magentacloud.NewClient(cfg.URL, cfg.Username, cfg.Password, cfg.ANID)
	
	// Ablauf wie Nextcloud-Test mit ANID-spezifischen Pfaden
	testDir := "/performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists
	err := client.EnsureDirectory(testDir)
	if err != nil {
		Logger.LogOperation(ERROR, "magentacloud", cfg.InstanceName, "directory", "error", 
			"Could not create test directory", 
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
	Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "upload", "start", 
		"Starting file upload", 
		WithSize(fileSize))
		
	err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
	uploadDuration := time.Since(startUpload)
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())
	// Always record duration
	TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())

	if err != nil {
		Logger.LogOperation(ERROR, "magentacloud", cfg.InstanceName, "upload", "error", 
			"Upload failed", 
			WithError(err),
			WithDuration(uploadDuration),
			WithSize(fileSize))
		uploadErrCode = ExtractErrorCode(err, "upload")
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(0)
		// Continue with cleanup attempt
	} else {
		uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
		// Only record speed for successful uploads
		TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadSpeed)
		Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "upload", "success", 
			"Upload completed", 
			WithDuration(uploadDuration),
			WithSize(fileSize),
			WithSpeed(uploadSpeed))
		TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(1)
	}

	// 3. Download test (only if upload was successful)
	downloadErrCode := "none"
	if err == nil {
		startDownload := time.Now()
		Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "download", "start", 
			"Starting file download")
			
		downloadReader, downloadErr := client.DownloadFile(fullPath)
		if downloadErr != nil {
			Logger.LogOperation(ERROR, "magentacloud", cfg.InstanceName, "download", "error", 
				"Download failed", 
				WithError(downloadErr))
			downloadErrCode = ExtractErrorCode(downloadErr, "download")
			TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Inc()
			TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
		} else {
			// Read and validate file size - measure total time including HTTP overhead
			downloadedBytes, readErr := io.Copy(io.Discard, downloadReader)
			downloadReader.Close()
			downloadDuration := time.Since(startDownload)
			
			// Record histogram data
			TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
			// Always record duration
			TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())

			if readErr != nil {
				Logger.LogOperation(ERROR, "magentacloud", cfg.InstanceName, "download", "error", 
					"Download read failed", 
					WithError(readErr),
					WithDuration(downloadDuration))
				downloadErrCode = ExtractErrorCode(readErr, "download")
				TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Inc()
				TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
			} else if downloadedBytes != fileSize {
				Logger.LogOperation(ERROR, "magentacloud", cfg.InstanceName, "download", "error", 
					fmt.Sprintf("Size mismatch: expected %d, got %d", fileSize, downloadedBytes),
					WithSize(downloadedBytes))
				TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Inc()
				TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "size_mismatch").Set(0)
			} else {
				downloadSpeed := float64(downloadedBytes) / (1024 * 1024) / downloadDuration.Seconds()
				// Only record speed for successful downloads
				TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
				Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "download", "success", 
					"Download completed", 
					WithDuration(downloadDuration),
					WithSize(downloadedBytes),
					WithSpeed(downloadSpeed))
				TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(1)
			}
		}
	}

	// 4. Cleanup
	Logger.LogOperation(DEBUG, "magentacloud", cfg.InstanceName, "cleanup", "start", 
		"Deleting test file")
	cleanupErr := client.DeleteFile(fullPath)
	if cleanupErr != nil {
		Logger.LogOperation(WARN, "magentacloud", cfg.InstanceName, "cleanup", "warning", 
			"Cleanup failed", 
			WithError(cleanupErr))
	} else {
		Logger.LogOperation(DEBUG, "magentacloud", cfg.InstanceName, "cleanup", "success", 
			"Cleanup completed")
	}

	Logger.LogOperation(INFO, "magentacloud", cfg.InstanceName, "test", "complete", 
		"MagentaCLOUD test completed")
	return err
}
