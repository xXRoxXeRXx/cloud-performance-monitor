package nextcloud

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/agent"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// UploadFileChunkedV2WithResume uploads a file using chunked upload v2 with resume capability
// This implementation follows the Nextcloud Desktop Client's logic for robustness
func (c *Client) UploadFileChunkedV2WithResume(filePath string, remotePath string, 
	stateManager *agent.StateManager, chunkSizer *agent.ChunkSizer) error {
	
	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime()

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "start",
		"Starting chunked upload with resume capability", map[string]interface{}{
			"file":        filePath,
			"remote_path": remotePath,
			"file_size":   fileSize,
			"mod_time":    modTime.Format(time.RFC3339),
		})

	// 1. Check for existing upload state
	state := stateManager.GetUploadState("nextcloud", c.BaseURL, filePath, fileSize, modTime)
	
	var transferID string
	var uploadedSize int64
	var nextChunk int

	if state != nil {
		// 2. Try to resume existing upload
		c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "resume_attempt",
			"Attempting to resume existing upload", map[string]interface{}{
				"transfer_id":   state.TransferID,
				"uploaded_size": state.UploadedSize,
				"last_chunk":    state.LastChunk,
			})

		resumeInfo, err := c.ResumeChunkedUpload(state.TransferID, fileSize)
		if err != nil {
			c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "upload", "resume_failed",
				"Failed to resume upload, starting fresh", map[string]interface{}{
					"transfer_id": state.TransferID,
					"error":       err.Error(),
				})

			// Clean up old upload folder (fire and forget like Nextcloud Client)
			go c.CleanupUploadFolder(state.TransferID)
		} else if resumeInfo.UploadedSize < fileSize {
			// Resume is possible
			transferID = state.TransferID
			uploadedSize = resumeInfo.UploadedSize
			nextChunk = resumeInfo.NextChunk

			// Delete stale chunks if any (fire and forget like Nextcloud Client)
			if len(resumeInfo.StaleChunks) > 0 {
				go c.DeleteStaleChunks(transferID, resumeInfo.StaleChunks)
			}

			c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "resuming",
				"Resuming upload", map[string]interface{}{
					"transfer_id":   transferID,
					"uploaded_size": uploadedSize,
					"next_chunk":    nextChunk,
					"stale_chunks":  len(resumeInfo.StaleChunks),
					"progress":      float64(uploadedSize) / float64(fileSize) * 100,
				})
		}
	}

	if transferID == "" {
		// 3. Start new upload (like Nextcloud Client's startNewUpload)
		transferID = c.generateTransferID(filePath, fileSize, modTime)
		uploadedSize = 0
		nextChunk = 1

		c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "new_start",
			"Starting new chunked upload", map[string]interface{}{
				"transfer_id": transferID,
				"file":        filePath,
				"file_size":   fileSize,
			})

		// Create upload folder (MKCOL)
		if err := c.createUploadFolder(transferID, fileSize, remotePath); err != nil {
			return fmt.Errorf("failed to create upload folder: %w", err)
		}
	}

	// 4. Save/update initial state
	uploadState := agent.UploadState{
		TransferID:   transferID,
		FilePath:     filePath,
		RemotePath:   remotePath,
		FileSize:     fileSize,
		ModTime:      modTime,
		UploadedSize: uploadedSize,
		ChunkSize:    chunkSizer.GetChunkSize(),
		LastChunk:    nextChunk - 1,
		CreatedAt:    time.Now(),
		Service:      "nextcloud",
		Instance:     c.BaseURL,
	}

	if err := stateManager.SaveUploadState(uploadState); err != nil {
		c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "upload", "state_save_error",
			"Failed to save upload state", map[string]interface{}{
				"error": err.Error(),
			})
	}

	// 5. Upload remaining chunks with dynamic sizing
	for uploadedSize < fileSize {
		// Check if file changed during upload (like Nextcloud Client does)
		currentFileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file disappeared during upload: %w", err)
		}

		if currentFileInfo.Size() != fileSize || !currentFileInfo.ModTime().Equal(modTime) {
			return fmt.Errorf("file changed during upload (size: %d->%d, modtime: %v->%v)",
				fileSize, currentFileInfo.Size(), modTime, currentFileInfo.ModTime())
		}

		// Calculate chunk size (dynamic or remaining bytes)
		chunkSize := chunkSizer.GetChunkSize()
		remaining := fileSize - uploadedSize
		if int64(chunkSize) > remaining {
			chunkSize = int(remaining)
		}

		// Upload chunk with timing
		startTime := time.Now()
		err := c.uploadChunk(filePath, transferID, nextChunk, uploadedSize, chunkSize, fileSize, remotePath)
		uploadDuration := time.Since(startTime)

		if err != nil {
			c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "chunk_upload", "failed",
				"Chunk upload failed", map[string]interface{}{
					"transfer_id":    transferID,
					"chunk_number":   nextChunk,
					"chunk_size":     chunkSize,
					"uploaded_size":  uploadedSize,
					"duration":       uploadDuration.String(),
					"error":          err.Error(),
				})
			return fmt.Errorf("chunk %d upload failed: %w", nextChunk, err)
		}

		uploadedSize += int64(chunkSize)
		nextChunk++

		// Adjust chunk size based on performance (like Nextcloud Client)
		stats := chunkSizer.AdjustChunkSize(uploadDuration, chunkSize)
		
		// Update state after each successful chunk
		uploadState.UploadedSize = uploadedSize
		uploadState.LastChunk = nextChunk - 1
		uploadState.ChunkSize = stats.ChunkSize
		
		if err := stateManager.SaveUploadState(uploadState); err != nil {
			c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "upload", "state_save_error",
				"Failed to save upload state after chunk", map[string]interface{}{
					"chunk_number": nextChunk - 1,
					"error":        err.Error(),
				})
		}

		c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "chunk_upload", "success",
			"Chunk uploaded successfully", map[string]interface{}{
				"transfer_id":     transferID,
				"chunk_number":    nextChunk - 1,
				"chunk_size":      chunkSize,
				"uploaded_size":   uploadedSize,
				"total_size":      fileSize,
				"duration":        uploadDuration.String(),
				"throughput_mbps": stats.ThroughputMBps,
				"progress":        float64(uploadedSize) / float64(fileSize) * 100,
			})
	}

	// 6. Final MOVE to assemble chunks
	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "move_start",
		"Starting final MOVE to assemble chunks", map[string]interface{}{
			"transfer_id":  transferID,
			"total_chunks": nextChunk - 1,
			"file_size":    fileSize,
		})

	moveStartTime := time.Now()
	err = c.moveChunksToFinalFile(transferID, remotePath, fileSize)
	moveDuration := time.Since(moveStartTime)

	if err != nil {
		c.logger.LogOperation(utils.ERROR, "nextcloud", c.BaseURL, "move", "failed",
			"Final MOVE failed", map[string]interface{}{
				"transfer_id": transferID,
				"duration":    moveDuration.String(),
				"error":       err.Error(),
			})
		return fmt.Errorf("final MOVE failed: %w", err)
	}

	// 7. Clean up state on successful completion
	if err := stateManager.RemoveUploadState("nextcloud", c.BaseURL, filePath); err != nil {
		c.logger.LogOperation(utils.WARN, "nextcloud", c.BaseURL, "upload", "state_cleanup_error",
			"Failed to clean up upload state", map[string]interface{}{
				"error": err.Error(),
			})
	}

	totalDuration := time.Since(uploadState.CreatedAt)
	averageThroughputMBps := float64(fileSize) / (1024 * 1024) / totalDuration.Seconds()

	c.logger.LogOperation(utils.INFO, "nextcloud", c.BaseURL, "upload", "completed",
		"Chunked upload completed successfully", map[string]interface{}{
			"transfer_id":            transferID,
			"file":                   filePath,
			"remote_path":            remotePath,
			"file_size":              fileSize,
			"total_chunks":           nextChunk - 1,
			"total_duration":         totalDuration.String(),
			"move_duration":          moveDuration.String(),
			"average_throughput_mbps": averageThroughputMBps,
		})

	return nil
}

