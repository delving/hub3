// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// TODO(kiivihal): Delete whole file

package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/delving/hub3/config"
	"github.com/rs/zerolog/log"
)

// orm is the storm db entry point
var orm *storm.DB

// CloseStorm close the underlying BoltDB for Storm
func CloseStorm() {
	err := ORM().Close()
	if err != nil {
		log.Fatal().Msgf("Unable to close BoltDB. %s", err)
	}
	orm = nil
}

func ResetStorm() {
	CloseStorm()
	os.Remove("hub3.db")
}

func ResetEADCache() {
	os.RemoveAll(config.Config.EAD.CacheDir)
}

func ORM() *storm.DB {
	if orm == nil {
		orm = newDB("")
		orm.WithBatch(true)
	}
	return orm
}

func newDB(dbName string) *storm.DB {
	if dbName == "" {
		dbName = "hub3.db"
	}
	if !strings.HasSuffix(dbName, ".db") {
		dbName = fmt.Sprintf("%s.db", dbName)
	}
	db, err := storm.Open(dbName)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to open the BoltDB database file.")
	}
	ex, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	exPath := filepath.Dir(ex)
	log.Info().
		Str("full_path", exPath).
		Str("db_name", db.Bolt.Path()).
		Msg("starting boldDB")
	//defer db.Close()
	return db
}
