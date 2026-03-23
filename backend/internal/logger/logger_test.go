package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: true,
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if logger.inner == nil {
		t.Fatal("New() created logger with nil inner")
	}
}

func TestParseSlogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"debug", "debug", "DEBUG"},
		{"info", "info", "INFO"},
		{"warn", "warn", "WARN"},
		{"warning", "warning", "WARN"},
		{"error", "error", "ERROR"},
		{"unknown", "unknown", "INFO"},
		{"empty", "", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSlogLevel(tt.level)
			if result.String() != tt.expected {
				t.Errorf("parseSlogLevel(%q) = %q, want %q", tt.level, result.String(), tt.expected)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	logger.Debug(ctx, "test debug message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test debug message") {
		t.Errorf("expected output to contain 'test debug message', got: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "info",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	logger.Info(ctx, "test info message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test info message") {
		t.Errorf("expected output to contain 'test info message', got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "warn",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	ctx := context.Background()
	logger.Warn(ctx, "test warn message")

	output := buf.String()
	if !strings.Contains(output, "test warn message") {
		t.Errorf("expected output to contain 'test warn message', got: %s", output)
	}
}

func TestLogger_Error_WithStackTrace(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "error",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	ctx := context.Background()
	testErr := errors.New("test error")
	logger.Error(ctx, "test error message", testErr, "extra_key", "extra_value")

	output := buf.String()

	// Verify JSON is valid
	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", output)
	}

	// Check for required fields
	if entry["msg"] != "test error message" {
		t.Errorf("expected message 'test error message', got: %v", entry["msg"])
	}
	if entry["error"] != "test error" {
		t.Errorf("expected error 'test error', got: %v", entry["error"])
	}
	if _, ok := entry["stack_trace"]; !ok {
		t.Error("expected stack_trace field in error log")
	} else {
		stackTrace := entry["stack_trace"].(string)
		if !strings.Contains(stackTrace, "in") {
			t.Errorf("stack_trace should contain function info, got: %s", stackTrace)
		}
	}
}

func TestLogger_Errorf(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "error",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	ctx := context.Background()
	testErr := errors.New("formatted error")
	logger.Errorf(ctx, "error: %s", testErr, "something")

	output := buf.String()

	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", output)
	}

	if entry["error"] != "formatted error" {
		t.Errorf("expected error 'formatted error', got: %v", entry["error"])
	}
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	loggerWith := logger.With("service", "test-service")
	loggerWith.Info(ctx, "test message with context")

	output := buf.String()

	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", output)
	}

	// Check if the additional context is present
	service, ok := entry["service"].(string)
	if !ok || service != "test-service" {
		t.Errorf("expected service context 'test-service', got: %v", entry["service"])
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "warn",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	ctx := context.Background()

	// Debug and Info should be filtered out
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")

	output := buf.String()
	if output != "" {
		t.Errorf("expected no output for filtered levels, got: %s", output)
	}

	// Reset buffer
	buf.Reset()
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	// Warn should pass through
	logger.Warn(ctx, "warn message")

	output = buf.String()
	if !strings.Contains(output, "warn message") {
		t.Errorf("expected warn message to pass through, got: %s", output)
	}
}

func TestDefaultLogger(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Error("Default() returned nil")
	}

	SetDefault(logger)
	if defaultLogger != logger {
		t.Error("SetDefault() did not set the default logger")
	}
}

func TestGetStackTrace(t *testing.T) {
	trace := getStackTrace(0)

	if trace == "" {
		t.Error("getStackTrace() returned empty string")
	}

	if !strings.Contains(trace, "in ") {
		t.Errorf("stack trace should contain function info, got: %s", trace)
	}

	// Verify it doesn't contain logger internal frames
	if strings.Contains(trace, "/internal/logger/logger.go") {
		t.Error("stack trace should not contain internal logger frames")
	}
}

func TestNew_WithStderr(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stderr",
		AddSource: false,
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New() with stderr returned nil")
	}
}

func TestNew_WithFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.log"
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "file",
		FilePath:  tmpFile,
		AddSource: false,
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New() with file returned nil")
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); err != nil {
		t.Error("log file was not created")
	}
}

func TestNew_WithInvalidFilePath(t *testing.T) {
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "file",
		FilePath:  "/nonexistent/directory/test.log",
		AddSource: false,
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New() with invalid file path returned nil")
	}
	// Should fallback to stdout
	if logger.inner == nil {
		t.Error("logger.inner is nil when file path is invalid")
	}
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	loggerWithGroup := logger.WithGroup("test-group")
	loggerWithGroup.Info(ctx, "test message in group", "key", "value")

	output := buf.String()

	var entry map[string]any
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("invalid JSON output: %v", output)
	}

	// WithGroup creates a nested group in slog
	// The group should appear in the output
	t.Logf("output: %s", output)
}

func TestLogger_OutputToWriter(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.LoggerConfig{
		Level:     "debug",
		Format:    "json",
		Output:    "stdout",
		AddSource: false,
	}

	logger := New(cfg)
	logger.inner = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	logger.Info(ctx, "test message")

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty string")
	}
}
