package sqlite3

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // register sqlite
)

func NewDB(connect string) (db *gorm.DB, err error) {
	db, err = gorm.Open("sqlite3", connect)

	return db, err
}
