package postgres

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // register sqlite
)

func NewDB(connect string) (db *gorm.DB, err error) {
	db, err = gorm.Open("postgres", connect)

	return db, err
}
