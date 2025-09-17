package agent

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/MarcelWMeyer/cloud-performance-monitor/internal/nextcloud"
)

// randomReader generates random data on-the-fly to avoid large memory allocations
type randomReader struct{}

func (r *randomReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// RunTest performs a single performance test run.
func RunTest(cfg *Config, ncClient *nextcloud.Client) {
	log.Printf("Starting performance test for instance: %s", cfg.URL)
	testDir := "performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists
	if err := ncClient.EnsureDirectory(testDir); err != nil {
		log.Printf("ERROR: Could not create test directory for %s: %v", cfg.URL, err)
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", "directory_creation").Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "setup", "mkdir_error").Set(0)
		return
	}

	// 1. Generate temp file using streaming reader to avoid large memory allocation
	fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
	reader := io.LimitReader(&randomReader{}, fileSize)
	chunkSizeBytes := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	
	// Record chunk size and initialize circuit breaker state for monitoring
	ChunkSize.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Set(float64(chunkSizeBytes))
	CircuitBreakerState.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Set(0)

	// 2. Upload test with enhanced metrics
	startUpload := time.Now()
	err := ncClient.UploadFile(fullPath, reader, fileSize, chunkSizeBytes)
	uploadDuration := time.Since(startUpload)
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())

	if err != nil {
		log.Printf("ERROR: Upload failed for %s: %v", cfg.URL, err)
		uploadErrCode := ExtractErrorCode(err, "upload")
		TestErrors.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", uploadErrCode).Inc()
		TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", uploadErrCode).Set(0)
		// Try to clean up the failed chunking directory
		_ = ncClient.DeleteFile(fullPath)
		return
	}
	
	// Calculate expected chunks for monitoring
	expectedChunks := (fileSize + chunkSizeBytes - 1) / chunkSizeBytes // Ceiling division
	ChunksUploaded.WithLabelValues(cfg.ServiceType, cfg.InstanceName).Add(float64(expectedChunks))

	uploadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / uploadDuration.Seconds()
	TestDuration.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload").Set(uploadSpeedMBs)
	TestSuccess.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "upload", "none").Set(1)
	log.Printf("Upload finished in %v (%.2f MB/s)", uploadDuration, uploadSpeedMBs)

	// 3. Download test with enhanced metrics
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

		if bytesDownloaded == fileSize {
			downloadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / downloadDuration.Seconds()
			TestDuration.WithLabelValues(cfg.ServiceType, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
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
	err = ncClient.DeleteFile(fullPath)
	if err != nil {
		log.Printf("WARN: Failed to delete test file %s: %v", fullPath, err)
	}
}