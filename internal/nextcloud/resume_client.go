package nextcloud

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/upload"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// ResumeClient wraps the existing Nextcloud client to implement upload.ResumeCapableClient
type ResumeClient struct {
	*Client
}

// NewResumeClient creates a new resume-capable Nextcloud client
func NewResumeClient(baseURL, username, password string, logger utils.ClientLogger) *ResumeClient {
	client := &Client{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 10 * time.Minute}, // Default timeout
		logger:     logger,
	}
	
	return &ResumeClient{Client: client}
}

// ResumeChunkedUpload implements upload.ResumeCapableClient interface
func (rc *ResumeClient) ResumeChunkedUpload(transferID string, fileSize int64) (upload.ResumeInfo, error) {
	resumeInfo, err := rc.Client.ResumeChunkedUpload(transferID, fileSize)
	if err != nil {
		return upload.ResumeInfo{}, err
	}

	// Convert internal ResumeInfo to upload.ResumeInfo
	chunks := make(map[int]upload.ChunkInfo)
	for k, v := range resumeInfo.Chunks {
		chunks[k] = upload.ChunkInfo{
			Number: v.Number,
			Size:   v.Size,
			Name:   v.Name,
		}
	}

	return upload.ResumeInfo{
		TransferID:   resumeInfo.TransferID,
		UploadedSize: resumeInfo.UploadedSize,
		NextChunk:    resumeInfo.NextChunk,
		Chunks:       chunks,
		StaleChunks:  resumeInfo.StaleChunks,
	}, nil
}

// DeleteStaleChunks implements upload.ResumeCapableClient interface
func (rc *ResumeClient) DeleteStaleChunks(transferID string, staleChunks []int) {
	rc.Client.DeleteStaleChunks(transferID, staleChunks)
}

// CleanupUploadFolder implements upload.ResumeCapableClient interface
func (rc *ResumeClient) CleanupUploadFolder(transferID string) error {
	return rc.Client.CleanupUploadFolder(transferID)
}

// CreateUploadFolder implements upload.ResumeCapableClient interface
func (rc *ResumeClient) CreateUploadFolder(transferID string, fileSize int64, remotePath string) error {
	chunkDir := fmt.Sprintf("/remote.php/dav/uploads/%s/%s", rc.Username, transferID)
	chunkDirURL := rc.BaseURL + chunkDir
	destinationURL := rc.BaseURL + fmt.Sprintf("/remote.php/dav/files/%s%s", rc.Username, remotePath)

	req, err := http.NewRequest("MKCOL", chunkDirURL, nil)
	if err != nil {
		return fmt.Errorf("could not create MKCOL request: %w", err)
	}
	
	req.SetBasicAuth(rc.Username, rc.Password)
	req.Header.Set("User-Agent", "NextcloudMonitor/1.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Destination", destinationURL)
	req.Header.Set("OC-Total-Length", fmt.Sprintf("%d", fileSize))
	
	resp, err := rc.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("MKCOL request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MKCOL for chunks failed with status %s, response: %s", resp.Status, string(body))
	}

	rc.logger.LogOperation(utils.INFO, "nextcloud", rc.BaseURL, "mkcol", "success",
		"Upload folder created", map[string]interface{}{
			"transfer_id": transferID,
			"chunk_dir":   chunkDir,
			"file_size":   fileSize,
		})

	return nil
}

// UploadSingleChunk implements upload.ResumeCapableClient interface  
func (rc *ResumeClient) UploadSingleChunk(filePath, transferID string, chunkNumber int, offset int64, 
	chunkSize int, fileSize int64, remotePath string) error {
	
	chunkName := fmt.Sprintf("%05d", chunkNumber)
	chunkURL := fmt.Sprintf("%s/remote.php/dav/uploads/%s/%s/%s", rc.BaseURL, rc.Username, transferID, chunkName)

	// Open file and seek to offset
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", filePath, err)
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("could not seek to offset %d in file %s: %w", offset, filePath, err)
	}

	// Read the chunk data
	chunkData := make([]byte, chunkSize)
	n, err := io.ReadFull(file, chunkData)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("could not read chunk %d from file %s: %w", chunkNumber, filePath, err)
	}
	if n != chunkSize {
		chunkData = chunkData[:n] // Adjust for last chunk
	}

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("PUT", chunkURL, bytes.NewReader(chunkData))
		if err != nil {
			return fmt.Errorf("could not create PUT request for chunk %d: %w", chunkNumber, err)
		}

		req.SetBasicAuth(rc.Username, rc.Password)
		req.Header.Set("User-Agent", "NextcloudMonitor/1.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = int64(n)

		if attempt > 0 && lastErr != nil {
			// Add If-Match header for retries, especially for 409 conflicts
			req.Header.Set("If-Match", "*")
		}

		resp, err := rc.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			backoffDelay := time.Duration(1<<uint(attempt)) * time.Second
			rc.logger.LogOperation(utils.WARN, "nextcloud", rc.BaseURL, "upload_chunk", "retry",
				fmt.Sprintf("Chunk upload failed, retrying in %v", backoffDelay), map[string]interface{}{
					"transfer_id":    transferID,
					"chunk_number":   chunkNumber,
					"chunk_name":     chunkName,
					"attempt":        attempt + 1,
					"max_retries":    maxRetries,
					"error":          err.Error(),
					"backoff_delay":  backoffDelay.String(),
				})
			time.Sleep(backoffDelay)
			continue
		}
		defer resp.Body.Close()

		// Success status codes for chunk upload
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			rc.logger.LogOperation(utils.INFO, "nextcloud", rc.BaseURL, "upload_chunk", "success",
				"Chunk uploaded successfully", map[string]interface{}{
					"transfer_id":   transferID,
					"chunk_number":  chunkNumber,
					"chunk_name":    chunkName,
					"chunk_size":    n,
					"offset":        offset,
					"status_code":   resp.StatusCode,
					"attempt":       attempt + 1,
				})
			return nil
		}

		// Handle error status codes
		body, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("chunk upload failed with status %s, response: %s", resp.Status, string(body))

		backoffDelay := time.Duration(1<<uint(attempt)) * time.Second
		rc.logger.LogOperation(utils.WARN, "nextcloud", rc.BaseURL, "upload_chunk", "retry",
			fmt.Sprintf("Chunk upload failed with status %s, retrying in %v", resp.Status, backoffDelay), map[string]interface{}{
				"transfer_id":    transferID,
				"chunk_number":   chunkNumber,
				"status_code":    resp.StatusCode,
				"attempt":        attempt + 1,
				"max_retries":    maxRetries,
				"response_body":  string(body),
				"backoff_delay":  backoffDelay.String(),
			})
		time.Sleep(backoffDelay)
	}

	return fmt.Errorf("chunk upload failed after %d attempts: %w", maxRetries, lastErr)
}

