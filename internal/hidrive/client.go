package hidrive

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"
)

// Client for interacting with the HiDrive Next WebDAV API
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

const (
	DefaultTimeout = 300 * time.Second
)

// NewClient creates a new HiDrive WebDAV client
func NewClient(baseURL, username, password string) *Client {
       t := http.DefaultTransport.(*http.Transport).Clone()
       t.MaxIdleConns = 100
       t.MaxConnsPerHost = 100
       t.MaxIdleConnsPerHost = 100
       return &Client{
	       BaseURL:    baseURL,
	       Username:   username,
	       Password:   password,
	       HTTPClient: &http.Client{
		       Timeout:   DefaultTimeout,
		       Transport: t,
	       },
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
	transferID := uuid.New().String()
	chunkDir := path.Join("/remote.php/dav/uploads/", c.Username, transferID)
	chunkDirURL := c.BaseURL + chunkDir
	destinationURL := c.BaseURL + path.Join("/remote.php/dav/files/", c.Username, filePath)

	log.Printf("[HiDrive] Starting chunked upload for %s (size: %d bytes, chunk size: %d bytes, transfer ID: %s)", filePath, size, chunkSize, transferID)

	// 1. Create temporary directory for chunks on the server
	log.Printf("[HiDrive] Creating chunk directory: %s", chunkDir)
	mkcolStart := time.Now()
	
	req, err := http.NewRequest("MKCOL", chunkDirURL, nil)
	if err != nil {
		log.Printf("[HiDrive] ERROR: Could not create MKCOL request: %v", err)
		return fmt.Errorf("could not create MKCOL request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	// Add Destination header like bash script does
	req.Header.Set("Destination", destinationURL)
	
	resp, err := c.HTTPClient.Do(req)
	mkcolDuration := time.Since(mkcolStart)
	
	if err != nil {
		log.Printf("[HiDrive] ERROR: MKCOL request failed after %v: %v", mkcolDuration, err)
		return fmt.Errorf("MKCOL request failed after %v: %w", mkcolDuration, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[HiDrive] ERROR: MKCOL failed with status %s after %v, response: %s", resp.Status, mkcolDuration, string(body))
		return fmt.Errorf("MKCOL for chunks failed with status %s after %v", resp.Status, mkcolDuration)
	}
	
	log.Printf("[HiDrive] SUCCESS: Chunk directory created in %v (status: %s)", mkcolDuration, resp.Status)

	// 2. Upload file in chunks
	log.Printf("[HiDrive] Starting chunk upload phase...")
	if err := c.uploadChunks(chunkDir, reader, chunkSize, destinationURL); err != nil {
		log.Printf("[HiDrive] ERROR: Chunk upload failed: %v", err)
		return err
	}
	log.Printf("[HiDrive] All chunks uploaded successfully")

	// 3. Assemble chunks by moving the directory
	moveSource := c.BaseURL + chunkDir + "/.file"
	
	log.Printf("[HiDrive] Starting MOVE operation from %s to %s", moveSource, destinationURL)
	req, err = http.NewRequest("MOVE", moveSource, nil)
	if err != nil {
		return fmt.Errorf("could not create MOVE request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Destination", destinationURL)
	// CRITICAL: Add OC-Total-Length header like bash script does!
	req.Header.Set("OC-Total-Length", fmt.Sprintf("%d", size))
	
	// For large files, MOVE operation can take very long server-side
	// Increase timeout specifically for this operation
	moveClient := &http.Client{
		Timeout: 10 * time.Minute, // 10 minutes for large file assembly
		Transport: c.HTTPClient.Transport,
	}
	
	log.Printf("[HiDrive] Executing MOVE operation (this may take several minutes for large files)...")
	moveStart := time.Now()
	resp, err = moveClient.Do(req)
	moveDuration := time.Since(moveStart)
	
	if err != nil {
		log.Printf("[HiDrive] MOVE operation failed after %v: %v", moveDuration, err)
		return fmt.Errorf("MOVE request failed after %v: %w", moveDuration, err)
	}
	defer resp.Body.Close()

	log.Printf("[HiDrive] MOVE operation completed in %v with status %s", moveDuration, resp.Status)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		// Read response body for more detailed error information
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[HiDrive] MOVE failed with status %s, response: %s", resp.Status, string(body))
		return fmt.Errorf("final MOVE to assemble chunks failed with status %s", resp.Status)
	}

	log.Printf("Chunked upload successful for %s", filePath)
	return nil
}

// uploadChunks uploads the file in chunks to the server
func (c *Client) uploadChunks(chunkDir string, reader io.Reader, chunkSize int64, destinationURL string) error {
	chunk := make([]byte, chunkSize)
	totalChunks := 0
	successfulChunks := 0
	
	log.Printf("[HiDrive] Starting chunk upload to %s (chunk size: %d bytes)", chunkDir, chunkSize)
	
	// CRITICAL: Start chunk numbering at 1, not 0! (HiDrive requires 1-based indexing)
	chunkNumber := 1
	for {
		bytesRead, readErr := reader.Read(chunk)
		if bytesRead > 0 {
			totalChunks++
			// Use 5-digit padded chunk names like bash script: 00001, 00002, 00003, etc.
			chunkPath := fmt.Sprintf("%s/%05d", chunkDir, chunkNumber)
			chunkURL := c.BaseURL + chunkPath

			log.Printf("[HiDrive] Uploading chunk %d: %d bytes to %s", chunkNumber, bytesRead, chunkPath)
			chunkStart := time.Now()

			// Retry logic for individual chunks
			var resp *http.Response
			var chunkErr error
			maxRetries := 3
			for attempt := 1; attempt <= maxRetries; attempt++ {
				req, err := http.NewRequest("PUT", chunkURL, bytes.NewReader(chunk[:bytesRead]))
				if err != nil {
					log.Printf("[HiDrive] ERROR: Could not create PUT request for chunk %d (attempt %d): %v", chunkNumber, attempt, err)
					if attempt == maxRetries {
						return fmt.Errorf("could not create PUT request for chunk %d after %d attempts: %w", chunkNumber, maxRetries, err)
					}
					continue
				}
				req.SetBasicAuth(c.Username, c.Password)
				req.Header.Set("Content-Type", "application/octet-stream")
				// CRITICAL: Add Destination header like bash script does for each chunk!
				req.Header.Set("Destination", destinationURL)
				req.ContentLength = int64(bytesRead)

				resp, chunkErr = c.HTTPClient.Do(req)
				if chunkErr != nil {
					log.Printf("[HiDrive] WARNING: PUT request for chunk %d failed (attempt %d/%d) after %v: %v", chunkNumber, attempt, maxRetries, time.Since(chunkStart), chunkErr)
					if attempt < maxRetries {
						time.Sleep(time.Duration(attempt) * time.Second) // Progressive backoff
						continue
					}
					return fmt.Errorf("PUT request for chunk %d failed after %d attempts: %w", chunkNumber, maxRetries, chunkErr)
				}

				// Check response status - Accept both 201 Created and 200 OK
				if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
					break // Success
				} else {
					// Read response body for detailed error information
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					log.Printf("[HiDrive] WARNING: Chunk %d upload failed (attempt %d/%d) with status %s after %v, response: %s", chunkNumber, attempt, maxRetries, resp.Status, time.Since(chunkStart), string(body))
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
			
			log.Printf("[HiDrive] SUCCESS: Chunk %d uploaded successfully in %v (status: %s)", chunkNumber, chunkDuration, resp.Status)
			chunkNumber++ // Increment for next chunk
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			log.Printf("[HiDrive] ERROR: Failed to read chunk %d: %v", chunkNumber, readErr)
			return fmt.Errorf("failed to read chunk %d: %w", chunkNumber, readErr)
		}
	}
	
	log.Printf("[HiDrive] Chunk upload summary: %d/%d chunks uploaded successfully", successfulChunks, totalChunks)
	return nil
}

// DownloadFile downloads a file
func (c *Client) DownloadFile(filePath string) (io.ReadCloser, error) {
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
	return nil
}
