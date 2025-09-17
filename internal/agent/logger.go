package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string into a LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	Level        string    `json:"level"`
	Service      string    `json:"service,omitempty"`
	Instance     string    `json:"instance,omitempty"`
	Operation    string    `json:"operation,omitempty"`
	Phase        string    `json:"phase,omitempty"`
	Message      string    `json:"message"`
	Error        string    `json:"error,omitempty"`
	Duration     string    `json:"duration,omitempty"`
	Status       string    `json:"status,omitempty"`
	StatusCode   int       `json:"status_code,omitempty"`
	Size         int64     `json:"size,omitempty"`
	Speed        float64   `json:"speed_mbps,omitempty"`
	ChunkNumber  int       `json:"chunk_number,omitempty"`
	TotalChunks  int       `json:"total_chunks,omitempty"`
	Attempt      int       `json:"attempt,omitempty"`
	MaxAttempts  int       `json:"max_attempts,omitempty"`
	TransferID   string    `json:"transfer_id,omitempty"`
	File         string    `json:"file,omitempty"`
	Function     string    `json:"function,omitempty"`
}

// StructuredLogger provides structured logging functionality
type StructuredLogger struct {
	level      LogLevel
	writer     io.Writer
	service    string
	jsonFormat bool
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level LogLevel, service string, jsonFormat bool) *StructuredLogger {
	return &StructuredLogger{
		level:      level,
		writer:     os.Stdout,
		service:    service,
		jsonFormat: jsonFormat,
	}
}

// SetOutput sets the output writer
func (sl *StructuredLogger) SetOutput(w io.Writer) {
	sl.writer = w
}

// SetLevel sets the logging level
func (sl *StructuredLogger) SetLevel(level LogLevel) {
	sl.level = level
}

// log writes a log entry
func (sl *StructuredLogger) log(level LogLevel, entry LogEntry) {
	if level < sl.level {
		return
	}

	entry.Timestamp = time.Now().UTC()
	entry.Level = level.String()

	// Set default service if not provided
	if entry.Service == "" {
		entry.Service = sl.service
	}

	// Add caller information for ERROR level
	if level == ERROR {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.File = fmt.Sprintf("%s:%d", file, line)
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Function = fn.Name()
			}
		}
	}

	if sl.jsonFormat {
		sl.writeJSON(entry)
	} else {
		sl.writeText(entry)
	}
}

// writeJSON writes the log entry as JSON
func (sl *StructuredLogger) writeJSON(entry LogEntry) {
	if data, err := json.Marshal(entry); err == nil {
		fmt.Fprintln(sl.writer, string(data))
	} else {
		// Fallback to standard logging if JSON marshaling fails
		log.Printf("Failed to marshal log entry: %v", err)
	}
}

// writeText writes the log entry as formatted text
func (sl *StructuredLogger) writeText(entry LogEntry) {
	var builder strings.Builder
	
	builder.WriteString(fmt.Sprintf("[%s] %s", 
		entry.Timestamp.Format("2006-01-02T15:04:05.000Z"), 
		entry.Level))
	
	if entry.Service != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.Service))
	}
	
	if entry.Instance != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.Instance))
	}

	if entry.Operation != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.Operation))
	}

	if entry.Phase != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.Phase))
	}
	
	builder.WriteString(fmt.Sprintf(" %s", entry.Message))

	if entry.TransferID != "" {
		builder.WriteString(fmt.Sprintf(" (transfer_id: %s)", entry.TransferID))
	}

	if entry.ChunkNumber > 0 {
		if entry.TotalChunks > 0 {
			builder.WriteString(fmt.Sprintf(" (chunk: %d/%d)", entry.ChunkNumber, entry.TotalChunks))
		} else {
			builder.WriteString(fmt.Sprintf(" (chunk: %d)", entry.ChunkNumber))
		}
	}

	if entry.Attempt > 0 {
		if entry.MaxAttempts > 0 {
			builder.WriteString(fmt.Sprintf(" (attempt: %d/%d)", entry.Attempt, entry.MaxAttempts))
		} else {
			builder.WriteString(fmt.Sprintf(" (attempt: %d)", entry.Attempt))
		}
	}
	
	if entry.Duration != "" {
		builder.WriteString(fmt.Sprintf(" (duration: %s)", entry.Duration))
	}

	if entry.Size > 0 {
		builder.WriteString(fmt.Sprintf(" (size: %d bytes)", entry.Size))
	}

	if entry.Speed > 0 {
		builder.WriteString(fmt.Sprintf(" (speed: %.2f MB/s)", entry.Speed))
	}
	
	if entry.Status != "" {
		builder.WriteString(fmt.Sprintf(" (status: %s)", entry.Status))
	}

	if entry.StatusCode > 0 {
		builder.WriteString(fmt.Sprintf(" (http_status: %d)", entry.StatusCode))
	}
	
	if entry.Error != "" {
		builder.WriteString(fmt.Sprintf(" - Error: %s", entry.Error))
	}
	
	if entry.File != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.File))
	}
	
	fmt.Fprintln(sl.writer, builder.String())
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(message string) {
	sl.log(DEBUG, LogEntry{
		Message: message,
	})
}