// MoveChunksToFinalFile implements upload.ResumeCapableClient interface
func (rc *ResumeClient) MoveChunksToFinalFile(transferID, remotePath string, fileSize int64) error {
	// Fix double slashes by ensuring BaseURL doesn't end with slash
	baseURL := rc.BaseURL
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	
	// According to Nextcloud Chunking v2, the source should be the upload folder itself
	sourceURL := fmt.Sprintf("%s/remote.php/dav/uploads/%s/%s", baseURL, rc.Username, transferID)
	
	// Ensure remotePath starts with / for proper WebDAV path construction
	if !strings.HasPrefix(remotePath, "/") {
		remotePath = "/" + remotePath
	}
	destinationURL := fmt.Sprintf("%s/remote.php/dav/files/%s%s", baseURL, rc.Username, remotePath)

	// First, ensure the destination directory exists
	destDir := filepath.Dir(remotePath)
	if destDir != "." && destDir != "/" && destDir != "" {
		// Ensure destDir starts with / for proper WebDAV path construction
		if !strings.HasPrefix(destDir, "/") {
			destDir = "/" + destDir
		}
		dirURL := fmt.Sprintf("%s/remote.php/dav/files/%s%s", baseURL, rc.Username, destDir)
		if err := rc.ensureDirectoryExists(dirURL); err != nil {
			rc.logger.LogOperation(utils.WARN, "nextcloud", rc.BaseURL, "mkdir", "warn",
				"Could not ensure destination directory exists", map[string]interface{}{
					"dir_url": dirURL,
					"dest_dir": destDir,
					"remote_path": remotePath,
					"error":   err.Error(),
				})
		}
	}

	req, err := http.NewRequest("MOVE", sourceURL, nil)
	if err != nil {
		return fmt.Errorf("could not create MOVE request: %w", err)
	}

	req.SetBasicAuth(rc.Username, rc.Password)
	req.Header.Set("User-Agent", "NextcloudMonitor/1.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Destination", destinationURL)
	req.Header.Set("OC-Total-Length", fmt.Sprintf("%d", fileSize))

	// Set extended timeout for MOVE operations (large file assembly can take time)
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	rc.logger.LogOperation(utils.INFO, "nextcloud", rc.BaseURL, "move_chunks", "start",
		"Starting MOVE operation", map[string]interface{}{
			"source_url":      sourceURL,
			"destination_url": destinationURL,
			"transfer_id":     transferID,
			"file_size":       fileSize,
		})

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("MOVE request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MOVE operation failed with status %s, response: %s", resp.Status, string(body))
	}

	rc.logger.LogOperation(utils.INFO, "nextcloud", rc.BaseURL, "move_chunks", "success",
		"File assembled successfully", map[string]interface{}{
			"transfer_id":   transferID,
			"remote_path":   remotePath,
			"file_size":     fileSize,
			"source_url":    sourceURL,
			"destination":   destinationURL,
			"status_code":   resp.StatusCode,
		})

	return nil
}

// GenerateTransferID implements upload.ResumeCapableClient interface
func (rc *ResumeClient) GenerateTransferID(filePath string, fileSize int64, modTime time.Time) string {
	// Generate transfer ID like Nextcloud Desktop Client
	// Formula: rand() ^ modtime ^ (size << 16) ^ hash(filename)
	h := md5.New()
	h.Write([]byte(filePath))
	
	// Mix with modtime and size like Nextcloud Client
	seed := uint64(modTime.Unix()) ^ uint64(fileSize<<16) ^ uint64(h.Sum(nil)[0])
	
	return fmt.Sprintf("%d", seed)
}

// ensureDirectoryExists creates a directory if it doesn't exist
func (rc *ResumeClient) ensureDirectoryExists(dirURL string) error {
	req, err := http.NewRequest("MKCOL", dirURL, nil)
	if err != nil {
		return fmt.Errorf("could not create MKCOL request: %w", err)
	}

	req.SetBasicAuth(rc.Username, rc.Password)
	req.Header.Set("User-Agent", "NextcloudMonitor/1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("MKCOL request failed: %w", err)
	}
	defer resp.Body.Close()

	// MKCOL returns 201 Created if directory was created, 405 Method Not Allowed if it already exists
	// 403 Forbidden might also mean directory already exists in some Nextcloud versions
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MKCOL failed with status %s, response: %s", resp.Status, string(body))
	}

	// Log successful directory creation
	if resp.StatusCode == http.StatusCreated {
		rc.logger.LogOperation(utils.INFO, "nextcloud", rc.BaseURL, "mkdir", "success",
			"Directory created successfully", map[string]interface{}{
				"dir_url": dirURL,
			})
	}

	return nil
}