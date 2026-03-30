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
		dbc.DB, &sqlite3.Config{MigrationsTable: "migrations"})
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

// hasFTS5 checks if FTS5 extension is available in SQLite
func hasFTS5() bool {
	var result int
	err := dbc.QueryRow("SELECT sqlite_compileoption_used('ENABLE_FTS5')").Scan(&result)
	if err != nil {
		// Fallback: try to create a dummy FTS5 table
		_, err = dbc.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS _fts5_check USING fts5(content)")
		if err != nil {
			dbc.Exec("DROP TABLE IF EXISTS _fts5_check")
			return false
		}
		dbc.Exec("DROP TABLE IF EXISTS _fts5_check")
		return true
	}
	return result == 1
}

// MaxMigrationVersion is the latest migration version (12 = llm conversations)
const MaxMigrationVersion = 12

// MaxMigrationVersionWithoutFTS5 is the max version when FTS5 is unavailable
const MaxMigrationVersionWithoutFTS5 = 10

func Migrate() error {
	m, err := migrateSetup()
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			slog.Warn("db not versioned, performing migration")
		} else {
			return err
		}
	}

	if dirty {
		slog.Error("dirty migration state", "version", version)
		return errors.New("dirty migration state")
	}

	// Determine max migration version based on FTS5 availability
	maxVersion := uint(MaxMigrationVersion)
	if !hasFTS5() {
		maxVersion = MaxMigrationVersionWithoutFTS5
		slog.Debug("FTS5 not available, limiting migrations", "max_version", maxVersion)
	}

	slog.Debug("db migration", "version", version, "dirty", dirty, "target_version", maxVersion)

	// Migrate to the target version
	if version < maxVersion {
		return m.Migrate(maxVersion)
	}

	return nil
}
