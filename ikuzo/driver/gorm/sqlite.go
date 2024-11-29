package gorm

import (
	"log/slog"

	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"
)

// OpenSqliteDB  opens a SQLite database connection using the pure Go implementation
func OpenSqliteDB(dbPath string, logger *slog.Logger) (*gorm.DB, error) {
	// If no logger is provided, use the default logger
	if logger == nil {
		logger = slog.Default()
	}

	// Create a custom GORM logger that inherits settings from the provided slog logger
	gormLogger := NewSlogLogger(logger)

	// Use glebarez driver with modernc.org/sqlite (pure Go implementation)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	return db, nil
}
