package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// UploadState represents the persistent state of a chunked upload
type UploadState struct {
	TransferID     string    `json:"transfer_id"`
	FilePath       string    `json:"file_path"`
	RemotePath     string    `json:"remote_path"`
	FileSize       int64     `json:"file_size"`
	ModTime        time.Time `json:"mod_time"`
	UploadedSize   int64     `json:"uploaded_size"`
	ChunkSize      int       `json:"chunk_size"`
	LastChunk      int       `json:"last_chunk"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdated    time.Time `json:"last_updated"`
	Service        string    `json:"service"`
	Instance       string    `json:"instance"`
}

// StateManager manages persistent upload states across restarts
type StateManager struct {
	stateFile string
	mutex     sync.RWMutex
	logger    utils.ClientLogger
}

// NewStateManager creates a new state manager with the given state file path
func NewStateManager(stateFile string, logger utils.ClientLogger) *StateManager {
	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.LogOperation(utils.ERROR, "state_manager", "local", "mkdir", "error", 
			"Failed to create state directory", map[string]interface{}{
				"dir": dir,
				"error": err.Error(),
			})
	}

	return &StateManager{
		stateFile: stateFile,
		logger:    logger,
	}
}

// SaveUploadState saves the upload state to persistent storage
func (sm *StateManager) SaveUploadState(state UploadState) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	states := make(map[string]UploadState)
	if data, err := os.ReadFile(sm.stateFile); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			sm.logger.LogOperation(utils.WARN, "state_manager", "local", "unmarshal", "error",
				"Failed to parse existing state file", map[string]interface{}{
					"error": err.Error(),
				})
		}
	}

	// Use combination of service, instance, and file path as key
	key := fmt.Sprintf("%s:%s:%s", state.Service, state.Instance, state.FilePath)
	state.LastUpdated = time.Now()
	states[key] = state

	// Clean up old states (older than 24 hours)
	sm.cleanupOldStates(states)

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal states: %w", err)
	}

	if err := os.WriteFile(sm.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	sm.logger.LogOperation(utils.DEBUG, "state_manager", "local", "save", "success",
		"Upload state saved", map[string]interface{}{
			"service":  state.Service,
			"instance": state.Instance,
			"file":     state.FilePath,
			"uploaded": state.UploadedSize,
			"total":    state.FileSize,
			"chunk":    state.LastChunk,
		})

	return nil
}

// GetUploadState retrieves the upload state for a file if it's still valid
func (sm *StateManager) GetUploadState(service, instance, filePath string, fileSize int64, modTime time.Time) *UploadState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	states := make(map[string]UploadState)
	if data, err := os.ReadFile(sm.stateFile); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			sm.logger.LogOperation(utils.WARN, "state_manager", "local", "parse", "error",
				"Failed to parse state file", map[string]interface{}{
					"error": err.Error(),
				})
			return nil
		}
	}

	key := fmt.Sprintf("%s:%s:%s", service, instance, filePath)
	state, exists := states[key]
	if !exists {
		return nil
	}

	// Validate state is still valid (same file size and modification time)
	if state.FileSize == fileSize && state.ModTime.Equal(modTime) {
		// Check if state is not too old (max 24 hours)
		if time.Since(state.CreatedAt) < 24*time.Hour {
			sm.logger.LogOperation(utils.INFO, "state_manager", service, "resume", "found",
				"Found valid upload state for resume", map[string]interface{}{
					"instance": instance,
					"file":     filePath,
					"uploaded": state.UploadedSize,
					"total":    fileSize,
					"age":      time.Since(state.CreatedAt).String(),
				})
			return &state
		} else {
			sm.logger.LogOperation(utils.DEBUG, "state_manager", service, "resume", "too_old",
				"Upload state too old, ignoring", map[string]interface{}{
					"instance": instance,
					"file":     filePath,
					"age":      time.Since(state.CreatedAt).String(),
				})
		}
	} else {
		sm.logger.LogOperation(utils.DEBUG, "state_manager", service, "resume", "invalid",
			"Upload state invalid (file changed)", map[string]interface{}{
				"instance":     instance,
				"file":         filePath,
				"old_size":     state.FileSize,
				"new_size":     fileSize,
				"old_modtime":  state.ModTime.String(),
				"new_modtime":  modTime.String(),
			})
	}

	return nil
}

// RemoveUploadState removes the upload state for a completed upload
func (sm *StateManager) RemoveUploadState(service, instance, filePath string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	states := make(map[string]UploadState)
	if data, err := os.ReadFile(sm.stateFile); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			return fmt.Errorf("failed to parse state file: %w", err)
		}
	}

	key := fmt.Sprintf("%s:%s:%s", service, instance, filePath)
	delete(states, key)

	data, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal states: %w", err)
	}

	if err := os.WriteFile(sm.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	sm.logger.LogOperation(utils.DEBUG, "state_manager", service, "remove", "success",
		"Upload state removed", map[string]interface{}{
			"instance": instance,
			"file":     filePath,
		})

	return nil
}

// cleanupOldStates removes states older than 24 hours
func (sm *StateManager) cleanupOldStates(states map[string]UploadState) {
	now := time.Now()
	for key, state := range states {
		if now.Sub(state.CreatedAt) > 24*time.Hour {
			delete(states, key)
			sm.logger.LogOperation(utils.DEBUG, "state_manager", state.Service, "cleanup", "removed",
				"Cleaned up old upload state", map[string]interface{}{
					"file": state.FilePath,
					"age":  now.Sub(state.CreatedAt).String(),
				})
		}
	}
}

// ListActiveUploads returns all currently tracked upload states
func (sm *StateManager) ListActiveUploads() ([]UploadState, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	states := make(map[string]UploadState)
	if data, err := os.ReadFile(sm.stateFile); err == nil {
		if err := json.Unmarshal(data, &states); err != nil {
			return nil, fmt.Errorf("failed to parse state file: %w", err)
		}
	}

	var activeStates []UploadState
	now := time.Now()
	for _, state := range states {
		if now.Sub(state.CreatedAt) < 24*time.Hour {
			activeStates = append(activeStates, state)
		}
	}

	return activeStates, nil
}