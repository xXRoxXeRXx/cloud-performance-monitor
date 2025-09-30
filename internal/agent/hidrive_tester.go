
package agent

import (
       "context"
       "fmt"
       "io"
       "time"

       hidrive "github.com/xXRoxXeRXx/cloud-performance-monitor/internal/hidrive"
)

// RunHiDriveTest führt einen Upload/Download-Test für HiDrive Next durch
func RunHiDriveTest(ctx context.Context, cfg *Config) error {
       serviceLabel := "hidrive"
       uploadErrCode := "none"
       
       Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "test", "start", 
              "Starting HiDrive performance test")
       
       client := hidrive.NewClient(cfg.URL, cfg.Username, cfg.Password)
       // Ablauf wie Nextcloud-Test
       testDir := "/performance_tests"
       testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
       fullPath := testDir + "/" + testFileName

       // 0. Ensure directory exists
       err := client.EnsureDirectory(testDir)
       if err != nil {
              Logger.LogOperation(ERROR, "hidrive", cfg.InstanceName, "directory", "error", 
                     "Could not create test directory", 
                     WithError(err))
              directoryErrCode := ExtractErrorCode(err, "directory")
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", "directory_creation").Inc()
              // Set failed test metrics to trigger alerts
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", directoryErrCode).Set(0)
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", directoryErrCode).Set(0)
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
       Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "upload", "start", 
              "Starting file upload", 
              WithSize(fileSize))
              
       err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
       uploadDuration := time.Since(startUpload)
       
       // Record histogram data
       TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Observe(uploadDuration.Seconds())
       // Always record duration
       TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadDuration.Seconds())
       
       if err != nil {
              Logger.LogOperation(ERROR, "hidrive", cfg.InstanceName, "upload", "error", 
                     "Upload failed", 
                     WithError(err),
                     WithDuration(uploadDuration),
                     WithSize(fileSize))
              uploadErrCode = ExtractErrorCode(err, "upload")
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(0)
              return err
       }
       
       // Calculate expected chunks for monitoring
       expectedChunks := (fileSize + chunkSize - 1) / chunkSize // Ceiling division
       ChunksUploaded.WithLabelValues(serviceLabel, cfg.InstanceName).Add(float64(expectedChunks))
       
       uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
       // Only record speed for successful uploads
       TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "upload").Set(uploadSpeed)
       TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", uploadErrCode).Set(1)
       Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "upload", "success", 
              "Upload completed", 
              WithDuration(uploadDuration),
              WithSize(fileSize),
              WithSpeed(uploadSpeed))

       // 3. Download test with enhanced metrics
       startDownload := time.Now()
       Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "download", "start", 
              "Starting file download")
              
       body, err := client.DownloadFile(fullPath)
       downloadErrCode := "none"
       if err != nil {
              Logger.LogOperation(ERROR, "hidrive", cfg.InstanceName, "download", "error", 
                     "Download failed", 
                     WithError(err))
              downloadErrCode = ExtractErrorCode(err, "download")
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
              return err
       }
       // Measure download time including HTTP overhead
       bytesDownloaded, _ := io.Copy(io.Discard, body)
       body.Close()
       downloadDuration := time.Since(startDownload)
       
       // Record histogram data for download
       TestDurationHistogram.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Observe(downloadDuration.Seconds())
       // Always record duration
       TestDuration.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadDuration.Seconds())
       
       if bytesDownloaded == fileSize {
              // Calculate speed based on actually downloaded bytes including overhead
              downloadSpeed := float64(bytesDownloaded) / (1024 * 1024) / downloadDuration.Seconds()
              // Only record speed for successful downloads
              TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.InstanceName, "download").Set(downloadSpeed)
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(1)
              Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "download", "success", 
                     "Download completed", 
                     WithDuration(downloadDuration),
                     WithSize(bytesDownloaded),
                     WithSpeed(downloadSpeed))
       } else {
              downloadErrCode = "incomplete_download"
              TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "download", "incomplete_download").Inc()
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", downloadErrCode).Set(0)
              Logger.LogOperation(ERROR, "hidrive", cfg.InstanceName, "download", "error", 
                     fmt.Sprintf("Download incomplete: expected %d bytes, got %d", fileSize, bytesDownloaded))
       }

       // 4. Cleanup
       Logger.LogOperation(DEBUG, "hidrive", cfg.InstanceName, "cleanup", "start", 
              "Deleting test file")
       err = client.DeleteFile(fullPath)
       if err != nil {
              Logger.LogOperation(WARN, "hidrive", cfg.InstanceName, "cleanup", "warning", 
                     "Delete failed", 
                     WithError(err))
       } else {
              Logger.LogOperation(DEBUG, "hidrive", cfg.InstanceName, "cleanup", "success", 
                     "Test file cleanup completed")
       }
       
       Logger.LogOperation(INFO, "hidrive", cfg.InstanceName, "test", "complete", 
              "HiDrive test completed successfully")
       
       // Reset any previous error states to prevent false alerts
       // This ensures old failed test metrics don't trigger false alarms
       errorCodesToReset := []string{
              "http_404_not_found",
              "http_403_forbidden", 
              "http_401_unauthorized",
              "http_500_server_error",
              "http_503_unavailable",
              "network_timeout",
              "upload_failed",
              "download_failed",
       }
       
       for _, errorCode := range errorCodesToReset {
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "upload", errorCode).Set(1)
              TestSuccess.WithLabelValues(serviceLabel, cfg.InstanceName, "download", errorCode).Set(1)
       }
       
       Logger.LogOperation(DEBUG, "hidrive", cfg.InstanceName, "metrics", "reset", 
              "Reset previous error states to prevent false alarms")
       
       return nil
}
