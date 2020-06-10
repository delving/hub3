// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"fmt"

	"github.com/delving/hub3/ikuzo"
	storage "github.com/delving/hub3/ikuzo/storage/x/gorm"
	"github.com/jinzhu/gorm"
)

type DB struct {
	// supported types are "sqlite3" "postgres"
	Type string
	// go sql compatible connection string, e.g. "/tmp/test.db" for sqlit3 or
	// "host=myhost port=myport user=hub3 dbname=hub3 password=mypassword"
	Connect string
	// database
	db *gorm.DB
}

func (db *DB) AddOptions(cfg *Config) error {
	var err error

	if db.Type == "" || db.Connect == "" {
		return fmt.Errorf("DB.Type and DB.Connect config options must not be empty")
	}

	db.db, err = storage.NewDB(db.Type, db.Connect)
	if err != nil {
		return fmt.Errorf("failed to connect to database; %w", err)
	}

	cfg.options = append(cfg.options, ikuzo.SetShutdownHook("db", db))

	return nil
}

func (db *DB) getDB() (*gorm.DB, error) {
	if db.db != nil {
		return db.db, nil
	}

	return nil, fmt.Errorf("call DB.AddOptions first")
}

func (db *DB) Shutdown(ctx context.Context) error {
	if db.db != nil {
		return db.db.Close()
	}

	return nil
}
