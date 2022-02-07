package postgresql

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var migrations embed.FS

const schemaVersion = 1

func EnsureSchema(dsn string) error {
	sourceInstance, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("invalid source instance, %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceInstance, dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance, %w", err)
	}

	err = m.Migrate(schemaVersion)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return sourceInstance.Close()
}
