package middleware

import (
	"io"
	"log"
	"os"
)

// LogLevel represents the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Logger is a simple logger implementation
type Logger struct {
	level LogLevel
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	var logLevel LogLevel
	switch level {
	case "debug":
		logLevel = LogLevelDebug
	case "info":
		logLevel = LogLevelInfo
	case "warn":
		logLevel = LogLevelWarn
	case "error":
		logLevel = LogLevelError
	default:
		logLevel = LogLevelInfo
	}

	return &Logger{
		level: logLevel,
	}
}

// Debug logs a debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an info message
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		log.Printf("[INFO] "+format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		log.Printf("[WARN] "+format, args...)
	}
}

// Error logs an error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.shouldLog(LogLevelError) {
		log.Printf("[ERROR] "+format, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	log.Fatalf("[FATAL] "+format, args...)
}

// shouldLog checks if the given log level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levelPriority := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	return levelPriority[level] >= levelPriority[l.level]
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// SetFlags sets the flags for the logger
func (l *Logger) SetFlags(flag int) {
	log.SetFlags(flag)
}

// FileLogger creates a logger that writes to a file
func FileLogger(filename string, level string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger := NewLogger(level)
	logger.SetOutput(file)
	logger.SetFlags(log.LstdFlags | log.Lshortfile)

	return logger, nil
}