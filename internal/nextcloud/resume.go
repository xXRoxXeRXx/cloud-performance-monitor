package nextcloud

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// ResumeInfo contains information about a resumable upload
type ResumeInfo struct {
	TransferID   string               `json:"transfer_id"`
	UploadedSize int64                `json:"uploaded_size"`
	NextChunk    int                  `json:"next_chunk"`
	Chunks       map[int]ChunkInfo    `json:"chunks"`
	StaleChunks  []int                `json:"stale_chunks"` // chunks that need to be deleted
}

// ChunkInfo represents information about an uploaded chunk
type ChunkInfo struct {
	Number int   `json:"number"`
	Size   int64 `json:"size"`
	Name   string `json:"name"`
}

// PropfindResponse represents the XML response from a PROPFIND request
type PropfindResponse struct {
	XMLName   xml.Name   `xml:"multistatus"`
	Namespace string     `xml:"xmlns,attr"`
	Responses []Response `xml:"response"`
}

// Response represents a single response in PROPFIND
type Response struct {
	Href  string `xml:"href"`
	Props Props  `xml:"propstat>prop"`
}

// Props represents properties of a WebDAV resource
type Props struct {
	ResourceType    ResourceType `xml:"resourcetype"`
	ContentLength   string       `xml:"getcontentlength"`
}

// ResourceType indicates if the resource is a collection (directory)
type ResourceType struct {
	Collection *struct{} `xml:"collection"`
}

// ResumeChunkedUpload checks for existing chunks and returns resume information
// This implements the same logic as Nextcloud Desktop Client's slotPropfindFinished()
func (c *Client) ResumeChunkedUpload(transferID string, fileSize int64) (*ResumeInfo, error) {
	uploadPath := fmt.Sprintf("/remote.php/dav/uploads/%s/%s", c.Username, transferID)
	url := c.BaseURL + uploadPath

	// Create PROPFIND request like Nextcloud Desktop Client
	propfindBody := `<?xml version="1.0" encoding="UTF-8"?>
<d:propfind xmlns:d="DAV:">
	<d:prop>
		<d:resourcetype/>
		<d:getcontentlength/>
	</d:prop>
</d:propfind>`

	req, err := http.NewRequest("PROPFIND", url, strings.NewReader(propfindBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create PROPFIND request: %w", err)
	}

	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(c.Username, c.Password)

	// Set timeout for PROPFIND (shorter than chunk uploads)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PROPFIND request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 207 { // Multi-Status
		return nil, fmt.Errorf("PROPFIND failed with status %d", resp.StatusCode)
	}

	// Parse PROPFIND response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PROPFIND response: %w", err)
	}

	var propfindResp PropfindResponse
	if err := xml.Unmarshal(body, &propfindResp); err != nil {
		return nil, fmt.Errorf("failed to parse PROPFIND response: %w", err)
	}

	// Parse chunks from response (like Nextcloud Client's slotPropfindIterate)
	chunks := make(map[int]ChunkInfo)
	uploadFolderPath := uploadPath + "/"

	for _, response := range propfindResp.Responses {
		// Skip the upload folder itself
		if response.Href == uploadPath || response.Href == uploadFolderPath {
			continue
		}

		// Skip directories (collections)
		if response.Props.ResourceType.Collection != nil {
			continue
		}

		// Extract chunk name from href
		chunkName := strings.TrimPrefix(response.Href, uploadFolderPath)
		chunkName = strings.TrimSuffix(chunkName, "/")

		// Parse chunk number (Nextcloud v2 chunks are numbered 00001, 00002, etc.)
		chunkNum, err := strconv.Atoi(chunkName)
		if err != nil {
			c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "propfind", "parse_error",
				"Failed to parse chunk number", map[string]interface{}{
					"chunk_name": chunkName,
					"href":       response.Href,
					"error":      err.Error(),
				})
			continue
		}

		// Parse content length
		contentLength, err := strconv.ParseInt(response.Props.ContentLength, 10, 64)
		if err != nil {
			c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "propfind", "parse_error",
				"Failed to parse content length", map[string]interface{}{
					"chunk_name":     chunkName,
					"content_length": response.Props.ContentLength,
					"error":          err.Error(),
				})
			continue
		}

		chunks[chunkNum] = ChunkInfo{
			Number: chunkNum,
			Size:   contentLength,
			Name:   chunkName,
		}
	}

	// Calculate resume position (like Nextcloud Client's logic)
	resumeInfo := &ResumeInfo{
		TransferID: transferID,
		Chunks:     chunks,
	}

	// Find consecutive chunks from beginning (like Nextcloud Desktop Client)
	currentChunk := 1
	uploadedSize := int64(0)

	for {
		if chunkInfo, exists := chunks[currentChunk]; exists {
			uploadedSize += chunkInfo.Size
			currentChunk++
		} else {
			break
		}
	}

	// Check for inconsistency (like Nextcloud Client does)
	if uploadedSize > fileSize {
		c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "resume", "inconsistency",
			"Inconsistency while resuming: server has more data than file size", map[string]interface{}{
				"uploaded_size": uploadedSize,
				"file_size":     fileSize,
				"transfer_id":   transferID,
			})

		// Return error to trigger cleanup and restart (like Nextcloud Client)
		return nil, fmt.Errorf("inconsistency detected: uploaded %d bytes > file size %d bytes", 
			uploadedSize, fileSize)
	}

	// Identify stale chunks (gaps) that need to be deleted (like Nextcloud Client)
	var staleChunks []int
	for chunkNum := range chunks {
		if chunkNum >= currentChunk {
			staleChunks = append(staleChunks, chunkNum)
		}
	}

	resumeInfo.UploadedSize = uploadedSize
	resumeInfo.NextChunk = currentChunk
	resumeInfo.StaleChunks = staleChunks

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "resume", "calculated",
		"Resume information calculated", map[string]interface{}{
			"transfer_id":    transferID,
			"uploaded_size":  uploadedSize,
			"file_size":      fileSize,
			"next_chunk":     currentChunk,
			"total_chunks":   len(chunks),
			"stale_chunks":   len(staleChunks),
			"resume_percent": float64(uploadedSize) / float64(fileSize) * 100,
		})

	return resumeInfo, nil
}

