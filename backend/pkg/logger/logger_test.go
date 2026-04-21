//go:build unit

package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogger_Info tests the Info method
func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := &Logger{Logger: slog.New(handler)}

	logger.Info("test message", "key1", "value1", "key2", "value2")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "test message", result["msg"])
	assert.Equal(t, "INFO", result["level"])
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
}

// TestLogger_Error tests the Error method
func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := &Logger{Logger: slog.New(handler)}

	logger.Error("error message", "error", "test error", "code", 500)

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "error message", result["msg"])
	assert.Equal(t, "ERROR", result["level"])
	assert.Equal(t, "test error", result["error"])
	assert.Equal(t, float64(500), result["code"])
}

// TestLogger_Warn tests the Warn method
func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := &Logger{Logger: slog.New(handler)}

	logger.Warn("warning message", "severity", "high")

	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "warning message", result["msg"])
	assert.Equal(t, "WARN", result["level"])
	assert.Equal(t, "high", result["severity"])
}

// TestConvertArgs tests the convertArgs helper function
func TestConvertArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		wantLen  int
		validate func(result []interface{})
	}{
		{
			name:    "empty args",
			args:    []interface{}{},
			wantLen: 0,
		},
		{
			name:    "single pair",
			args:    []interface{}{"key", "value"},
			wantLen: 1,
			validate: func(result []interface{}) {
				attr, ok := result[0].(slog.Attr)
				assert.True(t, ok)
				assert.Equal(t, "key", attr.Key)
			},
		},
		{
			name:    "multiple pairs",
			args:    []interface{}{"key1", "value1", "key2", "value2"},
			wantLen: 2,
		},
		{
			name:    "odd number of args",
			args:    []interface{}{"key1", "value1", "key2"},
			wantLen: 3, // Kept as-is by slog
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertArgs(tt.args)
			assert.Equal(t, tt.wantLen, len(result))

			if tt.validate != nil {
				tt.validate(result)
			}
		})
	}
}

// TestLogger_Singleton tests that logger is a singleton
func TestLogger_Singleton(t *testing.T) {
	InitLogger()
	logger1 := GetLogger()

	InitLogger() // Call again
	logger2 := GetLogger()

	// Should be the same instance
	assert.Equal(t, logger1, logger2)
}
