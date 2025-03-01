package db

import (
	"embed"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFs embed.FS

func migrateSetup() (*migrate.Migrate, error) {
	driver, err := sqlite3.WithInstance(
		dbc, &sqlite3.Config{MigrationsTable: "migrations"})
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(migrationsFs, "migrations")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"sqlite",
		driver,
	)
	if err != nil {
		return m, err
	}

	return m, nil
}

func MigrateTo(version uint) error {
	m, err := migrateSetup()
	if err != nil {
		return err
	}

	return m.Migrate(version)
}

func MigrateDown() error {
	m, err := migrateSetup()
	if err != nil {
		return err
	}
	return m.Down()
}

func Drop() error {
	m, err := migrateSetup()
	if err != nil {
		return err
	}
	return m.Drop()
}

func Migrate() error {
	m, err := migrateSetup()
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			slog.Warn("db not versioned, performing migration")
			return m.Up()
		}
		return err
	}

	slog.Debug("db migration", "version", version, "dirty", dirty)
	return m.Up()
}
