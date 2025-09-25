package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	dropbox "github.com/xXRoxXeRXx/cloud-performance-monitor/internal/dropbox"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// clientLoggerAdapter adapts StructuredLogger to ClientLogger interface
type clientLoggerAdapter struct {
	logger *StructuredLogger
}

func (a *clientLoggerAdapter) LogOperation(level utils.LogLevel, service, instance, operation, phase, message string, fields map[string]interface{}) {
	// Convert utils.LogLevel to agent.LogLevel
	var agentLevel LogLevel
	switch level {
	case utils.DEBUG:
		agentLevel = DEBUG
	case utils.INFO:
		agentLevel = INFO
	case utils.WARN:
		agentLevel = WARN
	case utils.ERROR:
		agentLevel = ERROR
	default:
		agentLevel = INFO
	}
	
	// Convert common fields to LogOptions
	var opts []LogOption
	for key, value := range fields {
		switch key {
		case "error":
			if err, ok := value.(error); ok {
				opts = append(opts, WithError(err))
			}
		case "duration":
			if d, ok := value.(time.Duration); ok {
				opts = append(opts, WithDuration(d))
			}
		case "status":
			if s, ok := value.(string); ok {
				opts = append(opts, WithStatus(s))
			}
		case "status_code":
			if code, ok := value.(int); ok {
				opts = append(opts, WithStatusCode(code))
			}
		case "size", "file_size", "chunk_size":
			if size, ok := value.(int64); ok {
				opts = append(opts, WithSize(size))
			} else if size, ok := value.(int); ok {
				opts = append(opts, WithSize(int64(size)))
			}
		case "speed_mbps":
			if speed, ok := value.(float64); ok {
				opts = append(opts, WithSpeed(speed))
			}
		case "chunk_num":
			if chunkNum, ok := value.(int); ok {
				if totalChunks, exists := fields["total_chunks"]; exists {
					if total, ok := totalChunks.(int); ok {
						opts = append(opts, WithChunk(chunkNum, total))
					}
				}
			}
		case "transfer_id", "session_id":
			if id, ok := value.(string); ok {
				opts = append(opts, WithTransferID(id))
			}
		// Other fields are ignored for now since LogEntry has specific structure
		}
	}
	
	a.logger.LogOperation(agentLevel, service, instance, operation, phase, message, opts...)
}

// RunDropboxTest führt einen Upload/Download-Test für Dropbox durch
func RunDropboxTest(ctx context.Context, cfg *Config) error {
	serviceLabel := "dropbox"
	uploadErrCode := "none"
	
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "test", "start", 
		"Starting Dropbox performance test")
	
	// Create OAuth2 client - all Dropbox instances use OAuth2 now
	Logger.LogOperation(DEBUG, "dropbox", cfg.InstanceName, "auth", "oauth2_init", 
		"Using OAuth2 client with refresh token")
	// Create logger adapter for client
	loggerAdapter := &clientLoggerAdapter{logger: Logger}
	client := dropbox.NewClientWithOAuth2("", cfg.RefreshToken, cfg.AppKey, cfg.AppSecret, loggerAdapter)
	
	// Generate initial access token from refresh token
	if err := client.RefreshAccessToken(); err != nil {
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "auth", "error", 
			"Failed to generate initial access token", 
			WithError(err))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "connection", "oauth2_failed").Inc()
		return err
	}
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "auth", "success", 
		"OAuth2 access token generated successfully")
	
	// Ablauf wie Nextcloud-Test
	testDir := "/performance_tests"
	testFileName := fmt.Sprintf("testfile_%d.tmp", time.Now().UnixNano())
	fullPath := testDir + "/" + testFileName

	// 0. Ensure directory exists (Dropbox creates directories automatically)
	err := client.EnsureDirectory(testDir)
	if err != nil {
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "directory", "error", 
			"Could not validate test directory", 
			WithError(err))
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
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "upload", "start", 
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
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "upload", "error", 
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
	
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "upload", "success", 
		"Upload completed", 
		WithDuration(uploadDuration),
		WithSize(fileSize),
		WithSpeed(uploadSpeed))

	// 3. Download test with metrics
	downloadErrCode := "none"
	startDownload := time.Now()
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "download", "start", 
		"Starting file download")
		
	downloadReader, err := client.DownloadFile(fullPath)
	if err != nil {
		downloadErrCode = ExtractErrorCode(err, "download")
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "download", "error", 
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
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "download", "error", 
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
		Logger.LogOperation(ERROR, "dropbox", cfg.InstanceName, "download", "error", 
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
	
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "download", "success", 
		"Download completed", 
		WithDuration(downloadDuration),
		WithSize(downloadedBytes),
		WithSpeed(downloadSpeed))

	// 4. Cleanup - delete test file
	Logger.LogOperation(DEBUG, "dropbox", cfg.InstanceName, "cleanup", "start", 
		"Deleting test file")
	err = client.DeleteFile(fullPath)
	if err != nil {
		Logger.LogOperation(WARN, "dropbox", cfg.InstanceName, "cleanup", "warning", 
			"Could not delete test file", 
			WithError(err))
		TestErrors.WithLabelValues(serviceLabel, cfg.InstanceName, "cleanup", "delete_failed").Inc()
		// Don't return error for cleanup failure
	} else {
		Logger.LogOperation(DEBUG, "dropbox", cfg.InstanceName, "cleanup", "success", 
			"Test file cleanup completed")
	}

	// Record overall test success
	// Total successful tests counter (if needed add this metric to metrics.go)
	
	// Circuit breaker: Close on success
	CircuitBreakerState.WithLabelValues(serviceLabel, cfg.InstanceName).Set(0)
	
	Logger.LogOperation(INFO, "dropbox", cfg.InstanceName, "test", "complete", 
		"Dropbox test completed successfully")
	return nil
}
