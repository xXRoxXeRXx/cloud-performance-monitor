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
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service,omitempty"`
	Instance  string    `json:"instance,omitempty"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Status    string    `json:"status,omitempty"`
	File      string    `json:"file,omitempty"`
	Function  string    `json:"function,omitempty"`
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
func (sl *StructuredLogger) log(level LogLevel, service, instance, message, errorMsg, duration, status string) {
	if level < sl.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
		Service:   service,
		Instance:  instance,
		Message:   message,
		Error:     errorMsg,
		Duration:  duration,
		Status:    status,
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
	
	builder.WriteString(fmt.Sprintf(" %s", entry.Message))
	
	if entry.Duration != "" {
		builder.WriteString(fmt.Sprintf(" (duration: %s)", entry.Duration))
	}
	
	if entry.Status != "" {
		builder.WriteString(fmt.Sprintf(" (status: %s)", entry.Status))
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
	sl.log(DEBUG, sl.service, "", message, "", "", "")
}

// DebugWithFields logs a debug message with additional fields
func (sl *StructuredLogger) DebugWithFields(service, instance, message string) {
	sl.log(DEBUG, service, instance, message, "", "", "")
}

// Info logs an info message
func (sl *StructuredLogger) Info(message string) {
	sl.log(INFO, sl.service, "", message, "", "", "")
}

// InfoWithFields logs an info message with additional fields
func (sl *StructuredLogger) InfoWithFields(service, instance, message, duration, status string) {
	sl.log(INFO, service, instance, message, "", duration, status)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(message string) {
	sl.log(WARN, sl.service, "", message, "", "", "")
}

// WarnWithFields logs a warning message with additional fields
func (sl *StructuredLogger) WarnWithFields(service, instance, message, errorMsg string) {
	sl.log(WARN, service, instance, message, errorMsg, "", "")
}

// Error logs an error message
func (sl *StructuredLogger) Error(message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	sl.log(ERROR, sl.service, "", message, errorMsg, "", "")
}

// ErrorWithFields logs an error message with additional fields
func (sl *StructuredLogger) ErrorWithFields(service, instance, message string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	sl.log(ERROR, service, instance, message, errorMsg, "", "")
}

// TestResult logs a test result with structured fields
func (sl *StructuredLogger) TestResult(service, instance, testType, message, duration, status string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	
	level := INFO
	if status == "FAILED" || err != nil {
		level = ERROR
	}
	
	sl.log(level, service, instance, 
		fmt.Sprintf("%s %s", testType, message), 
		errorMsg, duration, status)
}

// Global logger instance
var Logger *StructuredLogger

// InitLogger initializes the global logger
func InitLogger(levelStr, service string, jsonFormat bool) {
	level := ParseLogLevel(levelStr)
	Logger = NewStructuredLogger(level, service, jsonFormat)
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
