package agent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/upload"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// UploadManager handles resumable uploads for all services
type UploadManager struct {
	stateManager upload.StateManager
	logger       utils.ClientLogger
}

// NewUploadManager creates a new upload manager with resume capability
func NewUploadManager(stateDir, service, instance string, logger utils.ClientLogger) *UploadManager {
	// Ensure the state file path includes the filename
	stateFile := filepath.Join(stateDir, "upload_states.json")
	stateManager := NewStateManager(stateFile, logger)

	return &UploadManager{
		stateManager: stateManager,
		logger:       logger,
	}
}

// UploadWithResume performs an upload with resume capability
func (um *UploadManager) UploadWithResume(client upload.ResumeCapableClient, 
	filePath, remotePath, service, instance string, fileSize int64, chunkSizeMB int) error {
	
	// Get file modification time for transfer ID generation
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("could not stat file %s: %w", filePath, err)
	}

	transferID := client.GenerateTransferID(filePath, fileSize, fileInfo.ModTime())
	
	// Check for existing upload state
	state := um.stateManager.GetUploadState(service, instance, filePath, fileSize, fileInfo.ModTime())
	if state != nil {
		um.logger.LogOperation(utils.INFO, "upload_manager", instance, "resume", "info",
			"Resuming upload from existing state", map[string]interface{}{
				"transfer_id":    transferID,
				"uploaded_size":  state.UploadedSize,
				"last_chunk":     state.LastChunk,
				"chunk_size":     state.ChunkSize,
			})
	}

	// Start fresh upload if no valid state
	if state == nil {
		// Create upload folder
		if err := client.CreateUploadFolder(transferID, fileSize, remotePath); err != nil {
			return fmt.Errorf("could not create upload folder: %w", err)
		}

		// Initialize new state
		chunkSizeBytes := chunkSizeMB * 1024 * 1024
		state = &upload.UploadState{
			TransferID:     transferID,
			FilePath:       filePath,
			RemotePath:     remotePath,
			FileSize:       fileSize,
			ModTime:        fileInfo.ModTime(),
			UploadedSize:   0,
			ChunkSize:      chunkSizeBytes,
			LastChunk:      -1,
			CreatedAt:      time.Now(),
			LastUpdated:    time.Now(),
			Service:        service,
			Instance:       instance,
		}

		if err := um.stateManager.SaveUploadState(*state); err != nil {
			um.logger.LogOperation(utils.WARN, "upload_manager", instance, "save", "warn",
				"Could not save initial upload state", map[string]interface{}{
					"transfer_id": transferID,
					"error":       err.Error(),
				})
		}
	} else {
		um.logger.LogOperation(utils.INFO, "upload_manager", "state", "resume", "info",
			"Resuming upload from existing state", map[string]interface{}{
				"transfer_id":    transferID,
				"uploaded_size":  state.UploadedSize,
				"last_chunk":     state.LastChunk,
				"chunk_size":     state.ChunkSize,
			})
	}

	// Perform chunked upload with resume
	currentOffset := state.UploadedSize
	
	for chunkIndex := state.LastChunk + 1; currentOffset < fileSize; chunkIndex++ {
		remainingBytes := fileSize - currentOffset
		currentChunkSize := state.ChunkSize
		if remainingBytes < int64(currentChunkSize) {
			currentChunkSize = int(remainingBytes)
		}
		
		// Ensure we don't have negative chunk size
		if currentChunkSize <= 0 {
			break
		}

		// Calculate total chunks for logging (approximate)
		estimatedTotalChunks := int((fileSize + int64(state.ChunkSize) - 1) / int64(state.ChunkSize))

		um.logger.LogOperation(utils.INFO, "upload_manager", "chunk", "upload", "info",
			"Uploading chunk", map[string]interface{}{
				"transfer_id":   transferID,
				"chunk_index":   chunkIndex,
				"chunk_size":    currentChunkSize,
				"offset":        currentOffset,
				"total_chunks":  estimatedTotalChunks,
			})

		startTime := time.Now()
		
		// Upload the chunk
		err := client.UploadSingleChunk(filePath, transferID, chunkIndex, currentOffset, 
			currentChunkSize, fileSize, remotePath)
		if err != nil {
			return fmt.Errorf("chunk upload failed for chunk %d: %w", chunkIndex, err)
		}

		duration := time.Since(startTime)
		
		// Keep using the same chunk size (no dynamic adjustment)
		// Use the original chunk size from config

		// Update state
		currentOffset += int64(currentChunkSize)
		state.UploadedSize = currentOffset
		state.LastChunk = chunkIndex
		// Keep the same chunk size - don't change: state.ChunkSize = newChunkSize
		state.LastUpdated = time.Now()

		// Save state after each successful chunk
		if err := um.stateManager.SaveUploadState(*state); err != nil {
			um.logger.LogOperation(utils.WARN, "upload_manager", instance, "save", "warn",
				"Could not save upload state after chunk", map[string]interface{}{
					"transfer_id":  transferID,
					"chunk_index":  chunkIndex,
					"error":        err.Error(),
				})
		}

		um.logger.LogOperation(utils.INFO, "upload_manager", "chunk", "success", "success",
			"Chunk uploaded successfully", map[string]interface{}{
				"transfer_id":     transferID,
				"chunk_index":     chunkIndex,
				"duration_ms":     duration.Milliseconds(),
				"chunk_size":      currentChunkSize,
				"uploaded_size":   state.UploadedSize,
				"progress_pct":    float64(state.UploadedSize*100) / float64(fileSize),
			})
	}

	// Move chunks to final file
	um.logger.LogOperation(utils.INFO, "upload_manager", "finalize", "start", "info",
		"Assembling chunks into final file", map[string]interface{}{
			"transfer_id":   transferID,
			"final_size":    currentOffset,
			"file_size":     fileSize,
		})

	if err := client.MoveChunksToFinalFile(transferID, remotePath, fileSize); err != nil {
		return fmt.Errorf("could not assemble final file: %w", err)
	}

	// Clean up state after successful upload
	if err := um.stateManager.RemoveUploadState(service, instance, filePath); err != nil {
		um.logger.LogOperation(utils.WARN, "upload_manager", instance, "cleanup", "warn",
			"Could not clean up upload state", map[string]interface{}{
				"transfer_id": transferID,
				"error":       err.Error(),
			})
	}

	um.logger.LogOperation(utils.INFO, "upload_manager", instance, "complete", "success",
		"Upload completed successfully", map[string]interface{}{
			"transfer_id": transferID,
			"file_size":   fileSize,
			"file_path":   filePath,
			"remote_path": remotePath,
		})

	return nil
}

// CreateTempFile creates a temporary file with random data for testing
func (um *UploadManager) CreateTempFile(size int64) (string, error) {
	tempDir := os.TempDir()
	fileName := fmt.Sprintf("cloud_test_%d.tmp", time.Now().UnixNano())
	filePath := filepath.Join(tempDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}
	defer file.Close()

	// Write random data
	reader := io.LimitReader(&randomReader{}, size)
	_, err = io.Copy(file, reader)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("could not write random data: %w", err)
	}

	return filePath, nil
}

// CleanupTempFile removes a temporary test file
func (um *UploadManager) CleanupTempFile(filePath string) {
	if err := os.Remove(filePath); err != nil {
		um.logger.LogOperation(utils.WARN, "upload_manager", "cleanup", "temp_file", "warn",
			"Could not remove temp file", map[string]interface{}{
				"file_path": filePath,
				"error":     err.Error(),
			})
	}
}