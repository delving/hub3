package gorm

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SlogLogger is a custom GORM logger that uses slog
type SlogLogger struct {
	SlowThreshold        time.Duration
	IgnoreRecordNotFound bool
	LogLevel             logger.LogLevel
	Logger               *slog.Logger
}

// NewSlogLogger creates a new GORM logger that uses the provided slog logger
func NewSlogLogger(slogger *slog.Logger) *SlogLogger {
	// Map slog level to GORM log level
	var logLevel logger.LogLevel

	// Determine the log level based on the slog logger's level
	switch slogger.Enabled(context.Background(), slog.LevelDebug) {
	case true:
		logLevel = logger.Info // If debug is enabled, set GORM to Info (most verbose)
	case false:
		if slogger.Enabled(context.Background(), slog.LevelInfo) {
			logLevel = logger.Info
		} else if slogger.Enabled(context.Background(), slog.LevelWarn) {
			logLevel = logger.Warn
		} else if slogger.Enabled(context.Background(), slog.LevelError) {
			logLevel = logger.Error
		} else {
			logLevel = logger.Silent
		}
	}

	return &SlogLogger{
		SlowThreshold:        time.Second, // Slow SQL threshold
		IgnoreRecordNotFound: true,        // Skip record not found errors
		LogLevel:             logLevel,
		Logger:               slogger,
	}
}

// LogMode sets the log level for the logger
func (l *SlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs info messages
func (l *SlogLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Logger.InfoContext(ctx, msg, "args", args)
	}
}

// Warn logs warn messages
func (l *SlogLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Logger.WarnContext(ctx, msg, "args", args)
	}
}

// Error logs error messages
func (l *SlogLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Logger.ErrorContext(ctx, msg, "args", args)
	}
}

// Trace logs SQL statements
func (l *SlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Skip logging if RecordNotFound error is ignored
	if l.IgnoreRecordNotFound && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	// Build log attributes
	attrs := []any{
		"elapsed", elapsed,
		"rows", rows,
		"sql", sql,
	}

	// Log based on error and execution time
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		l.Logger.ErrorContext(ctx, "GORM error", append(attrs, "error", err)...)
	case elapsed > l.SlowThreshold && l.SlowThreshold > 0 && l.LogLevel >= logger.Warn:
		l.Logger.WarnContext(ctx, "GORM slow query", attrs...)
	case l.LogLevel >= logger.Info:
		l.Logger.InfoContext(ctx, "GORM query", attrs...)
	}
}

