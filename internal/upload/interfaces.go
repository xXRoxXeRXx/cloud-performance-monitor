package upload

import (
	"time"
)

// StateManager interface to avoid import cycles
type StateManager interface {
	SaveUploadState(state UploadState) error
	GetUploadState(service, instance, filePath string, fileSize int64, modTime time.Time) *UploadState
	RemoveUploadState(service, instance, filePath string) error
	ListActiveUploads() ([]UploadState, error)
}

// ChunkSizer interface for dynamic chunk sizing
type ChunkSizer interface {
	AdjustChunkSize(actualDuration time.Duration, chunkSize int) ChunkSizeStats
	GetChunkSize() int
	SetChunkSize(size int)
	IsEnabled() bool
	GetConfiguration() map[string]interface{}
}

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

// ChunkSizeStats holds statistics about chunk performance
type ChunkSizeStats struct {
	ChunkSize      int           `json:"chunk_size"`
	UploadTime     time.Duration `json:"upload_time"`
	ThroughputMBps float64       `json:"throughput_mbps"`
	Timestamp      time.Time     `json:"timestamp"`
}

// ResumeCapableClient interface for clients that support upload resume
type ResumeCapableClient interface {
	// Resume operations
	ResumeChunkedUpload(transferID string, fileSize int64) (ResumeInfo, error)
	DeleteStaleChunks(transferID string, staleChunks []int)
	CleanupUploadFolder(transferID string) error
	
	// Upload operations
	CreateUploadFolder(transferID string, fileSize int64, remotePath string) error
	UploadSingleChunk(filePath, transferID string, chunkNumber int, offset int64, 
		chunkSize int, fileSize int64, remotePath string) error
	MoveChunksToFinalFile(transferID, remotePath string, fileSize int64) error
	
	// Utility
	GenerateTransferID(filePath string, fileSize int64, modTime time.Time) string
}

// ResumeInfo contains information about a resumable upload
type ResumeInfo struct {
	TransferID   string            `json:"transfer_id"`
	UploadedSize int64             `json:"uploaded_size"`
	NextChunk    int               `json:"next_chunk"`
	Chunks       map[int]ChunkInfo `json:"chunks"`
	StaleChunks  []int             `json:"stale_chunks"`
}

// ChunkInfo represents information about an uploaded chunk
type ChunkInfo struct {
	Number int   `json:"number"`
	Size   int64 `json:"size"`
	Name   string `json:"name"`
}