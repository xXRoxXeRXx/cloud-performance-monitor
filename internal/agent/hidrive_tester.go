
package agent

import (
       "context"
       "fmt"
       "io"
       "log"
       "time"

       hidrive "github.com/MarcelWMeyer/cloud-performance-monitor/internal/hidrive"
)

// RunHiDriveTest führt einen Upload/Download-Test für HiDrive Next durch
func RunHiDriveTest(ctx context.Context, cfg *Config) error {
       serviceLabel := "hidrive"
       uploadErrCode := "none"
       log.Printf("[HiDrive] >>> RunHiDriveTest betreten für %s", cfg.URL)
       log.Printf("Starting HiDrive performance test for instance: %s", cfg.URL)
       client := hidrive.NewClient(cfg.URL, cfg.Username, cfg.Password)
       // Ablauf wie Nextcloud-Test
       testDir := "/performance_tests"
       testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
       fullPath := testDir + "/" + testFileName

       // 0. Ensure directory exists
       err := client.EnsureDirectory(testDir)
       if err != nil {
              log.Printf("[HiDrive] ERROR: Could not create test directory for %s: %v", cfg.URL, err)
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
              log.Printf("[HiDrive] ERROR: upload failed for %s: %v", cfg.URL, err)
              uploadErrCode = "upload_failed"
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(0)
              return err
       }
       
       // Calculate expected chunks for monitoring
       expectedChunks := (fileSize + chunkSize - 1) / chunkSize // Ceiling division
       ChunksUploaded.WithLabelValues(serviceLabel, cfg.InstanceName).Add(float64(expectedChunks))
       
       TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
       TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadSpeed)
       TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(1)
       log.Printf("[HiDrive] Upload finished in %v (%.2f MB/s)", uploadDuration, uploadSpeed)

       // 3. Download test with enhanced metrics
       startDownload := time.Now()
       body, err := client.DownloadFile(fullPath)
       downloadErrCode := "none"
       if err != nil {
              log.Printf("[HiDrive] ERROR: download failed for %s: %v", cfg.URL, err)
              downloadErrCode = "download_error"
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "download_failed").Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
              return err
       }
       // Measure download time including HTTP overhead
       bytesDownloaded, _ := io.Copy(io.Discard, body)
       body.Close()
       downloadDuration := time.Since(startDownload)
       
       // Calculate speed based on actually downloaded bytes including overhead
       downloadSpeed := float64(bytesDownloaded) / (1024 * 1024) / downloadDuration.Seconds()
       
       // Record histogram data for download
       TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
       
       if bytesDownloaded == fileSize {
              TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
              TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(1)
              log.Printf("[HiDrive] Download finished in %v (%.2f MB/s)", downloadDuration, downloadSpeed)
       } else {
              downloadErrCode = "incomplete_download"
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "incomplete_download").Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
              log.Printf("[HiDrive] ERROR: Download incomplete for %s: expected %d bytes, got %d", cfg.URL, fileSize, bytesDownloaded)
       }

       // 4. Cleanup
       err = client.DeleteFile(fullPath)
       if err != nil {
              log.Printf("[HiDrive] Warn: delete failed: %v", err)
       }
       return nil
}
