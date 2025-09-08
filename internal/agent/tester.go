package agent

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/MarcelWMeyer/nextcloud-performance-monitor/internal/nextcloud"
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
		TestSuccess.WithLabelValues("nextcloud", cfg.URL, "setup", "mkdir_error").Set(0)
		return
	}

	// 1. Generate temp file using streaming reader to avoid large memory allocation
	fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
	reader := io.LimitReader(&randomReader{}, fileSize)

	// 2. Upload test
	startUpload := time.Now()
	chunkSizeBytes := int64(cfg.TestChunkSizeMB) * 1024 * 1024
	err := ncClient.UploadFile(fullPath, reader, fileSize, chunkSizeBytes)
	uploadDuration := time.Since(startUpload)

	if err != nil {
		log.Printf("ERROR: Upload failed for %s: %v", cfg.URL, err)
		TestSuccess.WithLabelValues("nextcloud", cfg.URL, "upload", "upload_error").Set(0)
		// Try to clean up the failed chunking directory
		_ = ncClient.DeleteFile(fullPath)
		return
	}

	uploadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / uploadDuration.Seconds()
	TestDuration.WithLabelValues("nextcloud", cfg.URL, "upload").Set(uploadDuration.Seconds())
	TestSpeedMbytesPerSec.WithLabelValues("nextcloud", cfg.URL, "upload").Set(uploadSpeedMBs)
	TestSuccess.WithLabelValues("nextcloud", cfg.URL, "upload", "none").Set(1)
	log.Printf("Upload finished in %v (%.2f MB/s)", uploadDuration, uploadSpeedMBs)

	// 3. Download test
	startDownload := time.Now()
	body, err := ncClient.DownloadFile(fullPath)
	if err != nil {
	       log.Printf("ERROR: Download failed for %s: %v", cfg.URL, err)
		TestSuccess.WithLabelValues("nextcloud", cfg.URL, "download", "download_error").Set(0)
	} else {
	       // We need to read the body to get an accurate time measurement
	       bytesDownloaded, _ := io.Copy(io.Discard, body)
	       body.Close()
	       downloadDuration := time.Since(startDownload)

	       if bytesDownloaded == fileSize {
		       downloadSpeedMBs := (float64(fileSize) / (1024 * 1024)) / downloadDuration.Seconds()
			  TestDuration.WithLabelValues("nextcloud", cfg.URL, "download").Set(downloadDuration.Seconds())
			  TestSpeedMbytesPerSec.WithLabelValues("nextcloud", cfg.URL, "download").Set(downloadSpeedMBs)
			  TestSuccess.WithLabelValues("nextcloud", cfg.URL, "download", "none").Set(1)
		       log.Printf("Download finished in %v (%.2f MB/s)", downloadDuration, downloadSpeedMBs)
	       } else {
		       log.Printf("ERROR: Download incomplete for %s: expected %d bytes, got %d", cfg.URL, fileSize, bytesDownloaded)
			  TestSuccess.WithLabelValues("nextcloud", cfg.URL, "download", "incomplete_download").Set(0)
	       }
	}

	// 4. Cleanup
	err = ncClient.DeleteFile(fullPath)
	if err != nil {
		log.Printf("WARN: Failed to delete test file %s: %v", fullPath, err)
	}
}