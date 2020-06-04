package gorm

import (
	"fmt"

	"github.com/delving/hub3/ikuzo/storage/x/gorm/internal/postgres"
	"github.com/jinzhu/gorm"
)

func NewDB(dbType, connect string) (*gorm.DB, error) {
	if dbType == "postgres" {
		return postgres.NewDB(connect)
	}

	return nil, fmt.Errorf("unsupported database type %s", dbType)
}
