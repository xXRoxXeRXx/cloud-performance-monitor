
package agent

import (
       "context"
       "fmt"
       "io"
       "log"
       "time"

       hidrive "github.com/MarcelWMeyer/nextcloud-performance-monitor/internal/hidrive"
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
              return err
       }

       // 1. Generate temp file using streaming reader
       fileSize := int64(cfg.TestFileSizeMB) * 1024 * 1024
       reader := io.LimitReader(&randomReader{}, fileSize)
       chunkSize := int64(cfg.TestChunkSizeMB) * 1024 * 1024

       // 2. Upload test
       startUpload := time.Now()
       err = client.UploadFile(fullPath, reader, fileSize, chunkSize)
       uploadDuration := time.Since(startUpload)
       uploadSpeed := float64(fileSize) / (1024 * 1024) / uploadDuration.Seconds()
       if err != nil {
              log.Printf("[HiDrive] ERROR: upload failed for %s: %v", cfg.URL, err)
              uploadErrCode = "upload_error"
              TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "upload", uploadErrCode).Set(0)
              return err
       }
       TestDuration.WithLabelValues(serviceLabel, cfg.URL, "upload").Set(uploadDuration.Seconds())
       TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.URL, "upload").Set(uploadSpeed)
       TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "upload", uploadErrCode).Set(1)
       log.Printf("[HiDrive] Upload finished in %v (%.2f MB/s)", uploadDuration, uploadSpeed)

       // 3. Download test
       startDownload := time.Now()
       body, err := client.DownloadFile(fullPath)
       downloadErrCode := "none"
       if err != nil {
              log.Printf("[HiDrive] ERROR: download failed for %s: %v", cfg.URL, err)
              downloadErrCode = "download_error"
              TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Set(0)
              return err
       }
       bytesDownloaded, _ := io.Copy(io.Discard, body)
       body.Close()
       downloadDuration := time.Since(startDownload)
       downloadSpeed := float64(fileSize) / (1024 * 1024) / downloadDuration.Seconds()
       if bytesDownloaded == fileSize {
              TestDuration.WithLabelValues(serviceLabel, cfg.URL, "download").Set(downloadDuration.Seconds())
              TestSpeedMbytesPerSec.WithLabelValues(serviceLabel, cfg.URL, "download").Set(downloadSpeed)
              TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Set(1)
              log.Printf("[HiDrive] Download finished in %v (%.2f MB/s)", downloadDuration, downloadSpeed)
       } else {
              downloadErrCode = "incomplete_download"
              TestSuccess.WithLabelValues(serviceLabel, cfg.URL, "download", downloadErrCode).Set(0)
              log.Printf("[HiDrive] ERROR: Download incomplete for %s: expected %d bytes, got %d", cfg.URL, fileSize, bytesDownloaded)
       }

       // 4. Cleanup
       err = client.DeleteFile(fullPath)
       if err != nil {
              log.Printf("[HiDrive] Warn: delete failed: %v", err)
       }
       return nil
}