// DeleteStaleChunks removes stale chunks from the server (fire and forget like Nextcloud Client)
func (c *Client) DeleteStaleChunks(transferID string, staleChunks []int) {
	if len(staleChunks) == 0 {
		return
	}

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "cleanup", "start",
		"Deleting stale chunks", map[string]interface{}{
			"transfer_id":  transferID,
			"chunk_count":  len(staleChunks),
			"chunk_numbers": staleChunks,
		})

	// Delete chunks in parallel (fire and forget)
	for _, chunkNum := range staleChunks {
		go func(chunk int) {
			chunkName := fmt.Sprintf("%05d", chunk)
			chunkURL := fmt.Sprintf("%s/remote.php/dav/uploads/%s/%s/%s", 
				c.BaseURL, c.Username, transferID, chunkName)

			req, err := http.NewRequest("DELETE", chunkURL, nil)
			if err != nil {
				c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "cleanup", "request_error",
					"Failed to create DELETE request for stale chunk", map[string]interface{}{
						"chunk_number": chunk,
						"error":        err.Error(),
					})
				return
			}

			req.SetBasicAuth(c.Username, c.Password)
			
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "cleanup", "network_error",
					"Failed to delete stale chunk", map[string]interface{}{
						"chunk_number": chunk,
						"error":        err.Error(),
					})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 204 && resp.StatusCode != 404 {
				c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "cleanup", "status_error",
					"Unexpected status when deleting stale chunk", map[string]interface{}{
						"chunk_number": chunk,
						"status_code":  resp.StatusCode,
					})
			} else {
				c.logger.LogOperation(utils.DEBUG, "nextcloud", c.BaseURL, "cleanup", "success",
					"Stale chunk deleted", map[string]interface{}{
						"chunk_number": chunk,
						"status_code":  resp.StatusCode,
					})
			}
		}(chunkNum)
	}
}

// CleanupUploadFolder removes the entire upload folder (used when starting fresh)
func (c *Client) CleanupUploadFolder(transferID string) error {
	uploadURL := fmt.Sprintf("%s/remote.php/dav/uploads/%s/%s", c.BaseURL, c.Username, transferID)

	req, err := http.NewRequest("DELETE", uploadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	req.SetBasicAuth(c.Username, c.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE request failed: %w", err)
	}
	defer resp.Body.Close()

	// 204 No Content or 404 Not Found are both acceptable
	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		return fmt.Errorf("cleanup failed with status %d", resp.StatusCode)
	}

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "cleanup", "folder_deleted",
		"Upload folder cleaned up", map[string]interface{}{
			"transfer_id": transferID,
			"status_code": resp.StatusCode,
		})

	return nil
}
