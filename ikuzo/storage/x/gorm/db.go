package gorm

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/storage/x/gorm/internal/postgres"
	"github.com/delving/hub3/ikuzo/storage/x/gorm/internal/sqlite3"
	"github.com/jinzhu/gorm"
)

func NewDB(dbType, connect string) (*gorm.DB, error) {
	switch dbType {
	case "sqlite3":
		return sqlite3.NewDB(connect)
	case "postgres":
		return postgres.NewDB(connect)
	}

	return nil, fmt.Errorf("unsupported database type %s", dbType)
}
