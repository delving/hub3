package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo"
	"github.com/delving/hub3/ikuzo/logger"
	"github.com/delving/hub3/ikuzo/storage/x/postgresql"
)

type DB struct {
	DSN          string `json:"dsn"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxIdleTime  string `json:"maxIdleTime"`
	AutoMigrate  bool   `json:"autoMigrate"`
	db           *sql.DB
	log          *logger.CustomLogger
}

func (db *DB) Shutdown(ctx context.Context) error {
	db.log.Info().Msg("stopping the database connectionpool")
	return db.db.Close()
}

func (db *DB) AddOptions(cfg *Config) error {
	if !strings.HasPrefix(db.DSN, "postgres") {
		return fmt.Errorf("only postgresql is supported for now")
	}

	dbConfig := postgresql.Config{
		DSN:          db.DSN,
		MaxOpenConns: db.MaxOpenConns,
		MaxIdleConns: db.MaxIdleConns,
		MaxIdleTime:  db.MaxIdleTime,
		AutoMigrate:  db.AutoMigrate,
	}

	database, err := postgresql.OpenDB(dbConfig)
	if err != nil {
		return fmt.Errorf("unable to open DB; %w", err)
	}

	cfg.logger.Info().Msg("connected to db")
	db.log = &cfg.logger

	db.db = database

	cfg.options = append(cfg.options, ikuzo.SetShutdownHook("db", db))

	return nil
}
