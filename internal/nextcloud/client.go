package nextcloud

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// Client for interacting with the Nextcloud WebDAV API
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
	logger     utils.ClientLogger
}

const (
	DefaultTimeout = 300 * time.Second
	// User-Agent string that mimics the official Nextcloud desktop client
	NextcloudUserAgent = "Mozilla/5.0 (Windows) mirall/3.15.3 (build 20250107) (Nextcloud, windows-10.0.20348 ClientArchitecture: x86_64 OsArchitecture: x86_64)"
)

// NewClient creates a new Nextcloud WebDAV client
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: DefaultTimeout}, // Generous timeout for large files
		logger:     &utils.DefaultClientLogger{},
	}
}

// newRequest is a helper to create authenticated WebDAV requests
func (c *Client) newRequest(method, urlPath string, body io.Reader) (*http.Request, error) {
	fullURL := c.BaseURL + urlPath
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	// Set User-Agent to mimic official Nextcloud desktop client
	req.Header.Set("User-Agent", NextcloudUserAgent)
	// Add additional headers that Nextcloud desktop client sends
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

// EnsureDirectory ensures the test directory exists
func (c *Client) EnsureDirectory(dirPath string) error {
	fullPath := path.Join("/remote.php/dav/files/", c.Username, dirPath)
	req, err := c.newRequest("MKCOL", fullPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 405 is returned if the directory already exists, which is not an error for us.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusMethodNotAllowed {
		return fmt.Errorf("failed to create directory %s, status: %s", dirPath, resp.Status)
	}
	return nil
}

// UploadFile uploads a file using the chunking API
func (c *Client) UploadFile(filePath string, reader io.Reader, size int64, chunkSize int64) error {
	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "start", 
		fmt.Sprintf("Starting upload for %s (size: %d bytes, chunk size: %d bytes)", filePath, size, chunkSize), 
		map[string]interface{}{"file_path": filePath, "file_size": size, "chunk_size": chunkSize})

	transferID := uuid.New().String()
	chunkDir := path.Join("/remote.php/dav/uploads/", c.Username, transferID)
	chunkDirURL := c.BaseURL + chunkDir
	destinationURL := c.BaseURL + path.Join("/remote.php/dav/files/", c.Username, filePath)

	// 1. Create temporary directory for chunks on the server
	mkcolStart := time.Now()
	
	req, err := http.NewRequest("MKCOL", chunkDirURL, nil)
	if err != nil {
		return fmt.Errorf("could not create MKCOL request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("User-Agent", NextcloudUserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	// Add Destination header like bash script does
	req.Header.Set("Destination", destinationURL)
	
	resp, err := c.HTTPClient.Do(req)
	mkcolDuration := time.Since(mkcolStart)
	
	if err != nil {
		return fmt.Errorf("MKCOL request failed after %v: %w", mkcolDuration, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("MKCOL for chunks failed with status %s after %v", resp.Status, mkcolDuration)
	}

	// 2. Upload file in chunks
	if err := c.uploadChunks(chunkDir, reader, chunkSize, destinationURL); err != nil {
		return err
	}

	// 3. Assemble chunks by moving the directory
	moveSource := c.BaseURL + chunkDir + "/.file"
	
	fmt.Printf("[Nextcloud] Starting MOVE operation from %s to %s\n", moveSource, destinationURL)
	req, err = http.NewRequest("MOVE", moveSource, nil)
	if err != nil {
		return fmt.Errorf("could not create MOVE request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("User-Agent", NextcloudUserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Destination", destinationURL)
	// CRITICAL: Add OC-Total-Length header like bash script does!
	req.Header.Set("OC-Total-Length", fmt.Sprintf("%d", size))
	
	// For large files, MOVE operation can take very long server-side
	// Increase timeout specifically for this operation
	moveClient := &http.Client{
		Timeout: 10 * time.Minute, // 10 minutes for large file assembly
		Transport: c.HTTPClient.Transport,
	}
	
	fmt.Printf("[Nextcloud] Executing MOVE operation (this may take several minutes for large files)...\n")
	moveStart := time.Now()
	resp, err = moveClient.Do(req)
	moveDuration := time.Since(moveStart)
	
	if err != nil {
		c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "move", "failed", 
			fmt.Sprintf("MOVE operation failed after %v: %v", moveDuration, err), 
			map[string]interface{}{"duration": moveDuration.String(), "error": err.Error()})
		return fmt.Errorf("MOVE request failed after %v: %w", moveDuration, err)
	}
	defer resp.Body.Close()

	fmt.Printf("[Nextcloud] MOVE operation completed in %v with status %s\n", moveDuration, resp.Status)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		// Read response body for more detailed error information
		body, _ := io.ReadAll(resp.Body)
		c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "move", "failed", 
			fmt.Sprintf("MOVE failed with status %s, response: %s", resp.Status, string(body)), 
			map[string]interface{}{"status": resp.Status, "response_body": string(body)})
		return fmt.Errorf("final MOVE to assemble chunks failed with status %s", resp.Status)
	}

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "completed", 
		fmt.Sprintf("Chunked upload successful for %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}

// uploadChunks uploads the file in chunks to the server
func (c *Client) uploadChunks(chunkDir string, reader io.Reader, chunkSize int64, destinationURL string) error {
	chunk := make([]byte, chunkSize)
	totalChunks := 0
	successfulChunks := 0
	
	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "chunk_upload", "start", 
		fmt.Sprintf("Starting chunk upload to %s (chunk size: %d bytes)", chunkDir, chunkSize), 
		map[string]interface{}{"chunk_dir": chunkDir, "chunk_size": chunkSize})
	
	// CRITICAL: Start chunk numbering at 1, not 0! (Nextcloud requires 1-based indexing like HiDrive)
	chunkNumber := 1
	for {
		bytesRead, readErr := reader.Read(chunk)
		if bytesRead > 0 {
			totalChunks++
			// Use 5-digit padded chunk names like bash script: 00001, 00002, 00003, etc.
			chunkPath := fmt.Sprintf("%s/%05d", chunkDir, chunkNumber)
			chunkURL := c.BaseURL + chunkPath

			c.logger.LogOperation(utils.DEBUG, "nextcloud", c.BaseURL, "chunk_upload", "chunk_progress", 
				fmt.Sprintf("Uploading chunk %d: %d bytes to %s", chunkNumber, bytesRead, chunkPath), 
				map[string]interface{}{"chunk_number": chunkNumber, "bytes": bytesRead, "chunk_path": chunkPath})

			chunkStart := time.Now()

			// Retry logic for individual chunks
			var resp *http.Response
			var chunkErr error
			maxRetries := 3
			for attempt := 1; attempt <= maxRetries; attempt++ {
				req, err := http.NewRequest("PUT", chunkURL, bytes.NewReader(chunk[:bytesRead]))
				if err != nil {
					c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "chunk_upload", "request_error", 
						fmt.Sprintf("Could not create PUT request for chunk %d (attempt %d): %v", chunkNumber, attempt, err), 
						map[string]interface{}{"chunk_number": chunkNumber, "attempt": attempt, "error": err.Error()})
					if attempt == maxRetries {
						return fmt.Errorf("could not create PUT request for chunk %d after %d attempts: %w", chunkNumber, maxRetries, err)
					}
					continue
				}
				req.SetBasicAuth(c.Username, c.Password)
				req.Header.Set("User-Agent", NextcloudUserAgent)
				req.Header.Set("Accept", "*/*")
				req.Header.Set("Accept-Language", "en-US,en;q=0.9")
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Content-Type", "application/octet-stream")
				// CRITICAL: Add Destination header like bash script does for each chunk!
				req.Header.Set("Destination", destinationURL)
				req.ContentLength = int64(bytesRead)

				resp, chunkErr = c.HTTPClient.Do(req)
				if chunkErr != nil {
					c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "chunk_upload", "http_error", 
						fmt.Sprintf("PUT request for chunk %d failed (attempt %d/%d) after %v: %v", chunkNumber, attempt, maxRetries, time.Since(chunkStart), chunkErr), 
						map[string]interface{}{"chunk_number": chunkNumber, "attempt": attempt, "max_retries": maxRetries, "error": chunkErr.Error()})
					if attempt < maxRetries {
						time.Sleep(time.Duration(attempt) * time.Second) // Progressive backoff
						continue
					}
					return fmt.Errorf("PUT request for chunk %d failed after %d attempts: %w", chunkNumber, maxRetries, chunkErr)
				}

				// Check response status - Accept both 201 Created, 200 OK and 204 No Content
				if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
					break // Success
				} else {
					// Read response body for detailed error information
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "chunk_upload", "status_error", 
						fmt.Sprintf("Chunk %d upload failed (attempt %d/%d) with status %s after %v, response: %s", chunkNumber, attempt, maxRetries, resp.Status, time.Since(chunkStart), string(body)), 
						map[string]interface{}{"chunk_number": chunkNumber, "attempt": attempt, "max_retries": maxRetries, "status": resp.Status, "response_body": string(body)})
					if attempt < maxRetries {
						time.Sleep(time.Duration(attempt) * time.Second) // Progressive backoff
						continue
					}
					return fmt.Errorf("upload of chunk %d failed with status %s after %d attempts", chunkNumber, resp.Status, maxRetries)
				}
			}
			
			chunkDuration := time.Since(chunkStart)
			
			// Immediately close response body to avoid resource leaks
			resp.Body.Close()
			successfulChunks++
			
			c.logger.LogOperation(utils.DEBUG, "nextcloud", c.BaseURL, "chunk_upload", "chunk_success", 
				fmt.Sprintf("Chunk %d uploaded successfully in %v (status: %s)", chunkNumber, chunkDuration, resp.Status), 
				map[string]interface{}{"chunk_number": chunkNumber, "duration": chunkDuration.String(), "status": resp.Status})
			chunkNumber++ // Increment for next chunk
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "chunk_upload", "read_error", 
				fmt.Sprintf("Failed to read chunk %d: %v", chunkNumber, readErr), 
				map[string]interface{}{"chunk_number": chunkNumber, "error": readErr.Error()})
			return fmt.Errorf("failed to read chunk %d: %w", chunkNumber, readErr)
		}
	}
	
	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "chunk_upload", "completed", 
		fmt.Sprintf("Chunk upload summary: %d/%d chunks uploaded successfully", successfulChunks, totalChunks), 
		map[string]interface{}{"successful_chunks": successfulChunks, "total_chunks": totalChunks})
	return nil
}

// DownloadFile downloads a file
func (c *Client) DownloadFile(filePath string) (io.ReadCloser, error) {
	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "download", "started", 
		fmt.Sprintf("Download started for %s", filePath), 
		map[string]interface{}{"file_path": filePath})

	fullPath := path.Join("/remote.php/dav/files/", c.Username, filePath)
	req, err := c.newRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed, status: %s", resp.Status)
	}
	return resp.Body, nil
}

// DeleteFile deletes a file or directory
func (c *Client) DeleteFile(filePath string) error {
	fullPath := path.Join("/remote.php/dav/files/", c.Username, filePath)
	req, err := c.newRequest("DELETE", fullPath, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed, status: %s", resp.Status)
	}

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "delete", "success", 
		fmt.Sprintf("File deleted successfully: %s", filePath), 
		map[string]interface{}{"file_path": filePath})
	return nil
}
