package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"

	"spotify-insights/migrations"
)

// RunMigrations applies all pending SQL migrations embedded in the
// migrations package. It is safe to call on every startup: a schema that is
// already up to date is reported by migrate as ErrNoChange, which is treated
// as success rather than an error.
func RunMigrations(databaseURL string) error {
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
