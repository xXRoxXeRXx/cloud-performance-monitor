package upload

import (
	"fmt"
	"os"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// Manager handles resumable uploads for any service
type Manager struct {
	stateManager StateManager
	logger       utils.ClientLogger
}

// NewManager creates a new upload manager
func NewManager(stateManager StateManager, logger utils.ClientLogger) *Manager {
	return &Manager{
		stateManager: stateManager,
		logger:       logger,
	}
}

// UploadFileWithResume performs a resumable chunked upload
// This implements the same logic as Nextcloud Desktop Client for maximum robustness
func (m *Manager) UploadFileWithResume(
	client ResumeCapableClient,
	chunkSizer ChunkSizer,
	filePath string,
	remotePath string,
	service string,
	instance string,
) error {
	
	// Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()
	modTime := fileInfo.ModTime()

	m.logger.LogOperation(utils.INFO, service, instance, "upload", "start",
		"Starting chunked upload with resume capability", map[string]interface{}{
			"file":        filePath,
			"remote_path": remotePath,
			"file_size":   fileSize,
			"mod_time":    modTime.Format(time.RFC3339),
		})

	// 1. Check for existing upload state
	state := m.stateManager.GetUploadState(service, instance, filePath, fileSize, modTime)
	
	var transferID string
	var uploadedSize int64
	var nextChunk int

	if state != nil {
		// 2. Try to resume existing upload
		m.logger.LogOperation(utils.INFO, service, instance, "upload", "resume_attempt",
			"Attempting to resume existing upload", map[string]interface{}{
				"transfer_id":   state.TransferID,
				"uploaded_size": state.UploadedSize,
				"last_chunk":    state.LastChunk,
			})

		resumeInfo, err := client.ResumeChunkedUpload(state.TransferID, fileSize)
		if err != nil {
			m.logger.LogOperation(utils.WARN, service, instance, "upload", "resume_failed",
				"Failed to resume upload, starting fresh", map[string]interface{}{
					"transfer_id": state.TransferID,
					"error":       err.Error(),
				})

			// Clean up old upload folder (fire and forget like Nextcloud Client)
			go client.CleanupUploadFolder(state.TransferID)
		} else if resumeInfo.UploadedSize < fileSize {
			// Resume is possible
			transferID = state.TransferID
			uploadedSize = resumeInfo.UploadedSize
			nextChunk = resumeInfo.NextChunk

			// Delete stale chunks if any (fire and forget like Nextcloud Client)
			if len(resumeInfo.StaleChunks) > 0 {
				go client.DeleteStaleChunks(transferID, resumeInfo.StaleChunks)
			}

			m.logger.LogOperation(utils.INFO, service, instance, "upload", "resuming",
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
		// 3. Start new upload 
		transferID = client.GenerateTransferID(filePath, fileSize, modTime)
		uploadedSize = 0
		nextChunk = 1

		m.logger.LogOperation(utils.INFO, service, instance, "upload", "new_start",
			"Starting new chunked upload", map[string]interface{}{
				"transfer_id": transferID,
				"file":        filePath,
				"file_size":   fileSize,
			})

		// Create upload folder
		if err := client.CreateUploadFolder(transferID, fileSize, remotePath); err != nil {
			return fmt.Errorf("failed to create upload folder: %w", err)
		}
	}

	// 4. Save/update initial state
	uploadState := UploadState{
		TransferID:   transferID,
		FilePath:     filePath,
		RemotePath:   remotePath,
		FileSize:     fileSize,
		ModTime:      modTime,
		UploadedSize: uploadedSize,
		ChunkSize:    chunkSizer.GetChunkSize(),
		LastChunk:    nextChunk - 1,
		CreatedAt:    time.Now(),
		Service:      service,
		Instance:     instance,
	}

	if err := m.stateManager.SaveUploadState(uploadState); err != nil {
		m.logger.LogOperation(utils.WARN, service, instance, "upload", "state_save_error",
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
		err = client.UploadSingleChunk(filePath, transferID, nextChunk, uploadedSize, chunkSize, fileSize, remotePath)
		uploadDuration := time.Since(startTime)

		if err != nil {
			m.logger.LogOperation(utils.ERROR, service, instance, "chunk_upload", "failed",
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
		
		if err := m.stateManager.SaveUploadState(uploadState); err != nil {
			m.logger.LogOperation(utils.WARN, service, instance, "upload", "state_save_error",
				"Failed to save upload state after chunk", map[string]interface{}{
					"chunk_number": nextChunk - 1,
					"error":        err.Error(),
				})
		}

		m.logger.LogOperation(utils.INFO, service, instance, "chunk_upload", "success",
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
	m.logger.LogOperation(utils.INFO, service, instance, "upload", "move_start",
		"Starting final MOVE to assemble chunks", map[string]interface{}{
			"transfer_id":  transferID,
			"total_chunks": nextChunk - 1,
			"file_size":    fileSize,
		})

	moveStartTime := time.Now()
	err = client.MoveChunksToFinalFile(transferID, remotePath, fileSize)
	moveDuration := time.Since(moveStartTime)

	if err != nil {
		m.logger.LogOperation(utils.ERROR, service, instance, "move", "failed",
			"Final MOVE failed", map[string]interface{}{
				"transfer_id": transferID,
				"duration":    moveDuration.String(),
				"error":       err.Error(),
			})
		return fmt.Errorf("final MOVE failed: %w", err)
	}

	// 7. Clean up state on successful completion
	if err := m.stateManager.RemoveUploadState(service, instance, filePath); err != nil {
		m.logger.LogOperation(utils.WARN, service, instance, "upload", "state_cleanup_error",
			"Failed to clean up upload state", map[string]interface{}{
				"error": err.Error(),
			})
	}

	totalDuration := time.Since(uploadState.CreatedAt)
	averageThroughputMBps := float64(fileSize) / (1024 * 1024) / totalDuration.Seconds()

	m.logger.LogOperation(utils.INFO, service, instance, "upload", "completed",
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