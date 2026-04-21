package logger

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

type Logger struct {
	*slog.Logger
}

var (
	globalLogger *Logger
	once         sync.Once
)

// InitLogger initializes the global logger (call this once at startup)
func InitLogger() {
	once.Do(func() {
		globalLogger = NewLogger()
	})
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		InitLogger()
	}
	return globalLogger
}

func NewLogger() *Logger {
	// Create a JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{
		Logger: slog.New(handler),
	}
}

// Info logs an information message with key-value pairs
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Info(msg, convertArgs(args)...)
}

// Error logs an error message with key-value pairs
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Logger.Error(msg, convertArgs(args)...)
}

// Debug logs a debug message with key-value pairs
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Logger.Debug(msg, convertArgs(args)...)
}

// Warn logs a warning message with key-value pairs
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Logger.Warn(msg, convertArgs(args)...)
}

// convertArgs converts variadic arguments to slog.Attr slice
// Expects pairs of key-value arguments
func convertArgs(args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}
	
	// Convert key-value pairs to slog format
	// slog.Any(key, value) or slog.Attr pattern
	attrs := make([]interface{}, 0, len(args))
	
	// If odd number of arguments, pass them as-is for slog to handle
	if len(args)%2 != 0 {
		return args
	}
	
	// Convert to key, value pairs
	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprintf("%v", args[i])
		value := args[i+1]
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}
