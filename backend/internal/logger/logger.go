package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
)

const goMod = "github.com/PineappleBond/TradingEino/backend/"

// Logger wraps slog.Logger with additional stack trace support for errors
type Logger struct {
	inner       *slog.Logger
	addSource   bool
	skipCallers int
	closer      io.Closer // closer holds the resource to close on shutdown
}

// New creates a new Logger instance with JSON output
func New(cfg config.LoggerConfig, skipCallers int) *Logger {
	var output io.Writer
	var closer io.Closer

	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		if cfg.FilePath == "" {
			output = os.Stdout
		} else {
			f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				output = os.Stdout
			} else {
				output = f
				closer = f // Save file handle for closing later
			}
		}
	default:
		output = os.Stdout
	}

	level := parseSlogLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false, // We'll handle source ourselves with correct skip
	}

	handler := slog.NewJSONHandler(output, opts)

	return &Logger{
		inner:       slog.New(handler),
		addSource:   cfg.AddSource,
		skipCallers: skipCallers,
		closer:      closer,
	}
}

// parseSlogLevel converts string level to slog.Level
func parseSlogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getStackTrace returns a formatted stack trace, skipping the specified number of frames
func getStackTrace(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder
	foundRelevantFrame := false

	for {
		frame, more := frames.Next()
		// Skip runtime frames only
		if strings.Contains(frame.File, "/runtime/") {
			continue
		}
		// Skip goroot frames
		if strings.Contains(frame.File, "/src/runtime/") {
			continue
		}
		// Only start recording after we pass the logger internal frames
		if !foundRelevantFrame && strings.Contains(frame.File, "/internal/logger/") {
			continue
		}
		foundRelevantFrame = true
		sb.WriteString(fmt.Sprintf("%s:%d in %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	result := strings.TrimSuffix(sb.String(), "\n")
	if result == "" {
		return "no stack trace available"
	}
	return result
}

// getCallerInfo returns the file and line number of the caller, skipping logger internal frames
func getCallerInfo(skip int) string {
	pcs := [13]uintptr{}
	// the third caller usually from gorm internal
	l := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:l])
	for i := 0; i < l; i++ {
		// second return value is "more", not "ok"
		frame, _ := frames.Next()
		shortFile := frame.File
		if !strings.HasSuffix(frame.File, ".gen.go") && !strings.Contains(frame.File, "gorm.io") && !strings.Contains(frame.File, "go-sqlite3") {
			split := strings.Split(shortFile, goMod)
			if len(split) == 2 {
				shortFile = split[1] + ":" + strconv.Itoa(frame.Line)
			}
			return shortFile
		}
	}

	return ""
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	if l.addSource {
		for _, arg := range args {
			msg += fmt.Sprintf(" %+v", arg)
		}
		l.inner.LogAttrs(ctx, slog.LevelDebug, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.DebugContext(ctx, msg, args...)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	if l.addSource {
		msg := fmt.Sprintf(format, args...)
		l.inner.LogAttrs(ctx, slog.LevelDebug, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.DebugContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	if l.addSource {
		for _, arg := range args {
			msg += fmt.Sprintf(" %+v", arg)
		}
		l.inner.LogAttrs(ctx, slog.LevelInfo, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.InfoContext(ctx, msg, args...)
	}
}

// Infof logs a formatted info message
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	if l.addSource {
		msg := fmt.Sprintf(format, args...)
		l.inner.LogAttrs(ctx, slog.LevelInfo, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.InfoContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if l.addSource {
		for _, arg := range args {
			msg += fmt.Sprintf(" %+v", arg)
		}
		l.inner.LogAttrs(ctx, slog.LevelWarn, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.WarnContext(ctx, msg, args...)
	}
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	if l.addSource {
		msg := fmt.Sprintf(format, args...)
		l.inner.LogAttrs(ctx, slog.LevelWarn, msg, slog.Attr{
			Key:   "source",
			Value: slog.StringValue(getCallerInfo(l.skipCallers)),
		})
	} else {
		l.inner.WarnContext(ctx, fmt.Sprintf(format, args...))
	}
}

// Error logs an error message with stack trace
func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
	for _, arg := range args {
		msg += fmt.Sprintf(" %+v", arg)
	}
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	l.inner.LogAttrs(ctx, slog.LevelError, msg, slog.Attr{
		Key:   "source",
		Value: slog.StringValue(getCallerInfo(l.skipCallers)),
	}, slog.Attr{
		Key:   "error",
		Value: slog.StringValue(errString),
	})
}

// Errorf logs a formatted error message with stack trace
func (l *Logger) Errorf(ctx context.Context, format string, err error, args ...any) {
	msg := fmt.Sprintf(format, args...)
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	l.inner.LogAttrs(ctx, slog.LevelError, msg, slog.Attr{
		Key:   "source",
		Value: slog.StringValue(getCallerInfo(l.skipCallers)),
	}, slog.Attr{
		Key:   "error",
		Value: slog.StringValue(errString),
	})
}

// With creates a new Logger with additional context
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		inner: l.inner.With(args...),
	}
}

// WithGroup creates a new Logger with a group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		inner: l.inner.WithGroup(name),
	}
}

// Close closes any resources held by the logger (e.g., file handles)
func (l *Logger) Close() error {
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// Global default logger
var defaultLogger = New(config.LoggerConfig{
	Level:     "info",
	Output:    "stdout",
	AddSource: true,
}, 4)

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// Convenience functions using the default logger

// SetGlobalDefault sets the global default logger and also sets slog.Default
func SetGlobalDefault(logger *Logger) {
	defaultLogger = logger
	slog.SetDefault(logger.inner)
}

func Debug(ctx context.Context, msg string, args ...any) {
	defaultLogger.Debug(ctx, msg, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	defaultLogger.Debugf(ctx, format, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	defaultLogger.Info(ctx, msg, args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	defaultLogger.Infof(ctx, format, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	defaultLogger.Warn(ctx, msg, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	defaultLogger.Warnf(ctx, format, args...)
}

func Error(ctx context.Context, msg string, err error, args ...any) {
	defaultLogger.Error(ctx, msg, err, args...)
}

func Errorf(ctx context.Context, format string, err error, args ...any) {
	defaultLogger.Errorf(ctx, format, err, args...)
}
