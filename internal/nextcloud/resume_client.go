package nextcloud

import (
	"crypto/md5"
	"fmt"
	"net/http"
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
	// TODO: Implement using the existing MKCOL logic from UploadFile
	return fmt.Errorf("CreateUploadFolder not yet implemented - need to extract from UploadFile method")
}

// UploadSingleChunk implements upload.ResumeCapableClient interface  
func (rc *ResumeClient) UploadSingleChunk(filePath, transferID string, chunkNumber int, offset int64, 
	chunkSize int, fileSize int64, remotePath string) error {
	// TODO: Use the existing uploadChunk method or extract chunking logic
	return fmt.Errorf("UploadSingleChunk not yet implemented - need to refactor existing chunking logic")
}

// MoveChunksToFinalFile implements upload.ResumeCapableClient interface
func (rc *ResumeClient) MoveChunksToFinalFile(transferID, remotePath string, fileSize int64) error {
	// TODO: Implement using the existing MOVE logic from UploadFile  
	return fmt.Errorf("MoveChunksToFinalFile not yet implemented - need to extract from UploadFile method")
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