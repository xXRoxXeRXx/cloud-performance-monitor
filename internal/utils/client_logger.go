package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// ClientLogger provides a simple logging interface for clients to avoid import cycles
type ClientLogger interface {
	LogOperation(level LogLevel, service, instance, operation, phase, message string, fields map[string]interface{})
}

// DefaultClientLogger provides basic stdout logging
type DefaultClientLogger struct{}

func (dl *DefaultClientLogger) LogOperation(level LogLevel, service, instance, operation, phase, message string, fields map[string]interface{}) {
	// Respect LOG_LEVEL environment variable like the agent logger does
	configuredLevel := getLogLevelFromEnv()
	if level < configuredLevel {
		return // Skip logging if level is below configured threshold
	}

	levelStr := "INFO"
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}
	
	fmt.Printf("[%s] [%s] [%s] [%s] [%s] %s", 
		time.Now().Format("2006-01-02T15:04:05.000Z"),
		levelStr, service, operation, phase, message)
	
	// Print additional fields
	for key, value := range fields {
		fmt.Printf(" (%s: %v)", key, value)
	}
	fmt.Println()
}

// getLogLevelFromEnv gets the configured log level from environment
func getLogLevelFromEnv() LogLevel {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "INFO"
	}
	
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Helper functions for creating field maps
func WithError(err error) map[string]interface{} {
	fields := make(map[string]interface{})
	if err != nil {
		fields["error"] = err.Error()
	}
	return fields
}

func WithDuration(d time.Duration) map[string]interface{} {
	return map[string]interface{}{"duration": d.String()}
}

func WithSize(size int64) map[string]interface{} {
	return map[string]interface{}{"size": size}
}

func WithSpeed(mbps float64) map[string]interface{} {
	return map[string]interface{}{"speed_mbps": mbps}
}

func WithChunk(chunkNumber, totalChunks int) map[string]interface{} {
	return map[string]interface{}{
		"chunk_number": chunkNumber,
		"total_chunks": totalChunks,
	}
}

func WithAttempt(attempt, maxAttempts int) map[string]interface{} {
	return map[string]interface{}{
		"attempt": attempt,
		"max_attempts": maxAttempts,
	}
}

func WithTransferID(transferID string) map[string]interface{} {
	return map[string]interface{}{"transfer_id": transferID}
}

func WithStatusCode(code int) map[string]interface{} {
	return map[string]interface{}{"status_code": code}
}

func MergeFields(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fieldMap := range fieldMaps {
		for k, v := range fieldMap {
			result[k] = v
		}
	}
	return result
}