// generateTransferID generates a transfer ID like Nextcloud Desktop Client
// Formula: rand() ^ modtime ^ (size << 16) ^ hash(filename)
func (c *Client) generateTransferID(filePath string, fileSize int64, modTime time.Time) string {
	// Simple implementation - in real world you'd want crypto/rand
	h := md5.New()
	h.Write([]byte(filePath))
	
	// Mix with modtime and size like Nextcloud Client
	seed := uint64(modTime.Unix()) ^ uint64(fileSize<<16) ^ uint64(h.Sum(nil)[0])
	
	return fmt.Sprintf("%d", seed)
}

// createUploadFolder creates the upload folder with MKCOL (like Nextcloud Client)
func (c *Client) createUploadFolder(transferID string, fileSize int64, remotePath string) error {
	return c.CreateChunkDirectory(transferID, fileSize, remotePath)
}

// uploadChunk uploads a single chunk (wrapper for existing method)
func (c *Client) uploadChunk(filePath, transferID string, chunkNumber int, offset int64, 
	chunkSize int, fileSize int64, remotePath string) error {
	
	return c.UploadChunk(filePath, transferID, chunkNumber, offset, chunkSize, fileSize, remotePath)
}

// moveChunksToFinalFile performs the final MOVE operation (wrapper for existing method)
func (c *Client) moveChunksToFinalFile(transferID, remotePath string, fileSize int64) error {
	return c.MoveChunksToFile(transferID, remotePath, fileSize)
}