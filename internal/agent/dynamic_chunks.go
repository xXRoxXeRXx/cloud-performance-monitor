package agent

import (
	"math"
	"time"

	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/upload"
	"github.com/xXRoxXeRXx/cloud-performance-monitor/internal/utils"
)

// ChunkSizer manages dynamic chunk sizing based on upload performance
// Inspired by Nextcloud Desktop Client's dynamic chunk sizing algorithm
// Implements upload.ChunkSizer interface
type ChunkSizer struct {
	targetDuration time.Duration
	currentSize    int
	minSize        int
	maxSize        int
	alpha          float64 // exponential moving average factor
	logger         utils.ClientLogger
	service        string
	instance       string
}

// ChunkSizeStats holds statistics about chunk performance
type ChunkSizeStats struct {
	ChunkSize      int           `json:"chunk_size"`
	UploadTime     time.Duration `json:"upload_time"`
	ThroughputMBps float64       `json:"throughput_mbps"`
	Timestamp      time.Time     `json:"timestamp"`
}

// NewChunkSizer creates a new dynamic chunk sizer
func NewChunkSizer(targetDuration time.Duration, service, instance string, logger utils.ClientLogger) upload.ChunkSizer {
	return &ChunkSizer{
		targetDuration: targetDuration,
		currentSize:    10 * 1024 * 1024, // 10MB start (like Nextcloud default)
		minSize:        1 * 1024 * 1024,  // 1MB minimum
		maxSize:        100 * 1024 * 1024, // 100MB maximum  
		alpha:          0.5, // exponential moving average factor (like Nextcloud)
		logger:         logger,
		service:        service,
		instance:       instance,
	}
}

// NewChunkSizerWithConfig creates a chunk sizer with custom configuration
func NewChunkSizerWithConfig(targetDuration time.Duration, initialSize, minSize, maxSize int, 
	service, instance string, logger utils.ClientLogger) *ChunkSizer {
	return &ChunkSizer{
		targetDuration: targetDuration,
		currentSize:    initialSize,
		minSize:        minSize,
		maxSize:        maxSize,
		alpha:          0.5,
		logger:         logger,
		service:        service,
		instance:       instance,
	}
}

// AdjustChunkSize adjusts the chunk size based on upload performance
// Algorithm inspired by Nextcloud Desktop Client:
// predictedGoodSize = (currentChunkSize * targetDuration) / actualUploadTime
// newChunkSize = (oldChunkSize * alpha + predictedGoodSize * (1-alpha))
func (cs *ChunkSizer) AdjustChunkSize(actualDuration time.Duration, chunkSize int) upload.ChunkSizeStats {
	if cs.targetDuration <= 0 {
		// Dynamic sizing disabled
		stats := upload.ChunkSizeStats{
			ChunkSize:      chunkSize,
			UploadTime:     actualDuration,
			ThroughputMBps: calculateThroughput(chunkSize, actualDuration),
			Timestamp:      time.Now(),
		}
		return stats
	}

	// Add small value to avoid division by zero (like Nextcloud client does)
	uploadTimeMs := actualDuration.Milliseconds() + 1
	targetTimeMs := cs.targetDuration.Milliseconds()

	// Calculate predicted good size using Nextcloud's formula
	predictedGoodSize := int64(chunkSize) * targetTimeMs / uploadTimeMs

	// Apply exponential moving average (like Nextcloud client)
	// targetSize = currentSize/2 + predictedGoodSize/2
	targetSize := int(float64(cs.currentSize)*cs.alpha + float64(predictedGoodSize)*(1-cs.alpha))

	// Bound the size within configured limits
	if targetSize < cs.minSize {
		targetSize = cs.minSize
	}
	if targetSize > cs.maxSize {
		targetSize = cs.maxSize
	}

	oldSize := cs.currentSize
	cs.currentSize = targetSize

	throughput := calculateThroughput(chunkSize, actualDuration)

	// Log chunk size adjustment (like Nextcloud client does)
	cs.logger.LogOperation(utils.INFO, cs.service, cs.instance, "chunk_sizing", "adjusted",
		"Dynamic chunk size adjustment", map[string]interface{}{
			"old_chunk_size":      oldSize,
			"new_chunk_size":      cs.currentSize,
			"actual_duration_ms":  uploadTimeMs - 1, // subtract the +1 we added
			"target_duration_ms":  targetTimeMs,
			"predicted_good_size": predictedGoodSize,
			"throughput_mbps":     throughput,
			"chunk_bytes":         chunkSize,
		})

	stats := upload.ChunkSizeStats{
		ChunkSize:      cs.currentSize,
		UploadTime:     actualDuration,
		ThroughputMBps: throughput,
		Timestamp:      time.Now(),
	}

	return stats
}

// GetChunkSize returns the current optimal chunk size
func (cs *ChunkSizer) GetChunkSize() int {
	return cs.currentSize
}

// SetChunkSize manually sets the chunk size (for testing or manual override)
func (cs *ChunkSizer) SetChunkSize(size int) {
	if size < cs.minSize {
		size = cs.minSize
	}
	if size > cs.maxSize {
		size = cs.maxSize
	}
	
	oldSize := cs.currentSize
	cs.currentSize = size

	cs.logger.LogOperation(utils.INFO, cs.service, cs.instance, "chunk_sizing", "manual",
		"Manual chunk size change", map[string]interface{}{
			"old_size": oldSize,
			"new_size": cs.currentSize,
		})
}

// GetConfiguration returns current chunk sizer configuration
func (cs *ChunkSizer) GetConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"target_duration_ms": cs.targetDuration.Milliseconds(),
		"current_size":       cs.currentSize,
		"min_size":           cs.minSize,
		"max_size":           cs.maxSize,
		"alpha":              cs.alpha,
		"service":            cs.service,
		"instance":           cs.instance,
	}
}

// IsEnabled returns true if dynamic chunk sizing is enabled
func (cs *ChunkSizer) IsEnabled() bool {
	return cs.targetDuration > 0
}

// calculateThroughput calculates upload throughput in MB/s
func calculateThroughput(bytes int, duration time.Duration) float64 {
	if duration.Seconds() == 0 {
		return 0
	}
	
	megabytes := float64(bytes) / (1024 * 1024)
	return megabytes / duration.Seconds()
}

// GetOptimalChunkSize calculates optimal chunk size based on current network conditions
// This is a helper function that can be used to initialize chunk size based on quick tests
func GetOptimalChunkSize(estimatedBandwidthMbps float64, targetDuration time.Duration, 
	minSize, maxSize int) int {
	if targetDuration <= 0 || estimatedBandwidthMbps <= 0 {
		return 10 * 1024 * 1024 // Default 10MB
	}

	// Convert bandwidth from Mbps to bytes per second
	bytesPerSecond := estimatedBandwidthMbps * 1024 * 1024 / 8

	// Calculate optimal size: bandwidth * targetDuration
	optimalSize := int(bytesPerSecond * targetDuration.Seconds())

	// Round to nearest MB
	optimalSize = int(math.Round(float64(optimalSize)/(1024*1024))) * 1024 * 1024

	// Apply bounds
	if optimalSize < minSize {
		return minSize
	}
	if optimalSize > maxSize {
		return maxSize
	}

	return optimalSize
}