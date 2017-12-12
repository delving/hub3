package models

import (
	"fmt"
	"log"
	"strings"

	"github.com/asdine/storm"
)

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

// orm is the storm db entry point
var orm *storm.DB

func init() {
	orm = newDB("")
	orm.WithBatch(true)
}

func newDB(dbName string) *storm.DB {
	if dbName == "" {
		dbName = "rapid.db"
	}
	if !strings.HasSuffix(dbName, ".db") {
		dbName = fmt.Sprintf("%s.db", dbName)
	}
	db, err := storm.Open(dbName, storm.Batch())
	if err != nil {
		log.Fatal("Unable to open the BoltDB database file.")
	}
	//defer db.Close()
	return db
}
