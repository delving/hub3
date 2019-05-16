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

package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
)

// orm is the storm db entry point
var orm *storm.DB

func init() {
	orm = newDB("")
	//orm.WithBatch(true)
}

// CloseStorm close the underlying BoltDB for Storm
func CloseStorm() {
	err := orm.Close()
	if err != nil {
		log.Fatalf("Unable to close BoltDB. %s", err)
	}
}

func ResetStorm() {
	CloseStorm()
	os.Remove("hub3.db")
	orm = newDB("")
}

func ORM() *storm.DB {
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
		log.Fatal("Unable to open the BoltDB database file.")
	}
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)
	log.Printf("Running from %s\n", exPath)
	log.Printf("Using Storm/BoltDB path: %s\n", db.Bolt.Path())
	//defer db.Close()
	return db
}
