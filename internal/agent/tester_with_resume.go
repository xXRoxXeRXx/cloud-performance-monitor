package agent

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/nextcloud"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// RunTestWithResume performs a single performance test run with upload resume capability
func RunTestWithResume(cfg *Config, ncClient *nextcloud.Client, logger utils.ClientLogger) {
	log.Printf("Starting performance test with resume capability for instance: %s", cfg.URL)
	testDir := "performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName
	
	// Initialize upload manager
	uploadManager := NewUploadManager("./upload_states", cfg.ServiceType, cfg.InstanceName, logger)

	// 0. Ensure directory exists
	if err := ncClient.EnsureDirectory(testDir); err != nil {
		log.Printf("ERROR: Could not create test directory for %s: %v", cfg.URL, err)
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", "directory_creation").Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "setup", "mkdir_error").Set(0)
		return
	}

	// 1. Create temporary file with random data
	fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
	tempFilePath, err := uploadManager.CreateTempFile(fileSize)
	if err != nil {
		log.Printf("ERROR: Could not create temp file for %s: %v", cfg.URL, err)
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", "temp_file_creation").Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "setup", "temp_file_error").Set(0)
		return
	}
	defer uploadManager.CleanupTempFile(tempFilePath)

	// Record initial chunk size
	chunkSizeBytes := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	ChunkSize.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Set(float64(chunkSizeBytes))
	CircuitBreakerState.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Set(0)

	// 2. Create resume capable client wrapper  
	resumeClient := nextcloud.NewResumeClient(cfg.URL, cfg.Username, cfg.Password, logger)

	// Upload test with resume capability
	startUpload := time.Now()
	err = uploadManager.UploadWithResume(resumeClient, tempFilePath, fullPath, 
		cfg.ServiceType, cfg.InstanceName, fileSize, cfg.TestChunkSizeMB)
	uploadDuration := time.Since(startUpload)
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())

	if err != nil {
		log.Printf("ERROR: Upload with resume failed for %s: %v", cfg.URL, err)
		uploadErrCode := ExtractErrorCode(err, "upload")
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", uploadErrCode).Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", uploadErrCode).Set(0)
		// Try to clean up the failed chunking directory
		_ = ncClient.DeleteFile(fullPath)
		return
	}
	
	// Calculate expected chunks for monitoring
	initialChunkSize := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	expectedChunks := (fileSize + initialChunkSize - 1) / initialChunkSize // Ceiling division
	ChunksUploaded.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Add(float64(expectedChunks))

	uploadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / uploadDuration.Seconds()
	TestDuration.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Set(uploadSpeedMBs)
	TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", "none").Set(1)
	log.Printf("Upload with resume finished in %v (%.2f MB/s)", uploadDuration, uploadSpeedMBs)

	// 3. Download test (unchanged)
	downloadErrCode := "none"
	startDownload := time.Now()
	body, err := ncClient.DownloadFile(fullPath)
	if err != nil {
		log.Printf("ERROR: Download failed for %s: %v", cfg.URL, err)
		downloadErrCode = ExtractErrorCode(err, "download")
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download", downloadErrCode).Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download", downloadErrCode).Set(0)
	} else {
		// We need to read the body to get an accurate time measurement
		bytesDownloaded, _ := io.Copy(io.Discard, body)
		body.Close()
		downloadDuration := time.Since(startDownload)
		
		// Record histogram data for download
		TestDurationHistogram.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
		TestDuration.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download").Set(downloadDuration.Seconds())

		if bytesDownloaded == fileSize {
			downloadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / downloadDuration.Seconds()
			TestSpeedMbytesPerSec.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download").Set(downloadSpeedMBs)
			TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download", downloadErrCode).Set(1)
			log.Printf("Download finished in %v (%.2f MB/s)", downloadDuration, downloadSpeedMBs)
		} else {
			log.Printf("ERROR: Download incomplete for %s: expected %d bytes, got %d", cfg.URL, fileSize, bytesDownloaded)
			downloadErrCode = ExtractErrorCode(fmt.Errorf("download incomplete: expected %d bytes, got %d", fileSize, bytesDownloaded), "download")
			TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download", downloadErrCode).Set(0)
		}
	}

	// 4. Cleanup
	cleanupErr := ncClient.DeleteFile(fullPath)
	if cleanupErr != nil {
		log.Printf("WARN: Failed to delete test file %s: %v", fullPath, cleanupErr)
	}
	
	// Reset previous error states after successful test
	if downloadErrCode == "none" {
		for _, errorCode := range GetAllErrorCodes() {
			TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", errorCode).Set(1)
			TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download", errorCode).Set(1)
		}
		log.Printf("Reset all previous error states for Nextcloud instance %s", cfg.InstanceName)
	}
}