// DebugWithFields logs a debug message with additional fields
func (sl *StructuredLogger) DebugWithFields(service, instance, message string) {
	sl.log(DEBUG, LogEntry{
		Service:  service,
		Instance: instance,
		Message:  message,
	})
}

// Info logs an info message
func (sl *StructuredLogger) Info(message string) {
	sl.log(INFO, LogEntry{
		Message: message,
	})
}

// InfoWithFields logs an info message with additional fields
func (sl *StructuredLogger) InfoWithFields(service, instance, message, duration, status string) {
	sl.log(INFO, LogEntry{
		Service:  service,
		Instance: instance,
		Message:  message,
		Duration: duration,
		Status:   status,
	})
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(message string) {
	sl.log(WARN, LogEntry{
		Message: message,
	})
}

// WarnWithFields logs a warning message with additional fields
func (sl *StructuredLogger) WarnWithFields(service, instance, message, errorMsg string) {
	sl.log(WARN, LogEntry{
		Service:  service,
		Instance: instance,
		Message:  message,
		Error:    errorMsg,
	})
}

// Error logs an error message
func (sl *StructuredLogger) Error(message string, err error) {
	entry := LogEntry{
		Message: message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	sl.log(ERROR, entry)
}

// ErrorWithFields logs an error message with additional fields
func (sl *StructuredLogger) ErrorWithFields(service, instance, message string, err error) {
	entry := LogEntry{
		Service:  service,
		Instance: instance,
		Message:  message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	sl.log(ERROR, entry)
}

// TestResult logs a test result with structured fields
func (sl *StructuredLogger) TestResult(service, instance, testType, message, duration, status string, err error) {
	entry := LogEntry{
		Service:   service,
		Instance:  instance,
		Operation: testType,
		Message:   message,
		Duration:  duration,
		Status:    status,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	
	level := INFO
	if status == "FAILED" || err != nil {
		level = ERROR
	}
	
	sl.log(level, entry)
}

// LogOperation logs a general operation with optional fields
func (sl *StructuredLogger) LogOperation(level LogLevel, service, instance, operation, phase, message string, opts ...LogOption) {
	entry := LogEntry{
		Service:   service,
		Instance:  instance,
		Operation: operation,
		Phase:     phase,
		Message:   message,
	}
	
	// Apply optional fields
	for _, opt := range opts {
		opt(&entry)
	}
	
	sl.log(level, entry)
}

// LogOption is a function type for setting optional log entry fields
type LogOption func(*LogEntry)

// WithError adds an error to the log entry
func WithError(err error) LogOption {
	return func(e *LogEntry) {
		if err != nil {
			e.Error = err.Error()
		}
	}
}

// WithDuration adds a duration to the log entry
func WithDuration(d time.Duration) LogOption {
	return func(e *LogEntry) {
		e.Duration = d.String()
	}
}

// WithStatus adds a status to the log entry
func WithStatus(status string) LogOption {
	return func(e *LogEntry) {
		e.Status = status
	}
}

// WithStatusCode adds an HTTP status code to the log entry
func WithStatusCode(code int) LogOption {
	return func(e *LogEntry) {
		e.StatusCode = code
	}
}

// WithSize adds a file size to the log entry
func WithSize(size int64) LogOption {
	return func(e *LogEntry) {
		e.Size = size
	}
}

// WithSpeed adds a transfer speed to the log entry
func WithSpeed(mbps float64) LogOption {
	return func(e *LogEntry) {
		e.Speed = mbps
	}
}

// WithChunk adds chunk information to the log entry
func WithChunk(chunkNumber, totalChunks int) LogOption {
	return func(e *LogEntry) {
		e.ChunkNumber = chunkNumber
		e.TotalChunks = totalChunks
	}
}

// WithAttempt adds retry attempt information to the log entry
func WithAttempt(attempt, maxAttempts int) LogOption {
	return func(e *LogEntry) {
		e.Attempt = attempt
		e.MaxAttempts = maxAttempts
	}
}

// WithTransferID adds a transfer ID to the log entry
func WithTransferID(transferID string) LogOption {
	return func(e *LogEntry) {
		e.TransferID = transferID
	}
}

// Global logger instance
var Logger *StructuredLogger

// InitLogger initializes the global logger
func InitLogger(levelStr, service string, jsonFormat bool) {
	level := ParseLogLevel(levelStr)
	Logger = NewStructuredLogger(level, service, jsonFormat)
}

// LogServiceOperation logs a service operation (convenience function for services)
func LogServiceOperation(level LogLevel, service, instance, operation, phase, message string, opts ...LogOption) {
	if Logger == nil {
		return
	}
	Logger.LogOperation(level, service, instance, operation, phase, message, opts...)
}

// GetLogLevel returns the current log level from environment
func GetLogLevel() string {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "INFO"
	}
	return level
}

// GetLogFormat returns whether to use JSON format from environment
func GetLogFormat() bool {
	format := os.Getenv("LOG_FORMAT")
	return strings.ToLower(format) == "json"
}
