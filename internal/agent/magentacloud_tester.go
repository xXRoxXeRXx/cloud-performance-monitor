package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	magentacloud "github.com/MarcelWMeyer/cloud-performance-monitor/internal/magentacloud"
)

// RunMagentaCloudTest führt einen Upload/Download-Test für MagentaCLOUD durch
func RunMagentaCloudTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "magentacloud"
	uploadErrCode := "none"
	log.Printf("[MagentaCLOUD] >>> RunMagentaCloudTest betreten für %s", cfg.URL)
	log.Printf("Starting MagentaCLOUD performance test for instance: %s", cfg.URL)
	client := magentacloud.NewClient(cfg.URL, cfg.Username, cfg.Password, cfg.ANID)
	
	// Ablauf wie Nextcloud-Test mit ANID-spezifischen Pfaden
	testDir := "/performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists
	err := client.EnsureDirectory(testDir)
	if err != nil {
		log.Printf("[MagentaCLOUD] ERROR: Could not create test directory for %s: %v", cfg.URL, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.URL, "upload", "directory_creation").Inc()
		return err
	}

	// 1. Generate temp file using streaming reader
	fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
	reader := io.LimitReader(&randomReader{}, fileSize)
	chunkSize := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	
	// Record chunk size for monitoring
	ChunkSize.WithLabelValues(serviceLabel, cfg.URL).Set(float64(chunkSize))
	// Initialize circuit breaker state (0 = closed)
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.URL).Set(0)

	// 2. Upload test with enhanced metrics
	startUpload := time.Now()
	err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
	uploadDuration := time.Since(startUpload)
	uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
	
	// Record histogram data
	TestDurationHistogram.WithLabelValues(serviceLabel, cfg.URL, "upload").Observe(uploadDuration.Seconds())
	// Record standard metrics
	TestDuration.WithLabelValues(serviceLabel, cfg.URL, "upload").Set(uploadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.URL, "upload").Set(uploadSpeed)

	if err != nil {
		log.Printf("[MagentaCLOUD] ERROR: Upload failed for %s: %v", cfg.URL, err)
		TestErrors.WithLabelValues(serviceLabel, cfg.URL, "upload", uploadErrCode).Inc()
		TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "upload", uploadErrCode).Set(0)
		// Continue with cleanup attempt
	} else {
		log.Printf("[MagentaCLOUD] Upload completed for %s in %v (speed: %.2f MB/s)", cfg.URL, uploadDuration, uploadSpeed)
		TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "upload", "none").Set(1)
	}

	// 3. Download test (only if upload was successful)
	downloadErrCode := "none"
	if err == nil {
		startDownload := time.Now()
		downloadReader, downloadErr := client.DownloadFile(fullPath)
		if downloadErr != nil {
			log.Printf("[MagentaCLOUD] ERROR: Download failed for %s: %v", cfg.URL, downloadErr)
			TestErrors.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Inc()
			TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Set(0)
		} else {
			// Read and validate file size
			downloadedBytes, readErr := io.Copy(io.Discard, downloadReader)
			downloadReader.Close()
			downloadDuration := time.Since(startDownload)
			downloadSpeed := float64(downloadedBytes) / (1024 * 1024) / downloadDuration.Seconds()
			
			// Record histogram data
			TestDurationHistogram.WithLabelValues(serviceLabel, cfg.URL, "download").Observe(downloadDuration.Seconds())
			// Record standard metrics
			TestDuration.WithLabelValues(serviceLabel, cfg.URL, "download").Set(downloadDuration.Seconds())
			TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.URL, "download").Set(downloadSpeed)

			if readErr != nil {
				log.Printf("[MagentaCLOUD] ERROR: Download read failed for %s: %v", cfg.URL, readErr)
				TestErrors.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Inc()
				TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", "read_error").Set(0)
			} else if downloadedBytes != fileSize {
				log.Printf("[MagentaCLOUD] ERROR: Size mismatch for %s: expected %d, got %d", cfg.URL, fileSize, downloadedBytes)
				TestErrors.WithLabelValues(serviceLabel, cfg.URL, "download", "size_mismatch").Inc()
				TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", "size_mismatch").Set(0)
			} else {
				log.Printf("[MagentaCLOUD] Download completed for %s in %v (speed: %.2f MB/s)", cfg.URL, downloadDuration, downloadSpeed)
				TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", "none").Set(1)
			}
		}
	}

	// 4. Cleanup
	cleanupErr := client.DeleteFile(fullPath)
	if cleanupErr != nil {
		log.Printf("[MagentaCLOUD] WARNING: Cleanup failed for %s: %v", cfg.URL, cleanupErr)
	} else {
		log.Printf("[MagentaCLOUD] Cleanup completed for %s", cfg.URL)
	}

	return err
}
